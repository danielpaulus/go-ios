import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// Global HTTP proxy (CLI: ios httpproxy). Setting a proxy is supervised.
// All under /device/{udid}.
// ===========================================================================

/**
 * Configure a global HTTP proxy (CLI: `ios httpproxy`). Supervised: send
 * multipart form-data with `host`, `port`, a `p12` supervisor identity and
 * optional `user`/`pass`/`password` fields.
 */
@route("/httpproxy")
@put
@summary("Set HTTP proxy (supervised)")
op setHttpProxy(
  ...UdidPath,

  @multipartBody body: {
    /** Proxy host. */
    host: HttpPart<string>;

    /** Proxy port. */
    port: HttpPart<string>;

    /** Supervision identity certificate. */
    p12: HttpPart<File>;

    /** Proxy username. */
    user?: HttpPart<string>;

    /** Proxy password. */
    pass?: HttpPart<string>;

    /** Passphrase for the `.p12` identity. */
    password?: HttpPart<string>;
  },
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Clear the global HTTP proxy (CLI: `ios httpproxy remove`).
 */
@route("/httpproxy")
@delete
@summary("Remove HTTP proxy")
op removeHttpProxy(...UdidPath): GenericResponse | DeviceErrors;
