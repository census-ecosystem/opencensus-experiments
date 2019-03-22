/* Copyright 2019 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.example.pubsub;

import com.google.cloud.ServiceOptions;

import io.opencensus.common.Duration;
import io.opencensus.common.Scope;
import io.opencensus.exporter.stats.stackdriver.StackdriverStatsConfiguration;
import io.opencensus.exporter.stats.stackdriver.StackdriverStatsExporter;
import io.opencensus.stats.Aggregation.Distribution;
import io.opencensus.stats.BucketBoundaries;
import io.opencensus.stats.Measure;
import io.opencensus.stats.Stats;
import io.opencensus.stats.StatsRecorder;
import io.opencensus.stats.View;
import io.opencensus.tags.TagContext;
import io.opencensus.tags.TagKey;
import io.opencensus.tags.TagValue;
import io.opencensus.tags.Tagger;
import io.opencensus.tags.Tags;

import java.io.IOException;
import java.util.Arrays;

/**
 * A utility class for the OpenCensus Stats Pub/Sub example.
 */
final class OpenCensusStatsUtil {
  private static final Tagger tagger = Tags.getTagger();

  private static final StatsRecorder statsRecorder = Stats.getStatsRecorder();

  // The method key allows us to break down the recorded latencies.
  private static final TagKey METHOD = TagKey.create("METHOD");

  // The method key allows us to break down the recorded latencies.
  private static final TagKey PUBLISHER = TagKey.create("PUBLISHER");

  private static final TagKey SUBSCRIBER = TagKey.create("SUBSCRIBER");

  private static int pubCount = 0;

  static Scope createPublisherScope() {
    TagValue val = TagValue.create("publisher-" + pubCount);
    pubCount = (pubCount + 1) % 10;
    return tagger.currentBuilder().put(PUBLISHER, val).buildScoped();
  }

  private static int subCount = 0;

  static Scope createSubscriberScope() {
    TagValue val = TagValue.create("subscriber-" + subCount);
    subCount = (subCount + 1) % 10;
    return tagger.currentBuilder().put(SUBSCRIBER, val).buildScoped();
  }

  /**
   * Creates a Scope that captures the current system time and records latency when Scope is
   * close()'d.
   */
  static final Scope createLatencyScope() {
    return new LatencyStatsRecorder();
  }

  /*
   * A utility class to simplify recording latency. When a LatencyStatsRecorder is created,
   * the current time is captured and the specified key/value pair are added to the current
   * tag scope. When the close method is called, the latency is calculated and recorded
   * against the LATENCY measure (and, implicitly, the specified tag).
   */
  private static final class LatencyStatsRecorder implements Scope {
    private final long startTimeMs;

    private LatencyStatsRecorder() {
      this.startTimeMs = System.currentTimeMillis();
    }

    /**
     * On close, records latency against the LATENCY measure.
     */
    @Override
    public void close() {
      long latencyMs = System.currentTimeMillis() - startTimeMs;
      statsRecorder.newMeasureMap().put(LATENCY, latencyMs).record();
    }
  }

  // Measure for Pub/Sub latency in milliseconds.
  private static final Measure.MeasureDouble LATENCY =
      Measure.MeasureDouble.create("my.io/cloud-pubsub/latency", "Latency in milliseconds.", "ms");

  private static final View LATENCY_VIEW =
      View.create(
          View.Name.create("my.io/cloud-pubsub/latency"),
          "Latency in milliseconds",
          LATENCY,
          Distribution.create(BucketBoundaries.create(Arrays.asList(
              1.0, 2.0, 4.0, 8.0, 16.0, 32.0, 64.0, 128.0, 256.0, 512.0, 1024.0, 2048.0, 4096.0))),
          Arrays.asList(PUBLISHER, SUBSCRIBER));

  static {
    String projectId = ServiceOptions.getDefaultProjectId();

    if (projectId != null && !projectId.isEmpty()) {
      // Register latency view.
      Stats.getViewManager().registerView(LATENCY_VIEW);
      try {
        // Initialize stats exporter.
        StackdriverStatsExporter.createAndRegister(
            StackdriverStatsConfiguration.builder()
            .setProjectId(PROJECT_ID)
            .setExportInterval(Duration.create(15, 0))
            .build());
      } catch (IOException exn) {
        if (exn == null) {
          System.err.println("Null exn");
        }
      }
    }
  }
}
