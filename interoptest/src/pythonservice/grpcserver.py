#!/usr/bin/env python
# Copyright 2019 Google LLC
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

from concurrent import futures
from contextlib import contextmanager
import logging
import sys

from opencensus.trace.exporters.ocagent import trace_exporter
from opencensus.trace.ext.grpc import server_interceptor
from opencensus.trace.samplers import always_on
import grpc

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc
import service
import util

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler(sys.stdout))

GRPC_TPE_WORKERS = 10
SERVICE_NAME = "interop test python grpc binary service"


class GRPCBinaryTestServer(pb2_grpc.TestExecutionServiceServicer):

    def test(self, request, context):
        """Handle a test request by calling other test services"""
        logger.debug("grpc service recieved: %s", request)
        if not request.service_hops:
            response = pb2.TestResponse(
                id=request.id,
                status=[pb2.CommonResponseStatus(
                    status=pb2.SUCCESS,
                )],
            )
        else:
            status = ([pb2.CommonResponseStatus(status=pb2.SUCCESS)] +
                      list(service.call_next(request).status))
            response = pb2.TestResponse(id=request.id, status=status)
        return response


def register(host='localhost', port=pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT):
    """Register the server at host:port with the test registration service."""
    request = pb2.RegistrationRequest(
        server_name='python',
        services=[
            pb2.Service(
                name='python',
                port=port,
                host=host,
                spec=pb2.Spec(
                    transport=pb2.Spec.GRPC,
                    propagation=pb2.Spec.BINARY_FORMAT_PROPAGATION)
            )])
    client = pb2_grpc.RegistrationServiceStub(
        channel=grpc.insecure_channel(
            '{}:{}'.format(service.REGISTRATION_SERVER_HOST,
                           service.REGISTRATION_SERVER_PORT))
    )
    return client.register(request)


@contextmanager
def serve_grpc_binary(port=pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT):
    """Run the GRPC/binary server, shut down on exiting context."""
    interceptor = server_interceptor.OpenCensusServerInterceptor(
        always_on.AlwaysOnSampler(),
        trace_exporter.TraceExporter(SERVICE_NAME))
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=GRPC_TPE_WORKERS),
        interceptors=(interceptor,)
    )
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


def test_server(port=pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT):
    """Send a single multi-hop request to the server and shut it down."""

    test_request = pb2.TestRequest(
        name="python:grpc:binary",
        service_hops=[
            pb2.ServiceHop(
                service=pb2.Service(
                    name="python:grpc:binary",
                    port=port,
                    host="localhost",
                    spec=pb2.Spec(
                        transport=pb2.Spec.GRPC,
                        propagation=pb2.Spec.BINARY_FORMAT_PROPAGATION))),
            pb2.ServiceHop(
                service=pb2.Service(
                    name="python:grpc:binary",
                    port=port,
                    host="localhost",
                    spec=pb2.Spec(
                        transport=pb2.Spec.GRPC,
                        propagation=pb2.Spec.BINARY_FORMAT_PROPAGATION)))
        ])

    with serve_grpc_binary():
        return service.call_grpc_binary('localhost', port, test_request)


def main(host='localhost', port=pb2.PYTHON_GRPC_BINARY_PROPAGATION_PORT,
         exit_event=None):
    """Runs the service and registers it with the test coordinator."""
    with serve_grpc_binary():
        try:
            logger.debug("Registering with test coordinator")
            register(host, port)
        except grpc.RpcError as ex:
            logger.info(
                "Registration server call failed with exception: %s", ex)
        logger.debug("Serving...")
        if exit_event is not None:
            while not exit_event.is_set():
                exit_event.wait(60)


if __name__ == "__main__":
    with util.get_signal_exit() as _exit_event:
        main(exit_event=_exit_event)
