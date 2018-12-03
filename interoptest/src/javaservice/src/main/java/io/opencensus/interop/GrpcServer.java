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

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;

import io.opencensus.trace.AttributeValue;
import io.opencensus.trace.Span;
import io.opencensus.trace.SpanBuilder;
import io.opencensus.trace.Status;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.samplers.Samplers;

import java.io.IOException;
import java.util.logging.Logger;

final class GrpcServer {
  private static final Logger logger = Logger.getLogger(GrpcServer.class.getName());
  private static final Tracer tracer = Tracing.getTracer();

  private final int serverPort;
  private Server server;

  GrpcServer(int serverPort) {
    this.serverPort = serverPort;
  }

  void start() throws IOException {
    server = ServerBuilder.forPort(serverPort).addService(new TestExecutionServiceImpl()).build().start();
    logger.info("Server started, listening on " + serverPort);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                // Use stderr here since the logger may have been reset by its JVM shutdown hook.
                System.err.println("*** shutting down gRPC server since JVM is shutting down");
                GrpcServer.this.stop();
                System.err.println("*** server shut down");
              }
            });
  }

  private void stop() {
    if (server != null) {

    }
  }

  static class TestExecutionServiceImpl extends TestExecutionServiceGrpc.TestExecutionServiceImplBase {
    @Override
    public void test(TestRequest req, StreamObserver<TestResponse> responseObserver) {
      long id = req.getId();
      String name = req.getName();
      for (ServiceHop hop : req.getServiceHopsList()) {

      }
      return;
    }
  }
}
