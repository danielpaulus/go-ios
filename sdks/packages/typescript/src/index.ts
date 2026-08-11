import { createClient, createConfig, type Client } from "@hey-api/client-fetch";

import * as api from "./generated/sdk.gen";
import type {
  AgentShutdown,
  AppInfo,
  AssistiveTouchState,
  AxAuditIssue,
  AxElement,
  BatteryInfo,
  BatteryRegistry,
  CloudConfig,
  CrashListing,
  DeviceDate,
  DeviceInfo,
  DeviceList,
  DevModeState,
  Diagnostics,
  DiskSpaceInfo,
  FileDomain,
  FileListing,
  FilePushResult,
  ForwardRequest,
  FsyncListing,
  FsyncMessage,
  FsyncPushResult,
  FsyncTreeListing,
  GenericResponse,
  IconLayout,
  InstalledProfiles,
  Job,
  LanguageConfiguration,
  LockdownValues,
  MemLimitResult,
  MobileGestalt,
  MountedImages,
  NetworkInfo,
  PasteboardContent,
  PrepareResult,
  PrepareSkipOptions,
  ProcessInfo,
  ProfileType,
  ProvisioningResult,
  RsdServices,
  RunTestRequest,
  SecurityInfo,
  StatusOk,
  SupervisionCert,
  TimeFormatState,
  Tunnel,
  TunnelStopped,
  UiResponse,
  UnlockToken,
  VoiceOverState,
  WdaConfig,
  WdaSession,
  WebInspectorEvalResult,
  WebInspectorLaunchResult,
  WebInspectorPage,
  ZoomTouchState,
} from "./generated/types.gen";
import { IosApiError, unwrap } from "./errors";
import { openBinaryStream, type BinaryStream } from "./binary";
import { parseSseStream, type SseEvent } from "./sse";
import type {
  JobLogEventMap,
  ListenEventMap,
  NotificationEventMap,
  OsTraceEventMap,
  SyslogEventMap,
  SysmontapEventMap,
} from "./events";

export { IosApiError } from "./errors";
export { openBinaryStream } from "./binary";
export type { BinaryStream } from "./binary";
export { SseFrameParser, parseSseStream, isSseEvent } from "./sse";
export type {
  SseEvent,
  SseFrame,
  KnownSseEvent,
  UnknownSseEvent,
} from "./sse";
export type {
  JobLogEventMap,
  ListenEventMap,
  NotificationEventMap,
  OsTraceEventMap,
  SyslogEventMap,
  SysmontapEventMap,
} from "./events";
export type * from "./generated/types.gen";

/** Options for constructing an {@link IosClient}. */
export interface IosClientOptions {
  /** Base URL of the go-ios REST server, e.g. `http://localhost:60105`. */
  baseUrl: string;
  /**
   * Bearer API key. Sent as `Authorization: Bearer <apiKey>` on every request.
   * Optional (the server may run with `--disable-auth`), but strongly encouraged
   * and always sent when present.
   */
  apiKey?: string;
  /** Optional custom `fetch` implementation (defaults to global `fetch`). */
  fetch?: typeof fetch;
}

/** Options accepted by streaming methods. */
export interface StreamOptions {
  /** Abort signal to cancel the stream and release the underlying connection. */
  signal?: AbortSignal;
}

/** Filters for {@link DeviceHandle.ostrace} (all optional, AND-combined). */
export interface OsTraceFilters {
  pid?: number;
  level?: string;
  subsystem?: string;
  match?: string;
  exclude?: string;
}

/** Options for {@link DeviceHandle.pair}. */
export interface PairOptions {
  supervised: boolean;
  /** Supervision identity `.p12` file (required when `supervised` is true). */
  p12file?: Blob | Uint8Array | ArrayBuffer;
  /** Supervision identity passphrase (`Supervision-Password` header). */
  supervisionPassword?: string;
}

/** Options for {@link DeviceHandle.installImage}. */
export interface InstallImageOptions {
  /** Auto-resolve and download the matching Developer Disk Image. */
  auto?: boolean;
  /** Base directory the server caches/looks up DDIs in when `auto` is true. */
  basedir?: string;
  /** Raw DDI bytes to upload (when not auto-resolving). */
  image?: Blob | Uint8Array | ArrayBuffer;
}

/** Supported binary inputs for multipart/raw request bodies. */
export type BinaryInput = Blob | Uint8Array | ArrayBuffer;

/** Options for {@link AppsHandle.install}. */
export type IpaInput = BinaryInput;

/** File-service domain selector shared by the {@link FilesHandle} methods. */
export interface FileScope {
  /** File service domain: `app`, `app-group`, `crash` or `temp`. */
  domain: FileDomain;
  /** Bundle/group id for the `app`/`app-group` domains. */
  identifier?: string;
}

/** A `.p12` supervision identity plus its optional passphrase. */
export interface SupervisionIdentity {
  /** The supervision identity `.p12` certificate bytes. */
  p12: BinaryInput;
  /** Passphrase for the `.p12` identity. */
  password?: string;
}

/** Options for {@link MediaHandle.setWallpaper}. */
export interface SetWallpaperOptions extends SupervisionIdentity {
  /** The wallpaper image bytes. */
  image: BinaryInput;
  /** Target screen (`home`, `lock`, `both`). */
  screen?: string;
}

/** Options for {@link ProfilesMixin.addProfile}. */
export interface AddProfileOptions {
  /** The `.mobileconfig` profile bytes. */
  profile: BinaryInput;
  /** Supervision identity for a supervised install. */
  p12?: BinaryInput;
  /** Passphrase for the `.p12` identity. */
  password?: string;
}

/** Options for {@link ProxyHandle.setHttpProxy}. */
export interface SetHttpProxyOptions extends SupervisionIdentity {
  /** Proxy host. */
  host: string;
  /** Proxy port. */
  port: string | number;
  /** Proxy username. */
  user?: string;
  /** Proxy password. */
  pass?: string;
}

/**
 * Fsync path scoping: which app container (if any) a path is resolved against.
 * Omitting `bundleId` targets the media directory (the default AFC service).
 */
export interface FsyncOptions {
  /** App bundle id to scope the path to that app's container. */
  bundleId?: string;
}

/** Options for {@link FsyncHandle.rm} — recursive removal + container scope. */
export interface FsyncRemoveOptions extends FsyncOptions {
  /** Remove directory contents recursively. */
  recursive?: boolean;
}

/**
 * Shared options for the {@link UiHandle} methods. Every UI route accepts a
 * backend selector, a forwarded backend URL and a per-request timeout.
 */
export interface UiOptions {
  /** Backend to target: `wda` (default) or `devicekit`. */
  backend?: string;
  /** Forwarded backend base URL (defaults per backend). */
  wdaUrl?: string;
  /** Per-request HTTP timeout in seconds (default 60). */
  timeout?: number;
}

/** Options for {@link UiHandle.stream} — codec plus backend selection + tuning. */
export interface UiStreamOptions extends UiOptions {
  /** Video codec: `mjpeg` (default) or `h264` (devicekit backend only). */
  codec?: string;
  /** Target frames per second (backend-dependent). */
  fps?: string | number;
  /** JPEG quality for the mjpeg codec. */
  quality?: string | number;
  /** Scale factor (backend-dependent). */
  scale?: string | number;
  /** Target bitrate for the h264 codec. */
  bitrate?: string | number;
  /** Abort signal to cancel the stream and release the connection. */
  signal?: AbortSignal;
}

