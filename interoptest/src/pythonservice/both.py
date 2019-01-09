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

"""
Run both Flask and grpc services, shut down on SIG{INT,TERM,HUP}.
"""

from concurrent import futures

import flaskserver
import grpcserver
import util

if __name__ == "__main__":
    with util.get_signal_exit() as exit_event:
        with futures.ThreadPoolExecutor(max_workers=2) as tpe:
            tpe.submit(flaskserver.main, exit_event=exit_event)
            tpe.submit(grpcserver.main, exit_event=exit_event)
