import "@typespec/http";
import "@typespec/openapi";

import "./models.tsp";

using Http;
using OpenAPI;

namespace GoIos.Devices;

// ===========================================================================
// Wallpaper, icon layout and pasteboard (CLI: ios get-wallpaper/set-wallpaper/
// get-icon-layout/set-icon-layout/pasteboard). All under /device/{udid}.
// ===========================================================================

/**
 * Get the home-screen wallpaper as PNG (CLI: `ios get-wallpaper`).
 */
@route("/wallpaper")
@get
@summary("Get wallpaper")
op getWallpaper(...UdidPath): PngImage | DeviceErrors;

/**
 * Set the wallpaper (CLI: `ios set-wallpaper`). Supervised: upload the image
 * and a `.p12` supervisor identity as multipart form-data.
 */
@route("/wallpaper")
@put
@summary("Set wallpaper (supervised)")
op setWallpaper(
  ...UdidPath,

  @multipartBody body: {
    /** The wallpaper image. */
    image: HttpPart<File>;

    /** The supervision identity certificate. */
    p12: HttpPart<File>;

    /** Passphrase for the `.p12` identity. */
    password?: HttpPart<string>;

    /** Target screen (`home`, `lock`, `both`). */
    screen?: HttpPart<string>;
  },
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Get the SpringBoard icon layout (CLI: `ios get-icon-layout`).
 */
@route("/icon-layout")
@get
@summary("Get icon layout")
op getIconLayout(...UdidPath): IconLayout | DeviceErrors;

/**
 * Restore a SpringBoard icon layout (CLI: `ios set-icon-layout`). Body is the
 * layout JSON as returned by GET.
 */
@route("/icon-layout")
@put
@summary("Set icon layout")
op setIconLayout(
  ...UdidPath,
  @body body: IconLayout,
): GenericResponse | BadRequestError | DeviceErrors;

/**
 * Get the pasteboard (clipboard) text (CLI: `ios pasteboard get`).
 */
@route("/pasteboard")
@get
@summary("Get pasteboard")
op getPasteboard(...UdidPath): PasteboardContent | DeviceErrors;

/**
 * Set the pasteboard text from the raw request body (CLI: `ios pasteboard set`).
 */
@route("/pasteboard")
@put
@summary("Set pasteboard")
op setPasteboard(
  ...UdidPath,

  /** The clipboard text to set. */
  @header contentType: "text/plain",
  @body text: string,
): GenericResponse | BadRequestError | DeviceErrors;
