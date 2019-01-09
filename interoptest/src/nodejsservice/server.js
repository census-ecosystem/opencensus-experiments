// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

const path = require('path');
const interop = require('./proto/interoperability_test_pb');
const http = require("http");
const httpServer = require('./src/testservice/http-server');
const httpPlugin = require('@opencensus/instrumentation-http');
const grpc = require('grpc');
const grpcServer = require('./src/testservice/grpc-server');
const grpcPlugin = require('@opencensus/instrumentation-grpc');
const tracing = require('@opencensus/nodejs');
const {logger, CoreTracer} = require('@opencensus/core');
const jaeger = require('@opencensus/exporter-jaeger');
const propagation = require('@opencensus/propagation-tracecontext');

function main () {
  // Setup Tracer
  const tracer = tracing.start({
    samplingRate: 1,
    propagation: new propagation.TraceContextFormat()
  }).tracer;

  // Setup Exporter, Enable GRPC and HTTP plugin
  enableJaegerTraceExporter(tracer);
  enableHttpPlugin(tracer);
  //enableGrpcPlugin(tracer);

  // Start GRPC Server
  grpcServer.start(interop.ServicePort.NODEJS_GRPC_BINARY_PROPAGATION_PORT, '0.0.0.0');

  // Start HTTP Server
  httpServer.start(interop.ServicePort.NODEJS_HTTP_TRACECONTEXT_PROPAGATION_PORT, '0.0.0.0');
}

function enableGrpcPlugin (tracer) {
  // 1. Define basedir and version
  const basedir = path.dirname(require.resolve('grpc'));
  const version = require(path.join(basedir, 'package.json')).version;

  // 2. Enable GRPC plugin: Method that enables the instrumentation patch.
  grpcPlugin.plugin.enable(grpc, tracer, version, basedir);
}

function enableHttpPlugin (tracer) {
  // 1. Define node version
  const version = process.versions.node;

  // 2. Enable HTTP plugin
  httpPlugin.plugin.enable(http, tracer, version, null);
}

function enableJaegerTraceExporter (tracer) {
  // 1. Define service name and jaeger options
  const service = 'nodejsservice';
  const jaegerOptions = {
    serviceName: service,
    host: 'localhost',
    port: 6832,
    bufferTimeout: 10
  };

  // 2. Configure exporter to export traces to Jaeger.
  const exporter = new jaeger.JaegerTraceExporter(jaegerOptions);
  tracer.registerSpanEventListener(exporter);
}

main();
