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

import com.google.protobuf.TextFormat;
import com.google.protobuf.TextFormat.ParseException;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.opencensus.common.Scope;
import io.opencensus.contrib.http.jetty.client.OcJettyHttpClient;
import io.opencensus.tags.TagContextBuilder;
import io.opencensus.tags.TagKey;
import io.opencensus.tags.TagValue;
import io.opencensus.tags.Tagger;
import io.opencensus.tags.Tags;
import java.util.List;
import org.apache.log4j.Logger;
import org.eclipse.jetty.client.api.ContentResponse;
import org.eclipse.jetty.client.HttpRequest;
import org.eclipse.jetty.client.util.StringContentProvider;
import org.eclipse.jetty.http.HttpMethod;

final class ServiceHopper {
  private static final Logger logger = Logger.getLogger(ServiceHopper.class.getName());
  private static final Tagger tagger = Tags.getTagger();
  private static final CommonResponseStatus SUCCESS =
      CommonResponseStatus.newBuilder().setStatus(Status.SUCCESS).build();

  static final TestResponse serviceHop(TestRequest request) {
    long id = request.getId();
    String name = request.getName();
    List<ServiceHop> hops = request.getServiceHopsList();
    if (hops.size() == 0) {
      return TestResponse.newBuilder().setId(id).addStatus(SUCCESS).build();
    }
    ServiceHop first = hops.get(0);
    String host = first.getService().getHost();
    int port = first.getService().getPort();
    Spec.Transport transport = first.getService().getSpec().getTransport();
    Spec.Propagation propagation = first.getService().getSpec().getPropagation();
    List<ServiceHop> rest = hops.subList(1, hops.size());
    TestRequest restRequest =
        TestRequest.newBuilder().setId(id).setName(name).addAllServiceHops(rest).build();
    try (Scope tagScope = scopeTags(first.getTagsList())) {
      switch (transport) {
        case HTTP:
          switch (propagation) {
            case B3_FORMAT_PROPAGATION:
              return httpServiceHop(id, name, host, port, restRequest, B3Format.httpClient);
            case TRACE_CONTEXT_FORMAT_PROPAGATION:
              return httpServiceHop(id, name, host, port, restRequest, TcFormat.httpClient);
            default:
              return setFailureStatus(id, "Unsupported propagation: " + propagation);
          }
        case GRPC:
          switch (propagation) {
            case BINARY_FORMAT_PROPAGATION:
              return grpcServiceHop(id, name, host, port, restRequest);
            default:
              return setFailureStatus(id, "Unsupported propagation: " + propagation);
          }
        default:
          return setFailureStatus(id, "Unknown transport: " + transport);
      }
    }
  }

  private static Scope scopeTags(List<Tag> tags) {
    TagContextBuilder builder = tagger.currentBuilder();
    for (Tag tag : tags) {
      builder.put(TagKey.create(tag.getKey()), TagValue.create(tag.getValue()));
    }
    return builder.buildScoped();
  }

  private static TestResponse setFailureStatus(long id, String msg) {
    CommonResponseStatus status =
        CommonResponseStatus.newBuilder().setStatus(Status.FAILURE).setError(msg).build();
    return TestResponse.newBuilder().setId(id).addStatus(status).build();
  }

  private static final TestResponse addSuccessStatus(TestResponse response) {
    return TestResponse.newBuilder()
        .setId(response.getId())
        .addStatus(SUCCESS)
        .addAllStatus(response.getStatusList())
        .build();
  }

  private static TestResponse grpcServiceHop(
      long id, String name, String host, int port, TestRequest request) {
    ManagedChannel channel = ManagedChannelBuilder.forAddress(host, port)
                             // Channels are secure by default (via SSL/TLS).
                             // For the test, we disable TLS to avoid
                             // needing certificates.
                             .usePlaintext(true)
                             .build();
    TestExecutionServiceGrpc.TestExecutionServiceBlockingStub blockingStub =
        TestExecutionServiceGrpc.newBlockingStub(channel);
    TestResponse response = blockingStub.test(request);
    channel.shutdown();
    return addSuccessStatus(response);
  }

  private static TestResponse httpServiceHop(long id, String name, String host, int port,
      TestRequest request, OcJettyHttpClient httpClient) {
    try {
      HttpRequest httpRequest = (HttpRequest) httpClient
          .newRequest("http://" + host + ":" + port + "/test/request")
          .method(HttpMethod.POST);
      httpRequest.content(new StringContentProvider(request.toString()));
      ContentResponse httpResponse = httpRequest.send();
      TestResponse.Builder responseBuilder = TestResponse.newBuilder();
      TextFormat.merge(httpResponse.getContentAsString(), responseBuilder);
      return addSuccessStatus(responseBuilder.build());
    } catch (Exception exn) {
          return setFailureStatus(id, "HTTP Service Hopper Error: " + exn);
    }
  }

  private static class B3Format {
    // TODO(dpo): specify propagation format when available.
    private static OcJettyHttpClient httpClient = new OcJettyHttpClient();
    static {
      try {
        httpClient.start();
        logger.info("HTTP B3 Client Started");
      } catch (Exception exn) {
        logger.info("HTTP B3 Client Initialization Failed:  " + exn);
      }
    }
  }

  private static class TcFormat {
    // TODO(dpo): specify propagation format when available.
    private static OcJettyHttpClient httpClient = new OcJettyHttpClient();
    static {
      try {
        httpClient.start();
        logger.info("HTTP Trace Context Client Started");
      } catch (Exception exn) {
        logger.info("HTTP Trace Context Client Initialization Failed:  " + exn);
      }
    }
  }
}
