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

import static io.netty.handler.codec.http.HttpResponseStatus.INTERNAL_SERVER_ERROR;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;

import com.google.protobuf.util.JsonFormat;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.buffer.Unpooled;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpRequest;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.codec.http.QueryStringDecoder;
import io.netty.handler.logging.LoggingHandler;
import io.netty.util.CharsetUtil;
import io.opencensus.interop.EchoResponse;
import io.opencensus.interop.util.TestUtils;
import io.opencensus.interop.http.netty.NettyUtils;
import io.opencensus.interop.http.netty.OpenCensusServerInboundOutboundHandler;
import io.opencensus.trace.Span;
import io.opencensus.trace.Tracer;
import io.opencensus.trace.Tracing;
import io.opencensus.trace.propagation.TextFormat;
import java.util.logging.Logger;

/** Server for HTTP interop testing. */
public final class HttpInteropTestServer {
  private static final Logger logger = Logger.getLogger(HttpInteropTestServer.class.getName());
  private static final Tracer tracer = Tracing.getTracer();
  private final int portNumber;

  private HttpInteropTestServer(int port) {
    portNumber = port;
  }

  /** Server main serving thread. */
  private void run() throws Exception {
    EventLoopGroup bossGroup = new NioEventLoopGroup(1);
    EventLoopGroup workerGroup = new NioEventLoopGroup();
    try {
      ServerBootstrap b = new ServerBootstrap();
      b.group(bossGroup, workerGroup)
          .channel(NioServerSocketChannel.class)
          .handler(new LoggingHandler())
          .childHandler(new HttpInteropTestServerInitializer());

      Channel ch = b.bind(portNumber).sync().channel();
      logger.info("Started on http://127.0.0.1:" + portNumber + "/");
      ch.closeFuture().sync();
    } finally {
      bossGroup.shutdownGracefully();
      workerGroup.shutdownGracefully();
    }
  }

  /** Customized {@link ChannelInitializer} for the interop test server. */
  private final class HttpInteropTestServerInitializer extends ChannelInitializer<SocketChannel> {
    @Override
    public void initChannel(SocketChannel ch) {
      ChannelPipeline pipeline = ch.pipeline();
      pipeline.addLast(new HttpServerCodec());
      pipeline.addLast(
          new OpenCensusServerInboundOutboundHandler(tracer) {
            // extract the propagator from query paramater "p"
            @Override
            public TextFormat getTextFormat(ChannelHandlerContext ctx, HttpRequest req) {
              try {
                QueryStringDecoder decoder = new QueryStringDecoder(req.uri());
                String propagator = decoder.parameters().get("p").get(0);
                return TestUtils.getTextFormat(propagator);
              } catch (Throwable error) {
                return null;
              }
            }
          });
      pipeline.addLast(new HttpInteropTestServerInboundHandler());
    }
  }

  /** A handler that serves the content. */
  private static final class HttpInteropTestServerInboundHandler
      extends SimpleChannelInboundHandler<HttpRequest> {
    @Override
    public void channelRead0(ChannelHandlerContext ctx, HttpRequest msg) throws Exception {
      Span span = NettyUtils.getCurrentSpan(ctx.channel());
      EchoResponse echo = TestUtils.buildResponse(span.getContext(), null);
      String result = JsonFormat.printer().print(echo);
      FullHttpResponse response =
          new DefaultFullHttpResponse(
              HTTP_1_1, HttpResponseStatus.OK, Unpooled.copiedBuffer(result, CharsetUtil.UTF_8));
      // Write and flush the response. Close the channel after the action is completed.
      ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
      FullHttpResponse response = new DefaultFullHttpResponse(HTTP_1_1, INTERNAL_SERVER_ERROR);
      ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }
  }

  /** Main launcher of the test server. */
  public static void main(String[] args) throws Exception {
    int port =
        TestUtils.getPortOrDefault(
            HttpInteropTestUtils.ENV_PORT_KEY_JAVA, HttpInteropTestUtils.DEFAULT_PORT_JAVA);
    new HttpInteropTestServer(port).run();
  }
}
