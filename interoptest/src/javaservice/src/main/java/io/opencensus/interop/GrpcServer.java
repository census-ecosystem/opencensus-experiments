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

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
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
import java.util.concurrent.TimeUnit;
import java.util.List;
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
    logger.info("Java gRPC Server started, listening on " + serverPort);
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

  static final TestResponse serviceHop(ServiceHop hop, List<ServiceHop> hops) {
    String name = hop.getService().getName();
    int port = hop.getService().getPort();
    String host = hop.getService().getHost();
    Spec spec = hop.getService().getSpec();
    if (spec.getTransport().equals(Spec.Transport.GRPC) &&
        spec.getPropagation().equals(Spec.Propagation.BINARY_FORMAT_PROPAGATION)) {


    }
    return null;
  }

  private static final CommonResponseStatus OK_STATUS = CommonResponseStatus
      .newBuilder().setStatus(io.opencensus.interop.Status.SUCCESS).setError("").build();

  static class TestExecutionServiceImpl extends TestExecutionServiceGrpc.TestExecutionServiceImplBase {
    @Override
    public void test(TestRequest req, StreamObserver<TestResponse> responseObserver) {
      logger.info("Java gRPC Test Server RPC: start");
      long id = req.getId();
      String name = req.getName();
      TestResponse.Builder respBuilder = TestResponse.newBuilder().setId(id).addStatus(OK_STATUS);
      if (req.getServiceHopsCount() != 0) {
        List<ServiceHop> hops = req.getServiceHopsList();
        ServiceHop hop = hops.remove(0); // dpo: this may not work.
        TestResponse rest = serviceHop(hop, hops);
        // null indicates unsupported service hop.
        if (rest != null) {
          respBuilder.addAllStatus(rest.getStatusList());
        }
      }
      responseObserver.onNext(respBuilder.build());
      responseObserver.onCompleted();
      logger.info("Java gRPC Test Server RPC: done");
    }
  }

  static class GrpcClient {
    private final String host;
    private final int port;

    // Construct client connecting to server at {@code host:port}.
    GrpcClient(String host, int port) {
      this.host = host;
      this.port = port;
    }

    TestResponse test(long id, String name, List<ServiceHop> serviceHops) {
      ManagedChannel channel = ManagedChannelBuilder.forAddress(host, port)
                               // Channels are secure by default (via SSL/TLS).
                               // For the example we disable TLS to avoid
                               // needing certificates.
                               .usePlaintext(true)
                               .build();
      TestExecutionServiceGrpc.TestExecutionServiceBlockingStub blockingStub =
          TestExecutionServiceGrpc.newBlockingStub(channel);
      TestRequest request =
          TestRequest.newBuilder().setId(id).setName(name).addAllServiceHops(serviceHops).build();
      TestResponse response = blockingStub.test(request);
      channel.shutdown();//.awaitTermination(5, TimeUnit.SECONDS);
      return response;
    }
  }
}