/** Options for {@link DeviceHandle.screenshotStream}. */
export interface ScreenshotStreamOptions {
  /** JPEG quality (1–100, default 80). */
  quality?: number;
  /** Abort signal to cancel the stream and release the connection. */
  signal?: AbortSignal;
}

/** Options for {@link DeviceHandle.pcap}. */
export interface PcapOptions {
  /** Capture duration in seconds (default 60, max 3600). */
  timeout?: number;
  /** Abort signal to cancel the capture and release the connection. */
  signal?: AbortSignal;
}

/** Options for {@link WebInspectorHandle.launch}. */
export interface WebInspectorLaunchOptions {
  /** Restrict the launch to a specific app by bundle id. */
  bundleId?: string;
}

/** Options for {@link WebInspectorHandle.eval}. */
export interface WebInspectorEvalOptions {
  /** Target inspector page id (defaults to the first page). */
  page?: string;
  /** Restrict evaluation to a specific app by bundle id. */
  bundleId?: string;
}

/** Options for {@link DeviceHandle.axAudit}. */
export interface AxAuditOptions {
  /** Audit timeout in seconds (default 60). */
  timeout?: number;
}

/** Options for {@link DeviceHandle.prepare} (device supervision preparation). */
export interface PrepareOptions {
  /** Supervision identity (DER/PEM/P12). Omit to prepare unsupervised. */
  cert?: BinaryInput;
  /** P12 password (when `cert` is a P12). */
  p12password?: string;
  /** Setup panes to skip (see {@link PrepareHandle.skipOptions}). */
  skip?: string[];
  /** Supervision organization name. */
  orgname?: string;
  /** Device locale (default `en_US`). */
  locale?: string;
  /** Device language. */
  lang?: string;
}

/** App Store Connect API credentials shared by the signing/provisioning ops. */
export interface AscCredentials {
  /** `.p8` App Store Connect API key bytes. */
  ascPrivateKey: BinaryInput;
  /** App Store Connect key id. */
  ascKeyId: string;
  /** App Store Connect issuer id. */
  ascIssuerId: string;
  /** Revoke existing certificates first. */
  revokeExisting?: boolean;
  /** Password to protect the generated P12. */
  p12password?: string;
}

/** Options for {@link SignHandle.provision}. */
export interface SignProvisionOptions extends AscCredentials {
  /** App bundle identifier. */
  bundleId: string;
  /** Target device udid to register against the profile. */
  udid: string;
  /** Human-readable bundle name. */
  bundleName?: string;
  /** Profile name. */
  profileName?: string;
  /** Device name to register. */
  deviceName?: string;
  /** Existing certificate id to reuse. */
  certificateId?: string;
}

/** Options for {@link SignHandle.app}. */
export interface SignAppOptions {
  /** App or `.ipa` bytes to resign. */
  ipa: BinaryInput;
  /** Signing identity (`.p12`) bytes. */
  p12: BinaryInput;
  /** Provisioning profile (`.mobileprovision`) bytes. */
  profile: BinaryInput;
  /** P12 password. */
  p12password?: string;
  /** Override bundle id. */
  bundleId?: string;
}

/**
 * App-management sub-facade, reachable via `client.device(udid).apps`.
 */
export class AppsHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** List installed applications (each entry is an open Info.plist map). */
  async list(): Promise<AppInfo[]> {
    return unwrap(
      await api.devicesListApps({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Launch an application by bundle id. */
  async launch(bundleId: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesLaunchApp({
        client: this.client,
        path: { udid: this.udid },
        query: { bundleID: bundleId },
      }),
    );
  }

  /** Kill a running application by bundle id. */
  async kill(bundleId: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesKillApp({
        client: this.client,
        path: { udid: this.udid },
        query: { bundleID: bundleId },
      }),
    );
  }

  /** Install an application from an `.ipa`/`.app` archive. */
  async install(ipa: IpaInput): Promise<GenericResponse> {
    return unwrap(
      await api.devicesInstallApp({
        client: this.client,
        path: { udid: this.udid },
        body: { file: toBlob(ipa) },
      }),
    );
  }

  /** Uninstall an application by bundle id. */
  async uninstall(bundleId: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesUninstallApp({
        client: this.client,
        path: { udid: this.udid },
        query: { bundleID: bundleId },
      }),
    );
  }
}

/**
 * WebDriverAgent (XCUITest) session sub-facade, reachable via
 * `client.device(udid).wda`.
 */
export class WdaHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Start a WDA session with the given runner config. */
  async createSession(config: WdaConfig): Promise<WdaSession> {
    return unwrap(
      await api.devicesCreateWdaSession({
        client: this.client,
        path: { udid: this.udid },
        body: config,
      }),
    );
  }

  /** Read a running WDA session by id. */
  async readSession(id: string): Promise<WdaSession> {
    return unwrap(
      await api.devicesGetWdaSession({
        client: this.client,
        path: { udid: this.udid, sessionId: id },
      }),
    );
  }

  /** Stop a running WDA session by id. */
  async deleteSession(id: string): Promise<WdaSession> {
    return unwrap(
      await api.devicesDeleteWdaSession({
        client: this.client,
        path: { udid: this.udid, sessionId: id },
      }),
    );
  }
}

/**
 * On-device file-service sub-facade, reachable via `client.device(udid).files`.
 * Every call targets a {@link FileScope} (domain + optional identifier).
 */
export class FilesHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** List a device directory (defaults to `.`). */
  async ls(scope: FileScope, path = "."): Promise<FileListing> {
    return unwrap(
      await api.devicesListFiles({
        client: this.client,
        path: { udid: this.udid },
        query: { domain: scope.domain, path, ...identifierQuery(scope) },
      }),
    );
  }

  /** Download a file from the device; resolves to the raw bytes as a `Blob`. */
  async pull(scope: FileScope, remote: string): Promise<Blob> {
    const result = await api.devicesPullFile({
      client: this.client,
      path: { udid: this.udid },
      query: { domain: scope.domain, remote, ...identifierQuery(scope) },
      parseAs: "blob",
    });
    return unwrap(result as { data?: Blob; error?: unknown; response: Response });
  }

  /** Upload bytes to a device path. */
  async push(
    scope: FileScope,
    remote: string,
    data: BinaryInput,
  ): Promise<FilePushResult> {
    return unwrap(
      await api.devicesPushFile({
        client: this.client,
        path: { udid: this.udid },
        query: { domain: scope.domain, remote, ...identifierQuery(scope) },
        body: toBlob(data),
        bodySerializer: null,
      }),
    );
  }
}

/**
 * Crash-report sub-facade, reachable via `client.device(udid).crashes`.
 */
export class CrashesHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** List crash reports, optionally filtered by a glob `pattern`. */
  async list(pattern?: string): Promise<CrashListing> {
    return unwrap(
      await api.devicesListCrashes({
        client: this.client,
        path: { udid: this.udid },
        ...(pattern !== undefined ? { query: { pattern } } : {}),
      }),
    );
  }

  /** Delete crash reports matching `pattern` under `cwd`. */
  async remove(pattern: string, cwd = "."): Promise<GenericResponse> {
    return unwrap(
      await api.devicesRemoveCrashes({
        client: this.client,
        path: { udid: this.udid },
        query: { cwd, pattern },
      }),
    );
  }
}

