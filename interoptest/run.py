#!/usr/bin/env python
#
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
"""
Run the OpenCensus interop test and print the results.

Usage:
  run [--host=<host> [--port=<port>]] [--verbose]
  run (-h | --help)

Options:
  -h --help         Show help (this screen)
  -v --verbose      Verbose output
  -o --host=<host>  Test service host [default: localhost]
  -p --port=<port>  Test service port [default: 10000]

"""

import logging
import sys
import time

from docopt import docopt
import grpc

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc

logger = logging.getLogger(__name__)
logger.setLevel(logging.WARN)
logger.addHandler(logging.StreamHandler(sys.stdout))

# Seconds between calls to get test result
POLL_INTERVAL = 1


def call_test(host, port):
    client = pb2_grpc.InteropTestServiceStub(
        channel=grpc.insecure_channel('{}:{}'.format(host, port)))
    logger.debug("Kicking off test run on %s:%s", host, port)
    run_id = int(client.run(pb2.InteropRunRequest()).id)
    logger.debug("Got test run ID %s", run_id)
    response = client.result(pb2.InteropResultRequest(id=run_id))
    while response is None or response.status.status == pb2.RUNNING:
        logger.debug("Waiting for test run %s to finish...", run_id)
        time.sleep(POLL_INTERVAL)
        response = client.result(pb2.InteropResultRequest(id=run_id))
    logger.debug("Done test run %s", run_id)
    return response


def main():
    args = docopt(__doc__)
    if args.get('--verbose'):
        logger.setLevel(logging.DEBUG)
    host = args.get('--host', 'localhost')
    port = args.get('--port', 10000)
    results = call_test(host, port)
    logger.debug("Results:")
    print(results)


if __name__ == "__main__":
    main()
