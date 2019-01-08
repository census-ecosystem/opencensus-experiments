/**
 * Copyright 2019, OpenCensus Authors
 *
 * Licensed under the Apache License, Version 2.0 (the 'License');
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an 'AS IS' BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const services = require('../../proto/interoperability_test_grpc_pb');
const serviceHopper = require('./service-hopper');
const grpc = require('grpc');
const {constants} = require('./util');

let server;
/**
 * Starts a RPC server that receives requests for the TestExecutionService
 * service at the sample server port
 */
function start (grpcPort, grpcHost) {
  const host = grpcHost || constants.DEFAULT_HOST;
  // Creates a server
  server = new grpc.Server();

  server.addService(services.TestExecutionServiceService, { test: test });

  const credentials = grpc.ServerCredentials.createInsecure();
  const boundPort = server.bind(`${host}:${grpcPort}`, credentials);

  if (boundPort !== grpcPort) {
    console.warn('Failed to bind to port', grpcPort);
  }

  // Starts the server !
  server.start();
  console.log(`Node GRPC listening on ${host}:${grpcPort}`);
}

/**
 * Gracefully shuts down the server. The server will stop receiving new calls
 * and any pending calls will complete.
 */
function close () {
  if (server) {
    server.tryShutdown(() => {
      console.log('GRPC Server shutdown completed!');
    });
  }
}

/**
 * Implements the test RPC method.
 */
function test (call, callback) {
  (async () => {
    const testResponse = await serviceHopper.serviceHop(call.request);
    callback(null, testResponse);
  })();
}

exports.start = start;
exports.close = close;
