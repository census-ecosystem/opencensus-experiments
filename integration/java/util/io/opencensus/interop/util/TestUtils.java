/*
 * Copyright 2018, OpenCensus Authors
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

package io.opencensus.interop.util;

import static com.google.common.base.Preconditions.checkNotNull;

import com.google.common.collect.Sets;
import com.google.protobuf.ByteString;
import io.opencensus.contrib.http.util.HttpPropagationUtil;
import io.opencensus.interop.EchoResponse;
import io.opencensus.interop.EchoServiceGrpc;
import io.opencensus.tags.InternalUtils;
import io.opencensus.tags.Tag;
import io.opencensus.tags.TagContext;
import io.opencensus.tags.TagKey;
import io.opencensus.tags.TagValue;
import io.opencensus.tags.Tags;
import io.opencensus.tags.propagation.TagContextBinarySerializer;
import io.opencensus.tags.propagation.TagContextDeserializationException;
import io.opencensus.tags.propagation.TagContextSerializationException;
import io.opencensus.trace.SpanContext;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.propagation.TextFormat;
import java.math.BigInteger;
import java.util.Set;
import java.util.logging.Level;
import java.util.logging.Logger;

/** Interop test utilities. */
public class TestUtils {

  private static final Logger logger = Logger.getLogger(TestUtils.class.getName());

  private static final TagKey METHOD_KEY = TagKey.create("method");
  private static final TagValue METHOD_VALUE =
      TagValue.create(EchoServiceGrpc.getEchoMethod().getFullMethodName());

  private TestUtils() {}

  /**
   * Try to get port configuration from the environment and if not specified, use the default one.
   *
   * @param varName the environment variable name.
   * @return the port.
   */
  public static int getPortOrDefault(String varName, int defaultPort) {
    int portNumber = defaultPort;
    String portStr = System.getenv(varName);
    if (portStr != null) {
      try {
        portNumber = Integer.parseInt(portStr);
      } catch (NumberFormatException e) {
        logger.warning(
            String.format("Port %s is invalid, use default port %d.", portStr, defaultPort));
      }
    }
    return portNumber;
  }

  /**
   * Build the {@link io.opencensus.interop.EchoResponse} according to the given {@link SpanContext}
   * and {@link TagContext}.
   *
   * @param spanContext the {@code SpanContext}.
   * @param tagContext the {@code TagContext}. Can be {@code null}.
   * @return a {@code EchoResponse}.
   */
  public static EchoResponse buildResponse(SpanContext spanContext, TagContext tagContext) {
    checkNotNull(spanContext, "spanContext");
    TagContextBinarySerializer serializer = Tags.getTagPropagationComponent().getBinarySerializer();
    EchoResponse.Builder builder = EchoResponse.newBuilder();

    try {
      byte[] traceIdBytes = spanContext.getTraceId().getBytes();
      byte[] spanIdBytes = spanContext.getSpanId().getBytes();
      int traceOptionInt = new BigInteger(spanContext.getTraceOptions().getBytes()).intValue();
      builder
          .setTraceId(ByteString.copyFrom(traceIdBytes))
          .setSpanId(ByteString.copyFrom(spanIdBytes))
          .setTraceOptions(traceOptionInt);

      if (tagContext != null) {
        byte[] tagContextBytes = serializer.toByteArray(tagContext);
        builder.setTagsBlob(ByteString.copyFrom(tagContextBytes));
      }
    } catch (TagContextSerializationException e) {
      logger.log(Level.SEVERE, "Serialization failed.", e);
    }
    return builder.build();
  }

  /**
   * Returns true if all tracing values and tags received in the response match the expected values.
   *
   * @param expectedSpanContext expected {@link SpanContext}. Can be {@code null}.
   * @param expectedTagContext expected {@link TagContext}. Can be {@code null}.
   * @return whether the values in the response matches the expected ones.
   */
  public static boolean verifyResponse(
      SpanContext expectedSpanContext, TagContext expectedTagContext, EchoResponse response) {
    boolean succeeded = true;
    if (expectedSpanContext != null) {
      ByteString expectedTraceIdByteStr =
          ByteString.copyFrom(expectedSpanContext.getTraceId().getBytes());
      int expectedTraceOption =
          new BigInteger(expectedSpanContext.getTraceOptions().getBytes()).intValue();

      if (!response.getTraceId().equals(expectedTraceIdByteStr)) {
        succeeded = false;
        logger.severe(
            String.format(
                "Client received bad trace id. got %s, want %s.",
                response.getTraceId(), expectedTraceIdByteStr));
      }

      int spanIdInt = new BigInteger(expectedSpanContext.getSpanId().getBytes()).intValue();
      if (spanIdInt == 0) {
        succeeded = false;
        logger.severe("Client received bad span id. Got 0, want non-zero.");
      }

      if (!(response.getTraceOptions() == expectedTraceOption)) {
        succeeded = false;
        logger.severe(
            String.format(
                "Client received bad trace options. got %d, want %d.",
                response.getTraceOptions(), expectedTraceOption));
      }
    }

    // HTTP does not support tag propagation.
    if (expectedTagContext != null) {
      // For HTTP, it is still not determined whether to use binary or plain-text in the
      // propagation. So for now just leave the binary serializer as it is.
      TagContextBinarySerializer serializer =
          Tags.getTagPropagationComponent().getBinarySerializer();

      try {
        TagContext actualTagContext =
            serializer.fromByteArray(response.getTagsBlob().toByteArray());
        Set<Tag> actualTags = Sets.<Tag>newHashSet(InternalUtils.getTags(actualTagContext));
        actualTags.remove(Tag.create(METHOD_KEY, METHOD_VALUE));
        Set<Tag> expectedTags = Sets.<Tag>newHashSet(InternalUtils.getTags(expectedTagContext));
        if (!actualTags.equals(expectedTags)) {
          succeeded = false;
          logger.severe(
              String.format(
                  "Client received wrong TagContext. Got %s, want %s.", actualTags, expectedTags));
        }
      } catch (TagContextDeserializationException e) {
        succeeded = false;
        logger.log(Level.SEVERE, "Bad binary format for TagContext.", e);
      }
    }

    return succeeded;
  }

  /**
   * Returns a {@link TextFormat} according to the propagator name, or {@code null} if not
   * applicable.
   *
   * @param propagator name of the propagator
   * @return a {@link TextFormat} according to the propagator name, or {@code null} if not
   *     applicable.
   */
  public static TextFormat getTextFormat(String propagator) {
    if ("b3".equals(propagator)) {
      return Tracing.getPropagationComponent().getB3Format();
    } else if ("google".equals(propagator)) {
      return HttpPropagationUtil.getCloudTraceFormat();
    } else {
      return null;
    }
  }
}
