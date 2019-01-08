// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// Copyright 2019, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
'use strict';
var grpc = require('grpc');
var proto_interoperability_test_pb = require('../proto/interoperability_test_pb.js');

function serialize_interop_InteropResultRequest (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.InteropResultRequest)) {
    throw new Error('Expected argument of type interop.InteropResultRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_InteropResultRequest (buffer_arg) {
  return proto_interoperability_test_pb.InteropResultRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_InteropResultResponse (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.InteropResultResponse)) {
    throw new Error('Expected argument of type interop.InteropResultResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_InteropResultResponse (buffer_arg) {
  return proto_interoperability_test_pb.InteropResultResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_InteropRunRequest (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.InteropRunRequest)) {
    throw new Error('Expected argument of type interop.InteropRunRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_InteropRunRequest (buffer_arg) {
  return proto_interoperability_test_pb.InteropRunRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_InteropRunResponse (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.InteropRunResponse)) {
    throw new Error('Expected argument of type interop.InteropRunResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_InteropRunResponse (buffer_arg) {
  return proto_interoperability_test_pb.InteropRunResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_RegistrationRequest (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.RegistrationRequest)) {
    throw new Error('Expected argument of type interop.RegistrationRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_RegistrationRequest (buffer_arg) {
  return proto_interoperability_test_pb.RegistrationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_RegistrationResponse (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.RegistrationResponse)) {
    throw new Error('Expected argument of type interop.RegistrationResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_RegistrationResponse (buffer_arg) {
  return proto_interoperability_test_pb.RegistrationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_TestRequest (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.TestRequest)) {
    throw new Error('Expected argument of type interop.TestRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_TestRequest (buffer_arg) {
  return proto_interoperability_test_pb.TestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_interop_TestResponse (arg) {
  if (!(arg instanceof proto_interoperability_test_pb.TestResponse)) {
    throw new Error('Expected argument of type interop.TestResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_interop_TestResponse (buffer_arg) {
  return proto_interoperability_test_pb.TestResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

// ------Test Execution Service (Runs on all Server)---------
var TestExecutionServiceService = exports.TestExecutionServiceService = {
  test: {
    path: '/interop.TestExecutionService/test',
    requestStream: false,
    responseStream: false,
    requestType: proto_interoperability_test_pb.TestRequest,
    responseType: proto_interoperability_test_pb.TestResponse,
    requestSerialize: serialize_interop_TestRequest,
    requestDeserialize: deserialize_interop_TestRequest,
    responseSerialize: serialize_interop_TestResponse,
    responseDeserialize: deserialize_interop_TestResponse
  }
};

exports.TestExecutionServiceClient = grpc.makeGenericClientConstructor(TestExecutionServiceService);
// ------Registration Service (Runs on Test-Coordinator)-------
var RegistrationServiceService = exports.RegistrationServiceService = {
  register: {
    path: '/interop.RegistrationService/register',
    requestStream: false,
    responseStream: false,
    requestType: proto_interoperability_test_pb.RegistrationRequest,
    responseType: proto_interoperability_test_pb.RegistrationResponse,
    requestSerialize: serialize_interop_RegistrationRequest,
    requestDeserialize: deserialize_interop_RegistrationRequest,
    responseSerialize: serialize_interop_RegistrationResponse,
    responseDeserialize: deserialize_interop_RegistrationResponse
  }
};

exports.RegistrationServiceClient = grpc.makeGenericClientConstructor(RegistrationServiceService);
// Interop Test Service
var InteropTestServiceService = exports.InteropTestServiceService = {
  result: {
    path: '/interop.InteropTestService/result',
    requestStream: false,
    responseStream: false,
    requestType: proto_interoperability_test_pb.InteropResultRequest,
    responseType: proto_interoperability_test_pb.InteropResultResponse,
    requestSerialize: serialize_interop_InteropResultRequest,
    requestDeserialize: deserialize_interop_InteropResultRequest,
    responseSerialize: serialize_interop_InteropResultResponse,
    responseDeserialize: deserialize_interop_InteropResultResponse
  },
  // Runs the test asynchronously.
  run: {
    path: '/interop.InteropTestService/run',
    requestStream: false,
    responseStream: false,
    requestType: proto_interoperability_test_pb.InteropRunRequest,
    responseType: proto_interoperability_test_pb.InteropRunResponse,
    requestSerialize: serialize_interop_InteropRunRequest,
    requestDeserialize: deserialize_interop_InteropRunRequest,
    responseSerialize: serialize_interop_InteropRunResponse,
    responseDeserialize: deserialize_interop_InteropRunResponse
  }
};

exports.InteropTestServiceClient = grpc.makeGenericClientConstructor(InteropTestServiceService);
