/* Copyright 2018 Google Inc.
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

package io.opencensus.appengine.flex;

import com.google.cloud.ServiceOptions;

import io.opencensus.common.Duration;
import io.opencensus.common.Scope;
import io.opencensus.contrib.zpages.ZPageHandlers;
import io.opencensus.exporter.stats.stackdriver.StackdriverStatsConfiguration;
import io.opencensus.exporter.stats.stackdriver.StackdriverStatsExporter;
import io.opencensus.stats.Aggregation.Distribution;
import io.opencensus.stats.BucketBoundaries;
import io.opencensus.stats.Measure;
import io.opencensus.stats.Stats;
import io.opencensus.stats.StatsRecorder;
import io.opencensus.stats.View;
import io.opencensus.tags.TagKey;
import io.opencensus.tags.TagValue;
import io.opencensus.tags.Tagger;
import io.opencensus.tags.Tags;

import java.io.IOException;
import java.util.Arrays;

/**
 * A utility class for using OpenCensus Stats in microservices.
 */
public final class OpenCensusStatsUtil {
  /** The project id, used by all microservices. */
  public static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();

  /** The client_method key allows us to break down the recorded latencies. */
  public static final TagKey CLIENT_METHOD = TagKey.create("client_method");

  /**
   * A utility class to simplify recording latency. When a LatencyStatsRecorder is created,
   * the current time is captured and the specified key/value pair are added to the current
   * tag scope. When the close method is called, the latency is calculated and recorded
   * against the CLIENT_LATENCY measure (and, implicitly, the specified tag).
   */
  public static final class LatencyStatsRecorder implements Scope {
    private final long startTimeMs;
    private final Scope tagScope;

    private LatencyStatsRecorder(long startTimeMs, Scope tagScope) {
      this.startTimeMs = startTimeMs;
      this.tagScope = tagScope;
    }

    /**
     * Factory method that 1. captures the current system time and 2. adds the specified key/value
     * pair to the current tag context.
     */
    public static final LatencyStatsRecorder create(TagKey key, String value) {
      return new LatencyStatsRecorder(
          System.currentTimeMillis(),
          tagger.currentBuilder().put(key, TagValue.create(value)).buildScoped());
    }

    /**
     * On close, records latency against the CLIENT_LATENCY measure and, implicitly, the specified
     * tag.
     */
    @Override
    public void close() {
      long latencyMs = System.currentTimeMillis() - startTimeMs;
      statsRecorder.newMeasureMap().put(CLIENT_LATENCY, latencyMs).record();
      tagScope.close();
    }
  }

  private static final Tagger tagger = Tags.getTagger();
  private static final StatsRecorder statsRecorder = Stats.getStatsRecorder();

  // Measure for Cloud Storage client latency in milliseconds.
  private static final Measure.MeasureDouble CLIENT_LATENCY =
      Measure.MeasureDouble.create(
          "my.org/cloud-storage-client/latency", "Latency in milliseconds.", "ms");

  private static final View CLIENT_LATENCY_VIEW =
      View.create(
          View.Name.create("my.org/cloud-storage-client/latency"),
          "Latency in milliseconds",
          CLIENT_LATENCY,
          Distribution.create(BucketBoundaries.create(Arrays.asList(
              1.0, 2.0, 4.0, 8.0, 16.0, 32.0, 64.0, 128.0, 256.0, 512.0, 1024.0, 2048.0, 4096.0))),
          Arrays.asList(CLIENT_METHOD));

  static {
    if (!PROJECT_ID.isEmpty()) {
      // Register client latency view.
      Stats.getViewManager().registerView(CLIENT_LATENCY_VIEW);
      try {
        // Initialize stats exporter.
        StackdriverStatsExporter.createAndRegister(
            StackdriverStatsConfiguration.builder()
            .setProjectId(PROJECT_ID)
            .setExportInterval(Duration.create(15, 0))
            .build());

        // Starts a HTTP server and registers all Zpages to it.
        ZPageHandlers.startHttpServerAndRegisterAll(3001);
      } catch (IOException exn) {
        if (exn == null) {
          System.err.println("Null exn");
        }
      }
    }
  }
}
