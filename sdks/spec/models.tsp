import "@typespec/http";

using Http;

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
