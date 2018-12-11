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

import static java.nio.charset.StandardCharsets.UTF_8;

import io.opencensus.contrib.grpc.metrics.RpcViews;
// TODO(dpo): uncomment when agent is available.
// import io.opencensus.exporter.trace.ocagent.OcAgentTraceExporter;
import io.opencensus.contrib.http.servlet.OcHttpServletFilter;
import io.opencensus.contrib.http.util.HttpViews;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.PrintWriter;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.EnumSet;
import java.util.List;
import java.util.logging.Logger;
import javax.servlet.DispatcherType;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.eclipse.jetty.server.Request;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.handler.AbstractHandler;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHandler;

public class JavaService {
  private static final Logger logger = Logger.getLogger(JavaService.class.getName());

  public static class HelloServlet extends HttpServlet {
    private static final long serialVersionUID = 1L;

    @Override
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
        throws ServletException, IOException {
      PrintWriter pout = response.getWriter();

      pout.print("<html><body>");
      pout.print("<h3>Hello Servlet</h3>");
      pout.print("</body></html>");
      return;
    }

    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
        throws ServletException, IOException {
      logger.info("POST");

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
      pout.print("<h3>Hello Servlet Post</h3>");
      pout.print("</body></html>");
      return;
    }
  }

  public void handle(
      String target, Request baseRequest, HttpServletRequest request, HttpServletResponse response)
      throws IOException, ServletException {
    response.setContentType("text/html;charset=utf-8");
    response.setStatus(HttpServletResponse.SC_OK);
    baseRequest.setHandled(true);
    response.getWriter().println("<h1>Hello World. default handle</h1>");
  }

  public static void main(String[] args) throws Exception {
    HttpViews.registerAllServerViews();
    HttpViews.registerAllClientViews();
    RpcViews.registerAllViews();
    // Initialize OpenCensus agent trace exporter.
    // TODO(dpo): uncomment when agent is available
    // OcAgentTraceExporter.createAndRegister();

    new GrpcServer(10101).start();

    ServletContextHandler context = new ServletContextHandler(ServletContextHandler.SESSIONS);
    context.setContextPath("/");

    Server server = new Server(10103);
    ServletHandler handler = new ServletHandler();
    server.setHandler(handler);

    handler.addFilterWithMapping(
        OcHttpServletFilter.class, "/*", EnumSet.of(DispatcherType.REQUEST));
    handler.addServletWithMapping(HelloServlet.class, "/*");
    logger.info("Java HTTP server started, listening on " + 10103);
    server.start();
    server.join();
  }
}
