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
const {constants, toBuffer, fromBuffer} = require('./util');
const http = require('http');

const URL_ENDPOINT = '/test/request';
const PROTOBUF_HEADER = {'Content-Type': 'application/x-protobuf'}
let server;

/**
 * Starts a HTTP server that receives requests on sample server port
 */
function start (httpPort, httpHost) {
  const host = httpHost || constants.DEFAULT_HOST;
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
  try {
    const url = request.url;
    if (url === constants.HTTP_URL_ENDPOINT) {
      let body = [];
      request.on('error', err => console.log(err));
      request.on('data', chunk => body.push(chunk));
      request.on('end', () => {
        const testRequest = fromBuffer(body, interop.TestRequest);
        (async () => {
          const testResponse = await serviceHopper.serviceHop(testRequest);
          console.log(`http hopper:${JSON.stringify(testResponse.toObject())}`);
          response.writeHead(200, constants.PROTOBUF_HEADER);
          response.write(toBuffer(testResponse));
          response.end();
        })();
      });
    } else {
      const testResponse = new interop.TestResponse();
      serviceHopper.setFailureStatus(testResponse, 'Bad Request');
      response.writeHead(400, constants.PROTOBUF_HEADER);
      response.write(toBuffer(testResponse));
      response.end();
    }
  } catch (err) {
    console.log(err);
  }
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
