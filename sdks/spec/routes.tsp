import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";
import "@typespec/sse";

import "./models.tsp";
import "./streaming.tsp";

using Http;
using OpenAPI;
using SSE;

namespace GoIos;

// ===========================================================================
// Reusable path parameter models
// ===========================================================================

/** The device udid (serial number). Resolved by the device middleware. */
alias UdidPath = {
  @path udid: string;
};

// A screenshot returned as raw PNG bytes.
model PngImage is File<"image/png", bytes>;

// ===========================================================================
// Global (non-device) routes
// ===========================================================================

/**
 * List all attached / reachable devices.
 */
@route("/list")
@get
@summary("List devices")
op listDevices(): DeviceList | UnauthorizedError | InternalError;

// ===========================================================================
// Device-scoped routes: /device/{udid}/...
// ===========================================================================

@route("/device/{udid}")
namespace Devices {
  // -------------------------------------------------------------------------
  // Device lifecycle / diagnostics
  // -------------------------------------------------------------------------

  /**
   * Activate the device (complete Setup Assistant / activation).
   */
  @route("/activate")
  @post
  @summary("Activate device")
  op activate(...UdidPath): GenericResponse | DeviceErrors;

  /**
   * Get lockdown values plus `instruments:*` keys for the device.
   * Returns an open dictionary of heterogeneous values.
   */
  @route("/info")
  @get
  @summary("Get device info")
  op getInfo(...UdidPath): DeviceInfo | DeviceErrors;

  /**
   * List installed configuration profiles. Returns an open dictionary.
   */
  @route("/profiles")
  @get
  @summary("List configuration profiles")
  op getProfiles(...UdidPath): InstalledProfiles | DeviceErrors;

  /**
   * Reset accessibility settings on the device.
   */
  @route("/resetaccessibility")
  @post
  @summary("Reset accessibility")
  op resetAccessibility(...UdidPath): GenericResponse | DeviceErrors;

  /**
   * Reset the simulated location back to the device's real GPS location.
   */
  @route("/resetlocation")
  @post
  @summary("Reset simulated location")
  op resetLocation(...UdidPath): GenericResponse | DeviceErrors;

  /**
   * Simulate a GPS location on the device.
   *
   * NOTE: the longitude parameter was historically misspelled `longtitude` on
   * the wire. This spec fixes it to `longitude`; the go-ios server accepts
   * `longitude` (and may keep `longtitude` as a deprecated alias).
   */
  @route("/setlocation")
  @put
  @summary("Set simulated location")
  op setLocation(
    ...UdidPath,

    /** Latitude in decimal degrees. */
    @query latitude: string,

    /** Longitude in decimal degrees. */
    @query longitude: string,
  ): GenericResponse | BadRequestError | DeviceErrors;

  // -------------------------------------------------------------------------
  // Pairing
  // -------------------------------------------------------------------------

  /**
   * Pair with the device.
   *
   * For a supervised pairing (`supervised=true`) upload the supervision
   * identity as `p12file` (multipart) and supply the passphrase in the
   * `Supervision-Password` header.
   *
   * Returns `423` when the device is locked and pairing cannot proceed.
   */
  @route("/pair")
  @post
  @summary("Pair device")
  op pair(
    ...UdidPath,

    /** Whether this is a supervised pairing. */
    @query supervised: boolean,

    /** Supervision identity passphrase (required when supervised). */
    @header("Supervision-Password") supervisionPassword?: string,

    /**
     * Multipart body carrying the supervision identity (`.p12`) when
     * `supervised=true`. Omit for unsupervised pairing.
     */
    @multipartBody body?: {
      /** The supervision identity certificate file. */
      p12file: HttpPart<File>;
    },
  ): GenericResponse | PairLockedError | DeviceErrors;

  // -------------------------------------------------------------------------
  // Developer Disk Image (DDI)
  // -------------------------------------------------------------------------

