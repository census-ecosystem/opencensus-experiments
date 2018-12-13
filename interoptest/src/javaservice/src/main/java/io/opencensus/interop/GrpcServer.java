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
import java.io.IOException;
import java.util.concurrent.TimeUnit;
import java.util.List;
import org.apache.log4j.Logger;

final class GrpcServer {
  private static final Logger logger = Logger.getLogger(GrpcServer.class.getName());

  private final int port;
  private Server server;

  private GrpcServer(int port) {
    this.port = port;
  }

  static GrpcServer createWithBinaryFormatPropagation(int port) {
    return new GrpcServer(port);
  }

  void start() throws IOException {
    server = ServerBuilder.forPort(port).addService(new TestExecutionServiceImpl()).build().start();
    logger.info("Java gRPC server started, listening on " + port);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                // Use stderr here since the logger may have been reset by its JVM shutdown hook.
                System.err.println("*** JVM shutting down, shutting down gRPC server");
                GrpcServer.this.stop();
                System.err.println("*** gRPC server shut down");
              }
            });
  }

  private void stop() {
    if (server != null) {
      server.shutdown();
    }
  }

  private static final class TestExecutionServiceImpl
      extends TestExecutionServiceGrpc.TestExecutionServiceImplBase {
    @Override
    public void test(TestRequest request, StreamObserver<TestResponse> responseObserver) {
      TestResponse response = ServiceHopper.serviceHop(request);
      responseObserver.onNext(response);
      responseObserver.onCompleted();
    }
  }
}
