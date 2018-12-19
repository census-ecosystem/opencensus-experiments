/**
 * Copyright 2018, OpenCensus Authors
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

const interop = require('../../proto/interoperability_test_pb');
const serviceHopper = require('./service-hopper');
const http = require('http');
const URL_ENDPOINT = '/test/request';

let server;

/**
 * Starts a HTTP server that receives requests on sample server port
 */
function start (httpPort, httpHost) {
  const host = httpHost || 'localhost';
  // Creates a server
  server = http.createServer(handleRequest);

  // Starts the server !
  server.listen(httpPort, err => {
    if (err) {
      throw err;
    }
    console.log(`Node HTTP listening on ${host}:${httpPort}`);
  });
}

// A function which handles requests and send response
function handleRequest (request, response) {
  // TODO(mayurkale) : Add this handler
}


/**
 * Gracefully shuts down the server. The server will stop receiving new calls
 * and any pending calls will complete.
 */
function close () {
  if (server) {
    server.close(() => {
      console.log('HTTP Server closed!');
    });
  }
}

exports.start = start;
exports.close = close;
