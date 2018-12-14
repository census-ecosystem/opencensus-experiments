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

from concurrent import futures
from contextlib import contextmanager
import logging
import sys

import interoperability_test_pb2 as pb2
import service

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


class HTTPTraceContextTestServer(BaseHTTPRequestHandler):

    def do_POST(self):
        if self.path != service.HTTP_POST_PATH:
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
            response = service.call_next(request)

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
