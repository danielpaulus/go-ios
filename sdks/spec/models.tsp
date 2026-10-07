import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace GoIos;

// ---------------------------------------------------------------------------
// Core response envelope
// ---------------------------------------------------------------------------

/**
 * The dominant response envelope used across the API. Success responses set
 * `message`; error responses set `error`. Streaming/middleware paths that emit
 * `gin.H{"error"|"message"}` are compatible with this shape.
 */
model GenericResponse {
  /** Human-readable success or status message. */
  message?: string;

  /** Human-readable error message. Present on failures. */
  error?: string;
}

// ---------------------------------------------------------------------------
// WebDriverAgent / XCUITest
// ---------------------------------------------------------------------------

/**
 * Configuration for launching a WebDriverAgent (XCUITest) runner session.
 */
model WdaConfig {
  /** Bundle id of the WDA runner host app (e.g. `com.facebook.WebDriverAgentRunner.xctrunner`). */
  bundleId: string;

  /** Bundle id of the XCTest test bundle. */
  testBundleId: string;

  /** Path/name of the `.xctestconfiguration` to use. */
  xcTestConfig: string;

  /** Extra process arguments passed to the runner. */
  args?: string[];

  /** Extra environment variables passed to the runner. */
  env?: Record<string>;
}

/**
 * A running WebDriverAgent session.
 */
model WdaSession {
  /** The configuration the session was started with. */
  config: WdaConfig;

  /** Opaque session identifier. */
  sessionId: string;

  /** The device udid the session runs on. */
  udid: string;
}

// ---------------------------------------------------------------------------
// Device listing
// ---------------------------------------------------------------------------

/**
 * Low-level device properties reported by usbmuxd / lockdown.
 */
model DeviceProperties {
  connectionSpeed?: int32;
  connectionType?: string;
  deviceID?: int32;
  locationID?: int32;
  productID?: int32;

  /** The device udid (serial number). This is what device-scoped routes key on. */
  serialNumber: string;
}

/**
 * A single device as returned by `GET /list`.
 */
model DeviceEntry {
  deviceID: int32;
  messageType?: string;
  properties: DeviceProperties;

  /** Network address for a device reached over the network / tunnel. */
  address?: string;

  /** True if reachable via the userspace TUN tunnel. */
  userspaceTUN?: boolean;
  userspaceTUNHost?: string;
  userspaceTUNPort?: int32;
}

/**
 * Response of `GET /list`.
 */
model DeviceList {
  deviceList: DeviceEntry[];
}

// ---------------------------------------------------------------------------
// Conditions (condition inducer profiles)
// ---------------------------------------------------------------------------

/**
 * A single condition profile within a `ProfileType`.
 */
model Profile {
  description?: string;
  identifier: string;
  name: string;
}

/**
 * A condition inducer profile type (e.g. thermal, network) with its variants.
 */
model ProfileType {
  activeProfile?: string;
  identifier: string;
  profilesSorted?: boolean;
  isActive?: boolean;
  name: string;
  isDestructive?: boolean;
  isInternal?: boolean;
  profiles: Profile[];
}

// ---------------------------------------------------------------------------
// Apps
// ---------------------------------------------------------------------------

/**
 * Installed application metadata. This is an open map: keys come straight from
 * the app's Info.plist. Common keys are surfaced for discoverability but any
 * additional keys may be present.
 */
model AppInfo is Record<unknown> {
  CFBundleIdentifier?: string;
  CFBundleExecutable?: string;
  CFBundleName?: string;
  CFBundleShortVersionString?: string;
  Path?: string;
  UIFileSharingEnabled?: boolean;
}

// ---------------------------------------------------------------------------
// Open dictionary responses
// ---------------------------------------------------------------------------

/**
 * `GET /device/{udid}/info` — lockdown values plus `instruments:*` keys.
 * Open dictionary; values are heterogeneous.
 */
model DeviceInfo is Record<unknown>;

/**
 * `GET /device/{udid}/profiles` — installed configuration profiles.
 * Open dictionary; values are heterogeneous.
 */
