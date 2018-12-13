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
from opencensus.trace.ext.flask import flask_middleware
import flask
import grpc
import requests

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc
import service
import test_service


logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
logger.addHandler(logging.StreamHandler(sys.stdout))

app = flask.Flask(__name__)

# Enable tracing the requests
flask_middleware.FlaskMiddleware(app)


REGISTRATION_SERVER_HOST = ''
REGISTRATION_SERVER_PORT = ''


@app.route(service.HTTP_POST_PATH, methods=['POST'])
def test():
    request = pb2.TestRequest.FromString(flask.request.get_data())
    logger.debug("Got request: %s", request)

    if not request.service_hops:
        response = pb2.TestResponse(
            id=request.id,
            status=[pb2.CommonResponseStatus(
                status=pb2.SUCCESS,
            )],
        )
    else:
        response = service.call_next(request)
    return response.SerializeToString()


# http://flask.pocoo.org/snippets/67/
def shutdown_server():
    func = flask.request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()


@app.route('/shutdown', methods=['POST'])
def shutdown():
    shutdown_server()
    return "Shutting down server"


@app.route('/register', methods=['POST'])
def register(port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
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
            '{}:{}'.format(REGISTRATION_SERVER_HOST, REGISTRATION_SERVER_PORT))
    )
    try:
        return client.register(request)
    except grpc.RpcError as ex:
        logger.info("Registration server call failed with exception: %s", ex)
        return "Failed to register"


@contextmanager
def serve_http_tracecontext(
        port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    host = 'localhost'
    with futures.ThreadPoolExecutor(max_workers=1) as tpe:
        tpe.submit(app.run, host=host, port=port)
        yield app
        requests.post('http://{}:{}{}'.format(host, port, '/shutdown'))
    logger.debug("Shut down flask server")


def test_server():
    """Send a single multi-hop request to the server and shut it down."""
    with serve_http_tracecontext():
        test_service.test_http_server()


def main(host='localhost', port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    """Runs the service and registers it with the test coordinator."""
    with serve_http_tracecontext():
        logger.debug("Registering with test coordinator")
        requests.post('http://{}:{}{}'.format(host, port, '/register'))
        logger.debug("Serving...")


if __name__ == "__main__":
    main()
