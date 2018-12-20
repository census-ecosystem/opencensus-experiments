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

package io.opencensus.pubsub;

import com.google.cloud.ServiceOptions;
import com.google.errorprone.annotations.MustBeClosed;
import io.opencensus.common.Scope;
import io.opencensus.exporter.trace.stackdriver.StackdriverTraceConfiguration;
import io.opencensus.exporter.trace.stackdriver.StackdriverTraceExporter;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.samplers.Samplers;
import java.io.IOException;
import java.util.Arrays;
import java.util.Map.Entry;
import java.util.logging.Level;
import java.util.logging.Logger;

final class OpenCensusUtil {
  private static final Logger logger = Logger.getLogger(OpenCensusUtil.class.getName());
  private static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();
  private static final Tracer tracer = Tracing.getTracer();

  public static void addAnnotation(String annotation) {
    tracer.getCurrentSpan().addAnnotation(annotation);
    logger.log(Level.INFO, annotation);
  }

  @MustBeClosed
  public static Scope createScopedSpan(String name) {
    return tracer
        .spanBuilderWithExplicitParent(name, tracer.getCurrentSpan())
        .setRecordEvents(true)
        .setSampler(Samplers.alwaysSample())
        .startScopedSpan();
  }

  static {
    if (!PROJECT_ID.isEmpty()) {
      try {
        // Initialize trace exporter.
        StackdriverTraceExporter.createAndRegister(
            StackdriverTraceConfiguration.builder().setProjectId(PROJECT_ID).build());
      } catch (IOException exn) {
        logger.log(Level.INFO, "Initializing OCU: Exception: " + exn);
      }
    }
  }
}