model InstalledProfiles is Record<unknown>;

// ---------------------------------------------------------------------------
// Shared error responses (consistent GenericResponse envelope)
// ---------------------------------------------------------------------------

/** 401 — missing/invalid bearer token (when auth is enabled). */
@error
model UnauthorizedError {
  @statusCode statusCode: 401;
  @body body: GenericResponse;
}

/** 404 — device (udid) not found. */
@error
model DeviceNotFoundError {
  @statusCode statusCode: 404;
  @body body: GenericResponse;
}

/** 422 — empty/invalid udid. */
@error
model UnprocessableUdidError {
  @statusCode statusCode: 422;
  @body body: GenericResponse;
}

/** 400 — malformed request (missing required query/body, bad payload). */
@error
model BadRequestError {
  @statusCode statusCode: 400;
  @body body: GenericResponse;
}

/** 500 — internal error while talking to the device. */
@error
model InternalError {
  @statusCode statusCode: 500;
  @body body: GenericResponse;
}

/**
 * Error set shared by every device-scoped route. Individual routes add
 * route-specific errors (e.g. 423 locked, 404 unknown session) on top.
 */
alias DeviceErrors = UnauthorizedError | DeviceNotFoundError | UnprocessableUdidError | InternalError;

/** 404 — the requested resource (e.g. a job) was not found for this device. */
@error
model NotFoundError {
  @statusCode statusCode: 404;
  @body body: GenericResponse;
}

/** 502 — the tunnel agent could not be reached or returned an error. */
@error
model BadGatewayError {
  @statusCode statusCode: 502;
  @body body: GenericResponse;
}

/**
 * 424 — a device-side prerequisite is missing. Used by the WebInspector routes
 * when Web Inspector / Remote Automation is not enabled on the device.
 */
@error
model FailedDependencyError {
  @statusCode statusCode: 424;
  @body body: GenericResponse;
}

/** 413 — the uploaded body exceeded the server's size cap. */
@error
model PayloadTooLargeError {
  @statusCode statusCode: 413;
  @body body: GenericResponse;
}

/** 501 — the selected UI-automation backend does not support this operation. */
@error
model NotImplementedError {
  @statusCode statusCode: 501;
  @body body: GenericResponse;
}

// ---------------------------------------------------------------------------
// Tunnel agent
// ---------------------------------------------------------------------------

/**
 * A running device tunnel as reported by the tunnel agent
 * (`GET /tunnels`, `POST /tunnels/{udid}/refresh`). Mirrors `tunnel.Tunnel`.
 */
model Tunnel {
  /** The device udid this tunnel serves. */
  Udid: string;

  /** Tunnel address (IPv6) reachable for RemoteXPC/RSD. */
  Address: string;

  /** RemoteServiceDiscovery port on the tunnel. */
  RsdPort: int32;

  /** Whether this tunnel is a userspace TUN. */
  UserspaceTUN?: boolean;

  /** Userspace TUN port, when `UserspaceTUN` is true. */
  UserspaceTUNPort?: int32;
}

/** `DELETE /tunnels/{udid}` — acknowledgement that the tunnel was stopped. */
model TunnelStopped {
  udid: string;

  /** Always `stopped`. */
  status: string;
}

/** `POST /tunnel-agent/shutdown` — acknowledgement. */
model AgentShutdown {
  /** Always `agent shutdown requested`. */
  status: string;
}

// ---------------------------------------------------------------------------
// Device info (read-only)
// ---------------------------------------------------------------------------

/** `GET /device/{udid}/devicename`. */
model DeviceName {
  devicename: string;
}

/** `GET /device/{udid}/date`. */
model DeviceDate {
  /** Human-readable RFC850 date on the device. */
  formatedDate: string;

  /** Device clock as Unix epoch seconds. */
  TimeIntervalSince1970: float64;
}

/**
 * `GET /device/{udid}/battery` — battery diagnostics (`ios.BatteryInfo`).
 * Open map; commonly-present keys are surfaced for discoverability.
 */
