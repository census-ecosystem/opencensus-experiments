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

package io.opencensus.interop.http;

import static io.netty.handler.codec.http.HttpMethod.GET;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;

import com.google.protobuf.util.JsonFormat;
import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.ByteBufInputStream;
import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.DefaultHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpClientCodec;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpRequest;
import io.opencensus.common.Scope;
import io.opencensus.interop.EchoResponse;
import io.opencensus.interop.TestUtils;
import io.opencensus.interop.http.HttpInteropTestUtils.Setup;
import io.opencensus.interop.http.netty.NettyUtils;
import io.opencensus.interop.http.netty.OpenCensusClientInboundOutboundHandler;
import io.opencensus.trace.SpanContext;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.propagation.TextFormat;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.Reader;
import java.util.Map.Entry;
import java.util.logging.Level;
import java.util.logging.Logger;

/**
 * Client for HTTP interop testing.
 *
 * <p>It will iterate through every setup in {@link HttpInteropTestUtils.SETUP_MAP} and send
 * requests according to the configuration.
 */
public final class HttpInteropTestClient {

  private static final String HOST = "localhost";
  private static final Logger logger = Logger.getLogger(HttpInteropTestClient.class.getName());
  private static final Tracer tracer = Tracing.getTracer();

  private final String setupName;
  private final int serverPort;
  private final String propagator;
  private final TextFormat textFormat;

  private final EchoResponse.Builder builder = EchoResponse.newBuilder();
  private boolean received = false;

  private HttpInteropTestClient(String setupName, int serverPort, String propagator) {
    this.setupName = setupName;
    this.serverPort = serverPort;
    this.propagator = propagator;
    this.textFormat = TestUtils.getTextFormat(propagator);
  }

  private void run() {
    boolean succeeded = false;
    EventLoopGroup group = new NioEventLoopGroup();
    Bootstrap b = new Bootstrap();
    b.group(group).channel(NioSocketChannel.class).handler(new HttpInteropTestClientInitializer());

    try {
      Channel ch = b.connect(HOST, serverPort).sync().channel();
      try (Scope scope = NettyUtils.withSpan(ch, tracer.spanBuilder("test").startSpan())) {
        HttpRequest request = new DefaultHttpRequest(HTTP_1_1, GET, "/?p=" + propagator, false);
        ch.writeAndFlush(request);

        synchronized (builder) {
          // Netty client is async, so we have to wait for the response before exit.
          if (!received) builder.wait();
          SpanContext spanContext = NettyUtils.getCurrentSpan(ch).getContext();
          succeeded = TestUtils.verifyResponse(spanContext, null, builder.build());
        }
      }
    } catch (Throwable error) {
      logger.log(Level.SEVERE, "Failed to get response.", error);
    } finally {
      logger.info(
          String.format(
              "%s-%s (%s:%d) - %s",
              setupName, propagator, HOST, serverPort, succeeded ? "PASSED" : "FAILED"));
      group.shutdownGracefully();
    }
  }

  private final class HttpInteropTestClientInitializer extends ChannelInitializer<SocketChannel> {
    @Override
    public void initChannel(SocketChannel ch) {
      ChannelPipeline pipeline = ch.pipeline();
      pipeline.addLast(new HttpClientCodec());
      // aggregate all chunked messages into one FullHttpResponse.
      pipeline.addLast(new HttpObjectAggregator(1048576));
      pipeline.addLast(new OpenCensusClientInboundOutboundHandler(tracer, textFormat));
      pipeline.addLast(new HttpInteropTestClientInboundHandler());
    }
  }

  private final class HttpInteropTestClientInboundHandler
      extends SimpleChannelInboundHandler<FullHttpResponse> {
    @Override
    // Note that in Netty 5+, this method is renamed to `messageReceived`.
    public void channelRead0(ChannelHandlerContext ctx, FullHttpResponse msg) {
      ByteBuf buf = msg.content();
      try (Reader reader = new BufferedReader(new InputStreamReader(new ByteBufInputStream(buf)))) {
        synchronized (builder) {
          JsonFormat.parser().merge(reader, builder);
          received = true;
          builder.notifyAll();
        }
      } catch (Throwable e) {
        logger.log(Level.SEVERE, "Error parsing the json result.", e);
      }
      ctx.close();
    }
  }

  /** Main launcher of the test client. */
  public static void main(String[] args) {
    for (Entry<String, Setup> setupConf : HttpInteropTestUtils.SETUP_MAP.entrySet()) {
      String setupName = setupConf.getKey();
      Setup setup = setupConf.getValue();
      int port = TestUtils.getPortOrDefault(setup.serverPortKey, setup.defaultServerPort);
      for (String propagator : setup.propagators) {
        new HttpInteropTestClient(setupName, port, propagator).run();
      }
    }
  }
}
