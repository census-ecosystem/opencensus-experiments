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
import com.google.errorprone.annotations.MustBeClosed;

import io.opencensus.common.Scope;
import io.opencensus.exporter.trace.stackdriver.StackdriverTraceConfiguration;
import io.opencensus.exporter.trace.stackdriver.StackdriverTraceExporter;
import io.opencensus.trace.SpanContext;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.samplers.Samplers;

import java.io.IOException;
import java.util.logging.Level;
import java.util.logging.Logger;

final class OpenCensusTraceUtil {
  private static final Logger logger = Logger.getLogger(OpenCensusTraceUtil.class.getName());
  private static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();
  private static final Tracer tracer = Tracing.getTracer();

  public static void addAnnotation(String annotation) {
    tracer.getCurrentSpan().addAnnotation(annotation);
    logger.log(Level.INFO, annotation);
  }

  @MustBeClosed
  public static Scope createScopedSampledSpan(String name) {
    return tracer
        .spanBuilderWithExplicitParent(name, tracer.getCurrentSpan())
        .setRecordEvents(true)
        .setSampler(Samplers.alwaysSample())
        .startScopedSpan();
  }

  public static void logCurrentSpan() {
    SpanContext ctxt = tracer.getCurrentSpan().getContext();
    logger.log(Level.INFO, "OpenCensusTraceUtil: logCurrentSpan(): "
        + "traceid=" + ctxt.getTraceId().toLowerBase16()
        + "&spanid=" + ctxt.getSpanId().toLowerBase16()
        + "&traceopt=" + (ctxt.getTraceOptions().isSampled() ? "t&" : "f&"));
  }

  static {
    if (!PROJECT_ID.isEmpty()) {
      try {
        // Initialize trace exporter.
        StackdriverTraceExporter.createAndRegister(
            StackdriverTraceConfiguration.builder().setProjectId(PROJECT_ID).build());
      } catch (IOException exn) {
        logger.log(Level.INFO, "Initializing OpenCensusTraceUtil: Exception: " + exn);
      }
    }
  }
}
