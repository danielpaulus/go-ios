import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 1d — WebInspector (device-scoped)
// Mirrors `ios webinspector list|launch|eval`. The interactive commands
// (js-shell, cdp) are intentionally not exposed. Under /device/{udid}/webinspector.
// `424` means Web Inspector / Remote Automation is not enabled on the device.
// ===========================================================================

@route("/device/{udid}/webinspector")
namespace WebInspector {
  /** List inspectable pages reported by the device (CLI: `ios webinspector list`). */
  @route("/pages")
  @get
  @summary("List inspectable pages")
  op webInspectorPages(...UdidPath): {
    @body pages: WebInspectorPage[];
  } | FailedDependencyError | DeviceErrors;

  /**
   * Open a URL in a new inspectable page via a remote automation session
   * (CLI: `ios webinspector launch <url>`). `url` may be a query param or in
   * the body; `bundleId` defaults to Safari.
   */
  @route("/launch")
  @post
  @summary("Open a URL in a new inspectable page")
  op webInspectorLaunch(
    ...UdidPath,

    /** URL to open (alternative to the request body). */
    @query url?: string,

    @body body?: WebInspectorLaunchRequest,
  ): WebInspectorLaunchResult | BadRequestError | FailedDependencyError | DeviceErrors;

  /**
   * Evaluate JavaScript in an inspectable page and return the result
   * (CLI: `ios webinspector eval`). `404` when no matching page exists.
   */
  @route("/eval")
  @post
  @summary("Evaluate JavaScript in a page")
  op webInspectorEval(
    ...UdidPath,
    @body body: WebInspectorEvalRequest,
  ): WebInspectorEvalResult | BadRequestError | NotFoundError | FailedDependencyError | DeviceErrors;
}
