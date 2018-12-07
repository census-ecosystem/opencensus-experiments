/*
 * Copyright 2018, Google LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.opencensus.interop;

import io.opencensus.contrib.grpc.metrics.RpcViews;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.PrintWriter;
import java.util.ArrayList;
import java.util.List;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHandler;

public class JavaService {

  public static class HelloServlet extends HttpServlet {

    private static final long serialVersionUID = 1L;

    @Override
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
        throws ServletException, IOException {

      System.err.println("GET");

      Spec grpcSpec = Spec.newBuilder()
                      .setTransport(Spec.Transport.GRPC)
                      .setPropagation(Spec.Propagation.BINARY_FORMAT_PROPAGATION)
                      .build();

      Service javaGrpcService = Service.newBuilder()
                                .setName("grpc java service")
                                .setPort(10101)
                                .setHost("localhost")
                                .setSpec(grpcSpec)
                                .build();

      Spec httpSpec = Spec.newBuilder()
                      .setTransport(Spec.Transport.HTTP)
                      .setPropagation(Spec.Propagation.B3_FORMAT_PROPAGATION)
                      .build();
      Service javaHttpService = Service.newBuilder()
                                .setName("http java service")
                                .setPort(10102)
                                .setHost("localhost")
                                .setSpec(httpSpec)
                                .build();
      ServiceHop hop1 = ServiceHop.newBuilder()
                        .setService(javaGrpcService)
                        .addTags(Tag.newBuilder().setKey("key1").setValue("val1").build())
                        .build();
      ServiceHop hop2 = ServiceHop.newBuilder()
                        .setService(javaGrpcService)
                        .addTags(Tag.newBuilder().setKey("key2").setValue("val2").build())
                        .build();
      ServiceHop hop3 = ServiceHop.newBuilder()
                        .setService(javaHttpService)
                        .addTags(Tag.newBuilder().setKey("key3").setValue("val3").build())
                        .build();
      ServiceHop hop4 = ServiceHop.newBuilder()
                        .setService(javaGrpcService)
                        .addTags(Tag.newBuilder().setKey("key4").setValue("val4").build())
                        .build();

      List<ServiceHop> hops = new ArrayList();
      hops.add(hop1);
      hops.add(hop2);
      hops.add(hop3);
      hops.add(hop4);
      TestResponse testResponse = ServiceHopper.serviceHop(4242, "JavaService", hops);
      System.err.println("GET: TestResponse: " + testResponse);
      PrintWriter pout = response.getWriter();

      pout.print("<html><body>");
      pout.print("<h3>Hello Servlet</h3>");
      pout.print("<h3>" + testResponse + "</h3>");
      pout.print("</body></html>");
      return;
    }

    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
        throws ServletException, IOException {
      System.err.println("POST");
      // Read from request
      StringBuilder buffer = new StringBuilder();
      BufferedReader reader = request.getReader();
      String line;
      while ((line = reader.readLine()) != null) {
        buffer.append(line);
      }
      String data = buffer.toString();

      PrintWriter pout = response.getWriter();

      pout.print("<html><body>");
      pout.print("<h3>Hello Servlet</h3>");
      pout.print("</body></html>");
      return;
    }
  }

  public static void main(String[] args) throws Exception {
    ServletContextHandler context = new ServletContextHandler(ServletContextHandler.SESSIONS);
    context.setContextPath("/");

    // Registers all gRPC views.
    RpcViews.registerAllViews();
    new GrpcServer(10101).start();

    Server server = new Server(10100);
    ServletHandler handler = new ServletHandler();
    server.setHandler(handler);

    handler.addServletWithMapping(HelloServlet.class, "/*");

    server.start();
    server.join();
  }
}
