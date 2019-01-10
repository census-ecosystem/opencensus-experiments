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

import io.opencensus.contrib.http.servlet.OcHttpServletFilter;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.EnumSet;
import javax.servlet.DispatcherType;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.apache.log4j.Logger;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletHandler;

final class HttpServer {
  private static final Logger logger = Logger.getLogger(HttpServer.class.getName());

  private final int port;
  private Server server;

  private HttpServer(int port) {
    this.port = port;
  }

  // TODO(dpo): specify propagation format when available.
  static HttpServer createWithB3FormatPropagation(int port) {
    return new HttpServer(port);
  }

  // TODO(dpo): specify propatation format when available.
  static HttpServer createWithTraceContextFormatPropagation(int port) {
    return new HttpServer(port);
  }

  void start() throws IOException {
    try {
      server = new Server(port);
      ServletHandler handler = new ServletHandler();
      server.setHandler(handler);
      handler.addFilterWithMapping(
          OcHttpServletFilter.class, "/*", EnumSet.of(DispatcherType.REQUEST));
      handler.addServletWithMapping(InteropTestHttpServlet.class, "/*");
      logger.info("Java HTTP server started, listening on " + port);
      Runtime.getRuntime()
          .addShutdownHook(
              new Thread() {
                @Override
                public void run() {
                  // Use stderr here since the logger may have been reset by its JVM shutdown hook.
                  System.err.println("*** JVM shutting down, shutting down HTTP server");
                  HttpServer.this.stop();
                  System.err.println("*** HTTP server shut down");
                }
              });
      server.start();
    } catch (Exception exn) {
      logger.info("Error starting HTTP server: " + exn);
    }
  }

  private void stop() {
    if (server != null) {
      try {
        server.stop();
      } catch (Exception exn) {
        // Use stderr here since the logger may have been reset by its JVM shutdown hook.
        System.err.println("Error shutting down HTTP server: " + exn);
      }
    }
  }

  public static final class InteropTestHttpServlet extends HttpServlet {
    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
        throws ServletException, IOException {
      int requestContentLength = request.getContentLength();
      byte[] requestContent = new byte[requestContentLength];
      if (request.getInputStream().readLine(requestContent, 0, requestContentLength) != requestContentLength) {
        logger.info("HttpServer: error reading request content");
        return;
      }
      TestRequest testRequest = TestRequest.parseFrom(ByteBuffer.wrap(requestContent));
      TestResponse testResponse = ServiceHopper.serviceHop(testRequest);
      response.getOutputStream().write(testResponse.toByteArray());
    }
  }
}
