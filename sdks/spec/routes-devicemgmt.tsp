import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// Device management (CLI: ios reboot/shutdown/erase/devmode/lang/memlimitoff).
// All under /device/{udid}.
// ===========================================================================

/**
 * Reboot the device (CLI: `ios reboot`).
 */
@route("/reboot")
@post
@summary("Reboot device")
op reboot(...UdidPath): GenericResponse | DeviceErrors;

/**
 * Shut down the device (CLI: `ios shutdown`).
 */
@route("/shutdown")
@post
@summary("Shut down device")
op shutdown(...UdidPath): GenericResponse | DeviceErrors;

/**
 * Erase all content and settings (CLI: `ios erase`). Destructive:
 * requires `confirm=true`.
 */
@route("/erase")
@post
@summary("Erase device")
op erase(
  ...UdidPath,

  /** Must be `true` to proceed with the destructive erase. */
  @query confirm: boolean,
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Get developer mode state (CLI: `ios devmode get`).
 */
@route("/devmode")
@get
@summary("Get developer mode")
op getDevMode(...UdidPath): DevModeState | DeviceErrors;

/**
 * Enable or reveal developer mode (CLI: `ios devmode enable|reveal`).
 */
@route("/devmode")
@post
@summary("Set developer mode")
op setDevMode(
  ...UdidPath,
  @body body: DevModeRequest,
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Get the device language/locale configuration (CLI: `ios lang`).
 */
@route("/lang")
@get
@summary("Get language")
op getLanguage(...UdidPath): LanguageConfiguration | DeviceErrors;

/**
 * Set the device language and/or locale (CLI: `ios lang --setlang --setlocale`).
 * Returns the resulting configuration.
 */
@route("/lang")
@put
@summary("Set language")
op setLanguage(
  ...UdidPath,
  @body body: SetLanguageRequest,
): LanguageConfiguration | BadRequestError | DeviceErrors;

/**
 * Waive the memory limit for a process (CLI: `ios memlimitoff`). The process
 * name may be given via the `process` query param or the JSON body.
 */
@route("/memlimitoff")
@post
@summary("Waive memory limit")
op memLimitOff(
  ...UdidPath,

  /** Process name whose memory limit should be waived. */
  @query process?: string,

  @body body?: MemLimitRequest,
): MemLimitResult | BadRequestError | DeviceErrors;
