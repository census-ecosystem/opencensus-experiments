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

import static com.google.common.base.Preconditions.checkNotNull;

import io.netty.channel.ChannelDuplexHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelPromise;
import io.netty.handler.codec.http.HttpRequest;
import io.netty.handler.codec.http.HttpResponse;
import io.netty.handler.codec.http.HttpStatusClass;
import io.netty.handler.codec.http.LastHttpContent;
import io.opencensus.common.Scope;
import io.opencensus.trace.Span;
import io.opencensus.trace.SpanBuilder;
import io.opencensus.trace.SpanContext;
import io.opencensus.trace.Status;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.propagation.TextFormat;

/**
 * A {@link ChannelDuplexHandler} that process both inbound and outbound messages for HTTP clients.
 *
 * <p>For outbound messages, it sets the propagation information into the {@link HttpRequest} and
 * starts a new {@link Span}.
 *
 * <p>For inbound messages, it closes the created {@link Span} after the last message of the
 * response is received.
 *
 * <p>This should be placed after the {@link io.netty.handler.codec.http.HttpClientCodec}.
 */
// TODO(hailongwen): moved this into `opencensus-instrumentation-http-netty` artifact.
public final class OpenCensusClientInboundOutboundHandler extends ChannelDuplexHandler {

  private final Tracer tracer;
  private final TextFormat textFormat;

  /**
   * Creates a handler.
   *
   * @param tracer the {@link Tracer} to use.
   * @param textFormat the {@link TextFormat} used in propagation.
   */
  public OpenCensusClientInboundOutboundHandler(Tracer tracer, TextFormat textFormat) {
    this.tracer = tracer;
    this.textFormat = textFormat;
  }

  @Override
  public void write(ChannelHandlerContext ctx, Object m, ChannelPromise promise) throws Exception {
    if (m instanceof HttpRequest) {
      HttpRequest req = (HttpRequest) m;
      if (textFormat != null) {
        Span span = ctx.channel().attr(NettyUtils.OPENCENSUS_SPAN).get();
        if (span != null) {
          try (Scope scope = tracer.withSpan(span)) {
            handleStart(req);
          }
        }
      }
    }
    ctx.write(m, promise);
  }

  @Override
  public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
    if (msg instanceof LastHttpContent) {
      Span span = ctx.channel().attr(NettyUtils.OPENCENSUS_HTTP_SPAN).get();
      if (span != null) {
        handleEnd(msg instanceof HttpResponse ? (HttpResponse) msg : null, /* error= */ null, span);
      }
    }
    ctx.fireChannelRead(msg);
  }

  /**
   * Start a new {@link Span] and inject its {@link SpanContext} information to HTTP header to be
   * propagated.
   *
   * <p>This method is a copy of
   * <a href="https://github.com/HailongWen/opencensus-java/blob/1b7864992078f331034b2b157c0c372f34a7ddb9/contrib/http_util/src/main/java/io/opencensus/contrib/http/HttpClientHandler.java#L74">
   * HttpClientHandler.handlerStart(TextFormat.Getter,Carrier,Request)</a> with some minor
   * modifications.
   */
  // TODO(hailongwen): remove this method once the HTTP util is merged.
  private Span handleStart(HttpRequest request) {
    checkNotNull(request, "request");
    // customize span name
    // String spanName = customizer.getSpanName(request, extractor);
    String spanName = "Client.send";
    // customize span builder
    // SpanBuilder builder =
    //     customizer.customizeSpanBuilder(request, tracer.spanBuilder(spanName), extractor);
    SpanBuilder builder = tracer.spanBuilder(spanName);
    Span span = builder.startSpan();

    // user-defined behaviors
    // customizer.customizeSpanStart(request, span, extractor);

    // inject propagation header
    SpanContext spanContext = span.getContext();
    if (!spanContext.equals(SpanContext.INVALID)) {
      textFormat.inject(spanContext, request, NettyUtils.HTTP_REQUEST_SETTER);
    }
    return span;
  }

  /**
   * Close the HTTP span.
   *
   * <p>This method is a copy of <a
   * href="https://github.com/HailongWen/opencensus-java/blob/1b7864992078f331034b2b157c0c372f34a7ddb9/contrib/http_util/src/main/java/io/opencensus/contrib/http/HttpHandler.java#L124">
   * HttpHandler.handleEnd(Response,Throwable,Span)</a> with some minor modifications.
   */
  // TODO(hailongwen): remove this method once the HTTP util is merged.
  private void handleEnd(HttpResponse response, Throwable error, Span span) {
    checkNotNull(span, "span");
    // user-customized handling
    // customizer.customizeSpanStart(response, error, span, extractor);
    if (response != null && response.status().codeClass() == HttpStatusClass.SUCCESS) {
      span.setStatus(Status.OK);
    } else {
      span.setStatus(Status.UNKNOWN);
    }
    span.end();
  }
}
