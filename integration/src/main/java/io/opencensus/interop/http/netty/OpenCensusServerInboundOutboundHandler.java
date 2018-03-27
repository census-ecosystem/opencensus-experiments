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
import io.opencensus.trace.Annotation;
import io.opencensus.trace.Span;
import io.opencensus.trace.SpanBuilder;
import io.opencensus.trace.SpanContext;
import io.opencensus.trace.Status;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.propagation.TextFormat;
import java.util.logging.Logger;

/**
 * A {@link ChannelDuplexHandler} that process both inbound and outbound messages for HTTP servers.
 *
 * <p>For inbound messages, it extracts propagation information from {@link HttpRequest} and set it
 * into current {@link ChannelHandlerContext} and starts a new {@link Span}.
 *
 * <p>For outbound messages, it closes the created {@link Span} after the last message of the
 * response is sent.
 *
 * <p>This class should be placed after {@link io.netty.handler.codec.http.HttpServerCodec}.
 */
// TODO(hailongwen): moved this into `opencensus-instrumentation-http-netty` artifact.
public abstract class OpenCensusServerInboundOutboundHandler extends ChannelDuplexHandler {
  private static final Logger logger =
      Logger.getLogger(OpenCensusServerInboundOutboundHandler.class.getName());

  private final Tracer tracer;

  /**
   * Creates a handler.
   *
   * @param tracer the {@link Tracer} to use.
   */
  public OpenCensusServerInboundOutboundHandler(Tracer tracer) {
    this.tracer = tracer;
  }

  /** Determine what {@link TextFormat} to use according to the context and request. */
  public abstract TextFormat getTextFormat(ChannelHandlerContext ctx, HttpRequest msg);

  @Override
  public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
    if (msg instanceof HttpRequest) {
      TextFormat textFormat = getTextFormat(ctx, (HttpRequest) msg);
      // try extract the information.
      if (textFormat != null) {
        Span span = handleStart(textFormat, (HttpRequest) msg);
        // set as current span.
        ctx.channel().attr(NettyUtils.OPENCENSUS_SPAN).set(span);
        ctx.channel().attr(NettyUtils.OPENCENSUS_HTTP_SPAN).set(span);
      }
    }
    // notify down-stream handlers.
    ctx.fireChannelRead(msg);
  }

  @Override
  public void write(ChannelHandlerContext ctx, Object msg, ChannelPromise promise)
      throws Exception {
    // last message
    if (msg instanceof LastHttpContent) {
      Span span = ctx.channel().attr(NettyUtils.OPENCENSUS_HTTP_SPAN).get();
      if (span != null) {
        handleEnd(msg instanceof HttpResponse ? (HttpResponse) msg : null, /* error= */ null, span);
      }
    }
    super.write(ctx, msg, promise);
  }

  /**
   * Extract the {@link SpanContext} from HTTP header and start a new span under this context.
   *
   * <p>This method is a copy of <a
   * href="https://github.com/HailongWen/opencensus-java/blob/1b7864992078f331034b2b157c0c372f34a7ddb9/contrib/http_util/src/main/java/io/opencensus/contrib/http/HttpServerHandler.java#L75">
   * HttpServerHandler.handlerStart(TextFormat.Getter,Carrier,Request)</a> with some minor
   * modifications.
   */
  // TODO(hailongwen): remove this method once the HTTP util is merged.
  private Span handleStart(TextFormat textFormat, HttpRequest request) {
    SpanBuilder spanBuilder = null;
    // customize the spanName
    // String spanName = customizer.getSpanName(request, extractor);
    String spanName = "Server.recv";
    String parseError = null;
    // de-serialize the context
    try {
      SpanContext spanContext = textFormat.extract(request, NettyUtils.HTTP_REQUEST_GETTER);
      spanBuilder = tracer.spanBuilderWithRemoteParent(spanName, spanContext);
    } catch (Exception e) {
      // record this exception
      spanBuilder = tracer.spanBuilder(spanName);
      parseError = e.getMessage() == null ? e.getClass().getSimpleName() : e.getMessage();
    }

    // customize the spanBulider
    // spanBuilder = customizer.customizeSpanBuilder(request, spanBuilder, extractor);
    Span span = spanBuilder.startSpan();
    // log an annotation to indicate the error
    if (parseError != null) {
      span.addAnnotation(Annotation.fromDescription("Error parsing span context: " + parseError));
    }
    // user-defined behaviors
    // customizer.customizeSpanStart(request, span, extractor);
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
    // customizer.customizeSpanEnd(response, error, span, extractor);
    if (response != null && response.status().codeClass() == HttpStatusClass.SUCCESS) {
      span.setStatus(Status.OK);
    } else {
      span.setStatus(Status.UNKNOWN);
    }
    span.end();
  }
}
