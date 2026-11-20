import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// MDM / supervision (CLI: ios mdm ...). Every endpoint needs a supervisor
// identity: send multipart/form-data with a `p12` file and optional `password`.
// All under /device/{udid}/mdm.
// ===========================================================================

/** Multipart fields carrying the supervisor identity, shared by MDM ops. */
alias SupervisorIdentity = {
  /** The supervision identity certificate. */
  p12: HttpPart<File>;

  /** Passphrase for the `.p12` identity. */
  password?: HttpPart<string>;
};

/**
 * Get device security info (CLI: `ios mdm security-info`).
 */
@route("/mdm/security-info")
@post
@summary("Get MDM security info (supervised)")
op mdmSecurityInfo(
  ...UdidPath,
  @multipartBody body: SupervisorIdentity,
): SecurityInfo | BadRequestError | DeviceErrors;

/**
 * Fetch the escrow unlock token, base64-encoded (CLI:
 * `ios mdm fetch-unlock-token`).
 */
@route("/mdm/fetch-unlock-token")
@post
@summary("Fetch unlock token (supervised)")
op mdmFetchUnlockToken(
  ...UdidPath,
  @multipartBody body: SupervisorIdentity,
): UnlockToken | BadRequestError | DeviceErrors;

/**
 * Clear the device passcode (CLI: `ios mdm clear-passcode`). Requires the
 * base64 unlock token as an additional `token` form field.
 */
@route("/mdm/clear-passcode")
@post
@summary("Clear passcode (supervised)")
op mdmClearPasscode(
  ...UdidPath,

  @multipartBody body: {
    ...SupervisorIdentity;

    /** Base64-encoded escrow unlock token. */
    token: HttpPart<string>;
  },
): StatusOk | BadRequestError | DeviceErrors;

/**
 * Clear the Screen Time password (CLI: `ios mdm clear-screen-time-password`).
 */
@route("/mdm/clear-screen-time-password")
@post
@summary("Clear Screen Time password (supervised)")
op mdmClearScreenTimePassword(
  ...UdidPath,
  @multipartBody body: SupervisorIdentity,
): StatusOk | BadRequestError | DeviceErrors;
