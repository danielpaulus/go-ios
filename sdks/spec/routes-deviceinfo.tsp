import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// Read-only device information (CLI: ios devicename/date/batterycheck/
// diagnostics/mobilegestalt/ps/lockdown). All under /device/{udid}.
// ===========================================================================

/**
 * Get the device name (CLI: `ios devicename`).
 */
@route("/devicename")
@get
@summary("Get device name")
op getDeviceName(...UdidPath): DeviceName | DeviceErrors;

/**
 * Get the device clock (CLI: `ios date`).
 */
@route("/date")
@get
@summary("Get device date")
op getDeviceDate(...UdidPath): DeviceDate | DeviceErrors;

/**
 * Get battery diagnostics (CLI: `ios batterycheck`).
 */
@route("/battery")
@get
@summary("Get battery info")
op getBattery(...UdidPath): BatteryInfo | DeviceErrors;

/**
 * List all IORegistry/diagnostic values (CLI: `ios diagnostics list`).
 */
@route("/diagnostics")
@get
@summary("List diagnostics")
op getDiagnostics(...UdidPath): Diagnostics | DeviceErrors;

/**
 * Query one or more MobileGestalt keys (CLI: `ios mobilegestalt <key>...`).
 * Pass repeated `key` query params.
 */
@route("/mobilegestalt")
@get
@summary("Query MobileGestalt")
op getMobileGestalt(
  ...UdidPath,

  /** One or more MobileGestalt keys to query. */
  @query key: string[],
): MobileGestalt | BadRequestError | DeviceErrors;

/**
 * List running processes (CLI: `ios ps [--apps]`).
 */
@route("/processes")
@get
@summary("List processes")
op getProcesses(
  ...UdidPath,

  /** Only return application processes. */
  @query apps?: boolean,
): {
  @body processes: ProcessInfo[];
} | DeviceErrors;

/**
 * Get all lockdown values (CLI: `ios lockdown get`).
 */
@route("/lockdown")
@get
@summary("Get lockdown values")
op getLockdownValues(...UdidPath): LockdownValues | DeviceErrors;
