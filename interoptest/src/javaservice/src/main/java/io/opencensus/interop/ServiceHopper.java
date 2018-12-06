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

import static io.opencensus.interop.Spec.Propagation.B3_FORMAT_PROPAGATION;
import static io.opencensus.interop.Spec.Propagation.BINARY_FORMAT_PROPAGATION;
import static io.opencensus.interop.Spec.Propagation.TRACE_CONTEXT_FORMAT_PROPAGATION;
import static io.opencensus.interop.Spec.Transport.HTTP;
import static io.opencensus.interop.Spec.Transport.GRPC;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.opencensus.common.Scope;
import io.opencensus.tags.TagContextBuilder;
import io.opencensus.tags.TagKey;
import io.opencensus.tags.TagValue;
import io.opencensus.tags.Tagger;
import io.opencensus.tags.Tags;

// TODO(dpo): decide whether we need this.
// import java.util.concurrent.TimeUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.logging.Logger;

final class ServiceHopper {
  private static final Logger logger = Logger.getLogger(ServiceHopper.class.getName());
  private static final Tagger tagger = Tags.getTagger();
  private static final CommonResponseStatus SUCCESS =
      CommonResponseStatus.newBuilder().setStatus(Status.SUCCESS).setError("").build();
  private static final CommonResponseStatus B3_FORMAT_FAILURE = CommonResponseStatus.newBuilder()
      .setStatus(Status.FAILURE).setError("B3 Format Unsupported").build();
  private static final CommonResponseStatus TC_FORMAT_FAILURE = CommonResponseStatus.newBuilder()
      .setStatus(Status.FAILURE).setError("Trace Context Format Unsupported").build();

  static final TestResponse serviceHop(long id, String name, List<ServiceHop> hops) {
    // TODO(dpo): verify base case.
    if (hops.size() == 0) {
      return TestResponse.newBuilder().setId(id).build();
    }
    ServiceHop first = hops.get(0);
    List<ServiceHop> rest = new ArrayList(hops.size() - 1);
    rest.addAll(1, hops);
    Spec.Transport transport = first.getService().getSpec().getTransport();
    Spec.Propagation propagation = first.getService().getSpec().getPropagation();
    try (Scope tagScope = scopeTags(first.getTagsList())) {
      switch (transport) {
        case HTTP:
          switch (propagation) {
            case B3_FORMAT_PROPAGATION:
              return httpB3FormatServiceHopper(id, name, first, rest);
            case TRACE_CONTEXT_FORMAT_PROPAGATION:
              return httpTraceContextFormatServiceHopper(id, name, first, rest);
            default:
              logger.info("Unsupported propagation: " + propagation);
              return null;
          }
        case GRPC:
          switch (propagation) {
            case BINARY_FORMAT_PROPAGATION:
              return grpcBinaryFormatServiceHopper(id, name, first, rest);
            default:
              logger.info("Unsupported propagation: " + propagation);
              return null;
          }
        default:
          logger.info("Unknown transport: " + transport);
      }
      return null;
    }
  }

  private static Scope scopeTags(List<Tag> tags) {
    TagContextBuilder builder = tagger.currentBuilder();
    for (Tag tag : tags) {
      builder.put(TagKey.create(tag.getKey()), TagValue.create(tag.getValue()));
    }
    return builder.buildScoped();
  }

  private static TestResponse grpcBinaryFormatServiceHopper(
      long id, String name, ServiceHop first, List<ServiceHop> rest) {
    String host = first.getService().getHost();
    int port = first.getService().getPort();
    // TODO(dpo): we could consider caching these channels and reusing them.
    ManagedChannel channel = ManagedChannelBuilder.forAddress(host, port)
                             // Channels are secure by default (via SSL/TLS).
                             // For the example we disable TLS to avoid
                             // needing certificates.
                             .usePlaintext(true)
                             .build();
    TestExecutionServiceGrpc.TestExecutionServiceBlockingStub blockingStub =
        TestExecutionServiceGrpc.newBlockingStub(channel);
    TestRequest request =
        TestRequest.newBuilder().setId(id).setName(name).addAllServiceHops(rest).build();
    TestResponse response = blockingStub.test(request);
    channel.shutdown();//.awaitTermination(5, TimeUnit.SECONDS);
    return addSuccessStatus(response);
  }

  private static TestResponse httpB3FormatServiceHopper(
      long id, String name, ServiceHop first, List<ServiceHop> rest) {
    return TestResponse.newBuilder().setId(id).addStatus(B3_FORMAT_FAILURE).build();
  }

  private static TestResponse httpTraceContextFormatServiceHopper(
      long id, String name, ServiceHop first, List<ServiceHop> rest) {
    return TestResponse.newBuilder().setId(id).addStatus(TC_FORMAT_FAILURE).build();
  }

  private static final TestResponse addSuccessStatus(TestResponse response) {
    return TestResponse.newBuilder()
        .setId(response.getId())
        .addStatus(SUCCESS)
        .addAllStatus(response.getStatusList())
        .build();
  }
}
