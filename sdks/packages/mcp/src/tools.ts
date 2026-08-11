/**
 * Curated MCP tool set for go-ios.
 *
 * This is a hand-picked, high-signal set of tools that proxy the go-ios REST
 * API — NOT a naive 1:1 OpenAPI→tool mapping (which produces too many
 * low-quality tools for LLMs). Each tool has an LLM-oriented description (what
 * it does, when to use it, its args, and what it returns), a typed zod input
 * schema, structured output, and explicit error surfacing.
 */
import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { GoIosClient, GoIosApiError } from "./client.js";
import { captureSse } from "./sse.js";

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

const udidArg = z
  .string()
  .min(1)
  .describe(
    "The device udid (serial number), as returned by list_devices. Device-scoped routes key on this.",
  );

const bundleIdArg = z
  .string()
  .min(1)
  .describe("The application bundle identifier, e.g. com.apple.Preferences.");

/** Encode a udid safely into a path segment. */
function seg(v: string): string {
  return encodeURIComponent(v);
}

/** Wrap a handler so any GoIosApiError becomes a clean MCP error result. */
type ToolResult = {
  content: Array<
    | { type: "text"; text: string }
    | { type: "image"; data: string; mimeType: string }
  >;
  structuredContent?: Record<string, unknown>;
  isError?: boolean;
};

async function guard(fn: () => Promise<ToolResult>): Promise<ToolResult> {
  try {
    return await fn();
  } catch (err) {
    if (err instanceof GoIosApiError) {
      return {
        isError: true,
        content: [{ type: "text", text: err.friendly }],
        structuredContent: {
          error: err.message,
          status: err.status,
          hint: err.friendly,
        },
      };
    }
    const message = err instanceof Error ? err.message : String(err);
    return { isError: true, content: [{ type: "text", text: message }] };
  }
}

/**
 * Open an SSE stream against the daemon, throwing a GoIosApiError for non-2xx.
 * Shared by every bounded-capture tool (stream_logs, sample_performance,
 * tail_job_logs) so they all get identical auth + error surfacing.
 */
async function openSse(
  client: GoIosClient,
  url: string,
  signal: AbortSignal,
): Promise<Response> {
  const res = await fetch(url, {
    headers: client.headers({ Accept: "text/event-stream" }),
    signal,
  });
  if (!res.ok) {
    throw new GoIosApiError({
      status: res.status,
      statusText: res.statusText,
      url,
      message: res.statusText,
    });
  }
  return res;
}

/** Build a text+structured result from a JSON value. */
function jsonResult(value: unknown): ToolResult {
  const structured =
    value !== null && typeof value === "object" && !Array.isArray(value)
      ? (value as Record<string, unknown>)
      : { result: value };
  return {
    content: [{ type: "text", text: JSON.stringify(value, null, 2) }],
    structuredContent: structured,
  };
}

// ---------------------------------------------------------------------------
// Domain response types (subset of the OpenAPI schemas we consume)
// ---------------------------------------------------------------------------

interface DeviceEntry {
  deviceID: number;
  properties: { serialNumber: string; connectionType?: string; [k: string]: unknown };
  address?: string;
  [k: string]: unknown;
}
interface DeviceList {
  deviceList: DeviceEntry[];
}
type DeviceInfo = Record<string, unknown>;
type AppInfo = Record<string, unknown>;
interface WdaSession {
  sessionId: string;
  udid: string;
  config: unknown;
}
type BatteryInfo = Record<string, unknown>;
type Diagnostics = Record<string, unknown>;
interface CrashListing {
  files: string[];
  count: number;
}
interface FileEntry {
  name?: string;
  path?: string;
  isDir?: boolean;
  size?: number;
}
interface FileListing {
  path: string;
  files: FileEntry[];
  count: number;
}
interface ProcessInfo {
  pid: number;
  name: string;
  realAppName?: string;
  isApplication?: boolean;
  startDate?: string;
}
interface PasteboardContent {
  present: boolean;
  text: string;
}
interface Job {
  id: string;
  kind: string;
  udid: string;
  status: string;
  startedAt: string;
  finishedAt?: string;
  error?: string;
  result?: unknown;
}

/** Bounded cap for pulled file/crash-report text (256 KiB). */
const MAX_PULL_BYTES = 256 * 1024;

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