model BatteryInfo is Record<unknown> {
  CurrentCapacity?: int32;
  ExternalConnected?: boolean;
  FullyCharged?: boolean;
  IsCharging?: boolean;
  Temperature?: int32;
}

/** `GET /device/{udid}/diagnostics` — all IORegistry/diagnostic values. Open map. */
model Diagnostics is Record<unknown>;

/** `GET /device/{udid}/mobilegestalt` — queried MobileGestalt keys. Open map. */
model MobileGestalt is Record<unknown>;

/** `GET /device/{udid}/lockdown` — all lockdown values. Open map. */
model LockdownValues is Record<unknown>;

/**
 * A running process entry (`instruments.ProcessInfo`) from
 * `GET /device/{udid}/processes`.
 */
model ProcessInfo {
  pid: int32;
  name: string;
  realAppName?: string;
  isApplication?: boolean;

  /** Process start time, ISO-8601. */
  startDate?: utcDateTime;
}

// ---------------------------------------------------------------------------
// Device management
// ---------------------------------------------------------------------------

/** `GET /device/{udid}/devmode` — developer mode state. */
model DevModeState {
  DeveloperModeEnabled: boolean;
}

/** `POST /device/{udid}/devmode` request. */
model DevModeRequest {
  /** `enable` to turn developer mode on, `reveal` to expose the settings menu. */
  action: string;

  /** When enabling, also arm developer mode to persist across the next reboot. */
  enablePostRestart?: boolean;
}

/**
 * Language/locale configuration (`ios.LanguageConfiguration`), returned by
 * `GET/PUT /device/{udid}/lang`.
 */
model LanguageConfiguration {
  Language?: string;
  Locale?: string;

  /** Supported locales advertised by the device. */
  SupportedLocales?: string[];

  /** Supported UI languages advertised by the device. */
  SupportedLanguages?: string[];
}

/** `PUT /device/{udid}/lang` request. */
model SetLanguageRequest {
  language?: string;
  locale?: string;
}

/** `POST /device/{udid}/memlimitoff` request. */
model MemLimitRequest {
  /** Process name whose memory limit should be waived. */
  process: string;
}

/** `POST /device/{udid}/memlimitoff` response. */
model MemLimitResult {
  process: string;
  pid: int32;
  disabled: boolean;
}

// ---------------------------------------------------------------------------
// Files & crash reports
// ---------------------------------------------------------------------------

/** Domain of the on-device file service. */
union FileDomain {
  string,
  app: "app",
  appGroup: "app-group",
  crash: "crash",
  temp: "temp",
}

/** A single entry in a device directory listing. */
model FileEntry is Record<unknown> {
  name?: string;
  path?: string;
  isDir?: boolean;
  size?: int64;
}

/** `GET /device/{udid}/files` — directory listing. */
model FileListing {
  path: string;
  files: FileEntry[];
  count: int32;
}

/** `POST /device/{udid}/files/push` — acknowledgement. */
model FilePushResult {
  remote: string;
  size: int64;
}

/** `GET /device/{udid}/crashes` — crash report names. */
model CrashListing {
  files: string[];
  count: int32;
}

// ---------------------------------------------------------------------------
// Media (wallpaper, icon layout, pasteboard)
// ---------------------------------------------------------------------------

/** `GET /device/{udid}/icon-layout` — SpringBoard icon layout. Open structure. */
model IconLayout is Record<unknown>;

/** `GET /device/{udid}/pasteboard` — clipboard contents. */
model PasteboardContent {
  /** Whether any text was present on the pasteboard. */
  present: boolean;

  /** The clipboard text (empty when `present` is false). */
  text: string;
}

// ---------------------------------------------------------------------------
// Settings (accessibility, time format, wifi)
// ---------------------------------------------------------------------------

/** `GET /device/{udid}/assistivetouch` — AssistiveTouch state. */
model AssistiveTouchState {
  AssistiveTouchEnabled: boolean;
}

/** `GET /device/{udid}/timeformat` — 24-hour clock state. */
model TimeFormatState {
  Uses24HourClock: boolean;
}

/** Request body for the `enabled`-toggle settings endpoints. */
model EnabledRequest {
  enabled: boolean;
}

