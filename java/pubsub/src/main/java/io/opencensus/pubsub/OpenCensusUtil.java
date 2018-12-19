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

import io.opencensus.exporter.trace.stackdriver.StackdriverTraceConfiguration;
import io.opencensus.exporter.trace.stackdriver.StackdriverTraceExporter;

import io.opencensus.common.Duration;
import io.opencensus.common.Scope;
import io.opencensus.trace.Link;
import io.opencensus.trace.Span;
import io.opencensus.trace.SpanId;
import io.opencensus.trace.SpanBuilder;
import io.opencensus.trace.SpanContext;
import io.opencensus.trace.TraceId;
import io.opencensus.trace.TraceOptions;
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
  public static final String OPEN_CENSUS_TRACE_CONTEXT = "OpenCensusTraceContext";
  private static final String PROJECT_ID = ServiceOptions.getDefaultProjectId();
  private static final TraceOptions SAMPLED = TraceOptions.builder().setIsSampled(true).build();
  private static final Tracer tracer = Tracing.getTracer();

  public static void addAnnotation(String annotation) {
    tracer.getCurrentSpan().addAnnotation(annotation);
    logger.log(Level.INFO, annotation);
  }

  public static SpanContext getCurrentSpanContext() {
    return tracer.getCurrentSpan().getContext();
  }

  @MustBeClosed
  public static Scope withSpanContext(String name, SpanContext spanContext) {
    return tracer.spanBuilderWithRemoteParent(name, spanContext).startScopedSpan();
  }


  @MustBeClosed
  public static Scope createScopedSpan(String name) {
    return tracer
        .spanBuilderWithExplicitParent(name, tracer.getCurrentSpan())
        .setRecordEvents(true)
        .setSampler(Samplers.alwaysSample())
        .startScopedSpan();
  }

  public static void addParentLink(String encodedParentSpanContext) {
    addParentLink(tracer.getCurrentSpan(), encodedParentSpanContext);
  }

  private static void addParentLink(Span span, String encodedParentSpanContext) {
    span.addLink(Link.fromSpanContext(
        createSpanContext(encodedParentSpanContext), Link.Type.PARENT_LINKED_SPAN));
  }

  private static SpanContext createSpanContext(String encodedSpanContext) {
    String traceId = getTraceId(encodedSpanContext);
    String spanId = getSpanId(encodedSpanContext);
    String traceOpt = getTraceOpt(encodedSpanContext);
    return SpanContext.create(
        TraceId.fromLowerBase16(traceId),
        SpanId.fromLowerBase16(spanId),
        traceOpt.equals("t") ? SAMPLED : TraceOptions.DEFAULT);
  }


  private static String getTraceId(String encodedSpan) {
    return lookupKey("traceid=", encodedSpan);
  }

  private static String getSpanId(String encodedSpan) {
    return lookupKey("spanid=", encodedSpan);
  }

  private static String getTraceOpt(String encodedSpan) {
    return lookupKey("traceopt=", encodedSpan);
  }

  // encodedSpan = (key=value&)*
  private static String lookupKey(String key, String encodedSpan) {
    int start = encodedSpan.indexOf(key, 0);
    if (start == -1) {
      return "";
    }
    start += key.length();
    int end = encodedSpan.indexOf("&", start);
    if (end == -1) {
      return "";
    }
    return encodedSpan.substring(start, end);
  }

  static {
    if (!PROJECT_ID.isEmpty()) {
      try {
        // Initialize trace exporter.
        StackdriverTraceExporter.createAndRegister(
            StackdriverTraceConfiguration.builder().setProjectId(PROJECT_ID).build());
      } catch (IOException exn) {
        System.err.println("Initializing OCU: Exception: " + exn);
      }
    }
  }
}
