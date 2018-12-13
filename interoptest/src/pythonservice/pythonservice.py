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
from concurrent import futures
from contextlib import contextmanager
from uuid import uuid4
import logging
import sys

import grpc
import requests

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc

try:
    from http.server import BaseHTTPRequestHandler
    from http.server import HTTPServer
    import socketserver
except ImportError:
    from BaseHTTPServer import BaseHTTPRequestHandler
    from BaseHTTPServer import HTTPServer
    import SocketServer as socketserver


logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler(sys.stdout))


GRPC_TPE_WORKERS = 10
HTTP_POST_PATH = "/test/request/"


PORT_MAP = {
    pb2.JAVA_GRPC_BINARY_PROPAGATION_PORT:
    (pb2.Spec.GRPC, pb2.Spec.BINARY_FORMAT_PROPAGATION),
    pb2.JAVA_HTTP_B3_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.B3_FORMAT_PROPAGATION),
    pb2.JAVA_HTTP_TRACECONTEXT_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION),
    pb2.GO_GRPC_BINARY_PROPAGATION_PORT:
    (pb2.Spec.GRPC, pb2.Spec.BINARY_FORMAT_PROPAGATION),
    pb2.GO_HTTP_B3_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.B3_FORMAT_PROPAGATION),
    pb2.GO_HTTP_TRACECONTEXT_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION),
    pb2.NODEJS_GRPC_BINARY_PROPAGATION_PORT:
    (pb2.Spec.GRPC, pb2.Spec.BINARY_FORMAT_PROPAGATION),
    pb2.NODEJS_HTTP_B3_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.B3_FORMAT_PROPAGATION),
    pb2.NODEJS_HTTP_TRACECONTEXT_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION),
    pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT:
    (pb2.Spec.GRPC, pb2.Spec.BINARY_FORMAT_PROPAGATION),
    pb2.PYTHON_HTTP_B3_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.B3_FORMAT_PROPAGATION),
    pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION),
    pb2.CPP_GRPC_BINARY_PROPAGATION_PORT:
    (pb2.Spec.GRPC, pb2.Spec.BINARY_FORMAT_PROPAGATION),
    pb2.CPP_HTTP_B3_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.B3_FORMAT_PROPAGATION),
    pb2.CPP_HTTP_TRACECONTEXT_PROPAGATION_PORT:
    (pb2.Spec.HTTP, pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION),
}


def rand63():
    """Get a random positive 63 bit int"""
    return uuid4().int >> 65


def call_http(host, port, request):
    """Call the HTTP service at host:port with the given request."""

    logger.debug("Called call_http at %s:%s, hops left: %s", host, port,
                 len(request.service_hops))
    data = request.SerializeToString()
    response = requests.post(
        'http://{}:{}{}'.format(host, port, HTTP_POST_PATH),
        data=data
    )
    response.raise_for_status()
    logger.debug("http service responded: %s", response)
    return pb2.TestResponse.FromString(response.content)


call_http_b3 = call_http
call_http_tracecontext = call_http


def call_grpc_binary(host, port, request):
    """Call the GRPC/binary service at host:port with the given request."""

    logger.debug("Called call_grpc_binary at %s:%s, hops left: %s", host, port,
                 len(request.service_hops))
    client = pb2_grpc.TestExecutionServiceStub(
        channel=grpc.insecure_channel('{}:{}'.format(host, port))
    )
    response = client.test(request)
    logger.debug("grpc service responded: %s", response)
    return response


def call_next(request):
    """Call the next service given by request.service_hops."""

    if not request.service_hops:
        raise ValueError()

    new_request = pb2.TestRequest(
        id=request.id,
        name=request.name,
        service_hops=request.service_hops[1:]
    )

    next_hop = request.service_hops[0]
    transport = next_hop.service.spec.transport
    prop = next_hop.service.spec.propagation
    host = next_hop.service.host
    port = next_hop.service.port

    if port not in PORT_MAP:
        raise ValueError()
    if (transport, prop) != PORT_MAP[port]:
        raise ValueError()

    if (transport == pb2.Spec.HTTP
            and prop == pb2.Spec.B3_FORMAT_PROPAGATION):
        return call_http_b3(host, port, new_request)
    if (transport == pb2.Spec.HTTP
            and prop == pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION):
        return call_http_tracecontext(host, port, new_request)
    if (transport == pb2.Spec.GRPC
            and prop == pb2.Spec.BINARY_FORMAT_PROPAGATION):
        return call_grpc_binary(host, port, new_request)
    raise ValueError("No service for transport/propagation")


class HTTPTraceContextTestServer(BaseHTTPRequestHandler):

    def do_POST(self):
        if self.path != HTTP_POST_PATH:
            self.send_error(404)
            return

        length = int(self.headers['Content-Length'])
        message = self.rfile.read(length)
        request = pb2.TestRequest.FromString(message)
        logger.debug("http service recieved: %s", request)

        if not request.service_hops:
            response = pb2.TestResponse(
                id=request.id,
                status=[pb2.CommonResponseStatus(
                    status=pb2.SUCCESS,
                )],
            )
        else:
            response = call_next(request)

        self.send_response(200)
        self.end_headers()
        self.wfile.write(response.SerializeToString())


# Stolen from http.server in python 3.7, here for backwards compatability. See
# https://docs.python.org/3.7/library/http.server.html.
class ThreadingHTTPServer(socketserver.ThreadingMixIn, HTTPServer):
    daemon_threads = True


@contextmanager
def serve_http_tracecontext(
        port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    httpd = ThreadingHTTPServer(('', port), HTTPTraceContextTestServer)
    with futures.ThreadPoolExecutor(max_workers=1) as tpe:
        tpe.submit(httpd.serve_forever)
        yield httpd
        httpd.shutdown()


class GRPCBinaryTestServer(pb2_grpc.TestExecutionServiceServicer):

    def test(self, request, context):
        logger.debug("grpc service recieved: %s", request)
        if not request.service_hops:
            response = pb2.TestResponse(
                id=request.id,
                status=[pb2.CommonResponseStatus(
                    status=pb2.SUCCESS,
                )],
            )
        else:
            response = call_next(request)
        return response


@contextmanager
def serve_grpc_binary(port=pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT):
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=GRPC_TPE_WORKERS))
    pb2_grpc.add_TestExecutionServiceServicer_to_server(
        GRPCBinaryTestServer(),
        server
    )
    server.add_insecure_port('[::]:{}'.format(port))
    try:
        server.start()
        yield server
    finally:
        server.stop(0)


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
        id_ = rand63()

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
    with serve_http_tracecontext():
        return call_http_tracecontext(OWN_HOST, OWN_HTTP_PORT, test_request)


def test_grpc_server():
    test_request = build_request([
        Hop(OWN_HOST, OWN_GRPC_PORT, GRPC, BINARY),
        Hop(OWN_HOST, OWN_GRPC_PORT, GRPC, BINARY),
    ])
    with serve_grpc_binary():
        return call_grpc_binary('localhost', OWN_GRPC_PORT, test_request)


if __name__ == '__main__':
    test_http_server()
    test_grpc_server()
