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
import requests

import interoperability_test_pb2 as pb2
import service
import test_service


logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
logger.addHandler(logging.StreamHandler(sys.stdout))

app = flask.Flask(__name__)

# Enable tracing the requests
flask_middleware.FlaskMiddleware(app)


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


@contextmanager
def serve_http_tracecontext(
        port=pb2.PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT):
    host = 'localhost'
    with futures.ThreadPoolExecutor(max_workers=1) as tpe:
        tpe.submit(app.run, host=host, port=port)
        yield app
        requests.post('http://{}:{}{}'.format(host, port, '/shutdown'))
    logger.debug("Shut down flask server")


if __name__ == '__main__':
    with serve_http_tracecontext():
        test_service.test_http_server()
