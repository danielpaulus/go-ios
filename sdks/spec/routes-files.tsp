import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// File transfer & crash reports (CLI: ios file / ios crash).
// File transfer streams to/from the HTTP body against the iOS 17+ file service;
// domain selects the container. All under /device/{udid}.
// ===========================================================================

/** Raw file bytes (`application/octet-stream`). */
model OctetStream is File<"application/octet-stream", bytes>;

/**
 * List a device directory (CLI: `ios file ls`).
 */
@route("/files")
@get
@summary("List files")
op listFiles(
  ...UdidPath,

  /** File service domain: `app`, `app-group`, `crash` or `temp`. */
  @query domain: FileDomain,

  /** Bundle/group id for the `app`/`app-group` domains. */
  @query identifier?: string,

  /** Directory path to list (defaults to `.`). */
  @query path?: string,
): FileListing | BadRequestError | DeviceErrors;

/**
 * Download a file from the device, streamed as the response body
 * (CLI: `ios file pull`).
 */
@route("/files/pull")
@get
@summary("Pull file")
op pullFile(
  ...UdidPath,

  /** File service domain: `app`, `app-group`, `crash` or `temp`. */
  @query domain: FileDomain,

  /** Bundle/group id for the `app`/`app-group` domains. */
  @query identifier?: string,

  /** Remote file path on the device. */
  @query remote: string,
): OctetStream | BadRequestError | DeviceErrors;

/**
 * Upload the request body to a device path (CLI: `ios file push`). A
 * `Content-Length` header is required.
 */
@route("/files/push")
@post
@summary("Push file")
op pushFile(
  ...UdidPath,

  /** File service domain: `app`, `app-group`, `crash` or `temp`. */
  @query domain: FileDomain,

  /** Bundle/group id for the `app`/`app-group` domains. */
  @query identifier?: string,

  /** Destination path on the device. */
  @query remote: string,

  /** Raw file bytes to upload. */
  @body content: bytes,
): FilePushResult | BadRequestError | DeviceErrors;

/**
 * List crash reports (CLI: `ios crash ls`).
 */
@route("/crashes")
@get
@summary("List crash reports")
op listCrashes(
  ...UdidPath,

  /** Optional glob pattern to filter reports. */
  @query pattern?: string,
): CrashListing | DeviceErrors;

/**
 * Delete crash reports (CLI: `ios crash rm`).
 */
@route("/crashes")
@delete
@summary("Delete crash reports")
op removeCrashes(
  ...UdidPath,

  /** Working directory on the device. */
  @query cwd: string,

  /** Glob pattern of reports to delete. */
  @query pattern: string,
): GenericResponse | BadRequestError | DeviceErrors;
