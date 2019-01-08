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

from contextlib import contextmanager
import logging
import sys
from concurrent import futures

from opencensus.trace.exporters.ocagent import trace_exporter
from opencensus.trace.ext.flask import flask_middleware
from opencensus.trace.propagation import trace_context_http_header_format
import flask
import grpc
import requests

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc
import service
import util


logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler(sys.stdout))

app = flask.Flask(__name__)


def instrument_flask():
    """Enable tracing via flask middleware."""
    oc_trace_config = app.config.get('OPENCENSUS_TRACE', {})
    oc_trace_config.update({
        'EXPORTER': trace_exporter.TraceExporter,
        'PROPAGATOR': trace_context_http_header_format.TraceContextPropagator
    })
    app.config.update(OPENCENSUS_TRACE=oc_trace_config)
    return flask_middleware.FlaskMiddleware(app)


instrument_flask()


@app.route('/', methods=['GET'])
def healthcheck():
    return "OK"


@app.route(service.HTTP_POST_PATH, methods=['POST'])
def test():
    """Handle a test request by calling other test services"""
    request = pb2.TestRequest.FromString(flask.request.get_data())
    logger.debug("Flask service received: %s", request)

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

    return response.SerializeToString()


# http://flask.pocoo.org/snippets/67/
def shutdown_server():
    """Shug down the server."""
    func = flask.request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()


@app.route('/shutdown', methods=['POST'])
def shutdown():
    """Handler to shut down the server."""
    shutdown_server()
    return "Shutting down server"


@app.route('/register', methods=['POST'])
def register(port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    """Register the server at host:port with the test registration service."""
    request = pb2.RegistrationRequest(
        server_name='python',
        services=[
            pb2.Service(
                name='python',
                port=port,
                host=flask.request.environ.get('SERVER_NAME'),
                spec=pb2.Spec(
                    transport=pb2.Spec.HTTP,
                    propagation=pb2.Spec.TRACE_CONTEXT_FORMAT_PROPAGATION),
            )])
    client = pb2_grpc.RegistrationServiceStub(
        channel=grpc.insecure_channel(
            '{}:{}'.format(service.REGISTRATION_SERVER_HOST,
                           service.REGISTRATION_SERVER_PORT))
    )
    try:
        return client.register(request)
    except grpc.RpcError as ex:
        logger.info("Registration server call failed with exception: %s", ex)
        return "Failed to register"


def block_until_ready(host, port, timeout=10):
    """Block until the server responds to a request, up to timeout."""
    def keep_pinging():
        ready = False
        while not ready:
            try:
                requests.get('http://{}:{}{}'.format(host, port, '/'),
                             timeout=.01)
                ready = True
            except requests.ConnectionError:
                pass
    with futures.ThreadPoolExecutor(max_workers=1) as tpe:
        ff = tpe.submit(keep_pinging)
        ff.result(timeout=timeout)


@contextmanager
def serve_http_tracecontext(
        port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    """Run the HTTP/tracecontext server, shut down on exiting context."""
    host = 'localhost'
    with futures.ThreadPoolExecutor(max_workers=1) as tpe:
        tpe.submit(app.run, host=host, port=port)
        block_until_ready(host, port)
        yield app
        requests.post('http://{}:{}{}'.format(host, port, '/shutdown'))
    logger.debug("Shut down flask server")


def test_server(port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    """Send a single multi-hop request to the server and shut it down."""

    test_request = pb2.TestRequest(
        name="python:http:tracecontext",
        service_hops=[
            pb2.ServiceHop(
                service=pb2.Service(
                    name="python:http:tracecontext",
                    port=port,
                    host="localhost",
                    spec=pb2.Spec(
                        transport=pb2.Spec.HTTP,
                        propagation=pb2.Spec.
                        TRACE_CONTEXT_FORMAT_PROPAGATION))),
            pb2.ServiceHop(
                service=pb2.Service(
                    name="python:http:tracecontext",
                    port=port,
                    host="localhost",
                    spec=pb2.Spec(
                        transport=pb2.Spec.HTTP,
                        propagation=pb2.Spec.
                        TRACE_CONTEXT_FORMAT_PROPAGATION)))
        ])

    with serve_http_tracecontext():
        return service.call_http_tracecontext('localhost', port, test_request)


def main(host='localhost', port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT,
         exit_event=None):
    """Runs the service and registers it with the test coordinator."""
    with serve_http_tracecontext():
        logger.debug("Registering with test coordinator")
        requests.post('http://{}:{}{}'.format(host, port, '/register'))
        logger.debug("Serving...")
        if exit_event is not None:
            while not exit_event.is_set():
                exit_event.wait(60)


if __name__ == "__main__":
    with util.get_signal_exit() as exit_event:
        main(exit_event=exit_event)
