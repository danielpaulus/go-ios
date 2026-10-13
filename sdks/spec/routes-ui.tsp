import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";
import "./routes.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 2 — UI automation (device-scoped, under /device/{udid}/ui)
//
// These endpoints proxy to a running, forwarded WebDriverAgent (WDA) or
// DeviceKit backend via ios/uidriver. They do NOT start WDA — the caller must
// first bring up and forward a backend (typically POST /jobs/runwda + POST
// /jobs/forward). Backend addressing is per request via the UIBackendParams
// query params (or same-named headers). Responses from the backend are
// forwarded verbatim, so most bodies are backend-defined (UIResponse).
//
// `501` means the selected backend does not support the operation (e.g. WDA
// foreground, most hardware buttons on WDA); `502` means the backend rejected
// the request or was unreachable.
// ===========================================================================

@route("/device/{udid}/ui")
namespace UI {
  /** Tap at absolute coordinates. */
  @route("/tap")
  @post
  @summary("Tap")
  op uiTap(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UITapRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Drag from (x1,y1) to (x2,y2). */
  @route("/swipe")
  @post
  @summary("Swipe")
  op uiSwipe(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UISwipeRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Press and hold at (x,y). */
  @route("/longpress")
  @post
  @summary("Long press")
  op uiLongPress(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UILongPressRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Send text as keyboard input. */
  @route("/type")
  @post
  @summary("Type text")
  op uiType(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UITypeRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Press a hardware button by name (WDA supports only `home`). */
  @route("/button")
  @post
  @summary("Press hardware button")
  op uiButton(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UIButtonRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Capture the screen and return raw PNG bytes. */
  @route("/screenshot")
  @get
  @summary("UI screenshot (PNG)")
  op uiScreenshot(
    ...UdidPath,
    ...UIBackendParams,
  ): PngImage | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Return the current view hierarchy (XML for WDA; backend Content-Type preserved). */
  @route("/source")
  @get
  @summary("UI source hierarchy")
  op uiSource(
    ...UdidPath,
    ...UIBackendParams,
  ): UISourceXml | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Return the device window/screen size payload (typically {width,height}). */
  @route("/size")
  @get
  @summary("UI window size")
  op uiWindowSize(
    ...UdidPath,
    ...UIBackendParams,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Get the current device orientation payload. */
  @route("/orientation")
  @get
  @summary("Get orientation")
  op uiGetOrientation(
    ...UdidPath,
    ...UIBackendParams,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Set the device orientation. */
  @route("/orientation")
  @put
  @summary("Set orientation")
  op uiSetOrientation(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UIOrientationRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Return the backend status/health payload (WDA /status or DeviceKit /health). */
  @route("/status")
  @get
  @summary("UI backend status")
  op uiStatus(
    ...UdidPath,
    ...UIBackendParams,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Launch the app identified by `bundleId`. */
  @route("/app/launch")
  @post
  @summary("Launch app (UI backend)")
  op uiAppLaunch(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UIAppRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /** Terminate the app identified by `bundleId`. */
  @route("/app/terminate")
  @post
  @summary("Terminate app (UI backend)")
  op uiAppTerminate(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UIAppRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /**
   * Bring the backgrounded app to the foreground. Only the devicekit backend
   * supports this; WDA returns `501`. The request body is ignored.
   */
  @route("/app/foreground")
  @post
  @summary("Foreground app (UI backend)")
  op uiAppForeground(
    ...UdidPath,
    ...UIBackendParams,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;

  /**
   * Raw passthrough to the backend. For WDA supply `method`/`path`/`body`; for
   * DeviceKit supply `rpcMethod`/`rpcParams`. The backend response is forwarded
   * verbatim.
   */
  @route("/api")
  @post
  @summary("Raw backend passthrough")
  op uiApi(
    ...UdidPath,
    ...UIBackendParams,
    @body body: UIAPIRequest,
  ): UIResponse | BadRequestError | BadGatewayError | NotImplementedError | DeviceErrors;
}
