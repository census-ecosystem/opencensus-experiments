/*
 * Copyright 2019, Google LLC.
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

package io.opencensus.interop;

import io.opencensus.contrib.grpc.metrics.RpcViews;
import io.opencensus.exporter.trace.ocagent.OcAgentTraceExporter;
import io.opencensus.contrib.http.util.HttpViews;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Level;
import org.apache.log4j.Logger;

public final class JavaService {
  public static void main(String[] args) throws Exception {
    BasicConfigurator.configure();
    Logger.getRootLogger().setLevel(Level.INFO);
    HttpViews.registerAllServerViews();
    HttpViews.registerAllClientViews();
    RpcViews.registerAllViews();
    // Initialize OpenCensus agent trace exporter.
    OcAgentTraceExporter.createAndRegister();

    GrpcServer.createWithBinaryFormatPropagation(
        ServicePort.JAVA_GRPC_BINARY_PROPAGATION_PORT.getNumber()).start();
    HttpServer.createWithB3FormatPropagation(
        ServicePort.JAVA_HTTP_B3_PROPAGATION_PORT.getNumber()).start();
    HttpServer.createWithTraceContextFormatPropagation(
        ServicePort.JAVA_HTTP_TRACECONTEXT_PROPAGATION_PORT.getNumber()).start();
  }
}