  /**
   * List the hex signatures of Developer Disk Images mounted on the device.
   */
  @route("/image")
  @get
  @summary("List mounted developer images")
  op listImages(...UdidPath): {
    @body signatures: string[];
  } | DeviceErrors;

  /**
   * Mount a Developer Disk Image.
   *
   * Either let the server auto-resolve and download the correct image
   * (`auto=true`, optionally with `basedir`), or stream the image bytes as the
   * raw request body (up to 2 GiB).
   */
  @route("/image")
  @put
  @summary("Mount a developer image")
  op mountImage(
    ...UdidPath,

    /** Auto-resolve and download the matching DDI for the device. */
    @query auto?: boolean,

    /** Base directory the server uses to cache/lookup DDIs when `auto=true`. */
    @query basedir?: string,

    /**
     * Raw Developer Disk Image bytes (used when not auto-resolving).
     * Content up to 2 GiB.
     */
    @body image?: bytes,
  ): GenericResponse | BadRequestError | DeviceErrors;

  // -------------------------------------------------------------------------
  // Screenshot (binary PNG)
  // -------------------------------------------------------------------------

  /**
   * Capture a screenshot. Returns raw PNG bytes (`image/png`).
   */
  @route("/screenshot")
  @get
  @summary("Capture screenshot")
  op screenshot(...UdidPath): PngImage | DeviceErrors;

  // -------------------------------------------------------------------------
  // Condition inducers
  // -------------------------------------------------------------------------

  /**
   * List available condition inducer profile types.
   */
  @route("/conditions")
  @get
  @summary("List conditions")
  op listConditions(...UdidPath): {
    @body conditions: ProfileType[];
  } | DeviceErrors;

  /**
   * Enable a condition inducer profile.
   */
  @route("/enable-condition")
  @put
  @summary("Enable condition")
  op enableCondition(
    ...UdidPath,

    /** Identifier of the condition profile type. */
    @query profileTypeID: string,

    /** Identifier of the specific profile to activate. */
    @query profileID: string,
  ): GenericResponse | BadRequestError | DeviceErrors;

  /**
   * Disable the currently active condition inducer profile.
   */
  @route("/disable-condition")
  @post
  @summary("Disable condition")
  op disableCondition(...UdidPath): GenericResponse | DeviceErrors;

  // -------------------------------------------------------------------------
  // WebDriverAgent / XCUITest sessions
  // -------------------------------------------------------------------------

  /**
   * Start a WebDriverAgent (XCUITest) session.
   */
  @route("/wda/session")
  @post
  @summary("Start WDA session")
  op createWdaSession(
    ...UdidPath,
    @body config: WdaConfig,
  ): WdaSession | BadRequestError | DeviceErrors;

  /**
   * Get a running WebDriverAgent session. Returns `404` for an unknown session.
   */
  @route("/wda/session/{sessionId}")
  @get
  @summary("Get WDA session")
  op getWdaSession(
    ...UdidPath,

    /** The WDA session id. */
    @path sessionId: string,
  ): WdaSession | SessionNotFoundError | DeviceErrors;

  /**
   * Stop a running WebDriverAgent session.
   */
  @route("/wda/session/{sessionId}")
  @delete
  @summary("Stop WDA session")
  op deleteWdaSession(
    ...UdidPath,

    /** The WDA session id. */
    @path sessionId: string,
  ): WdaSession | SessionNotFoundError | DeviceErrors;

  // -------------------------------------------------------------------------
  // Apps (serialized per-udid)
  // -------------------------------------------------------------------------

  /**
   * List installed applications. Each entry is an open Info.plist map.
   */
  @route("/apps/")
  @get
  @summary("List apps")
  op listApps(...UdidPath): {
    @body apps: AppInfo[];
  } | DeviceErrors;

  /**
   * Launch an application by bundle id.
   */
  @route("/apps/launch")
  @post
  @summary("Launch app")
  op launchApp(
    ...UdidPath,

    /** Bundle id of the app to launch. */
    @query bundleID: string,
  ): GenericResponse | BadRequestError | DeviceErrors;

