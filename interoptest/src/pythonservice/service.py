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

from uuid import uuid4
import logging
import sys

from opencensus.trace.tracers import context_tracer
from opencensus.trace.ext import requests as requests_wrapper
import grpc
import requests

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc


# Monkey patch the requests library
requests_wrapper.trace.trace_integration(context_tracer.ContextTracer())

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler(sys.stdout))

REGISTRATION_SERVER_HOST = ''
REGISTRATION_SERVER_PORT = ''

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
