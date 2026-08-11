import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// Configuration profiles & developer-image management (CLI: ios profile add/
// remove, ios image list/unmount). GET /profiles and GET/PUT /image are in
// routes.tsp; this adds the remaining verbs. All under /device/{udid}.
// ===========================================================================

/**
 * Install a configuration profile (CLI: `ios profile add`). Send the profile as
 * the raw request body, or as multipart with a `profile` file plus an optional
 * `p12` supervisor identity and `password` for a supervised install.
 */
@route("/profiles")
@post
@summary("Install profile")
op addProfile(
  ...UdidPath,

  @multipartBody body: {
    /** The `.mobileconfig` profile. */
    profile: HttpPart<File>;

    /** Supervision identity for a supervised install. */
    p12?: HttpPart<File>;

    /** Passphrase for the `.p12` identity. */
    password?: HttpPart<string>;
  },
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Remove a configuration profile by identifier (CLI: `ios profile remove`).
 */
@route("/profiles/{name}")
@delete
@summary("Remove profile")
op removeProfile(
  ...UdidPath,

  /** The profile identifier to remove. */
  @path name: string,
): GenericResponse | DeviceErrors;

/**
 * List mounted developer image signatures (CLI: `ios image list`).
 */
@route("/image/list")
@get
@summary("List mounted images")
op listMountedImages(...UdidPath): MountedImages | DeviceErrors;

/**
 * Unmount the developer disk image (CLI: `ios image unmount`).
 */
@route("/image")
@delete
@summary("Unmount developer image")
op unmountImage(...UdidPath): GenericResponse | DeviceErrors;
