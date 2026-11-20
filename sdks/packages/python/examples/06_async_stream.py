#!/usr/bin/env python3
"""06 — Async streaming.

The same streaming endpoints are available on the **async** client
(:class:`~go_ios_sdk.AsyncIosClient`) as ``async for`` iterators. This example
consumes ``device.sysmontap()`` — a stream of CPU-usage samples
(:class:`~go_ios_sdk.CpuUsageSample`) — bounded to a handful of samples or a few
seconds so it always terminates.

Two things worth noting for real async code:

* The stream is an async context manager, so ``async with`` (or cancelling the
  consuming task) closes the HTTP connection promptly.
* A device can be idle, so we guard the whole read with ``asyncio.wait_for``: if
  no samples arrive within the deadline the example finishes cleanly rather than
  hanging.

Run it::

    export GO_IOS_API_KEY=your-key
    uv run python examples/06_async_stream.py
"""

from __future__ import annotations

import asyncio

from _common import SkipExample, base_url, print_header, require_api_key, target_udid

from go_ios_sdk import AsyncIosClient

_MAX_SAMPLES = 5
_DEADLINE_SECONDS = 6.0


async def _resolve_udid(client: AsyncIosClient) -> str:
    """Async device resolution (mirrors the sync helper in ``_common``)."""
    udid = target_udid()
    if udid is None:
        udids = await client.devices.udids()
        if not udids:
            raise SkipExample("no devices attached (set GO_IOS_UDID or attach a device)")
        udid = udids[0]
        print(f"(no GO_IOS_UDID set; using first device: {udid})")
    return udid


async def _consume(client: AsyncIosClient) -> int:
    udid = await _resolve_udid(client)
    print(f"target udid: {udid}")
    device = client.device(udid)

    count = 0
    # ``async with`` closes the stream when we leave the block (including on break).
    async with device.sysmontap() as stream:
        async for sample in stream:
            count += 1
            print(
                f"  [{count}] total={sample.total_load} "
                f"system={sample.system_load} user={sample.user_load}"
            )
            if count >= _MAX_SAMPLES:
                break
    return count


async def main_async() -> None:
    print_header("06 async_stream")
    api_key = require_api_key()

    async with AsyncIosClient(base_url=base_url(), api_key=api_key) as client:
        print(f"streaming sysmontap (stop after {_MAX_SAMPLES} samples "
              f"or {_DEADLINE_SECONDS:.0f}s)...")
        try:
            count = await asyncio.wait_for(_consume(client), timeout=_DEADLINE_SECONDS)
        except asyncio.TimeoutError:
            # Idle device: we hit the deadline before enough samples arrived. That
            # is a successful, bounded run — not an error.
            print("deadline reached before all samples arrived (device idle?)")
            count = 0
        print(f"done: received {count} sysmontap sample(s)")


def main() -> None:
    asyncio.run(main_async())


if __name__ == "__main__":
    main()
