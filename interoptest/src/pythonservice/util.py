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

from contextlib import contextmanager
import signal
import threading

try:
    from contextlib import ExitStack
except ImportError:
    from contextlib2 import ExitStack


SIGETC = (signal.SIGINT, signal.SIGTERM, signal.SIGHUP)


@contextmanager
def set_signals(sigs, func):
    """Temporarily set func as the action for multiple signals."""

    with ExitStack() as stack:
        for sig in sigs:
            stack.enter_context(set_signal(sig, func))
        yield stack


@contextmanager
def set_signal(sig, func):
    """Temporarily change a signal's action to the given func."""
    old_action = signal.signal(sig, func)
    try:
        yield
    finally:
        if old_action is not None:
            signal.signal(sig, old_action)


@contextmanager
def get_signal_exit():
    """Get an event that's set on SIG{INT,TERM,HUP}.

    Restore the old signal action on exiting the context.
    """
    event = threading.Event()

    def set_exit(signal_num, stack_frame):
        event.set()

    with set_signals(SIGETC, set_exit):
        yield event
