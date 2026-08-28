import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 4 — Codesigning & device preparation
//
// The /sign/* routes and /prepare/create-cert are HOST-scoped (device-free):
// they live directly under /api/v1, not under /device/{udid}. /device/{udid}/prepare
// is device-scoped. Secret material (.p8 keys, P12 identities, passwords) is
// never logged by the server.
// ===========================================================================

/**
 * Create one App Store Connect signing certificate and return its P12
 * (certificate + private key) as a downloadable `application/x-pkcs12` file. The
 * P12 password is echoed in the `X-P12-Password` response header and the
 * certificate resource id in `X-Certificate-Id`. Host-scoped (device-free).
 */
@route("/sign/certificate")
@post
@summary("Create a signing certificate")
op signCertificate(
  @multipartBody body: {
    /** `.p8` App Store Connect API key. */
    `asc-private-key`: HttpPart<File>;

    /** App Store Connect key id. */
    `asc-key-id`: HttpPart<string>;

    /** App Store Connect issuer id. */
    `asc-issuer-id`: HttpPart<string>;

    /** Revoke existing iOS Development certificates first. */
    `revoke-existing`?: HttpPart<string>;

    /** Password to protect the generated P12. */
    p12password?: HttpPart<string>;
  },
): Pkcs12File | BadRequestError | BadGatewayError | UnauthorizedError | InternalError;

/**
 * Create a bundle id, development certificate and provisioning profile via App
 * Store Connect and return both artifacts base64-encoded in a JSON envelope.
 * The target device udid is supplied as a form field. Host-scoped.
 */
@route("/sign/provision")
@post
@summary("Create a provisioning profile + P12")
op signProvision(
  @multipartBody body: {
    /** `.p8` App Store Connect API key. */
    `asc-private-key`: HttpPart<File>;

    /** App Store Connect key id. */
    `asc-key-id`: HttpPart<string>;

    /** App Store Connect issuer id. */
    `asc-issuer-id`: HttpPart<string>;

    /** App bundle identifier. */
    bundleid: HttpPart<string>;

    /** Target device udid to register against the profile. */
    udid: HttpPart<string>;

    /** Bundle display name. */
    bundlename?: HttpPart<string>;

    /** Provisioning profile name. */
    profilename?: HttpPart<string>;

    /** Device display name. */
    devicename?: HttpPart<string>;

    /** Reuse an existing certificate (no new P12 is generated). */
    `certificate-id`?: HttpPart<string>;

    /** Revoke existing certificates first. */
    `revoke-existing`?: HttpPart<string>;

    /** Password to protect the generated P12. */
    p12password?: HttpPart<string>;
  },
): ProvisioningResult | BadRequestError | BadGatewayError | UnauthorizedError | InternalError;

/**
 * Resign an uploaded app/IPA with an uploaded P12 identity and provisioning
 * profile, returning the signed IPA. Synchronous. Host-scoped.
 */
@route("/sign/app")
@post
@summary("Resign an app/IPA")
op signApp(
  @multipartBody body: {
    /** App or `.ipa` to resign. */
    ipa: HttpPart<File>;

    /** Signing identity (`.p12`). */
    p12file: HttpPart<File>;

    /** Provisioning profile (`.mobileprovision`). */
    profile: HttpPart<File>;

    /** P12 password. */
    p12password?: HttpPart<string>;

    /** Override bundle id. */
    bundleid?: HttpPart<string>;
  },
): SignedIpaFile | BadRequestError | UnauthorizedError | InternalError;

/**
 * Generate a self-signed supervision identity (CLI: `ios prepare create-cert`)
 * and return the DER (base64) and PEM for both the certificate and private key.
 * Host-scoped (device-free).
 */
@route("/prepare/create-cert")
@post
@summary("Generate a supervision certificate")
op prepareCreateCert(): SupervisionCert | UnauthorizedError | InternalError;

// ---------------------------------------------------------------------------
// Device-scoped preparation route: /device/{udid}/prepare
// ---------------------------------------------------------------------------

@route("/device/{udid}")
namespace Prepare {
  /**
   * Run the device preparation/provisioning flow (CLI: `ios prepare`). Send
   * multipart/form-data. To supervise the device include a `cert` file
   * (DER/PEM/P12 supervision identity) and optional `p12password`; without a
   * cert the device is prepared without supervision.
   */
  @route("/prepare")
  @post
  @summary("Prepare (and optionally supervise) a device")
  op prepareDevice(
    ...UdidPath,

    @multipartBody body: {
      /** Supervision identity (DER/PEM/P12). Omit to prepare unsupervised. */
      cert?: HttpPart<File>;

      /** P12 password (when `cert` is a P12). */
      p12password?: HttpPart<string>;

      /** Setup panes to skip (see /prepare/skip-options). Repeatable. */
      skip?: HttpPart<string>[];

      /** Supervision organization name. */
      orgname?: HttpPart<string>;

      /** Device locale (default en_US). */
      locale?: HttpPart<string>;

      /** Device language (default en). */
      lang?: HttpPart<string>;
    },
  ): PrepareResult | BadRequestError | DeviceErrors;
}
