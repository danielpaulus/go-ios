import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// Device settings (CLI: ios assistivetouch/timeformat/wifi).
// All under /device/{udid}.
// ===========================================================================

/**
 * Get AssistiveTouch state (CLI: `ios assistivetouch get`).
 */
@route("/assistivetouch")
@get
@summary("Get AssistiveTouch")
op getAssistiveTouch(...UdidPath): AssistiveTouchState | DeviceErrors;

/**
 * Enable/disable AssistiveTouch (CLI: `ios assistivetouch enable|disable`).
 */
@route("/assistivetouch")
@put
@summary("Set AssistiveTouch")
op setAssistiveTouch(
  ...UdidPath,
  @body body: EnabledRequest,
): AssistiveTouchState | BadRequestError | DeviceErrors;

/**
 * Get the 24-hour clock state (CLI: `ios timeformat get`).
 */
@route("/timeformat")
@get
@summary("Get time format")
op getTimeFormat(...UdidPath): TimeFormatState | DeviceErrors;

/**
 * Set 24-hour / 12-hour clock (CLI: `ios timeformat 24h|12h`).
 */
@route("/timeformat")
@put
@summary("Set time format")
op setTimeFormat(
  ...UdidPath,
  @body body: TimeFormatRequest,
): TimeFormatState | BadRequestError | DeviceErrors;

/**
 * Provision a wifi network (CLI: `ios wifi`).
 */
@route("/wifi")
@put
@summary("Provision wifi")
op setWifi(
  ...UdidPath,
  @body body: WifiRequest,
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Remove a provisioned wifi network (CLI: `ios wifi --remove`).
 */
@route("/wifi")
@delete
@summary("Remove wifi")
op removeWifi(
  ...UdidPath,

  /** SSID of the network to remove. */
  @query ssid: string,
): GenericResponse | BadRequestError | DeviceErrors;