  /**
   * Kill a running application by bundle id.
   */
  @route("/apps/kill")
  @post
  @summary("Kill app")
  op killApp(
    ...UdidPath,

    /** Bundle id of the app to kill. */
    @query bundleID: string,
  ): GenericResponse | BadRequestError | DeviceErrors;

  /**
   * Install an application from an uploaded `.ipa`/`.app` archive.
   * The multipart `file` part must be 1 byte–200 MB.
   */
  @route("/apps/install")
  @post
  @summary("Install app")
  op installApp(
    ...UdidPath,

    /** Multipart body carrying the app archive to install. */
    @multipartBody body: {
      /** The `.ipa`/`.app` archive (1 B – 200 MB). */
      file: HttpPart<File>;
    },
  ): GenericResponse | BadRequestError | DeviceErrors;

  /**
   * Uninstall an application by bundle id.
   */
  @route("/apps/uninstall")
  @post
  @summary("Uninstall app")
  op uninstallApp(
    ...UdidPath,

    /** Bundle id of the app to uninstall. */
    @query bundleID: string,
  ): GenericResponse | BadRequestError | DeviceErrors;

  // -------------------------------------------------------------------------
  // Streaming (Server-Sent Events)
  // -------------------------------------------------------------------------

  /**
   * Stream application state-change notifications as Server-Sent Events.
   */
  @route("/notifications")
  @get
  @summary("Stream app-state notifications (SSE)")
  @extension(
    "x-sse-events",
    #{
      schema: "NotificationEvents",
      events: #{ appstate: "AppStateNotification", heartbeat: "Heartbeat" },
    }
  )
  op streamNotifications(...UdidPath): SSEStream<NotificationEvents> | DeviceErrors;

  /**
   * Stream device syslog lines as Server-Sent Events.
   */
  @route("/syslog")
  @get
  @summary("Stream syslog (SSE)")
  @extension(
    "x-sse-events",
    #{
      schema: "SyslogEvents",
      events: #{ syslog: "SyslogMessage", heartbeat: "Heartbeat" },
    }
  )
  op streamSyslog(...UdidPath): SSEStream<SyslogEvents> | DeviceErrors;

  /**
   * Stream structured os_log trace entries as Server-Sent Events.
   * All filters are optional and combine with AND semantics.
   */
  @route("/ostrace")
  @get
  @summary("Stream os_log trace (SSE)")
  @extension(
    "x-sse-events",
    #{
      schema: "OsTraceEvents",
      events: #{ ostrace: "OsTraceEntry", heartbeat: "Heartbeat" },
    }
  )
  op streamOsTrace(
    ...UdidPath,

    /** Only include entries from this process id. */
    @query pid?: int32,

    /** Minimum log level to include (e.g. `info`, `debug`, `error`). */
    @query level?: string,

    /** Only include entries from this subsystem. */
    @query subsystem?: string,

    /** Only include entries whose message matches this substring/pattern. */
    @query match?: string,

    /** Exclude entries whose message matches this substring/pattern. */
    @query exclude?: string,
  ): SSEStream<OsTraceEvents> | DeviceErrors;

  /**
   * Stream device attach/detach events as Server-Sent Events.
   */
  @route("/listen")
  @get
  @summary("Stream device attach/detach (SSE)")
  @extension(
    "x-sse-events",
    #{
      schema: "ListenEvents",
      events: #{ attachdetach: "AttachDetachEvent", heartbeat: "Heartbeat" },
    }
  )
  op streamListen(...UdidPath): SSEStream<ListenEvents> | DeviceErrors;
}

// ===========================================================================
// Route-specific error responses
// ===========================================================================

/** 423 — device is locked; pairing cannot proceed. */
@error
model PairLockedError {
  @statusCode statusCode: 423;
  @body body: GenericResponse;
}

/** 404 — WDA session id not found. */
@error
model SessionNotFoundError {
  @statusCode statusCode: 404;
  @body body: GenericResponse;
}