/**
 * SpringBoard / pasteboard media sub-facade, reachable via
 * `client.device(udid).media`.
 */
export class MediaHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Get the home-screen wallpaper as PNG bytes. */
  async getWallpaper(): Promise<Blob> {
    const result = await api.devicesGetWallpaper({
      client: this.client,
      path: { udid: this.udid },
      parseAs: "blob",
    });
    return unwrap(result as { data?: Blob; error?: unknown; response: Response });
  }

  /** Set the wallpaper (supervised; requires an image + `.p12` identity). */
  async setWallpaper(opts: SetWallpaperOptions): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetWallpaper({
        client: this.client,
        path: { udid: this.udid },
        body: {
          image: toBlob(opts.image),
          p12: toBlob(opts.p12),
          ...(opts.password !== undefined ? { password: opts.password } : {}),
          ...(opts.screen !== undefined ? { screen: opts.screen } : {}),
        },
      }),
    );
  }

  /** Get the SpringBoard icon layout. */
  async getIconLayout(): Promise<IconLayout> {
    return unwrap(
      await api.devicesGetIconLayout({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Restore a SpringBoard icon layout (as returned by {@link getIconLayout}). */
  async setIconLayout(layout: IconLayout): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetIconLayout({
        client: this.client,
        path: { udid: this.udid },
        body: layout,
      }),
    );
  }

  /** Get the pasteboard (clipboard) contents. */
  async getPasteboard(): Promise<PasteboardContent> {
    return unwrap(
      await api.devicesGetPasteboard({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Set the pasteboard text. */
  async setPasteboard(text: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetPasteboard({
        client: this.client,
        path: { udid: this.udid },
        body: text,
      }),
    );
  }
}

/**
 * Settings sub-facade (AssistiveTouch, time format, wifi), reachable via
 * `client.device(udid).settings`.
 */
export class SettingsHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Get the AssistiveTouch state. */
  async assistiveTouch(): Promise<AssistiveTouchState> {
    return unwrap(
      await api.devicesGetAssistiveTouch({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Enable/disable AssistiveTouch. Returns the resulting state. */
  async setAssistiveTouch(enabled: boolean): Promise<AssistiveTouchState> {
    return unwrap(
      await api.devicesSetAssistiveTouch({
        client: this.client,
        path: { udid: this.udid },
        body: { enabled },
      }),
    );
  }

  /** Get the 24-hour clock state. */
  async timeFormat(): Promise<TimeFormatState> {
    return unwrap(
      await api.devicesGetTimeFormat({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Set 24-hour (`true`) or 12-hour (`false`) clock. Returns the resulting state. */
  async setTimeFormat(uses24Hour: boolean): Promise<TimeFormatState> {
    return unwrap(
      await api.devicesSetTimeFormat({
        client: this.client,
        path: { udid: this.udid },
        body: { uses24Hour },
      }),
    );
  }

  /** Provision a wifi network. */
  async setWifi(
    ssid: string,
    password?: string,
    encType?: string,
  ): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetWifi({
        client: this.client,
        path: { udid: this.udid },
        body: {
          ssid,
          ...(password !== undefined ? { password } : {}),
          ...(encType !== undefined ? { encType } : {}),
        },
      }),
    );
  }

  /** Remove a provisioned wifi network by SSID. */
  async removeWifi(ssid: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesRemoveWifi({
        client: this.client,
        path: { udid: this.udid },
        query: { ssid },
      }),
    );
  }
}

/**
 * MDM (supervised) sub-facade, reachable via `client.device(udid).mdm`. Every
 * call needs a `.p12` supervision identity.
 */
export class MdmHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Get device security info. */
  async securityInfo(identity: SupervisionIdentity): Promise<SecurityInfo> {
    return unwrap(
      await api.devicesMdmSecurityInfo({
        client: this.client,
        path: { udid: this.udid },
        body: identityBody(identity),
      }),
    );
  }

  /** Fetch the escrow unlock token (base64-encoded). */
  async fetchUnlockToken(identity: SupervisionIdentity): Promise<UnlockToken> {
    return unwrap(
      await api.devicesMdmFetchUnlockToken({
        client: this.client,
        path: { udid: this.udid },
        body: identityBody(identity),
      }),
    );
  }

  /** Clear the device passcode (needs the base64 unlock `token`). */
  async clearPasscode(
    identity: SupervisionIdentity,
    token: string,
  ): Promise<StatusOk> {
    return unwrap(
      await api.devicesMdmClearPasscode({
        client: this.client,
        path: { udid: this.udid },
        body: { ...identityBody(identity), token },
      }),
    );
  }

  /** Clear the Screen Time password. */
  async clearScreenTimePassword(
    identity: SupervisionIdentity,
  ): Promise<StatusOk> {
    return unwrap(
      await api.devicesMdmClearScreenTimePassword({
        client: this.client,
        path: { udid: this.udid },
        body: identityBody(identity),
      }),
    );
  }
}

/**
 * Global HTTP-proxy sub-facade, reachable via `client.device(udid).proxy`.
 */
export class ProxyHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Configure a global HTTP proxy (supervised; requires a `.p12` identity). */
  async setHttpProxy(opts: SetHttpProxyOptions): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetHttpProxy({
        client: this.client,
        path: { udid: this.udid },
        body: {
          host: opts.host,
          port: String(opts.port),
          p12: toBlob(opts.p12),
          ...(opts.user !== undefined ? { user: opts.user } : {}),
          ...(opts.pass !== undefined ? { pass: opts.pass } : {}),
          ...(opts.password !== undefined ? { password: opts.password } : {}),
        },
      }),
    );
  }

  /** Clear the global HTTP proxy. */
  async removeHttpProxy(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesRemoveHttpProxy({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }
}

/**
 * AFC file-service sub-facade, reachable via `client.device(udid).fsync`. Every
 * path is optionally scoped to an app container via `{ bundleId }`; omit it to
 * target the media directory. This is the newer `fsync` surface (ls/tree/pull/
 * push/rm/mkdir) distinct from the domain-based {@link FilesHandle}.
 */
export class FsyncHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** List a directory (defaults to `.`). */
  async ls(path = ".", opts: FsyncOptions = {}): Promise<FsyncListing> {
    return unwrap(
      await api.fsyncFsyncLs({
        client: this.client,
        path: { udid: this.udid },
        query: { path, ...bundleIdQuery(opts) },
      }),
    );
  }

  /** Recursively list a directory tree (defaults to `.`). */
  async tree(path = ".", opts: FsyncOptions = {}): Promise<FsyncTreeListing> {
    return unwrap(
      await api.fsyncFsyncTree({
        client: this.client,
        path: { udid: this.udid },
        query: { path, ...bundleIdQuery(opts) },
      }),
    );
  }

  /** Download a file from the device; resolves to the raw bytes. */
  async pull(path: string, opts: FsyncOptions = {}): Promise<Uint8Array> {
    const result = await api.fsyncFsyncPull({
      client: this.client,
      path: { udid: this.udid },
      query: { path, ...bundleIdQuery(opts) },
      parseAs: "blob",
    });
    const blob = unwrap(
      result as { data?: Blob; error?: unknown; response: Response },
    );
    return new Uint8Array(await blob.arrayBuffer());
  }

  /** Upload bytes to a device path. */
  async push(
    path: string,
    data: BinaryInput,
    opts: FsyncOptions = {},
  ): Promise<FsyncPushResult> {
    return unwrap(
      await api.fsyncFsyncPush({
        client: this.client,
        path: { udid: this.udid },
        query: { path, ...bundleIdQuery(opts) },
        body: toBlob(data),
        bodySerializer: null,
      }),
    );
  }

  /** Remove a file or directory (pass `{ recursive: true }` for directories). */
  async rm(path: string, opts: FsyncRemoveOptions = {}): Promise<FsyncMessage> {
    return unwrap(
      await api.fsyncFsyncRm({
        client: this.client,
        path: { udid: this.udid },
        query: {
          path,
          ...bundleIdQuery(opts),
          ...(opts.recursive !== undefined ? { recursive: opts.recursive } : {}),
        },
      }),
    );
  }

  /** Create a directory. */
  async mkdir(path: string, opts: FsyncOptions = {}): Promise<FsyncMessage> {
    return unwrap(
      await api.fsyncFsyncMkdir({
        client: this.client,
        path: { udid: this.udid },
        query: { path, ...bundleIdQuery(opts) },
      }),
    );
  }
}

/**
 * WebInspector (Safari / WKWebView) sub-facade, reachable via
 * `client.device(udid).webinspector`.
 */
export class WebInspectorHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** List inspectable pages (open tabs / web views). */
  async pages(): Promise<WebInspectorPage[]> {
    return unwrap(
      await api.webInspectorWebInspectorPages({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Open a URL in Safari (optionally scoped to a bundle id). */
  async launch(
    url: string,
    opts: WebInspectorLaunchOptions = {},
  ): Promise<WebInspectorLaunchResult> {
    return unwrap(
      await api.webInspectorWebInspectorLaunch({
        client: this.client,
        path: { udid: this.udid },
        body: {
          url,
          ...(opts.bundleId !== undefined ? { bundleId: opts.bundleId } : {}),
        },
      }),
    );
  }

  /** Evaluate JavaScript in an inspected page. */
  async eval(
    script: string,
    opts: WebInspectorEvalOptions = {},
  ): Promise<WebInspectorEvalResult> {
    return unwrap(
      await api.webInspectorWebInspectorEval({
        client: this.client,
        path: { udid: this.udid },
        body: {
          script,
          ...(opts.page !== undefined ? { page: opts.page } : {}),
          ...(opts.bundleId !== undefined ? { bundleId: opts.bundleId } : {}),
        },
      }),
    );
  }
}

/**
 * UI-automation sub-facade, reachable via `client.device(udid).ui`. Drives taps,
 * gestures, orientation, screenshots and app lifecycle through the WDA (default)
 * or devicekit backend. Every method accepts {@link UiOptions} (`backend`,
 * `wdaUrl`, `timeout`).
 */
export class UiHandle {
  /** App-lifecycle operations (launch / terminate / foreground). */
  readonly app: UiAppHandle;

  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {
    this.app = new UiAppHandle(client, udid);
  }

  /** Tap at a coordinate. */
  async tap(x: number, y: number, opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiTap({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { x, y },
      }),
    );
  }

  /** Swipe from one coordinate to another (optional `duration` in seconds). */
  async swipe(
    x1: number,
    y1: number,
    x2: number,
    y2: number,
    duration?: number,
    opts: UiOptions = {},
  ): Promise<UiResponse> {
    return unwrap(
      await api.uiUiSwipe({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { x1, y1, x2, y2, ...(duration !== undefined ? { duration } : {}) },
      }),
    );
  }

  /** Long-press at a coordinate (optional `duration` in seconds). */
  async longPress(
    x: number,
    y: number,
    duration?: number,
    opts: UiOptions = {},
  ): Promise<UiResponse> {
    return unwrap(
      await api.uiUiLongPress({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { x, y, ...(duration !== undefined ? { duration } : {}) },
      }),
    );
  }

  /** Type text into the focused field. */
  async type(text: string, opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiType({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { text },
      }),
    );
  }

  /** Press a hardware/named button (e.g. `home`, `volumeup`). */
  async button(name: string, opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiButton({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { name },
      }),
    );
  }

  /** Take a PNG screenshot through the UI backend; resolves to raw bytes. */
  async screenshot(opts: UiOptions = {}): Promise<Uint8Array> {
    const result = await api.uiUiScreenshot({
      client: this.client,
      path: { udid: this.udid },
      query: uiQuery(opts),
      parseAs: "blob",
    });
    const blob = unwrap(
      result as { data?: Blob; error?: unknown; response: Response },
    );
    return new Uint8Array(await blob.arrayBuffer());
  }

  /** Get the current UI hierarchy as an XML source string. */
  async source(opts: UiOptions = {}): Promise<string> {
    const result = await api.uiUiSource({
      client: this.client,
      path: { udid: this.udid },
      query: uiQuery(opts),
      parseAs: "text",
    });
    return unwrap(
      result as { data?: string; error?: unknown; response: Response },
    );
  }

  /** Get the screen size / window bounds. */
  async size(opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiWindowSize({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
      }),
    );
  }

  /** Get the current device orientation. */
  async orientation(opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiGetOrientation({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
      }),
    );
  }

  /** Set the device orientation (e.g. `portrait`, `landscapeLeft`). */
  async setOrientation(
    orientation: string,
    opts: UiOptions = {},
  ): Promise<UiResponse> {
    return unwrap(
      await api.uiUiSetOrientation({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { orientation },
      }),
    );
  }

  /** Get the UI backend status (health / session info). */
  async status(opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiStatus({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
      }),
    );
  }

  /**
   * Make a raw pass-through call to the UI backend (escape hatch for WDA
   * endpoints not modeled by the typed methods above).
   */
  async api(
    request: {
      method?: string;
      path?: string;
      body?: string;
      rpcMethod?: string;
      rpcParams?: unknown;
    },
    opts: UiOptions = {},
  ): Promise<UiResponse> {
    return unwrap(
      await api.uiUiApi({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: request,
      }),
    );
  }

  /**
   * Open a live video stream of the device screen as a raw byte stream (mjpeg
   * by default; h264 via the devicekit backend). Returns a {@link BinaryStream}
   * — iterate the chunks with `for await`, `break` or pass `signal` to stop.
   * This is a binary stream, not SSE.
   */
  stream(opts: UiStreamOptions = {}): Promise<BinaryStream> {
    const query = {
      ...uiQuery(opts),
      ...(opts.codec !== undefined ? { codec: opts.codec } : {}),
      ...(opts.fps !== undefined ? { fps: String(opts.fps) } : {}),
      ...(opts.quality !== undefined ? { quality: String(opts.quality) } : {}),
      ...(opts.scale !== undefined ? { scale: String(opts.scale) } : {}),
      ...(opts.bitrate !== undefined ? { bitrate: String(opts.bitrate) } : {}),
    };
    return openBinaryStream(
      api.streamsUiStream({
        client: this.client,
        path: { udid: this.udid },
        query,
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }
}

/**
 * App-lifecycle sub-facade of {@link UiHandle}, reachable via
 * `client.device(udid).ui.app`.
 */
export class UiAppHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Launch an app by bundle id through the UI backend. */
  async launch(bundleId: string, opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiAppLaunch({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { bundleId },
      }),
    );
  }

  /** Terminate an app by bundle id through the UI backend. */
  async terminate(bundleId: string, opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiAppTerminate({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
        body: { bundleId },
      }),
    );
  }

  /**
   * Bring the backgrounded app to the foreground (devicekit backend only; WDA
   * returns `501`). Takes no bundle id — it foregrounds the current app.
   */
  async foreground(opts: UiOptions = {}): Promise<UiResponse> {
    return unwrap(
      await api.uiUiAppForeground({
        client: this.client,
        path: { udid: this.udid },
        query: uiQuery(opts),
      }),
    );
  }
}

/**
 * Device-scoped facade, reachable via `client.device(udid)`. Groups unary device
 * operations plus the `apps`, `wda`, `files`, `fsync`, `crashes`, `media`,
 * `settings`, `mdm`, `proxy`, `webinspector`, `ui` and `jobs` sub-facades, the
 * SSE streams and the binary streams.
 */
export class DeviceHandle {
  /** App-management operations for this device. */
  readonly apps: AppsHandle;
  /** WebDriverAgent session operations for this device. */
  readonly wda: WdaHandle;
  /** On-device file-service operations for this device (domain-based). */
  readonly files: FilesHandle;
  /** AFC file-sync operations for this device (`fsync`: ls/tree/pull/push/rm/mkdir). */
  readonly fsync: FsyncHandle;
  /** WebInspector (Safari/WKWebView) operations for this device. */
  readonly webinspector: WebInspectorHandle;
  /** UI-automation (tap/swipe/screenshot/app lifecycle) for this device. */
  readonly ui: UiHandle;
  /** Crash-report operations for this device. */
  readonly crashes: CrashesHandle;
  /** Wallpaper / icon-layout / pasteboard operations for this device. */
  readonly media: MediaHandle;
  /** AssistiveTouch / time format / wifi settings for this device. */
  readonly settings: SettingsHandle;
  /** Supervised MDM operations for this device. */
  readonly mdm: MdmHandle;
  /** Global HTTP-proxy operations for this device. */
  readonly proxy: ProxyHandle;
  /** Long-running job operations (runtest, runwda, forward) for this device. */
  readonly jobs: JobsHandle;

  constructor(
    private readonly client: Client,
    /** The device udid this handle is bound to. */
    readonly udid: string,
  ) {
    this.apps = new AppsHandle(client, udid);
    this.wda = new WdaHandle(client, udid);
    this.files = new FilesHandle(client, udid);
    this.fsync = new FsyncHandle(client, udid);
    this.webinspector = new WebInspectorHandle(client, udid);
    this.ui = new UiHandle(client, udid);
    this.crashes = new CrashesHandle(client, udid);
    this.media = new MediaHandle(client, udid);
    this.settings = new SettingsHandle(client, udid);
    this.mdm = new MdmHandle(client, udid);
    this.proxy = new ProxyHandle(client, udid);
    this.jobs = new JobsHandle(client, udid);
  }

  // ---- Device info --------------------------------------------------------

  /** Get lockdown values plus `instruments:*` keys (open dictionary). */
  async info(): Promise<DeviceInfo> {
    return unwrap(
      await api.devicesGetInfo({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Get the device name. */
  async deviceName(): Promise<string> {
    const { devicename } = await unwrap(
      await api.devicesGetDeviceName({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
    return devicename;
  }

  /** Get the device clock. */
  async date(): Promise<DeviceDate> {
    return unwrap(
      await api.devicesGetDeviceDate({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Get battery diagnostics. */
  async battery(): Promise<BatteryInfo> {
    return unwrap(
      await api.devicesGetBattery({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** List all IORegistry/diagnostic values (open map). */
  async diagnostics(): Promise<Diagnostics> {
    return unwrap(
      await api.devicesGetDiagnostics({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Query one or more MobileGestalt keys (open map). */
  async mobileGestalt(keys: string[]): Promise<MobileGestalt> {
    return unwrap(
      await api.devicesGetMobileGestalt({
        client: this.client,
        path: { udid: this.udid },
        query: { key: keys },
      }),
    );
  }

  /** List running processes. */
  async processes(): Promise<ProcessInfo[]> {
    return unwrap(
      await api.devicesGetProcesses({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Get lockdown values (open map), optionally scoped to a `domain`. */
  async lockdown(domain?: string): Promise<LockdownValues> {
    return unwrap(
      await api.devicesGetLockdownValues({
        client: this.client,
        path: { udid: this.udid },
        ...(domain !== undefined ? { query: { domain } } : {}),
      }),
    );
  }

  // ---- Diagnostics / network ---------------------------------------------

  /** Get free/total disk-space info for the device. */
  async diskSpace(): Promise<DiskSpaceInfo> {
    return unwrap(
      await api.diagnosticsNetGetDiskSpace({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Get the device's network interface / IP info. */
  async ip(): Promise<NetworkInfo> {
    return unwrap(
      await api.diagnosticsNetGetDeviceIp({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** List RemoteServiceDiscovery (RSD) services exposed over the tunnel. */
  async rsd(): Promise<RsdServices> {
    return unwrap(
      await api.diagnosticsNetGetRsdServices({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Get the IORegistry battery registry (detailed battery telemetry). */
  async batteryRegistry(): Promise<BatteryRegistry> {
    return unwrap(
      await api.diagnosticsNetGetBatteryRegistry({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Get the device cloud-configuration (MDM enrollment) info. */
  async cloudConfig(): Promise<CloudConfig> {
    return unwrap(
      await api.fsyncGetCloudConfig({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  // ---- Accessibility ------------------------------------------------------

  /** Get the accessibility element snapshot (the AX tree). */
  async ax(): Promise<AxElement> {
    return unwrap(
      await api.accessibilityGetAxSnapshot({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Run an accessibility audit; resolves to the list of issues found. */
  async axAudit(opts: AxAuditOptions = {}): Promise<AxAuditIssue[]> {
    return unwrap(
      await api.accessibilityRunAxAudit({
        client: this.client,
        path: { udid: this.udid },
        ...(opts.timeout !== undefined ? { query: { timeout: opts.timeout } } : {}),
      }),
    );
  }

  /** Get the VoiceOver state. */
  async voiceOver(): Promise<VoiceOverState> {
    return unwrap(
      await api.accessibilityGetVoiceOver({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Enable/disable VoiceOver. Returns the resulting state. */
  async setVoiceOver(enabled: boolean): Promise<VoiceOverState> {
    return unwrap(
      await api.accessibilitySetVoiceOver({
        client: this.client,
        path: { udid: this.udid },
        query: { enabled },
      }),
    );
  }

  /** Get the Zoom (accessibility zoom) state. */
  async zoom(): Promise<ZoomTouchState> {
    return unwrap(
      await api.accessibilityGetZoomTouch({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Enable/disable Zoom. Returns the resulting state. */
  async setZoom(enabled: boolean): Promise<ZoomTouchState> {
    return unwrap(
      await api.accessibilitySetZoomTouch({
        client: this.client,
        path: { udid: this.udid },
        query: { enabled },
      }),
    );
  }

  /** Replay a GPX file as simulated GPS movement (multipart upload). */
  async setLocationGpx(gpx: BinaryInput): Promise<GenericResponse> {
    return unwrap(
      await api.accessibilitySetLocationGpx({
        client: this.client,
        path: { udid: this.udid },
        body: { gpx: toBlob(gpx) },
      }),
    );
  }

  // ---- Preparation / binary streams --------------------------------------

  /** Prepare (supervise/configure) the device via a multipart upload. */
  async prepare(opts: PrepareOptions = {}): Promise<PrepareResult> {
    return unwrap(
      await api.preparePrepareDevice({
        client: this.client,
        path: { udid: this.udid },
        body: {
          ...(opts.cert !== undefined ? { cert: toBlob(opts.cert) } : {}),
          ...(opts.p12password !== undefined ? { p12password: opts.p12password } : {}),
          ...(opts.skip !== undefined ? { skip: opts.skip } : {}),
          ...(opts.orgname !== undefined ? { orgname: opts.orgname } : {}),
          ...(opts.locale !== undefined ? { locale: opts.locale } : {}),
          ...(opts.lang !== undefined ? { lang: opts.lang } : {}),
        },
      }),
    );
  }

  /**
   * Open a live MJPEG screenshot stream as a raw byte stream. Returns a
   * {@link BinaryStream} (not SSE) — iterate the JPEG frames with `for await`,
   * `break` or pass `signal` to stop.
   */
  screenshotStream(opts: ScreenshotStreamOptions = {}): Promise<BinaryStream> {
    return openBinaryStream(
      api.streamsScreenshotStream({
        client: this.client,
        path: { udid: this.udid },
        ...(opts.quality !== undefined ? { query: { quality: opts.quality } } : {}),
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /**
   * Capture a live packet trace as a raw pcap byte stream. Returns a
   * {@link BinaryStream} (not SSE) — pipe `.body` to a `.pcap` file or iterate
   * the chunks with `for await`; `break` or pass `signal` to stop.
   */
  pcap(opts: PcapOptions = {}): Promise<BinaryStream> {
    return openBinaryStream(
      api.streamsPcap({
        client: this.client,
        path: { udid: this.udid },
        ...(opts.timeout !== undefined ? { query: { timeout: opts.timeout } } : {}),
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /** Take a PNG screenshot; resolves to the raw image bytes as a `Blob`. */
  async screenshot(): Promise<Blob> {
    const result = await api.devicesScreenshot({
      client: this.client,
      path: { udid: this.udid },
      parseAs: "blob",
    });
    return unwrap(result as { data?: Blob; error?: unknown; response: Response });
  }

  // ---- Device management --------------------------------------------------

  /** Reboot the device. */
  async reboot(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesReboot({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Shut down the device. */
  async shutdown(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesShutdown({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Erase all content and settings (destructive; requires `confirm=true`). */
  async erase(confirm: boolean): Promise<GenericResponse> {
    return unwrap(
      await api.devicesErase({
        client: this.client,
        path: { udid: this.udid },
        query: { confirm },
      }),
    );
  }

  /** Get developer mode state. */
  async devmode(): Promise<DevModeState> {
    return unwrap(
      await api.devicesGetDevMode({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /**
   * Enable or reveal developer mode. `action` is `enable` or `reveal`; set
   * `enablePostRestart` to arm developer mode across the next reboot.
   */
  async setDevmode(
    action: string,
    enablePostRestart?: boolean,
  ): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetDevMode({
        client: this.client,
        path: { udid: this.udid },
        body: {
          action,
          ...(enablePostRestart !== undefined ? { enablePostRestart } : {}),
        },
      }),
    );
  }

  /** Get the device language/locale configuration. */
  async lang(): Promise<LanguageConfiguration> {
    return unwrap(
      await api.devicesGetLanguage({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Set the device language and/or locale. Returns the resulting config. */
  async setLang(
    language?: string,
    locale?: string,
  ): Promise<LanguageConfiguration> {
    return unwrap(
      await api.devicesSetLanguage({
        client: this.client,
        path: { udid: this.udid },
        body: {
          ...(language !== undefined ? { language } : {}),
          ...(locale !== undefined ? { locale } : {}),
        },
      }),
    );
  }

  /** Waive the memory limit for a process. */
  async memlimitoff(process: string): Promise<MemLimitResult> {
    return unwrap(
      await api.devicesMemLimitOff({
        client: this.client,
        path: { udid: this.udid },
        body: { process },
      }),
    );
  }

  /** Activate the device (complete Setup Assistant / activation). */
  async activate(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesActivate({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Pair with the device (optionally supervised). */
  async pair(opts: PairOptions): Promise<GenericResponse> {
    const headers: Record<string, string> = {};
    if (opts.supervisionPassword !== undefined) {
      headers["Supervision-Password"] = opts.supervisionPassword;
    }
    return unwrap(
      await api.devicesPair({
        client: this.client,
        path: { udid: this.udid },
        query: { supervised: opts.supervised },
        ...(opts.p12file ? { body: { p12file: toBlob(opts.p12file) } } : {}),
        ...(Object.keys(headers).length ? { headers } : {}),
      }),
    );
  }

  /** Reset accessibility settings on the device. */
  async resetAccessibility(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesResetAccessibility({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  // ---- Location -----------------------------------------------------------

  /** Clear the simulated GPS location. */
  async resetLocation(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesResetLocation({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Simulate a GPS location (decimal degrees). */
  async setLocation(latitude: number, longitude: number): Promise<GenericResponse> {
    return unwrap(
      await api.devicesSetLocation({
        client: this.client,
        path: { udid: this.udid },
        query: { latitude: String(latitude), longitude: String(longitude) },
      }),
    );
  }

  // ---- Conditions ---------------------------------------------------------

  /** List available condition inducer profile types. */
  async conditions(): Promise<ProfileType[]> {
    return unwrap(
      await api.devicesListConditions({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Enable a condition inducer profile. */
  async enableCondition(
    profileTypeId: string,
    profileId: string,
  ): Promise<GenericResponse> {
    return unwrap(
      await api.devicesEnableCondition({
        client: this.client,
        path: { udid: this.udid },
        query: { profileTypeID: profileTypeId, profileID: profileId },
      }),
    );
  }

  /** Disable the currently active condition inducer profile. */
  async disableCondition(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesDisableCondition({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  // ---- Developer disk images ----------------------------------------------

  /** List hex signatures of Developer Disk Images mounted on the device. */
  async images(): Promise<string[]> {
    return unwrap(
      await api.devicesListImages({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** List mounted developer image signatures (structured). */
  async mountedImages(): Promise<MountedImages> {
    return unwrap(
      await api.devicesListMountedImages({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  /** Mount a Developer Disk Image (auto-resolve or upload raw bytes). */
  async installImage(opts: InstallImageOptions = {}): Promise<GenericResponse> {
    const query: { auto?: boolean; basedir?: string } = {};
    if (opts.auto !== undefined) query.auto = opts.auto;
    if (opts.basedir !== undefined) query.basedir = opts.basedir;
    return unwrap(
      await api.devicesMountImage({
        client: this.client,
        path: { udid: this.udid },
        ...(Object.keys(query).length ? { query } : {}),
        ...(opts.image ? { body: toBlob(opts.image) } : {}),
      }),
    );
  }

  /** Unmount the developer disk image. */
  async unmountImage(): Promise<GenericResponse> {
    return unwrap(
      await api.devicesUnmountImage({
        client: this.client,
        path: { udid: this.udid },
      }),
    );
  }

  // ---- Configuration profiles ---------------------------------------------

  /** List installed configuration profiles (open dictionary). */
  async profiles(): Promise<InstalledProfiles> {
    return unwrap(
      await api.devicesGetProfiles({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Install a configuration profile (optionally supervised). */
  async addProfile(opts: AddProfileOptions): Promise<GenericResponse> {
    return unwrap(
      await api.devicesAddProfile({
        client: this.client,
        path: { udid: this.udid },
        body: {
          profile: toBlob(opts.profile),
          ...(opts.p12 ? { p12: toBlob(opts.p12) } : {}),
          ...(opts.password !== undefined ? { password: opts.password } : {}),
        },
      }),
    );
  }

  /** Remove a configuration profile by identifier. */
  async removeProfile(name: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesRemoveProfile({
        client: this.client,
        path: { udid: this.udid, name },
      }),
    );
  }

  // ---- Streaming (SSE) ----------------------------------------------------

  /** Stream syslog lines. Yields `{ event: "syslog", data: SyslogMessage }`. */
  syslog(opts: StreamOptions = {}): AsyncIterable<SseEvent<SyslogEventMap>> {
    return this.stream<SyslogEventMap>(
      api.devicesStreamSyslog({
        client: this.client,
        path: { udid: this.udid },
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /** Stream app-state notifications. Yields `{ event: "appstate", data }`. */
  notifications(
    opts: StreamOptions = {},
  ): AsyncIterable<SseEvent<NotificationEventMap>> {
    return this.stream<NotificationEventMap>(
      api.devicesStreamNotifications({
        client: this.client,
        path: { udid: this.udid },
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /** Stream os_log trace entries. Yields `{ event: "ostrace", data }`. */
  ostrace(
    filters: OsTraceFilters = {},
    opts: StreamOptions = {},
  ): AsyncIterable<SseEvent<OsTraceEventMap>> {
    return this.stream<OsTraceEventMap>(
      api.devicesStreamOsTrace({
        client: this.client,
        path: { udid: this.udid },
        query: filters,
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /** Stream device attach/detach events. Yields `{ event: "attachdetach", data }`. */
  listen(opts: StreamOptions = {}): AsyncIterable<SseEvent<ListenEventMap>> {
    return this.stream<ListenEventMap>(
      api.devicesStreamListen({
        client: this.client,
        path: { udid: this.udid },
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /** Stream CPU-usage samples. Yields `{ event: "sample", data: CpuUsageSample }`. */
  sysmontap(opts: StreamOptions = {}): AsyncIterable<SseEvent<SysmontapEventMap>> {
    return this.stream<SysmontapEventMap>(
      api.devicesStreamSysmontap({
        client: this.client,
        path: { udid: this.udid },
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /**
   * Shared driver for the SSE endpoints: awaits the streaming request, checks
   * the HTTP status (surfacing errors as {@link IosApiError}), then hands the
   * `text/event-stream` body to the typed frame parser.
   */
  private stream<TMap>(
    request: Promise<{ data?: unknown; response: Response }>,
    signal?: AbortSignal,
  ): AsyncGenerator<SseEvent<TMap>, void, unknown> {
    return streamSse<TMap>(request, signal);
  }
}

/**
 * Job-management sub-facade, reachable via `client.device(udid).jobs`.
 * Long-running operations (test runs, WDA runner, port forwards) are started
 * per-device and then tracked / streamed by job id.
 */
export class JobsHandle {
  constructor(
    private readonly client: Client,
    private readonly udid: string,
  ) {}

  /** Start an XCUITest/unit-test run as an async job. Returns the created job. */
  async runtest(config: RunTestRequest): Promise<Job> {
    return unwrap(
      await api.devicesStartRunTest({
        client: this.client,
        path: { udid: this.udid },
        body: config,
      }),
    );
  }

  /** Start the WebDriverAgent runner as an async job. */
  async runwda(config: RunTestRequest = {}): Promise<Job> {
    return unwrap(
      await api.devicesStartRunWda({
        client: this.client,
        path: { udid: this.udid },
        body: config,
      }),
    );
  }

  /** Start a TCP port forward (host→device) as an async job. */
  async forward(config: ForwardRequest): Promise<Job> {
    return unwrap(
      await api.devicesStartForward({
        client: this.client,
        path: { udid: this.udid },
        body: config,
      }),
    );
  }

  /** List jobs for this device. */
  async list(): Promise<Job[]> {
    return unwrap(
      await api.devicesListJobs({ client: this.client, path: { udid: this.udid } }),
    );
  }

  /** Get a job's status by id. */
  async get(id: string): Promise<Job> {
    return unwrap(
      await api.devicesGetJob({ client: this.client, path: { udid: this.udid, id } }),
    );
  }

  /**
   * Stream a job's log output: the buffered history first, then live lines until
   * the job ends. Yields `{ event: "log", data: JobLogLine }`.
   */
  logs(
    id: string,
    opts: StreamOptions = {},
  ): AsyncIterable<SseEvent<JobLogEventMap>> {
    return streamSse<JobLogEventMap>(
      api.devicesStreamJobLogs({
        client: this.client,
        path: { udid: this.udid, id },
        parseAs: "stream",
        signal: opts.signal,
      }),
      opts.signal,
    );
  }

  /** Stop a running job, or purge an already-terminal one from the registry. */
  async delete(id: string): Promise<GenericResponse> {
    return unwrap(
      await api.devicesStopJob({ client: this.client, path: { udid: this.udid, id } }),
    );
  }
}

/**
 * Tunnel-management facade, reachable via `client.tunnels`.
 */
export class TunnelsHandle {
  constructor(private readonly client: Client) {}

  /** List running device tunnels. */
  async list(): Promise<Tunnel[]> {
    return unwrap(await api.listTunnels({ client: this.client }));
  }

  /** Stop the tunnel for a device. */
  async delete(udid: string): Promise<TunnelStopped> {
    return unwrap(
      await api.stopTunnel({ client: this.client, path: { udid } }),
    );
  }

  /** Restart the tunnel for a device and wait for it to come up. */
  async refresh(udid: string): Promise<Tunnel> {
    return unwrap(
      await api.refreshTunnel({ client: this.client, path: { udid } }),
    );
  }

  /** Shut down the tunnel agent entirely. */
  async shutdownAgent(): Promise<AgentShutdown> {
    return unwrap(await api.shutdownTunnelAgent({ client: this.client }));
  }
}

/**
 * Host-level code-signing sub-facade, reachable via `client.sign`. These are
 * device-free operations that talk to Apple's App Store Connect API to mint a
 * signing certificate / provisioning profile, or to resign an app.
 */
export class SignHandle {
  constructor(private readonly client: Client) {}

  /**
   * Mint an iOS Development signing certificate; resolves to the certificate
   * bytes as a PKCS#12 (`.p12`) `Uint8Array`.
   */
  async certificate(creds: AscCredentials): Promise<Uint8Array> {
    const result = await api.signCertificate({
      client: this.client,
      body: ascBody(creds),
      parseAs: "blob",
    });
    const blob = unwrap(
      result as { data?: Blob; error?: unknown; response: Response },
    );
    return new Uint8Array(await blob.arrayBuffer());
  }

  /** Create a provisioning profile for a bundle id + device. */
  async provision(opts: SignProvisionOptions): Promise<ProvisioningResult> {
    return unwrap(
      await api.signProvision({
        client: this.client,
        body: {
          ...ascBody(opts),
          bundleid: opts.bundleId,
          udid: opts.udid,
          ...(opts.bundleName !== undefined ? { bundlename: opts.bundleName } : {}),
          ...(opts.profileName !== undefined ? { profilename: opts.profileName } : {}),
          ...(opts.deviceName !== undefined ? { devicename: opts.deviceName } : {}),
          ...(opts.certificateId !== undefined
            ? { "certificate-id": opts.certificateId }
            : {}),
        },
      }),
    );
  }

  /** Resign an app/`.ipa`; resolves to the resigned `.ipa` bytes. */
  async app(opts: SignAppOptions): Promise<Uint8Array> {
    const result = await api.signApp({
      client: this.client,
      body: {
        ipa: toBlob(opts.ipa),
        p12file: toBlob(opts.p12),
        profile: toBlob(opts.profile),
        ...(opts.p12password !== undefined ? { p12password: opts.p12password } : {}),
        ...(opts.bundleId !== undefined ? { bundleid: opts.bundleId } : {}),
      },
      parseAs: "blob",
    });
    const blob = unwrap(
      result as { data?: Blob; error?: unknown; response: Response },
    );
    return new Uint8Array(await blob.arrayBuffer());
  }
}

/**
 * Host-level device-preparation sub-facade, reachable via `client.prepare`.
 * (The per-device `prepare` step itself is `client.device(udid).prepare()`.)
 */
export class PrepareHandle {
  constructor(private readonly client: Client) {}

  /** Generate a self-signed supervision identity (certificate + private key). */
  async createCert(): Promise<SupervisionCert> {
    return unwrap(await api.prepareCreateCert({ client: this.client }));
  }

  /** List the setup panes that can be skipped by {@link DeviceHandle.prepare}. */
  async skipOptions(): Promise<PrepareSkipOptions> {
    return unwrap(await api.getPrepareSkipOptions({ client: this.client }));
  }
}

/**
 * Top-level go-ios SDK client.
 *
 * @example
 * ```ts
 * const client = new IosClient({ baseUrl: "http://localhost:60105", apiKey: "secret" });
 * const { deviceList } = await client.devices.list();
 * const udid = deviceList[0].properties.serialNumber;
 * for await (const ev of client.device(udid).syslog()) {
 *   if (ev.event === "syslog") console.log(ev.data.message);
 * }
 * ```
 */
export class IosClient {
  private readonly client: Client;

  /** Host-level operations (currently device listing). */
  readonly devices: {
    /** List all attached / reachable devices. */
    list(): Promise<DeviceList>;
  };

  /** Device-tunnel operations. */
  readonly tunnels: TunnelsHandle;

  /** Host-level code-signing operations (certificate / provision / resign). */
  readonly sign: SignHandle;

  /** Host-level device-preparation helpers (create-cert / skip-options). */
  readonly prepare: PrepareHandle;

  constructor(options: IosClientOptions) {
    this.client = createClient(
      createConfig({
        baseUrl: options.baseUrl,
        ...(options.fetch ? { fetch: options.fetch } : {}),
        // The `bearer` security scheme reads this to set the Authorization header.
        ...(options.apiKey ? { auth: options.apiKey } : {}),
        throwOnError: false,
      }),
    );

    this.devices = {
      list: async (): Promise<DeviceList> =>
        unwrap(await api.listDevices({ client: this.client })),
    };
    this.tunnels = new TunnelsHandle(this.client);
    this.sign = new SignHandle(this.client);
    this.prepare = new PrepareHandle(this.client);
  }

  /** Get a device-scoped facade bound to `udid`. */
  device(udid: string): DeviceHandle {
    return new DeviceHandle(this.client, udid);
  }
}

/**
 * Convenience accessor for the udid of a device list entry: the serial number
 * is what every device-scoped route keys on.
 */
export function deviceUdid(entry: {
  properties: { serialNumber: string };
}): string {
  return entry.properties.serialNumber;
}

/**
 * Shared driver for every SSE endpoint: awaits the streaming request, surfaces a
 * non-2xx status as an {@link IosApiError}, then hands the `text/event-stream`
 * body to the typed frame parser.
 */
async function* streamSse<TMap>(
  request: Promise<{ data?: unknown; response: Response }>,
  signal?: AbortSignal,
): AsyncGenerator<SseEvent<TMap>, void, unknown> {
  const { response } = await request;
  if (!response.ok) {
    let body: unknown;
    try {
      body = await response.json();
    } catch {
      /* non-JSON error body */
    }
    const envelope = body as Partial<GenericResponse> | undefined;
    const message =
      envelope?.error ??
      envelope?.message ??
      `go-ios stream failed with status ${response.status}`;
    throw new IosApiError(response.status, String(message), body);
  }
  if (!response.body) return;
  yield* parseSseStream<TMap>(response.body, { signal });
}

/** Build the optional `identifier` query fragment for file-service calls. */
function identifierQuery(scope: FileScope): { identifier?: string } {
  return scope.identifier !== undefined ? { identifier: scope.identifier } : {};
}

/** Build the optional `bundleID` query fragment for fsync calls. */
function bundleIdQuery(opts: { bundleId?: string }): { bundleID?: string } {
  return opts.bundleId !== undefined ? { bundleID: opts.bundleId } : {};
}

/** Build the shared `backend`/`wdaUrl`/`timeout` query fragment for UI calls. */
function uiQuery(
  opts: UiOptions,
): { backend?: string; wdaUrl?: string; timeout?: number } {
  return {
    ...(opts.backend !== undefined ? { backend: opts.backend } : {}),
    ...(opts.wdaUrl !== undefined ? { wdaUrl: opts.wdaUrl } : {}),
    ...(opts.timeout !== undefined ? { timeout: opts.timeout } : {}),
  };
}

/** Build the shared App Store Connect multipart body for signing calls. */
function ascBody(creds: AscCredentials): {
  "asc-private-key": Blob;
  "asc-key-id": string;
  "asc-issuer-id": string;
  "revoke-existing"?: string;
  p12password?: string;
} {
  return {
    "asc-private-key": toBlob(creds.ascPrivateKey),
    "asc-key-id": creds.ascKeyId,
    "asc-issuer-id": creds.ascIssuerId,
    ...(creds.revokeExisting !== undefined
      ? { "revoke-existing": String(creds.revokeExisting) }
      : {}),
    ...(creds.p12password !== undefined ? { p12password: creds.p12password } : {}),
  };
}

/** Build the shared multipart identity body for supervised MDM calls. */
function identityBody(identity: SupervisionIdentity): {
  p12: Blob;
  password?: string;
} {
  return {
    p12: toBlob(identity.p12),
    ...(identity.password !== undefined ? { password: identity.password } : {}),
  };
}

/** Coerce supported binary inputs into a `Blob` for multipart/raw bodies. */
function toBlob(input: BinaryInput): Blob {
  if (input instanceof Blob) return input;
  return new Blob([input as BlobPart]);
}
