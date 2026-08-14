import "@typespec/http";
import "@typespec/openapi";
import "@typespec/sse";

import "./models.tsp";
import "./streaming.tsp";

using Http;
using OpenAPI;
using SSE;

namespace GoIos.Devices;

// ===========================================================================
// Streaming monitoring (CLI: ios sysmontap). Under /device/{udid}.
//
// `ios pcap` is intentionally not exposed yet — the pcap package writes to a
// local file and blocks with no streaming API to hook a response writer to.
// ===========================================================================

/**
 * Stream CPU-usage samples as Server-Sent Events (CLI: `ios sysmontap`).
 */
@route("/sysmontap")
@get
@summary("Stream CPU usage (SSE)")
@extension(
  "x-sse-events",
  #{
    schema: "SysmontapEvents",
    events: #{ sample: "CpuUsageSample", heartbeat: "Heartbeat" },
  }
)
op streamSysmontap(...UdidPath): SSEStream<SysmontapEvents> | DeviceErrors;
