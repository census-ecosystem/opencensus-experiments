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

const interop = require('./proto/interoperability_test_pb');
const grpcServer = require('./src/testservice/grpc-server');
const httpServer = require('./src/testservice/http-server');

function main () {
  // GRPC Server
  const grpcPort = interop.ServicePort.NODEJS_GRPC_BINARY_PROPAGATION_PORT;
  grpcServer.start(grpcPort);

  // HTTP Server
  const httpPort = interop.ServicePort.NODEJS_HTTP_B3_PROPAGATION_PORT;
  httpServer.start(httpPort);

  // setTimeout(() => {
  //   httpServer.close();
  //   grpcServer.close();
  // }, 2000);
}

main();
