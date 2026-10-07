import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";
import "./routes.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 3 — BINARY streaming endpoints (device-scoped)
//
// These are long-lived BINARY streams, NOT Server-Sent Events. They are
// deliberately distinct from the text/event-stream SSE endpoints in
// streaming.tsp: there are no typed `event:` frames, no `@events`, and no
// `x-sse-events` extension. Each streams raw bytes until the client disconnects
// or the source ends:
//
//   - GET /ui/stream          UI video: MJPEG (multipart/x-mixed-replace) by
//                             default, or an H.264 elementary stream with
//                             ?codec=h264 (devicekit backend). Proxied from a
//                             forwarded WDA/DeviceKit backend.
//   - GET /screenshot/stream  MJPEG (multipart/x-mixed-replace) screenshot
//                             stream from the instruments screenshot service.
//   - GET /pcap               live libpcap capture (application/vnd.tcpdump.pcap).
// ===========================================================================

@route("/device/{udid}")
namespace Streams {
  /**
   * Open a live UI video stream against a forwarded WDA/DeviceKit backend and
   * pipe it straight through to the client. Default codec is MJPEG
   * (multipart/x-mixed-replace); `codec=h264` returns an H.264 elementary
   * stream (requires the devicekit backend). Streams until the client
   * disconnects or the backend ends.
   *
   * Requires a running, forwarded WDA/DeviceKit backend (see the UI routes).
   */
  @route("/ui/stream")
  @get
  @summary("Stream UI video (binary)")
  op uiStream(
    ...UdidPath,
    ...UIBackendParams,

    /** Video codec: `mjpeg` (default) or `h264` (devicekit backend only). */
    @query codec?: string,

    /** Target frames per second (backend-dependent). */
    @query fps?: string,

    /** JPEG quality for the mjpeg codec. */
    @query quality?: string,

    /** Scale factor (backend-dependent). */
    @query scale?: string,

    /** Target bitrate for the h264 codec. */
    @query bitrate?: string,
  ): UIVideoStream | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /**
   * Serve an MJPEG (multipart/x-mixed-replace) stream of device screenshots
   * captured via the instruments screenshot service. Streams until the client
   * disconnects or the source fails.
   */
  @route("/screenshot/stream")
  @get
  @summary("Stream screenshots as MJPEG (binary)")
  op screenshotStream(
    ...UdidPath,

    /** Optional JPEG quality (1–100, default 80). */
    @query quality?: int32,
  ): MjpegStream | BadRequestError | DeviceErrors;

  /**
   * Stream a live packet capture from the device as a libpcap byte stream
   * (pipeable into wireshark/tshark). Runs until `timeout` (seconds) elapses,
   * the default timeout is reached, or the client disconnects.
   */
  @route("/pcap")
  @get
  @summary("Stream a live pcap capture (binary)")
  op pcap(
    ...UdidPath,

    /** Capture duration in seconds (default 60, max 3600). */
    @query timeout?: int32,
  ): PcapStream | BadRequestError | DeviceErrors;
}
