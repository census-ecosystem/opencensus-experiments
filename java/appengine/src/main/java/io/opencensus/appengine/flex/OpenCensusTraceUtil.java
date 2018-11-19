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

import io.opencensus.exporter.trace.stackdriver.StackdriverTraceConfiguration;
import io.opencensus.exporter.trace.stackdriver.StackdriverTraceExporter;

import io.opencensus.trace.Span;
import io.opencensus.trace.SpanBuilder;
import io.opencensus.trace.TraceOptions;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.samplers.Samplers;

import java.io.IOException;

public class OpenCensusTraceUtil {
  public static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();

  // Tracing
  public static final TraceOptions SAMPLED = TraceOptions.builder().setIsSampled(true).build();
  public static final Tracer tracer = Tracing.getTracer();

  public static SpanBuilder createSpanBuilder(String name) {
    return tracer
        .spanBuilderWithExplicitParent(name, tracer.getCurrentSpan())
        .setRecordEvents(true)
        .setSampler(Samplers.alwaysSample());
  }

  public static Span current() {
    return tracer.getCurrentSpan();
  }

  // Initialization
  static {
    if (!PROJECT_ID.isEmpty()) {
      try {
        // Initialize trace exporter.
        StackdriverTraceExporter.createAndRegister(
            StackdriverTraceConfiguration.builder().setProjectId(PROJECT_ID).build());
      } catch (IOException exn) {
        if (exn == null) {
          System.err.println("Null exn");
        }
      }
    }
  }
}
