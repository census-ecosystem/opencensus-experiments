#!/usr/bin/env python

# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# pylint: disable=no-member
# pylint: disable=invalid-name
# pylint: disable=missing-docstring
# pylint: disable=too-few-public-methods

from collections import namedtuple

import interoperability_test_pb2 as pb2
import flaskserver
import grpcserver
import service

TRANSPORT_TO_TRANSPORT_NAME = {
    pb2.Spec.HTTP: 'http',
    pb2.Spec.GRPC: 'grpc',
}

PROP_TO_PROP_NAME = {
    pb2.Spec.B3_FORMAT_PROPAGATION: 'b3',
    pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION: 'tracecontext',
    pb2.Spec.BINARY_FORMAT_PROPAGATION: 'binary',
}

PORT_TO_LANG_NAME = {
    pb2.JAVA_GRPC_BINARY_PROPAGATION_PORT: 'java',
    pb2.JAVA_HTTP_B3_PROPAGATION_PORT: 'java',
    pb2.JAVA_HTTP_TRACECONTEXT_PROPAGATION_PORT: 'java',
    pb2.GO_GRPC_BINARY_PROPAGATION_PORT: 'go',
    pb2.GO_HTTP_B3_PROPAGATION_PORT: 'go',
    pb2.GO_HTTP_TRACECONTEXT_PROPAGATION_PORT: 'go',
    pb2.NODEJS_GRPC_BINARY_PROPAGATION_PORT: 'nodejs',
    pb2.NODEJS_HTTP_B3_PROPAGATION_PORT: 'nodejs',
    pb2.NODEJS_HTTP_TRACECONTEXT_PROPAGATION_PORT: 'nodejs',
    pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT: 'python',
    pb2.PYTHON_HTTP_B3_PROPAGATION_PORT: 'python',
    pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT: 'python',
    pb2.CPP_GRPC_BINARY_PROPAGATION_PORT: 'cpp',
    pb2.CPP_HTTP_B3_PROPAGATION_PORT: 'cpp',
    pb2.CPP_HTTP_TRACECONTEXT_PROPAGATION_PORT: 'cpp',
}

Hop = namedtuple('Hop', ('host', 'port', 'transport', 'prop'))


def get_name(hop):
    return "{}:{}:{}".format(
        PORT_TO_LANG_NAME.get(hop.port, 'unknown'),
        TRANSPORT_TO_TRANSPORT_NAME.get(hop.transport, 'unknown'),
        PROP_TO_PROP_NAME.get(hop.prop, 'unknown'),
    )


def build_request(hops, id_=None):
    if id_ is None:
        id_ = service.rand63()

    def build_service_hop(hop):
        return pb2.ServiceHop(
            service=pb2.Service(
                name=get_name(hop),
                port=hop.port,
                host=hop.host,
                spec=pb2.Spec(
                    transport=hop.transport,
                    propagation=hop.prop,
                )
            )
        )

    return pb2.TestRequest(
        id=id_,
        name=get_name(hops[0]),
        service_hops=[build_service_hop(hop) for hop in hops]
    )


HTTP = pb2.Spec.HTTP
GRPC = pb2.Spec.GRPC
B3 = pb2.Spec.B3_FORMAT_PROPAGATION
TRACECONTEXT = pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION
BINARY = pb2.Spec.BINARY_FORMAT_PROPAGATION

OWN_HOST = 'localhost'
OWN_HTTP_PORT = pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT
OWN_GRPC_PORT = pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT


def test_http_server():
    test_request = build_request([
        Hop(OWN_HOST, OWN_HTTP_PORT, HTTP, TRACECONTEXT),
        Hop(OWN_HOST, OWN_HTTP_PORT, HTTP, TRACECONTEXT),
    ])
    with flaskserver.serve_http_tracecontext():
        return service.call_http_tracecontext(
            OWN_HOST, OWN_HTTP_PORT, test_request)


def test_grpc_server():
    test_request = build_request([
        Hop(OWN_HOST, OWN_GRPC_PORT, GRPC, BINARY),
        Hop(OWN_HOST, OWN_GRPC_PORT, GRPC, BINARY),
    ])
    with grpcserver.serve_grpc_binary():
        return service.call_grpc_binary(
            'localhost', OWN_GRPC_PORT, test_request)


if __name__ == '__main__':
    test_http_server()
    test_grpc_server()
