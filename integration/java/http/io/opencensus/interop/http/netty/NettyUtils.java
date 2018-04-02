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

package io.opencensus.interop.http.netty;

import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.http.HttpRequest;
import io.netty.util.AttributeKey;
import io.opencensus.common.Scope;
import io.opencensus.trace.BlankSpan;
import io.opencensus.trace.Span;
import io.opencensus.trace.propagation.TextFormat;
import java.util.logging.Logger;

/** Utilities for netty instrumentation. */
// TODO(hailongwen): moved this into `opencensus-instrumentation-http-netty` artifact.
public class NettyUtils {
  private static final Logger logger = Logger.getLogger(NettyUtils.class.getName());

  private NettyUtils() {}

  /** Customized {@link TextFormat.Getter} for netty. */
  public static final TextFormat.Getter<HttpRequest> HTTP_REQUEST_GETTER =
      new TextFormat.Getter<HttpRequest>() {
        @Override
        public String get(HttpRequest carrier, String key) {
          return carrier.headers().get(key);
        }
      };

  /** Customized {@link TextFormat.Setter} for netty. */
  public static final TextFormat.Setter<HttpRequest> HTTP_REQUEST_SETTER =
      new TextFormat.Setter<HttpRequest>() {
        @Override
        public void put(HttpRequest carrier, String key, String value) {
          carrier.headers().set(key, value);
        }
      };

  /** Active {@link Span} in the current {@link ChannelHandlerContext}. */
  public static final AttributeKey<Span> OPENCENSUS_SPAN =
      AttributeKey.<Span>valueOf("OpenCensus.Span");

  /** Propagation {@link TextFormat} used in current {@link ChannelHandlerContext}. */
  public static final AttributeKey<TextFormat> OPENCENSUS_TEXT_FORMAT =
      AttributeKey.<TextFormat>valueOf("OpenCensus.TextFormat");

  /** {@link Span} for the HTTP request/response process. */
  public static final AttributeKey<Span> OPENCENSUS_HTTP_SPAN =
      AttributeKey.<Span>valueOf("OpenCensus.HTTP.Span");

  /**
   * Scope used in Netty {@link ChannelHandlerContext}.
   *
   * <p>Netty channel handlers are driven by events and may execute in different threads. This class
   * provides a way to enter/exit scope in Netty channel.
   *
   * <p>The implementation is basically a copy of OpenCensus's {@code ScopeInSpan}.
   */
  static final class ChannelScopeInSpan implements Scope {
    private final Channel channel;
    private final Span origSpan;
    private final Span span;
    private boolean endSpan;

    /**
     * Constructs a new {@link ScopeInSpan}.
     *
     * @param span is the {@code Span} to be added to the current {@code io.grpc.Context}.
     */
    private ChannelScopeInSpan(Channel channel, Span span, boolean endSpan) {
      this.origSpan = getCurrentSpan(channel);
      this.channel = channel;
      this.span = span;
      this.endSpan = endSpan;
      channel.attr(OPENCENSUS_SPAN).set(span);
    }

    @Override
    public void close() {
      channel.attr(OPENCENSUS_SPAN).set(origSpan);
      if (endSpan) {
        span.end();
      }
    }
  }

  /**
   * Enter the scope of a {@link Span}.
   *
   * @param channel the channel.
   * @param span the span.
   * @return A new scope.
   */
  public static Scope withSpan(Channel channel, Span span) {
    return new ChannelScopeInSpan(channel, span, true);
  }

  /**
   * Get current active span in the scope.
   *
   * @param channel the channel.
   * @return Active span.
   */
  public static Span getCurrentSpan(Channel channel) {
    Span span = channel.attr(OPENCENSUS_SPAN).get();
    return span == null ? BlankSpan.INSTANCE : span;
  }
}
