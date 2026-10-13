#!/usr/bin/env python3
"""05 — Stream syslog (SSE).

``device.syslog()`` opens a long-lived Server-Sent-Events stream and returns a
generator of typed :class:`~go_ios_sdk.SyslogMessage` events. It is also a context
manager: leaving the ``with`` block closes the underlying HTTP connection
promptly, which is exactly what you want for a bounded tail like this one.

We stop after ~20 events **or** ~5 seconds, whichever comes first, so the example
always terminates on its own.

Run it::

    export GO_IOS_API_KEY=your-key
    uv run python examples/05_stream_syslog.py
"""

from __future__ import annotations

import time

from _common import base_url, print_header, require_api_key, resolve_device

from go_ios_sdk import IosClient

_MAX_EVENTS = 20
_MAX_SECONDS = 5.0


def main() -> None:
    print_header("05 stream_syslog")
    api_key = require_api_key()

    with IosClient(base_url=base_url(), api_key=api_key) as client:
        udid, device = resolve_device(client)
        print(f"target udid: {udid}")
        print(f"streaming syslog (stop after {_MAX_EVENTS} events or {_MAX_SECONDS:.0f}s)...")

        started = time.monotonic()
        count = 0

        # ``with`` guarantees the stream is closed even if we break out early.
        with device.syslog() as stream:
            for event in stream:
                count += 1
                # Trim long lines so the output stays readable.
                message = event.message.strip()
                if len(message) > 120:
                    message = message[:117] + "..."
                print(f"  [{count:2d}] {message}")

                if count >= _MAX_EVENTS or (time.monotonic() - started) >= _MAX_SECONDS:
                    break  # closes the stream on exit

        print(f"done: received {count} syslog event(s)")


if __name__ == "__main__":
    main()
