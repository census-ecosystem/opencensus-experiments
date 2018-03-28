/*
 * Copyright 2018, OpenCensus Authors
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

package io.opencensus.interop.grpc;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import io.opencensus.interop.EchoRequest;
import io.opencensus.interop.EchoResponse;
import io.opencensus.interop.EchoServiceGrpc.EchoServiceImplBase;
import io.opencensus.interop.util.TestUtils;
import io.opencensus.tags.Tagger;
import io.opencensus.tags.Tags;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.logging.Level;
import java.util.logging.Logger;

/** Interop test server for gRPC interop testing. */
public final class GrpcInteropTestServer {

  private final int portNumber;

  private static final Logger logger = Logger.getLogger(GrpcInteropTestServer.class.getName());
  private static final Tagger tagger = Tags.getTagger();
  private static final Tracer tracer = Tracing.getTracer();

  private GrpcInteropTestServer(int port) {
    portNumber = port;
  }

  private void run() throws Exception {
    ScheduledExecutorService executor = Executors.newSingleThreadScheduledExecutor();
    EchoServiceImpl echoService = new EchoServiceImpl();
    Server server =
        ServerBuilder.forPort(portNumber)
            .addService(echoService)
            .executor(Executors.newFixedThreadPool(1))
            .build();
    server.start();
    logger.info("Started on " + server.getPort());

    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                logger.info("Shutting down");
                try {
                  server.shutdown();
                } catch (Exception e) {
                  logger.log(Level.SEVERE, "Exception thrown when shutting down.", e);
                }
              }
            });

    try {
      server.awaitTermination();
      logger.info("Server terminated");
      executor.shutdown();
      executor.awaitTermination(1, TimeUnit.SECONDS);
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      throw new RuntimeException(e);
    }
  }

  private static final class EchoServiceImpl extends EchoServiceImplBase {

    @Override
    public void echo(EchoRequest request, StreamObserver<EchoResponse> responseObserver) {
      EchoResponse response =
          TestUtils.buildResponse(
              tracer.getCurrentSpan().getContext(), tagger.getCurrentTagContext());
      responseObserver.onNext(response);
      responseObserver.onCompleted();
    }
  }

  /** Main launcher of the test server. */
  public static void main(String[] args) throws Exception {
    int port =
        TestUtils.getPortOrDefault(
            GrpcInteropTestUtils.ENV_PORT_KEY_JAVA, GrpcInteropTestUtils.DEFAULT_PORT_JAVA);
    new GrpcInteropTestServer(port).run();
  }
}
