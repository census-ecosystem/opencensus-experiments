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

import com.google.cloud.ServiceOptions;
import com.google.cloud.pubsub.v1.AckReplyConsumer;
import com.google.cloud.pubsub.v1.MessageReceiver;
import com.google.cloud.pubsub.v1.OpenCensusUtil.OpenCensusMessageReceiver;
import com.google.cloud.pubsub.v1.Subscriber;
import com.google.pubsub.v1.ProjectSubscriptionName;
import com.google.pubsub.v1.PubsubMessage;

import io.opencensus.common.Scope;
import io.opencensus.tags.TagContext;
import io.opencensus.trace.SpanContext;

public class SubscriberExample {

  // use the default project id
  private static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();

  static class MessageReceiverExample implements MessageReceiver {
    @Override
    public void receiveMessage(PubsubMessage message, AckReplyConsumer consumer) {
      try (Scope subScope =  OpenCensusStatsUtil.createSubscriberScope()) {
        try (Scope latencyScope = OpenCensusStatsUtil.createLatencyScope()) {
          OpenCensusTraceUtil.logCurrentSpan();
          System.out.println("Message Id: " + message.getMessageId());
          String data = message.getData().toStringUtf8();
          System.out.println("Data: " + data);
          OpenCensusTraceUtil.addAnnotationAndLog("Receiver:Message");
          OpenCensusTraceUtil.addAnnotationAndLog("Receiver:Ack: " + data);
          consumer.ack();
          OpenCensusTraceUtil.addAnnotationAndLog("Receiver:Done: " + data);
        }
      }
    }
  }

  /** Receive messages over a subscription. */
  public static void main(String... args) throws Exception {
    // set subscriber id, eg. my-sub
    String subscriptionId = args[0];
    ProjectSubscriptionName subscriptionName = ProjectSubscriptionName.of(
        PROJECT_ID, subscriptionId);
    Subscriber subscriber = null;
    try {
      // create a subscriber bound to the asynchronous OpenCensus message receiver
      MessageReceiver receiver = new OpenCensusMessageReceiver(new MessageReceiverExample());
      subscriber = Subscriber.newBuilder(subscriptionName, receiver).build();
      subscriber.startAsync().awaitRunning();
      // Continue to listen to messages
      while (true) {
        Thread.sleep(Long.MAX_VALUE);
      }
    } finally {
      if (subscriber != null) {
        subscriber.stopAsync();
      }
    }
  }
}
