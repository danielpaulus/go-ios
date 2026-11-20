import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 1b — Accessibility & location (device-scoped)
// Mirrors `ios voiceover`, `ios zoomtouch`, `ios ax`, `ios ax audit` and
// `ios setlocationgpx`. All under /device/{udid}.
// ===========================================================================

@route("/device/{udid}")
namespace Accessibility {
  /** Get VoiceOver enabled state (CLI: `ios voiceover get`). */
  @route("/voiceover")
  @get
  @summary("Get VoiceOver state")
  op getVoiceOver(...UdidPath): VoiceOverState | DeviceErrors;

  /**
   * Enable/disable VoiceOver (CLI: `ios voiceover enable|disable`). The desired
   * state comes from the JSON body or the `enabled` query param.
   */
  @route("/voiceover")
  @put
  @summary("Set VoiceOver state")
  op setVoiceOver(
    ...UdidPath,

    /** Desired state (alternative to the request body). */
    @query enabled?: boolean,

    @body body?: AXEnabledRequest,
  ): VoiceOverState | BadRequestError | DeviceErrors;

  /** Get ZoomTouch enabled state (CLI: `ios zoomtouch get`). */
  @route("/zoom")
  @get
  @summary("Get ZoomTouch state")
  op getZoomTouch(...UdidPath): ZoomTouchState | DeviceErrors;

  /**
   * Enable/disable ZoomTouch (CLI: `ios zoomtouch enable|disable`). The desired
   * state comes from the JSON body or the `enabled` query param.
   */
  @route("/zoom")
  @put
  @summary("Set ZoomTouch state")
  op setZoomTouch(
    ...UdidPath,

    /** Desired state (alternative to the request body). */
    @query enabled?: boolean,

    @body body?: AXEnabledRequest,
  ): ZoomTouchState | BadRequestError | DeviceErrors;

  /**
   * Run the accessibility audit against the focused app and return the issues
   * found (CLI: `ios ax audit`). Bounded by `timeout` (seconds, default 60).
   */
  @route("/ax/audit")
  @post
  @summary("Run accessibility audit")
  op runAxAudit(
    ...UdidPath,

    /** Audit timeout in seconds (default 60). */
    @query timeout?: int32,
  ): {
    @body issues: AXAuditIssue[];
  } | BadRequestError | DeviceErrors;

  /**
   * Get a snapshot of the currently focused accessibility element
   * (CLI: `ios ax`).
   */
  @route("/ax")
  @get
  @summary("Get accessibility element snapshot")
  op getAxSnapshot(...UdidPath): AXElement | DeviceErrors;

  /**
   * Simulate live location tracking from an uploaded GPX file
   * (CLI: `ios setlocationgpx`). Send multipart/form-data with a `gpx` file.
   */
  @route("/setlocation/gpx")
  @put
  @summary("Simulate location from a GPX file")
  op setLocationGpx(
    ...UdidPath,

    @multipartBody body: {
      /** The GPX file to replay. */
      gpx: HttpPart<File>;
    },
  ): GenericResponse | BadRequestError | DeviceErrors;
}