/** `PUT /device/{udid}/timeformat` request. */
model TimeFormatRequest {
  uses24Hour: boolean;
}

/** `PUT /device/{udid}/wifi` request. */
model WifiRequest {
  ssid: string;
  password?: string;

  /** Encryption type, e.g. `WPA2`, `WPA`, `WEP`, `None`. */
  encType?: string;
}

// ---------------------------------------------------------------------------
// Config (profiles & images)
// ---------------------------------------------------------------------------

/** `GET /device/{udid}/image/list` — mounted DDI signatures. */
model MountedImages {
  /** Hex-encoded image signatures. */
  signatures: string[];
  count: int32;
}

// ---------------------------------------------------------------------------
// MDM (supervised)
// ---------------------------------------------------------------------------

/** `POST /device/{udid}/mdm/security-info` — device security info. Open map. */
model SecurityInfo is Record<unknown>;

/** `POST /device/{udid}/mdm/fetch-unlock-token` — base64 escrow unlock token. */
model UnlockToken {
  /** Base64-encoded escrow unlock token. */
  token: string;
}

/** Simple `{ "status": "ok" }` acknowledgement used by MDM clear operations. */
model StatusOk {
  status: string;
}

// ---------------------------------------------------------------------------
// Async jobs
// ---------------------------------------------------------------------------

/** Job lifecycle state. */
union JobStatus {
  string,
  running: "running",
  succeeded: "succeeded",
  failed: "failed",
  stopped: "stopped",
}

/**
 * A long-running operation started via the REST API (test run, WDA runner,
 * port forward). Mirrors the server's `jobView`.
 */
model Job {
  /** Opaque job id, e.g. `runtest-3`. */
  id: string;

  /** Job kind: `runtest`, `runwda` or `forward`. */
  kind: string;

  /** The device udid the job runs on. */
  udid: string;

  status: JobStatus;

  /** When the job started (ISO-8601). */
  startedAt: utcDateTime;

  /** When the job reached a terminal state (absent while running). */
  finishedAt?: utcDateTime;

  /** Error message when `status` is `failed`. */
  error?: string;

  /** Terminal result payload (e.g. test suites) when the job succeeded. */
  result?: unknown;
}

/** `POST /device/{udid}/jobs/runtest` (and `runwda`) request. */
model RunTestRequest {
  /** Bundle id of the app under test. */
  bundleId?: string;

  /** Bundle id of the test runner. Defaults to `bundleId` if omitted. */
  testRunnerBundleId?: string;

  /** Name of the `.xctestconfiguration`. */
  xctestConfig?: string;

  /** Extra environment variables for the test runner. */
  env?: Record<unknown>;

  /** Extra process arguments for the test runner. */
  args?: string[];

  /** Only run these tests (`Class/method` identifiers). */
  testsToRun?: string[];

  /** Skip these tests. */
  testsToSkip?: string[];

  /** Run as a plain XCTest (vs XCUITest). */
  xctest?: boolean;
}

/** `POST /device/{udid}/jobs/forward` request. */
model ForwardRequest {
  /** Local (host) TCP port to listen on. */
  hostPort: uint16;

  /** Device TCP port to forward to. */
  targetPort: uint16;
}

// ===========================================================================
// Wave 1a — Diagnostics & network
// ===========================================================================

/**
 * `GET /device/{udid}/diskspace` — AFC filesystem info (`afc.DeviceInfo`).
 * Total/free/used bytes and block size. Open map; common keys surfaced.
 */
model DiskSpaceInfo is Record<unknown> {
  /** Total filesystem capacity in bytes. */
  FSTotalBytes?: int64;

  /** Free filesystem space in bytes. */
  FSFreeBytes?: int64;

  /** Filesystem block size in bytes. */
  FSBlockSize?: int64;

  /** AFC model identifier reported by the device. */
  Model?: string;
}

/**
 * `GET /device/{udid}/ip` — device network info discovered over pcapd
 * (`pcap.NetworkInfo`).
 */
