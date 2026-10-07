import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos;

// ===========================================================================
// Wave 1c — AFC filesystem (fsync) & provisioning
//
// The fsync surface is the classic AFC (com.apple.afc) filesystem, distinct
// from the iOS 17+ fileservice surface exposed by /files. By default it operates
// on the media directory; ?bundleID=<id> scopes operations to that app's data
// container via house_arrest. ?path= is a device-side path; ".." elements are
// rejected server-side (400). All device routes under /device/{udid}.
// ===========================================================================

// A downloadable file pulled over AFC.
model FsyncFile is File<"application/octet-stream", bytes>;

/**
 * Query params shared by the fsync routes. `path` is a device-side path;
 * `bundleID` optionally scopes to an app data container instead of the media
 * directory.
 */
alias FsyncParams = {
  /** App bundle id to scope to its container (else the media dir). */
  @query bundleID?: string;

  /** Device-side path (rejects `..` elements). */
  @query path?: string;
};

@route("/device/{udid}")
namespace Fsync {
  /** List a device directory over AFC (CLI: `ios fsync ls`). */
  @route("/fsync/ls")
  @get
  @summary("List a directory over AFC")
  op fsyncLs(...UdidPath, ...FsyncParams): FsyncListing | BadRequestError | DeviceErrors;

  /** Recursively list a device directory over AFC (CLI: `ios fsync tree`). */
  @route("/fsync/tree")
  @get
  @summary("Recursively list a directory over AFC")
  op fsyncTree(...UdidPath, ...FsyncParams): FsyncTreeListing | BadRequestError | DeviceErrors;

  /**
   * Download a file from the device over AFC (CLI: `ios fsync pull`). Returns
   * the raw file bytes. `path` is required.
   */
  @route("/fsync/pull")
  @get
  @summary("Download a file over AFC")
  op fsyncPull(
    ...UdidPath,

    /** App bundle id to scope to its container (else the media dir). */
    @query bundleID?: string,

    /** Remote file path on the device (required). */
    @query path: string,
  ): FsyncFile | BadRequestError | DeviceErrors;

  /**
   * Upload a file to the device over AFC (CLI: `ios fsync push`). Accepts either
   * raw bytes (application/octet-stream) or a multipart form with a `file`
   * field. `path` is required. Bounded server-side; oversized uploads get `413`.
   */
  @route("/fsync/push")
  @post
  @summary("Upload a file over AFC")
  op fsyncPush(
    ...UdidPath,

    /** App bundle id to scope to its container (else the media dir). */
    @query bundleID?: string,

    /** Destination path on the device (required). */
    @query path: string,

    /** Raw file bytes to upload (application/octet-stream). */
    @body content: bytes,
  ): FsyncPushResult | BadRequestError | PayloadTooLargeError | DeviceErrors;

  /**
   * Remove a file or directory over AFC (CLI: `ios fsync rm`). Pass
   * `recursive=true` to delete a non-empty directory.
   */
  @route("/fsync/rm")
  @delete
  @summary("Remove a file or directory over AFC")
  op fsyncRm(
    ...UdidPath,

    /** App bundle id to scope to its container (else the media dir). */
    @query bundleID?: string,

    /** Path to remove (required). */
    @query path: string,

    /** Remove directory contents recursively. */
    @query recursive?: boolean,
  ): FsyncMessage | BadRequestError | DeviceErrors;

  /** Create a directory over AFC (CLI: `ios fsync mkdir`). */
  @route("/fsync/mkdir")
  @post
  @summary("Create a directory over AFC")
  op fsyncMkdir(
    ...UdidPath,

    /** App bundle id to scope to its container (else the media dir). */
    @query bundleID?: string,

    /** Directory path to create (required). */
    @query path: string,
  ): FsyncMessage | BadRequestError | DeviceErrors;

  /**
   * Get the device cloud configuration (supervision status, skip-setup options,
   * organization info).
   */
  @route("/cloudconfig")
  @get
  @summary("Get device cloud configuration")
  op getCloudConfig(...UdidPath): CloudConfig | DeviceErrors;
}

// ---------------------------------------------------------------------------
// Host-scoped provisioning route (device-free): /api/v1/prepare/skip-options
// ---------------------------------------------------------------------------

/**
 * List all setup-pane skip options usable when preparing a device
 * (CLI: `ios prepare printskip`). Static, device-free list.
 */
@route("/prepare/skip-options")
@get
@summary("List setup skip options")
op getPrepareSkipOptions(): PrepareSkipOptions | UnauthorizedError | InternalError;
