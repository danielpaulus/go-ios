import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";
import "@typespec/sse";
import "@typespec/events";

import "./models.tsp";
import "./streaming.tsp";
import "./routes.tsp";
import "./routes-tunnel.tsp";
import "./routes-deviceinfo.tsp";
import "./routes-devicemgmt.tsp";
import "./routes-files.tsp";
import "./routes-media.tsp";
import "./routes-config.tsp";
import "./routes-settings.tsp";
import "./routes-monitoring.tsp";
import "./routes-mdm.tsp";
import "./routes-jobs.tsp";
import "./routes-proxy.tsp";

using Http;
using OpenAPI;

/**
 * go-ios REST API.
 *
 * This is the *ideal* contract for the go-ios REST server. It is authored
 * spec-first in TypeSpec and emitted to OpenAPI 3.1; the go-ios server conforms
 * to this document (there is no backward-compatibility constraint yet).
 *
 * ## Authentication
 *
 * Every route under `/api/v1` requires a bearer token:
 * `Authorization: Bearer <GO_IOS_API_KEY>`. The server refuses to start unless
 * either an API key is configured or it is launched with `--disable-auth`.
 *
 * When the server is started with `--disable-auth`, authentication is **not**
 * enforced and the `Authorization` header may be omitted. The Swagger UI
 * (`/swagger/*`) is always unauthenticated and lives outside `/api/v1`, so it is
 * not modeled here.
 *
 * ## Device routing
 *
 * Device-scoped routes live under `/device/{udid}`. A middleware resolves the
 * udid: an unknown udid yields `404`, an empty udid yields `422`. The
 * `/device/{udid}/apps` subgroup is serialized per-udid (one concurrent request
 * per device).
 *
 * ## Streaming
 *
 * Long-lived endpoints (`/notifications`, `/ostrace`, `/syslog`, `/listen`,
 * `/sysmontap` and the async `/jobs/{id}/logs`) are modeled as real Server-Sent
 * Events (`text/event-stream`) with typed event payloads. See streaming.tsp.
 */
@service(#{ title: "go-ios REST API" })
@info(#{
  version: "0.1.0",
  license: #{ name: "MIT", url: "https://opensource.org/license/mit" },
})
@server("http://localhost:60105", "Default go-ios REST server")
@useAuth(BearerAuth)
@route("/api/v1")
namespace GoIos;