model NetworkInfo is Record<unknown> {
  /** Hardware (MAC) address. */
  MacAddress?: string;

  /** IPv4 address, when discovered. */
  IPv4?: string;

  /** IPv6 address, when discovered. */
  IPv6?: string;
}

/** A single RSD (Remote Service Discovery) service entry. */
model RsdServiceEntry {
  /** TCP port the service is reachable on over the tunnel. */
  Port?: int32;

  /** Wire protocol (e.g. `tcp`). */
  ProtocolType?: string;
}

/**
 * `GET /device/{udid}/rsd` — the device's Remote Service Discovery service list
 * keyed by service name. Requires a running tunnel (iOS 17+); devices without
 * RSD return `400`.
 */
model RsdServices is Record<RsdServiceEntry>;

/**
 * `GET /device/{udid}/battery/registry` — battery IORegistry stats
 * (`diagnostics.IORegistry`). Open map; common keys surfaced.
 */
model BatteryRegistry is Record<unknown> {
  Temperature?: int32;
  Voltage?: int32;
  CurrentCapacity?: int32;
  InstantAmperage?: int64;
  IsCharging?: boolean;
  FullyCharged?: boolean;
}

// ===========================================================================
// Wave 1b — Accessibility & location
// ===========================================================================

/** `GET|PUT /device/{udid}/voiceover` — VoiceOver enabled state. */
model VoiceOverState {
  VoiceOverEnabled: boolean;
}

/** `GET|PUT /device/{udid}/zoom` — ZoomTouch enabled state. */
model ZoomTouchState {
  ZoomTouchEnabled: boolean;
}

/**
 * Body for the accessibility toggle PUTs (`/voiceover`, `/zoom`). The desired
 * state may also be supplied as an `enabled` query param; a parseable body wins.
 */
model AXEnabledRequest {
  enabled: boolean;
}

/**
 * One accessibility audit issue (`accessibility.AXAuditIssue`) from
 * `POST /device/{udid}/ax/audit`. Open map — shape depends on the audit type.
 */
model AXAuditIssue is Record<unknown>;

/**
 * `GET /device/{udid}/ax` — a snapshot of the currently focused accessibility
 * element. Open map (backend-defined element attributes).
 */
model AXElement is Record<unknown>;

// ===========================================================================
// Wave 1c — AFC filesystem (fsync) & provisioning
// ===========================================================================

/** One entry returned by the recursive `GET /device/{udid}/fsync/tree` walk. */
model FsyncTreeEntry {
  /** Full device-side path of this entry. */
  path: string;

  /** Base name of the entry. */
  name: string;

  /** Whether the entry is a directory. */
  isDir: boolean;

  /** Size in bytes. */
  size: int64;
}

/** `GET /device/{udid}/fsync/ls` — a directory listing over AFC. */
model FsyncListing {
  /** The listed (cleaned) device path. */
  path: string;

  /** File/directory names in the listed directory. */
  files: string[];

  /** Number of entries. */
  count: int32;
}

/** `GET /device/{udid}/fsync/tree` — a recursive directory walk over AFC. */
model FsyncTreeListing {
  /** The root (cleaned) device path. */
  path: string;

  /** Flattened list of entries in the subtree. */
  entries: FsyncTreeEntry[];

  /** Number of entries. */
  count: int32;
}

/** `POST /device/{udid}/fsync/push` — result of an upload over AFC. */
model FsyncPushResult {
  /** Destination device path written. */
  path: string;

  /** Number of bytes written. */
  size: int64;
}

/**
 * `POST /device/{udid}/fsync/mkdir` and `DELETE /device/{udid}/fsync/rm` —
 * simple message + path acknowledgement.
 */
model FsyncMessage {
  /** Human-readable result message (e.g. `created`, `removed`). */
  message: string;

  /** The (cleaned) device path acted on. */
  path: string;
}

/**
 * `GET /device/{udid}/cloudconfig` — the device cloud configuration
 * (`mcinstall` GetCloudConfiguration): supervision status, skip-setup options
 * and organization info. Open map.
 */
model CloudConfig is Record<unknown>;

