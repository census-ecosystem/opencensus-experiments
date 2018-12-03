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

import java.io.BufferedReader;
import java.io.IOException;
import java.io.PrintWriter;
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

      PrintWriter pout = response.getWriter();

      pout.print("<html><body>");
      pout.print("<h3>Hello Servlet</h3>");
      pout.print("</body></html>");
      return;
    }

    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
        throws ServletException, IOException {
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

    Server server = new Server(10101);
    ServletHandler handler = new ServletHandler();
    server.setHandler(handler);

    handler.addServletWithMapping(HelloServlet.class, "/*");

    server.start();
    server.join();
  }
}
