import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Tunnel agent routes (global, NOT device-scoped).
//
// These talk to the running tunnel agent by udid string over its info API, so
// they live at the /api/v1 level rather than under /device/{udid}. `ios tunnel
// start` is intentionally not exposed (it starts a long-running privileged
// daemon — a host process-lifecycle concern, not a REST call).
// ===========================================================================

/**
 * List running device tunnels (CLI: `ios tunnel ls`).
 */
@route("/tunnels")
@get
@summary("List tunnels")
op listTunnels(): {
  @body tunnels: Tunnel[];
} | UnauthorizedError | BadGatewayError;

/**
 * Stop the tunnel for a device (CLI: `ios tunnel stop --udid`).
 */
@route("/tunnels/{udid}")
@delete
@summary("Stop tunnel")
op stopTunnel(...UdidPath): TunnelStopped | UnauthorizedError | BadGatewayError;

/**
 * Restart the tunnel for a device and wait for it to come up
 * (CLI: `ios tunnel refresh`).
 */
@route("/tunnels/{udid}/refresh")
@post
@summary("Refresh tunnel")
op refreshTunnel(...UdidPath): Tunnel | UnauthorizedError | BadGatewayError;

/**
 * Shut down the tunnel agent (CLI: `ios tunnel stopagent`).
 */
@route("/tunnel-agent/shutdown")
@post
@summary("Shut down tunnel agent")
op shutdownTunnelAgent(): AgentShutdown | UnauthorizedError | BadGatewayError;