/**
 * `GET /prepare/skip-options` — the static list of setup-pane skip options
 * usable when preparing a device. Host-scoped (device-free).
 */
model PrepareSkipOptions {
  /** All available skip-option identifiers. */
  options: string[];

  /** Number of options. */
  count: int32;
}

// ===========================================================================
// Wave 1d — WebInspector
// ===========================================================================

/**
 * One inspectable page (`webinspector.ApplicationPage`) from
 * `GET /device/{udid}/webinspector/pages`. Open map — carries the application
 * and page descriptors as the device reports them.
 */
model WebInspectorPage is Record<unknown>;

/** `POST /device/{udid}/webinspector/launch` request body. */
model WebInspectorLaunchRequest {
  /** URL to open. May alternatively be supplied as the `url` query param. */
  url?: string;

  /** Bundle id to open the URL in. Defaults to Safari. */
  bundleId?: string;
}

/** `POST /device/{udid}/webinspector/launch` — result of opening a URL. */
model WebInspectorLaunchResult {
  /** Bundle id the page was opened in. */
  bundleId: string;

  /** The resolved current URL after navigation. */
  url: string;

  /** The page title after navigation. */
  title: string;
}

/** `POST /device/{udid}/webinspector/eval` request body. */
model WebInspectorEvalRequest {
  /**
   * Inspectable page key. When empty the first matching web/javascript page
   * (optionally scoped by `bundleId`) is used.
   */
  page?: string;

  /** Optional bundle id to scope page selection. */
  bundleId?: string;

  /** JavaScript source to evaluate. Required. */
  script: string;
}

/** `POST /device/{udid}/webinspector/eval` — evaluation result. */
model WebInspectorEvalResult {
  /** The page key the script ran in. */
  page: string;

  /** The evaluation result (JSON value; shape depends on the script). */
  result: unknown;
}

// ===========================================================================
// Wave 2 — UI automation (proxied to a forwarded WDA/DeviceKit backend)
// ===========================================================================

/**
 * Query params common to every UI-automation route. Each may also be supplied
 * as a same-named request header. The endpoints do NOT start WDA; a forwarded
 * WDA/DeviceKit backend must already be reachable (see the UI routes doc).
 */
alias UIBackendParams = {
  /** Backend to target: `wda` (default) or `devicekit`. */
  @query backend?: string;

  /** Forwarded backend base URL (defaults per backend). */
  @query wdaUrl?: string;

  /** Per-request HTTP timeout in seconds (default 60). */
  @query timeout?: int32;
};

/** `POST /device/{udid}/ui/tap` request — absolute coordinates. */
model UITapRequest {
  x: int32;
  y: int32;
}

/** `POST /device/{udid}/ui/swipe` request — drag from (x1,y1) to (x2,y2). */
model UISwipeRequest {
  x1: int32;
  y1: int32;
  x2: int32;
  y2: int32;

  /** Gesture duration in seconds. */
  duration?: float64;
}

/** `POST /device/{udid}/ui/longpress` request — press and hold at (x,y). */
model UILongPressRequest {
  x: int32;
  y: int32;

  /** Hold duration in seconds. */
  duration?: float64;
}

/** `POST /device/{udid}/ui/type` request — keyboard input. */
model UITypeRequest {
  text: string;
}

/** `POST /device/{udid}/ui/button` request — hardware button by name. */
model UIButtonRequest {
  /** Button name (e.g. `home`, `volumeup`). WDA supports only `home`. */
  name: string;
}

/** `PUT /device/{udid}/ui/orientation` request. */
model UIOrientationRequest {
  /** Target orientation (e.g. `PORTRAIT`, `LANDSCAPE`). */
  orientation: string;
}

/** `POST /device/{udid}/ui/app/{launch,terminate}` request. */
model UIAppRequest {
  bundleId: string;
}

/**
 * `POST /device/{udid}/ui/api` request — raw passthrough to the backend
 * (`uidriver.APIRequest`). For WDA supply `method`/`path`/`body`; for DeviceKit
 * supply `rpcMethod`/`rpcParams`.
 */
