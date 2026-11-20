import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 1a — Diagnostics & network (device-scoped)
// Mirrors the `ios diskspace`, `ios ip`, `ios rsd ls` and
// `ios diagnostics ioregistry` CLI commands. All under /device/{udid}.
// ===========================================================================

@route("/device/{udid}")
namespace DiagnosticsNet {
  /**
   * Get filesystem info for the device (total/free/used bytes, block size)
   * via AFC (CLI: `ios diskspace`).
   */
  @route("/diskspace")
  @get
  @summary("Get disk space info")
  op getDiskSpace(...UdidPath): DiskSpaceInfo | DeviceErrors;

  /**
   * Resolve the device's network addresses (MAC/IPv4/IPv6) by sniffing pcapd
   * (CLI: `ios ip`).
   */
  @route("/ip")
  @get
  @summary("Get device IP / network info")
  op getDeviceIp(...UdidPath): NetworkInfo | DeviceErrors;

  /**
   * Get the device's RSD (Remote Service Discovery) service list
   * (CLI: `ios rsd ls`). Requires a running tunnel (iOS 17+); devices without
   * RSD return `400`.
   */
  @route("/rsd")
  @get
  @summary("Get RSD service list")
  op getRsdServices(...UdidPath): RsdServices | BadRequestError | DeviceErrors;

  /**
   * Get the battery IORegistry stats (Temperature, Voltage, CurrentCapacity,
   * ...) via the diagnostics relay (CLI: `ios diagnostics ioregistry`).
   */
  @route("/battery/registry")
  @get
  @summary("Get battery IORegistry")
  op getBatteryRegistry(...UdidPath): BatteryRegistry | DeviceErrors;
}
