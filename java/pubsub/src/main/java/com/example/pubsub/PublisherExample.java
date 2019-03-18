/*
 * Copyright 2019 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.example.pubsub;

import com.google.api.core.ApiFuture;
import com.google.api.core.ApiFutures;
import com.google.cloud.ServiceOptions;
import com.google.cloud.pubsub.v1.OpenCensusUtil;
import com.google.cloud.pubsub.v1.Publisher;
import com.google.protobuf.ByteString;
import com.google.pubsub.v1.ProjectTopicName;
import com.google.pubsub.v1.PubsubMessage;

import io.opencensus.common.Scope;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class PublisherExample {

  // use the default project id
  private static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();

  /** Publish messages to a topic.
   * @param args topic name, number of messages
   */
  public static void main(String... args) throws Exception {
    // topic id, eg. "my-topic"
    String topicId = args[0];
    int messageCount = Integer.parseInt(args[1]);
    ProjectTopicName topicName = ProjectTopicName.of(PROJECT_ID, topicId);
    Publisher publisher = null;
    List<ApiFuture<String>> futures = new ArrayList<>();

    // Scope the span for the requests.
    try (Scope scope = OpenCensusTraceUtil.createScopedSampledSpan("Publisher")) {
      OpenCensusTraceUtil.addAnnotationAndLog("Publisher:Begin");
      // Create a publisher instance with default settings bound to the topic.
      publisher = Publisher
                  .newBuilder(topicName)
                  .setTransform(OpenCensusUtil.OPEN_CENSUS_MESSAGE_TRANSFORM)
                  .build();
      for (int i = 0; i < messageCount; i++) {
        try (
            Scope traceScope = OpenCensusTraceUtil.createScopedSampledSpan("PublisherRoot-" + i);
            Scope statsScope = OpenCensusStatsUtil.createPublisherScope()) {
          try (Scope latencyScope = OpenCensusStatsUtil.createLatencyScope()) {
            OpenCensusTraceUtil.addAnnotationAndLog("Publisher:message-" + i);
            // Propagate the span information with the request.
            String message = "message-" + i;
            // convert message to bytes
            ByteString data = ByteString.copyFromUtf8(message);
            PubsubMessage pubsubMessage = PubsubMessage.newBuilder()
                .setData(data)
                .build();
            // Schedule a message to be published. Messages are automatically batched.
            ApiFuture<String> future = publisher.publish(pubsubMessage);
            futures.add(future);
          }
        }
      }
    } finally {
      // Wait on any pending requests
      List<String> messageIds = ApiFutures.allAsList(futures).get();
      for (String messageId : messageIds) {
        System.out.println(messageId);
      }
      if (publisher != null) {
        // When finished with the publisher, shutdown to free up resources.
        publisher.shutdown();
      }
      OpenCensusTraceUtil.addAnnotationAndLog("Publisher:End");
      Thread.sleep(5000);
    }
  }
}