model UIAPIRequest {
  /** HTTP method for a WDA passthrough (defaults to GET). */
  method?: string;

  /** HTTP path for a WDA passthrough (required for the wda backend). */
  path?: string;

  /** Raw HTTP request body for a WDA passthrough (base64 bytes on the wire). */
  body?: bytes;

  /** JSON-RPC method name for a DeviceKit passthrough. */
  rpcMethod?: string;

  /** JSON-RPC params for a DeviceKit passthrough. */
  rpcParams?: unknown;
}

/**
 * A backend passthrough response. The body and content-type are forwarded from
 * WDA/DeviceKit verbatim, so the shape is backend-defined (open map).
 */
model UIResponse is Record<unknown>;

// A UI source dump returned as the backend supplies it (XML for WDA).
model UISourceXml is File<"application/xml", string>;

// ===========================================================================
// Wave 3 — Binary streams (NOT SSE)
// ===========================================================================

// A live UI video stream. The wire content-type is chosen at request time:
// MJPEG (multipart/x-mixed-replace) by default, or video/H264 with ?codec=h264
// (devicekit backend). Modeled as raw octet-stream bytes; the real media type is
// per-request and recorded in x-content-type.
@extension("x-content-type", "multipart/x-mixed-replace (mjpeg) or video/H264 (h264)")
@extension("x-stream", "binary")
model UIVideoStream is File<"application/octet-stream", bytes>;

// An MJPEG screenshot stream. On the wire the response is actually
// `multipart/x-mixed-replace; boundary=BoundaryString`, each part an
// `image/jpeg` frame. TypeSpec's HTTP library rejects a `multipart/*` body that
// is not a structured @multipartBody payload (this is a continuous byte stream,
// not a form), so the schema models it as `image/jpeg` bytes and records the
// real wire media type in `x-content-type` for downstream consumers.
@extension("x-content-type", "multipart/x-mixed-replace; boundary=BoundaryString")
@extension("x-stream", "binary")
model MjpegStream is File<"image/jpeg", bytes>;

// A live libpcap capture stream (application/vnd.tcpdump.pcap): global header
// followed by one record per packet, pipeable into wireshark/tshark.
@extension("x-stream", "binary")
model PcapStream is File<"application/vnd.tcpdump.pcap", bytes>;

// ===========================================================================
// Wave 4 — Codesigning & device preparation
// ===========================================================================

// A generated PKCS#12 identity (certificate + private key) download.
model Pkcs12File is File<"application/x-pkcs12", bytes>;

// A signed application archive (.ipa) download.
model SignedIpaFile is File<"application/octet-stream", bytes>;

/**
 * `POST /sign/provision` — provisioning assets envelope. The mobileprovision
 * (and optionally the P12) are base64-encoded so one JSON response can carry
 * both binary artifacts. Host-scoped (device-free).
 */
model ProvisioningResult {
  /** The app bundle identifier registered with App Store Connect. */
  bundleId: string;

  /** The signing certificate resource id. */
  certificateId: string;

  /** The `.mobileprovision`, base64-encoded. */
  mobileprovisionBase64: string;

  /** The generated `.p12`, base64-encoded (absent when reusing a certificate). */
  p12Base64?: string;

  /** The password protecting `p12Base64`, echoed back (client-supplied). */
  p12Password?: string;
}

/**
 * `POST /prepare/create-cert` — a generated self-signed supervision identity,
 * returned as DER (base64) and PEM for both the certificate and private key.
 * Host-scoped (device-free).
 */
model SupervisionCert {
  /** Certificate in DER form, base64-encoded. */
  certDerBase64: string;

  /** Certificate in PEM form. */
  certPem: string;

  /** Private key in DER form, base64-encoded. */
  privateKeyDerBase64: string;

  /** Private key in PEM form. */
  privateKeyPem: string;
}

/** `POST /device/{udid}/prepare` — device preparation acknowledgement. */
model PrepareResult {
  /** Always `prepared`. */
  status: string;

  /** Whether the device was supervised (a supervision cert was supplied). */
  supervised: boolean;
}
