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
  run [--host=<host> [--port=<port>]] [--timeout=<timeout>] [--verbose]
  run (-h | --help)

Options:
  -h --help               Show help (this screen)
  -v --verbose            Verbose output
  -t --timeout=<timeout>  Seconds to wait for service calls [default: 60]
  -o --host=<host>        Test service host [default: localhost]
  -p --port=<port>        Test service port [default: 10003]

"""

from concurrent import futures
from threading import Event
import logging
import sys
import time

from docopt import docopt
import grpc

import interoperability_test_pb2 as pb2
import interoperability_test_pb2_grpc as pb2_grpc


logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler(sys.stdout))

# Seconds between calls to get test result
POLL_INTERVAL = 1


def call_test(host, port, executor, timeout=None):
    client = pb2_grpc.InteropTestServiceStub(
        channel=grpc.insecure_channel('{}:{}'.format(host, port)))

    logger.debug("Kicking off test run on %s:%s", host, port)
    run_future = executor.submit(client.run, pb2.InteropRunRequest())
    start_time = time.time()
    run_id = run_future.result(timeout=timeout).id
    logger.debug("Got test run ID %s", run_id)

    timeout_event = Event()

    def poll_server():
        response = client.result(pb2.InteropResultRequest(id=run_id))
        while response.status.status == pb2.RUNNING:
            logger.debug("Waiting for test run %s to finish...", run_id)
            if timeout_event.wait(POLL_INTERVAL):
                raise futures.TimeoutError()
            response = client.result(pb2.InteropResultRequest(id=run_id))
        return response

    result_future = executor.submit(poll_server)
    if timeout is None:
        time_left = None
    else:
        time_left = time.time() - start_time + timeout
    try:
        response = result_future.result(timeout=time_left)
    except futures.TimeoutError:
        timeout_event.set()
        raise
    except KeyboardInterrupt:
        timeout_event.set()
        raise

    logger.debug("Done test run %s", run_id)
    return response


def main():
    args = docopt(__doc__)
    if args.get('--verbose'):
        logger.setLevel(logging.DEBUG)
    host = args.get('--host', 'localhost')
    port = args.get('--port', 10000)
    timeout = int(args.get('--timeout', 60))
    if timeout <= 0:
        timeout = None
    with futures.ThreadPoolExecutor(max_workers=2) as tpe:
        try:
            results = call_test(host, port, tpe, timeout=timeout)
        except futures.TimeoutError as timeout_ex:
            logger.fatal("Timed out waiting for response from test service")
            logger.debug(timeout_ex)
            sys.exit(-2)
        except grpc.RpcError as rpc_ex:
            logger.fatal("RPC error while calling test service")
            logger.debug(rpc_ex)
            sys.exit(-2)
    logger.debug("Results:")
    print(results)
    if results.status.status == pb2.FAILURE:
        sys.exit(-1)


if __name__ == "__main__":
    main()
