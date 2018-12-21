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
const services = require('../../proto/interoperability_test_grpc_pb');
const grpc = require('grpc');

function serviceHop (request) {
  return new Promise((resolve, reject) => {
    const id = request.getId();
    const name = request.getName();
    const hops = request.getServiceHopsList();

    // Creates a test response
    const response = new interop.TestResponse();
    response.setId(id);
    if (hops.length === 0) {
      resolve(setSuccessStatus(response));
    } else {
      // Extracts data from first service hop.
      const firstHop = hops[0];
      const host = firstHop.getService().getHost() || 'localhost';
      const port = firstHop.getService().getPort();
      const restHops = hops.slice(1);
      const transport = firstHop
        .getService()
        .getSpec()
        .getTransport();
      const propagation = firstHop
        .getService()
        .getSpec()
        .getPropagation();

      // Creates a new request.
      const newRequest = new interop.TestRequest();
      newRequest.setId(id);
      newRequest.setName(name);
      newRequest.setServiceHopsList(restHops);

      const spec = interop.Spec;
      if (
        transport === spec.Transport.GRPC &&
        propagation === spec.Propagation.BINARY_FORMAT_PROPAGATION
      ) {
        (async () => {
          const nextResponse = await grpcServiceHop(host, port, newRequest);
          resolve(combineStatus(setSuccessStatus(response), nextResponse));
        })();
      } else {
        resolve(
          setFailureStatus(
            response,
            `Unsupported propagation: ${propagation} or transport:  ${transport}`
          )
        );
      }
    }
  });
}

// This method sends grpc request to a server specified
// in serviceHop -> request.
function grpcServiceHop (host, port, request) {
  const url = `${host}:${port}`;
  console.log(`NextHop GRPC:${url}->${JSON.stringify(request.toObject())}`);

  return new Promise((resolve, reject) => {
    const client = new services.TestExecutionServiceClient(
      url,
      grpc.credentials.createInsecure()
    );

    return client.test(request, (err, response) => {
      if (err) {
        const response = new interop.TestResponse();
        resolve(setFailureStatus(response, 'GRPC Service Hopper Error'));
      } else {
        resolve(response);
      }
    });
  });
}

function combineStatus (currentResponse, nextResponse) {
  nextResponse.getStatusList().forEach((status) => {
    currentResponse.addStatus(status);
  });
  return currentResponse;
}

function setSuccessStatus (response) {
  const commonResponseStatus = new interop.CommonResponseStatus();
  commonResponseStatus.setStatus(interop.Status.SUCCESS);
  response.addStatus(commonResponseStatus);
  return response;
}

function setFailureStatus (response, msg) {
  const commonResponseStatus = new interop.CommonResponseStatus();
  commonResponseStatus.setStatus(interop.Status.FAILURE);
  commonResponseStatus.setError(msg);
  response.addStatus(commonResponseStatus);
  return response;
}

exports.serviceHop = serviceHop;
