import { createClient, createConfig, type Client } from "@hey-api/client-fetch";

import * as api from "./generated/sdk.gen";
import type {
  AgentShutdown,
  AppInfo,
  AssistiveTouchState,
  BatteryInfo,
  CrashListing,
  DeviceDate,
  DeviceInfo,
  DeviceList,
  DevModeState,
  Diagnostics,
  FileDomain,
  FileListing,
  FilePushResult,
  ForwardRequest,
  GenericResponse,
  IconLayout,
  InstalledProfiles,
  Job,
  LanguageConfiguration,
  LockdownValues,
  MemLimitResult,
  MobileGestalt,
  MountedImages,
  PasteboardContent,
  ProcessInfo,
  ProfileType,
  RunTestRequest,
  SecurityInfo,
  StatusOk,
  TimeFormatState,
  Tunnel,
  TunnelStopped,
  UnlockToken,
  WdaConfig,
  WdaSession,
} from "./generated/types.gen";
import { IosApiError, unwrap } from "./errors";
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
 * Device-scoped facade, reachable via `client.device(udid)`. Groups unary device
 * operations plus the `apps`, `wda`, `files`, `crashes`, `media`, `settings`,
 * `mdm`, `proxy` and `jobs` sub-facades and the SSE streams.
 */
export class DeviceHandle {
  /** App-management operations for this device. */
  readonly apps: AppsHandle;
  /** WebDriverAgent session operations for this device. */
  readonly wda: WdaHandle;
  /** On-device file-service operations for this device. */
  readonly files: FilesHandle;
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

  /** Get all lockdown values (open map). */
  async lockdown(): Promise<LockdownValues> {
    return unwrap(
      await api.devicesGetLockdownValues({
        client: this.client,
        path: { udid: this.udid },
      }),
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