/** The canonical, ordered list of curated tool names (used by tests + docs). */
export const CURATED_TOOL_NAMES = [
  // Discovery & info
  "list_devices",
  "device_info",
  "device_health",
  // Apps
  "list_apps",
  "launch_app",
  "terminate_app",
  "install_app",
  "uninstall_app",
  // Screen & logs
  "screenshot",
  "stream_logs",
  // WebDriverAgent session lifecycle
  "create_wda_session",
  "read_wda_session",
  "delete_wda_session",
  // Device management (disruptive)
  "reboot_device",
  "shutdown_device",
  // Diagnostics & health
  "device_battery",
  "list_processes",
  "device_diagnostics",
  // Crash logs
  "list_crash_reports",
  "pull_crash_report",
  // Files (read-only)
  "list_files",
  // Performance
  "sample_performance",
  // Jobs (drive/observe test runs)
  "run_wda",
  "get_job",
  "list_jobs",
  "stop_job",
  "tail_job_logs",
  // Pasteboard
  "get_pasteboard",
  "set_pasteboard",
] as const;

export function registerTools(server: McpServer, client: GoIosClient): void {
  // ---- Discovery ---------------------------------------------------------

  server.registerTool(
    "list_devices",
    {
      title: "List connected iOS devices",
      description:
        "List all iOS devices currently attached to or reachable by the go-ios daemon. " +
        "Use this FIRST to discover the udid you need for every other device-scoped tool. " +
        "Returns each device's udid (serialNumber), connection type, and network address if any. " +
        "Takes no arguments.",
      inputSchema: {},
    },
    async () =>
      guard(async () => {
        const data = await client.getJson<DeviceList>("/list");
        const devices = (data.deviceList ?? []).map((d) => ({
          udid: d.properties?.serialNumber,
          deviceID: d.deviceID,
          connectionType: d.properties?.connectionType,
          address: d.address,
        }));
        return jsonResult({ count: devices.length, devices });
      }),
  );

  server.registerTool(
    "device_info",
    {
      title: "Get device info",
      description:
        "Get lockdown values plus hardware/network/instruments info for one device. " +
        "Use this when you need details like product type, iOS version, name, or capabilities. " +
        "Returns an open dictionary of heterogeneous values (keys vary by device). " +
        "Requires a udid from list_devices.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const info = await client.getJson<DeviceInfo>(`/device/${seg(udid)}/info`);
        return jsonResult(info);
      }),
  );

  server.registerTool(
    "device_health",
    {
      title: "Check device health",
      description:
        "Quick reachability + health check for one device. Confirms the daemon can talk to " +
        "the device and returns a compact summary (reachable flag plus key identity fields such " +
        "as ProductType, ProductVersion, DeviceName when available) enriched with a battery " +
        "snapshot (charge level, charging/plugged-in, temperature) when the device exposes it. " +
        "Use this before a longer workflow to fail fast if the device is offline or unauthorized. " +
        "Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        try {
          const info = await client.getJson<DeviceInfo>(`/device/${seg(udid)}/info`);
          const pick = (k: string) => info[k];
          // Battery is a cheap, high-signal add; never let it fail the health check.
          let battery: Record<string, unknown> | undefined;
          try {
            const b = await client.getJson<BatteryInfo>(`/device/${seg(udid)}/battery`);
            battery = {
              currentCapacity: b["CurrentCapacity"],
              isCharging: b["IsCharging"],
              externalConnected: b["ExternalConnected"],
              fullyCharged: b["FullyCharged"],
              temperature: b["Temperature"],
            };
          } catch {
            /* battery unavailable — omit it, device is still reachable */
          }
          return jsonResult({
            udid,
            reachable: true,
            deviceName: pick("DeviceName"),
            productType: pick("ProductType"),
            productVersion: pick("ProductVersion"),
            buildVersion: pick("BuildVersion"),
            battery,
          });
        } catch (err) {
          if (err instanceof GoIosApiError) {
            return {
              content: [
                { type: "text", text: `Device ${udid} is not healthy: ${err.friendly}` },
              ],
              structuredContent: {
                udid,
                reachable: false,
                status: err.status,
                error: err.message,
              },
            };
          }
          throw err;
        }
      }),
  );

  // ---- Apps --------------------------------------------------------------

  server.registerTool(
    "list_apps",
    {
      title: "List installed apps",
      description:
        "List installed applications on a device. Each entry is the app's Info.plist map " +
        "(bundle id, name, version, path). Use this to find a bundleId before launch/terminate/" +
        "uninstall. Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const apps = await client.getJson<AppInfo[]>(`/device/${seg(udid)}/apps/`);
        const summary = (apps ?? []).map((a) => ({
          bundleId: a["CFBundleIdentifier"],
          name: a["CFBundleName"] ?? a["CFBundleDisplayName"],
          version: a["CFBundleShortVersionString"],
        }));
        return jsonResult({ count: summary.length, apps: summary, raw: apps });
      }),
  );

  server.registerTool(
    "launch_app",
    {
      title: "Launch app",
      description:
        "Launch (open) an installed application by its bundle id. Use list_apps to find the " +
        "bundleId. Returns the daemon's status message. Requires a udid and bundleId.",
      inputSchema: { udid: udidArg, bundleId: bundleIdArg },
    },
    async ({ udid, bundleId }) =>
      guard(async () => {
        const res = await client.postJson<unknown>(`/device/${seg(udid)}/apps/launch`, {
          query: { bundleID: bundleId },
        });
        return jsonResult(res);
      }),
  );

  server.registerTool(
    "terminate_app",
    {
      title: "Terminate (kill) app",
      description:
        "Kill a running application by its bundle id. Use this to force-stop an app before " +
        "relaunching it. Returns the daemon's status message. Requires a udid and bundleId.",
      inputSchema: { udid: udidArg, bundleId: bundleIdArg },
    },
    async ({ udid, bundleId }) =>
      guard(async () => {
        const res = await client.postJson<unknown>(`/device/${seg(udid)}/apps/kill`, {
          query: { bundleID: bundleId },
        });
        return jsonResult(res);
      }),
  );

  server.registerTool(
    "install_app",
    {
      title: "Install app from .ipa/.app",
      description:
        "Install an application onto the device from a local .ipa or .app archive. Provide the " +
        "absolute filesystem path to the archive that is accessible to THIS MCP server process " +
        "(1 byte–200 MB); the server uploads it to the go-ios daemon as multipart/form-data. " +
        "Returns the daemon's status message. Requires a udid and ipaPath.",
      inputSchema: {
        udid: udidArg,
        ipaPath: z
          .string()
          .min(1)
          .describe(
            "Absolute path to a .ipa/.app archive readable by the MCP server process. Max 200 MB.",
          ),
      },
    },
    async ({ udid, ipaPath }) =>
      guard(async () => {
        // Read the file locally and upload as multipart. Kept out of the shared
        // client because it is the only multipart call in the curated set.
        const { readFile } = await import("node:fs/promises");
        const { basename } = await import("node:path");
        let bytes: Buffer;
        try {
          bytes = await readFile(ipaPath);
        } catch (err) {
          return {
            isError: true,
            content: [
              {
                type: "text",
                text: `Could not read archive at ${ipaPath}: ${
                  err instanceof Error ? err.message : String(err)
                }`,
              },
            ],
          };
        }
        const form = new FormData();
        form.append(
          "file",
          new Blob([new Uint8Array(bytes)], { type: "application/octet-stream" }),
          basename(ipaPath),
        );
        const url = client.url(`/device/${seg(udid)}/apps/install`);
        const res = await fetch(url, {
          method: "POST",
          headers: client.headers(),
          body: form,
        });
        const text = await res.text();
        let parsed: unknown = text;
        try {
          parsed = JSON.parse(text);
        } catch {
          /* keep text */
        }
        if (!res.ok) {
          throw new GoIosApiError({
            status: res.status,
            statusText: res.statusText,
            url,
            body: parsed,
            message:
              (parsed as { error?: string; message?: string })?.error ??
              (parsed as { message?: string })?.message ??
              text ??
              res.statusText,
          });
        }
        return jsonResult(parsed);
      }),
  );

  server.registerTool(
    "uninstall_app",
    {
      title: "Uninstall app",
      description:
        "Uninstall (remove) an application by its bundle id. Returns the daemon's status " +
        "message. Requires a udid and bundleId.",
      inputSchema: { udid: udidArg, bundleId: bundleIdArg },
    },
    async ({ udid, bundleId }) =>
      guard(async () => {
        const res = await client.postJson<unknown>(`/device/${seg(udid)}/apps/uninstall`, {
          query: { bundleID: bundleId },
        });
        return jsonResult(res);
      }),
  );

  // ---- Screenshot --------------------------------------------------------

  server.registerTool(
    "screenshot",
    {
      title: "Capture screenshot",
      description:
        "Capture the current screen of a device and return it as a PNG image. Use this to SEE " +
        "what is on screen — e.g. to verify an app launched or to read UI state. Returns an MCP " +
        "image content block (base64 PNG). Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const { bytes, contentType } = await client.getBytes(
          `/device/${seg(udid)}/screenshot`,
        );
        const base64 = Buffer.from(bytes).toString("base64");
        return {
          content: [
            {
              type: "image",
              data: base64,
              mimeType: contentType.startsWith("image/") ? contentType : "image/png",
            },
          ],
          structuredContent: { udid, bytes: bytes.length, mimeType: "image/png" },
        };
      }),
  );

  // ---- Logs (bounded SSE capture) ---------------------------------------

  server.registerTool(
    "stream_logs",
    {
      title: "Capture device logs (bounded)",
      description:
        "Capture a BOUNDED, recent window of device logs and return the collected lines. This is " +
        "NOT an infinite stream — the go-ios log endpoints stream forever, so this tool collects " +
        "events only until a limit is reached, then returns them, so an agent gets a finite result. " +
        "Choose source='syslog' for raw syslog lines, or source='ostrace' for structured os_log " +
        "entries with optional AND-combined filters (pid, level, subsystem, match substring, " +
        "exclude substring). Capture stops at whichever comes first: duration_seconds (default 5, " +
        "max 30) or max_lines (default 100, max 1000). Requires a udid.",
      inputSchema: {
        udid: udidArg,
        source: z
          .enum(["syslog", "ostrace"])
          .default("syslog")
          .describe(
            "Log source: 'syslog' (raw lines) or 'ostrace' (structured os_log, supports filters).",
          ),
        duration_seconds: z
          .number()
          .int()
          .min(1)
          .max(30)
          .default(5)
          .describe("Max seconds to collect before returning. Capped at 30."),
        max_lines: z
          .number()
          .int()
          .min(1)
          .max(1000)
          .default(100)
          .describe("Max number of log events to return. Capped at 1000."),
        pid: z
          .number()
          .int()
          .optional()
          .describe("ostrace only: only include entries from this process id."),
        process: z
          .string()
          .optional()
          .describe(
            "ostrace only: only include entries whose processName contains this string " +
              "(applied client-side after capture).",
          ),
        level: z
          .string()
          .optional()
          .describe("ostrace only: minimum log level, e.g. info, debug, error."),
        subsystem: z
          .string()
          .optional()
          .describe("ostrace only: only include entries from this subsystem."),
        match: z
          .string()
          .optional()
          .describe("ostrace only: only include entries whose message matches this substring."),
        exclude: z
          .string()
          .optional()
          .describe("ostrace only: exclude entries whose message matches this substring."),
      },
    },
    async (args) =>
      guard(async () => {
        const {
          udid,
          source,
          duration_seconds,
          max_lines,
          pid,
          process: processName,
          level,
          subsystem,
          match,
          exclude,
        } = args;

        const query: Record<string, string | number | undefined> =
          source === "ostrace"
            ? { pid, level, subsystem, match, exclude }
            : {};
        const path = source === "ostrace" ? "/ostrace" : "/syslog";
        const keepEvent = source === "ostrace" ? "ostrace" : "syslog";

        const url = client.url(`/device/${seg(udid)}${path}`, query);
        const captured = await captureSse(
          (signal) => openSse(client, url, signal),
          {
            durationMs: duration_seconds * 1000,
            maxLines: max_lines,
            keepEvents: new Set([keepEvent]),
          },
        );

        // Client-side processName filter for ostrace (not a server query param).
        let lines = captured.events.map((e) => e.payload ?? { message: e.raw });
        if (source === "ostrace" && processName) {
          lines = lines.filter((l) => {
            const pn = (l as { processName?: string }).processName ?? "";
            return pn.includes(processName);
          });
        }

        return jsonResult({
          udid,
          source,
          bounded: true,
          stoppedBy: captured.stoppedBy,
          durationSeconds: duration_seconds,
          maxLines: max_lines,
          returned: lines.length,
          totalMatched: captured.totalMatched,
          heartbeats: captured.heartbeats,
          lines,
        });
      }),
  );

  // ---- WebDriverAgent session lifecycle ---------------------------------

  server.registerTool(
    "create_wda_session",
    {
      title: "Start a WebDriverAgent session",
      description:
        "Start a WebDriverAgent (XCUITest) session on a device, which is the prerequisite for UI " +
        "automation. Provide the runner host app bundleId, the test bundleId, and the " +
        "xcTestConfig name. Returns the created session (sessionId, udid, config). Use " +
        "read_wda_session to check it and delete_wda_session to stop it. Requires a udid.",
      inputSchema: {
        udid: udidArg,
        bundleId: z
          .string()
          .min(1)
          .describe(
            "Bundle id of the WDA runner host app, e.g. com.facebook.WebDriverAgentRunner.xctrunner.",
          ),
        testBundleId: z.string().min(1).describe("Bundle id of the XCTest test bundle."),
        xcTestConfig: z
          .string()
          .min(1)
          .describe("Name/path of the .xctestconfiguration to use, e.g. WebDriverAgentRunner.xctest."),
        args: z
          .array(z.string())
          .optional()
          .describe("Extra process arguments passed to the runner."),
        env: z
          .record(z.string())
          .optional()
          .describe("Extra environment variables passed to the runner."),
      },
    },
    async ({ udid, bundleId, testBundleId, xcTestConfig, args, env }) =>
      guard(async () => {
        const session = await client.postJson<WdaSession>(
          `/device/${seg(udid)}/wda/session`,
          { body: { bundleId, testBundleId, xcTestConfig, args, env } },
        );
        return jsonResult(session);
      }),
  );

  server.registerTool(
    "read_wda_session",
    {
      title: "Get a WebDriverAgent session",
      description:
        "Fetch a running WebDriverAgent session by its sessionId. Returns the session details, " +
        "or a not-found error if the session id is unknown. Requires a udid and sessionId.",
      inputSchema: {
        udid: udidArg,
        sessionId: z.string().min(1).describe("The WDA session id from create_wda_session."),
      },
    },
    async ({ udid, sessionId }) =>
      guard(async () => {
        const session = await client.getJson<WdaSession>(
          `/device/${seg(udid)}/wda/session/${seg(sessionId)}`,
        );
        return jsonResult(session);
      }),
  );

  server.registerTool(
    "delete_wda_session",
    {
      title: "Stop a WebDriverAgent session",
      description:
        "Stop (tear down) a running WebDriverAgent session by its sessionId. Call this when UI " +
        "automation is finished to free the device. Requires a udid and sessionId.",
      inputSchema: {
        udid: udidArg,
        sessionId: z.string().min(1).describe("The WDA session id to stop."),
      },
    },
    async ({ udid, sessionId }) =>
      guard(async () => {
        const res = await client.deleteJson<unknown>(
          `/device/${seg(udid)}/wda/session/${seg(sessionId)}`,
        );
        return jsonResult(res);
      }),
  );

  // ---- Device management (disruptive) ------------------------------------

  server.registerTool(
    "reboot_device",
    {
      title: "Reboot device (DISRUPTIVE)",
      description:
        "Reboot (restart) the device. DISRUPTIVE: the device goes offline for ~30–60s and any " +
        "running app, WDA session, or job is terminated. Only call this when the user explicitly " +
        "wants a restart or a workflow requires one. Returns the daemon's status message. " +
        "Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const res = await client.postJson<unknown>(`/device/${seg(udid)}/reboot`);
        return jsonResult(res);
      }),
  );

  server.registerTool(
    "shutdown_device",
    {
      title: "Shut down device (DISRUPTIVE)",
      description:
        "Power off the device. DISRUPTIVE and hard to undo remotely: a powered-off device cannot " +
        "be turned back on over USB/network — it needs physical interaction to boot again. Only " +
        "call this when the user explicitly asks to power the device off. Returns the daemon's " +
        "status message. Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const res = await client.postJson<unknown>(`/device/${seg(udid)}/shutdown`);
        return jsonResult(res);
      }),
  );

  // ---- Diagnostics & health ----------------------------------------------

  server.registerTool(
    "device_battery",
    {
      title: "Get battery info",
      description:
        "Get the device battery snapshot: charge level (CurrentCapacity, 0–100), IsCharging, " +
        "ExternalConnected (plugged in), FullyCharged, and Temperature. Use this to check whether " +
        "a device has enough charge for a long test run or is on the charger. Returns an open map " +
        "(commonly-present keys surfaced). Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const info = await client.getJson<BatteryInfo>(`/device/${seg(udid)}/battery`);
        return jsonResult(info);
      }),
  );

  server.registerTool(
    "list_processes",
    {
      title: "List running processes",
      description:
        "List processes running on the device (pid, name, whether it is an application, start " +
        "time). Use this to confirm an app is actually running, find its pid for log filtering, or " +
        "check what else is active. Set apps_only=true to return only application processes. " +
        "Requires a udid.",
      inputSchema: {
        udid: udidArg,
        apps_only: z
          .boolean()
          .default(false)
          .describe("Return only application processes (vs all system processes)."),
      },
    },
    async ({ udid, apps_only }) =>
      guard(async () => {
        const procs = await client.getJson<ProcessInfo[]>(
          `/device/${seg(udid)}/processes`,
          apps_only ? { apps: "true" } : undefined,
        );
        return jsonResult({ count: procs?.length ?? 0, processes: procs ?? [] });
      }),
  );

  server.registerTool(
    "device_diagnostics",
    {
      title: "Get device diagnostics",
      description:
        "Get the device's IORegistry / diagnostic values (an open map of hardware sensor and " +
        "state readings — battery internals, thermal state, etc.). Heavier and less structured " +
        "than device_battery; use it when you need low-level diagnostics beyond the battery " +
        "snapshot. Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const diag = await client.getJson<Diagnostics>(`/device/${seg(udid)}/diagnostics`);
        return jsonResult(diag);
      }),
  );

  // ---- Crash logs --------------------------------------------------------

  server.registerTool(
    "list_crash_reports",
    {
      title: "List crash reports",
      description:
        "List crash report file names on the device. Optionally filter with a glob pattern " +
        "(e.g. 'MyApp*'). Use this to discover which crash reports exist, then pull_crash_report " +
        "to read one. Returns { count, files }. Requires a udid.",
      inputSchema: {
        udid: udidArg,
        pattern: z
          .string()
          .optional()
          .describe("Optional glob pattern to filter report names, e.g. 'MyApp*.ips'."),
      },
    },
    async ({ udid, pattern }) =>
      guard(async () => {
        const listing = await client.getJson<CrashListing>(
          `/device/${seg(udid)}/crashes`,
          pattern ? { pattern } : undefined,
        );
        return jsonResult(listing);
      }),
  );

  server.registerTool(
    "pull_crash_report",
    {
      title: "Read a crash report",
      description:
        "Download and return the text of a single crash report by name (as listed by " +
        "list_crash_reports). The report content is returned as text, BOUNDED to 256 KiB — larger " +
        "reports are truncated (truncated=true). Use this to read a stack trace / crash cause. " +
        "Requires a udid and the report name.",
      inputSchema: {
        udid: udidArg,
        name: z
          .string()
          .min(1)
          .describe("Crash report file name from list_crash_reports, e.g. 'MyApp-2024-...ips'."),
      },
    },
    async ({ udid, name }) =>
      guard(async () => {
        const pulled = await client.getTextBounded(
          `/device/${seg(udid)}/files/pull`,
          { domain: "crash", remote: name },
          MAX_PULL_BYTES,
        );
        return {
          content: [{ type: "text", text: pulled.text }],
          structuredContent: {
            udid,
            name,
            bytes: pulled.bytes,
            truncated: pulled.truncated,
            text: pulled.text,
          },
        };
      }),
  );

  // ---- Files (read-only) -------------------------------------------------

  server.registerTool(
    "list_files",
    {
      title: "List files on the device",
      description:
        "List a directory in one of the device's file-service domains (READ-ONLY). Choose domain: " +
        "'app' or 'app-group' (require a bundleId/group identifier), 'crash' (crash reports), or " +
        "'temp'. 'path' selects the directory within the domain (defaults to the domain root). " +
        "Returns { path, count, files:[{name, path, isDir, size}] }. This tool only LISTS; there " +
        "is deliberately no file-write tool. Requires a udid.",
      inputSchema: {
        udid: udidArg,
        domain: z
          .enum(["app", "app-group", "crash", "temp"])
          .default("app")
          .describe(
            "File service domain: 'app'/'app-group' (need identifier), 'crash', or 'temp'.",
          ),
        identifier: z
          .string()
          .optional()
          .describe("Bundle id / app-group id — required for the 'app' and 'app-group' domains."),
        path: z
          .string()
          .optional()
          .describe("Directory path within the domain to list. Defaults to the domain root ('.')."),
      },
    },
    async ({ udid, domain, identifier, path }) =>
      guard(async () => {
        const listing = await client.getJson<FileListing>(`/device/${seg(udid)}/files`, {
          domain,
          identifier,
          path,
        });
        return jsonResult(listing);
      }),
  );

  // ---- Performance (bounded SSE capture) ---------------------------------

  server.registerTool(
    "sample_performance",
    {
      title: "Sample CPU performance (bounded)",
      description:
        "Capture a BOUNDED window of CPU-usage samples from the device (sysmontap) and return " +
        "them. Like stream_logs, this is NOT an infinite stream — it collects samples for " +
        "duration_seconds (default 5, max 30) or until max_samples (default 30, max 300) is " +
        "reached, whichever comes first, then returns them. Each sample has total/system/user CPU " +
        "load. Use this to check device load during or after an operation. Requires a udid.",
      inputSchema: {
        udid: udidArg,
        duration_seconds: z
          .number()
          .int()
          .min(1)
          .max(30)
          .default(5)
          .describe("Max seconds to sample before returning. Capped at 30."),
        max_samples: z
          .number()
          .int()
          .min(1)
          .max(300)
          .default(30)
          .describe("Max number of samples to return. Capped at 300."),
      },
    },
    async ({ udid, duration_seconds, max_samples }) =>
      guard(async () => {
        const url = client.url(`/device/${seg(udid)}/sysmontap`);
        const captured = await captureSse((signal) => openSse(client, url, signal), {
          durationMs: duration_seconds * 1000,
          maxLines: max_samples,
          keepEvents: new Set(["sample"]),
        });
        const samples = captured.events.map((e) => e.payload ?? { raw: e.raw });
        return jsonResult({
          udid,
          bounded: true,
          stoppedBy: captured.stoppedBy,
          durationSeconds: duration_seconds,
          maxSamples: max_samples,
          returned: samples.length,
          totalMatched: captured.totalMatched,
          heartbeats: captured.heartbeats,
          samples,
        });
      }),
  );

  // ---- Jobs (drive & observe long-running operations) --------------------

  server.registerTool(
    "run_wda",
    {
      title: "Start WebDriverAgent runner (job)",
      description:
        "Start the WebDriverAgent runner as an async background JOB and return the created job " +
        "(id, kind, status). Unlike create_wda_session (a single WDA test session), this launches " +
        "the long-running runner the way `ios runwda` does; it is how an agent boots WDA to drive " +
        "UI automation. All body fields are optional and default to the standard WDA bundle id " +
        "and config. Use get_job to poll it, tail_job_logs to watch output, and stop_job to end " +
        "it. Requires a udid.",
      inputSchema: {
        udid: udidArg,
        bundleId: z
          .string()
          .optional()
          .describe("Bundle id of the WDA runner. Defaults to the standard WDA bundle id."),
        testRunnerBundleId: z
          .string()
          .optional()
          .describe("Test runner bundle id. Defaults to bundleId when omitted."),
        xctestConfig: z
          .string()
          .optional()
          .describe("Name of the .xctestconfiguration. Defaults to the standard WDA config."),
        env: z
          .record(z.string())
          .optional()
          .describe("Extra environment variables for the runner."),
        args: z.array(z.string()).optional().describe("Extra process arguments for the runner."),
      },
    },
    async ({ udid, bundleId, testRunnerBundleId, xctestConfig, env, args }) =>
      guard(async () => {
        const body: Record<string, unknown> = {};
        if (bundleId !== undefined) body.bundleId = bundleId;
        if (testRunnerBundleId !== undefined) body.testRunnerBundleId = testRunnerBundleId;
        if (xctestConfig !== undefined) body.xctestConfig = xctestConfig;
        if (env !== undefined) body.env = env;
        if (args !== undefined) body.args = args;
        const job = await client.postJson<Job>(`/device/${seg(udid)}/jobs/runwda`, { body });
        return jsonResult(job);
      }),
  );

  server.registerTool(
    "list_jobs",
    {
      title: "List jobs",
      description:
        "List the async jobs (test runs, WDA runners, port forwards) known for a device, each " +
        "with id, kind, status (running|succeeded|failed|stopped), and timestamps. Use this to " +
        "see what is running before starting or stopping work. Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const jobs = await client.getJson<Job[]>(`/device/${seg(udid)}/jobs`);
        return jsonResult({ count: jobs?.length ?? 0, jobs: jobs ?? [] });
      }),
  );

  server.registerTool(
    "get_job",
    {
      title: "Get job status",
      description:
        "Get one job's current status by id: status (running|succeeded|failed|stopped), start/" +
        "finish times, an error message if it failed, and a terminal result payload if it " +
        "succeeded. Poll this to know when a run_wda / test job finishes. Requires a udid and the " +
        "job id.",
      inputSchema: {
        udid: udidArg,
        id: z.string().min(1).describe("The job id from run_wda / list_jobs, e.g. 'runwda-2'."),
      },
    },
    async ({ udid, id }) =>
      guard(async () => {
        const job = await client.getJson<Job>(`/device/${seg(udid)}/jobs/${seg(id)}`);
        return jsonResult(job);
      }),
  );

  server.registerTool(
    "stop_job",
    {
      title: "Stop a job",
      description:
        "Stop a running job, or purge an already-finished one from the registry, by id. Call this " +
        "to tear down a WDA runner or test job you started with run_wda. Returns the daemon's " +
        "status message. Requires a udid and the job id.",
      inputSchema: {
        udid: udidArg,
        id: z.string().min(1).describe("The job id to stop."),
      },
    },
    async ({ udid, id }) =>
      guard(async () => {
        const res = await client.deleteJson<unknown>(`/device/${seg(udid)}/jobs/${seg(id)}`);
        return jsonResult(res);
      }),
  );

  server.registerTool(
    "tail_job_logs",
    {
      title: "Tail a job's logs (bounded)",
      description:
        "Capture a BOUNDED window of a job's log output and return the collected lines. The job " +
        "log stream sends buffered history first, then live lines; this collects until " +
        "duration_seconds (default 5, max 30) or max_lines (default 200, max 2000) is reached, " +
        "whichever comes first, then returns — it does NOT block until the job ends. Use it to " +
        "watch a run_wda / test job's progress. Requires a udid and the job id.",
      inputSchema: {
        udid: udidArg,
        id: z.string().min(1).describe("The job id whose logs to tail."),
        duration_seconds: z
          .number()
          .int()
          .min(1)
          .max(30)
          .default(5)
          .describe("Max seconds to collect before returning. Capped at 30."),
        max_lines: z
          .number()
          .int()
          .min(1)
          .max(2000)
          .default(200)
          .describe("Max number of log lines to return. Capped at 2000."),
      },
    },
    async ({ udid, id, duration_seconds, max_lines }) =>
      guard(async () => {
        const url = client.url(`/device/${seg(udid)}/jobs/${seg(id)}/logs`);
        const captured = await captureSse((signal) => openSse(client, url, signal), {
          durationMs: duration_seconds * 1000,
          maxLines: max_lines,
          keepEvents: new Set(["log"]),
        });
        const lines = captured.events.map((e) => {
          const p = e.payload as { line?: string } | undefined;
          return p?.line ?? e.raw;
        });
        return jsonResult({
          udid,
          jobId: id,
          bounded: true,
          stoppedBy: captured.stoppedBy,
          durationSeconds: duration_seconds,
          maxLines: max_lines,
          returned: lines.length,
          totalMatched: captured.totalMatched,
          heartbeats: captured.heartbeats,
          lines,
        });
      }),
  );

  // ---- Pasteboard --------------------------------------------------------

  server.registerTool(
    "get_pasteboard",
    {
      title: "Read the pasteboard (clipboard)",
      description:
        "Read the device pasteboard (clipboard) text. Returns { present, text } — present=false " +
        "means the clipboard held no text. Useful for reading a value an app copied, or verifying " +
        "text you injected with set_pasteboard. Requires a udid.",
      inputSchema: { udid: udidArg },
    },
    async ({ udid }) =>
      guard(async () => {
        const pb = await client.getJson<PasteboardContent>(`/device/${seg(udid)}/pasteboard`);
        return jsonResult(pb);
      }),
  );

  server.registerTool(
    "set_pasteboard",
    {
      title: "Set the pasteboard (clipboard)",
      description:
        "Set the device pasteboard (clipboard) to the given text. Useful for injecting text into " +
        "an app under automation (copy here, then paste in the app). Overwrites the current " +
        "clipboard contents. Requires a udid and the text to set.",
      inputSchema: {
        udid: udidArg,
        text: z.string().describe("The text to place on the device clipboard (may be empty)."),
      },
    },
    async ({ udid, text }) =>
      guard(async () => {
        const res = await client.putText<unknown>(`/device/${seg(udid)}/pasteboard`, text);
        return jsonResult(res);
      }),
  );

  // ---- Extension point: UI / accessibility interaction -------------------
  //
  // TODO(ui-automation): the curated set can already START and OBSERVE UI
  // automation — create_wda_session / run_wda boot WebDriverAgent, screenshot
  // lets an agent SEE the screen, and tail_job_logs/get_job track the runner.
  // What is still missing is ACTING on the UI. Once the go-ios daemon exposes UI
  // interaction over REST (tap / type / query by accessibility id), add a small,
  // task-shaped find-then-act set of curated tools here, riding on an existing
  // WDA session (create_wda_session) or WDA runner job (run_wda), e.g.:
  //   - `tap_element`   { udid, sessionId, accessibilityId }
  //   - `type_into`     { udid, sessionId, accessibilityId, text }
  //   - `query_element` { udid, sessionId, accessibilityId }  -> element state
  //   - `dump_ui_tree`  { udid, sessionId }                   -> a11y hierarchy
  // These would POST to the future /device/{udid}/wda/session/{sessionId}/...
  // routes. Keep them curated and semantic (accessibility-id targeted), NOT a
  // raw WDA HTTP passthrough, so descriptions stay LLM-legible. Pair each
  // action tool with screenshot so an agent can verify the result it produced.
}
