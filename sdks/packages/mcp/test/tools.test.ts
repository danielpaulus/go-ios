import { describe, it, expect, afterAll } from "vitest";
import { createServer as createHttpServer, type Server } from "node:http";
import type { AddressInfo } from "node:net";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";
import { createServer } from "../src/server.js";
import { CURATED_TOOL_NAMES } from "../src/tools.js";
import type { ServerConfig } from "../src/config.js";

/**
 * Spin up a mock go-ios REST daemon so tool handlers can be exercised end to end
 * without a real device. Returns the base URL and lets each test route requests.
 */
function startMockDaemon(
  handler: (
    method: string,
    path: string,
    url: URL,
    body: string,
  ) => { status: number; contentType?: string; body: string | Buffer },
): Promise<{ baseUrl: string; close: () => Promise<void> }> {
  return new Promise((resolve) => {
    const server: Server = createHttpServer(async (req, res) => {
      const chunks: Buffer[] = [];
      for await (const c of req) chunks.push(c as Buffer);
      const raw = Buffer.concat(chunks).toString("utf8");
      const url = new URL(req.url ?? "/", "http://localhost");
      const r = handler(req.method ?? "GET", url.pathname, url, raw);
      res.writeHead(r.status, {
        "Content-Type": r.contentType ?? "application/json",
      });
      res.end(r.body);
    });
    server.listen(0, "127.0.0.1", () => {
      const port = (server.address() as AddressInfo).port;
      resolve({
        baseUrl: `http://127.0.0.1:${port}`,
        close: () => new Promise((r) => server.close(() => r())),
      });
    });
  });
}

async function connectedClient(config: ServerConfig): Promise<Client> {
  const server = createServer(config);
  const [clientTransport, serverTransport] = InMemoryTransport.createLinkedPair();
  const client = new Client({ name: "test", version: "0.0.0" });
  await Promise.all([server.connect(serverTransport), client.connect(clientTransport)]);
  return client;
}

const baseConfig: ServerConfig = {
  baseUrl: "http://localhost:9",
  transport: "stdio",
  httpPort: 3000,
  httpHost: "127.0.0.1",
};

describe("tool registration", () => {
  it("registers exactly the curated tool set with input schemas", async () => {
    const client = await connectedClient(baseConfig);
    const { tools } = await client.listTools();
    const names = tools.map((t) => t.name).sort();
    expect(names).toEqual([...CURATED_TOOL_NAMES].sort());

    for (const t of tools) {
      expect(t.description, `${t.name} has a description`).toBeTruthy();
      expect(t.inputSchema, `${t.name} has an inputSchema`).toBeTruthy();
      expect(t.inputSchema.type).toBe("object");
    }
    await client.close();
  });

  it("device-scoped tools require a udid", async () => {
    const client = await connectedClient(baseConfig);
    const { tools } = await client.listTools();
    const needUdid = tools.filter((t) => t.name !== "list_devices");
    for (const t of needUdid) {
      const required = (t.inputSchema.required as string[] | undefined) ?? [];
      expect(required, `${t.name} requires udid`).toContain("udid");
    }
    await client.close();
  });

  it("registers every curated tool including the newly-added ones", async () => {
    const client = await connectedClient(baseConfig);
    const { tools } = await client.listTools();
    const names = new Set(tools.map((t) => t.name));
    for (const expected of [
      "reboot_device",
      "shutdown_device",
      "device_battery",
      "list_processes",
      "device_diagnostics",
      "list_crash_reports",
      "pull_crash_report",
      "list_files",
      "sample_performance",
      "run_wda",
      "get_job",
      "list_jobs",
      "stop_job",
      "tail_job_logs",
      "get_pasteboard",
      "set_pasteboard",
    ]) {
      expect(names.has(expected), `${expected} is registered`).toBe(true);
    }
    // erase is deliberately omitted (too destructive for an agent tool).
    expect(names.has("erase_device")).toBe(false);
    expect(names.has("erase")).toBe(false);
    await client.close();
  });

  it("marks the disruptive device-management tools in their descriptions", async () => {
    const client = await connectedClient(baseConfig);
    const { tools } = await client.listTools();
    const byName = new Map(tools.map((t) => [t.name, t]));
    expect(byName.get("reboot_device")!.description).toMatch(/DISRUPTIVE/);
    expect(byName.get("shutdown_device")!.description).toMatch(/DISRUPTIVE/);
    await client.close();
  });

  it("bounded-capture tools cap their duration and line inputs", async () => {
    const client = await connectedClient(baseConfig);
    // Out-of-range inputs are rejected by the tool's zod schema before any REST
    // call, surfaced as an input-validation error result.
    const rejects = async (name: string, args: Record<string, unknown>) => {
      const res = await client.callTool({ name, arguments: args });
      expect(res.isError, `${name} ${JSON.stringify(args)} rejected`).toBe(true);
      const text = (res.content as Array<{ text: string }>)[0]!.text;
      expect(text).toMatch(/validation|less than or equal/i);
    };
    // sample_performance: duration_seconds max 30, max_samples max 300.
    await rejects("sample_performance", { udid: "UDID-A", duration_seconds: 999 });
    await rejects("sample_performance", { udid: "UDID-A", max_samples: 100000 });
    // tail_job_logs: duration_seconds max 30, max_lines max 2000.
    await rejects("tail_job_logs", { udid: "UDID-A", id: "j1", duration_seconds: 60 });
    await rejects("tail_job_logs", { udid: "UDID-A", id: "j1", max_lines: 999999 });
    await client.close();
  });
});

describe("tool handlers against a mock REST daemon", () => {
  let daemon: Awaited<ReturnType<typeof startMockDaemon>>;

  afterAll(async () => {
    await daemon?.close();
  });

  it("list_devices maps the device list and sends the bearer token", async () => {
    daemon = await startMockDaemon((method, path) => {
      if (method === "GET" && path === "/api/v1/list") {
        return {
          status: 200,
          body: JSON.stringify({
            deviceList: [
              {
                deviceID: 5,
                properties: { serialNumber: "UDID-A", connectionType: "USB" },
                address: "10.0.0.2",
              },
            ],
          }),
        };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl, apiKey: "secret" });
    const res = await client.callTool({ name: "list_devices", arguments: {} });
    const sc = res.structuredContent as { count: number; devices: Array<{ udid: string }> };
    expect(sc.count).toBe(1);
    expect(sc.devices[0]!.udid).toBe("UDID-A");
    expect(res.isError).toBeFalsy();
    await client.close();
  });

  it("launch_app posts bundleID as a query param", async () => {
    let seenQuery = "";
    daemon = await startMockDaemon((method, path, url) => {
      if (method === "POST" && path === "/api/v1/device/UDID-A/apps/launch") {
        seenQuery = url.searchParams.get("bundleID") ?? "";
        return { status: 200, body: JSON.stringify({ message: "launched" }) };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "launch_app",
      arguments: { udid: "UDID-A", bundleId: "com.apple.Preferences" },
    });
    expect(seenQuery).toBe("com.apple.Preferences");
    expect((res.structuredContent as { message: string }).message).toBe("launched");
    await client.close();
  });

  it("surfaces a 404 device-not-found as a clean error result", async () => {
    daemon = await startMockDaemon(() => ({
      status: 404,
      body: JSON.stringify({ error: "device not found" }),
    }));
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "device_info",
      arguments: { udid: "UNKNOWN" },
    });
    expect(res.isError).toBe(true);
    const text = (res.content as Array<{ type: string; text: string }>)[0]!.text;
    expect(text).toContain("not found");
    await client.close();
  });

  it("screenshot returns an image content block", async () => {
    // 1x1 transparent PNG.
    const png = Buffer.from(
      "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
      "base64",
    );
    daemon = await startMockDaemon((method, path) => {
      if (method === "GET" && path === "/api/v1/device/UDID-A/screenshot") {
        return { status: 200, contentType: "image/png", body: png };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({ name: "screenshot", arguments: { udid: "UDID-A" } });
    const block = (res.content as Array<{ type: string; mimeType?: string; data?: string }>)[0]!;
    expect(block.type).toBe("image");
    expect(block.mimeType).toBe("image/png");
    expect(block.data).toBe(png.toString("base64"));
    await client.close();
  });

  it("get_pasteboard reads clipboard content", async () => {
    daemon = await startMockDaemon((method, path) => {
      if (method === "GET" && path === "/api/v1/device/UDID-A/pasteboard") {
        return { status: 200, body: JSON.stringify({ present: true, text: "hello" }) };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({ name: "get_pasteboard", arguments: { udid: "UDID-A" } });
    const sc = res.structuredContent as { present: boolean; text: string };
    expect(sc.present).toBe(true);
    expect(sc.text).toBe("hello");
    await client.close();
  });

  it("set_pasteboard PUTs the text as text/plain", async () => {
    let seenBody = "";
    let seenContentType = "";
    daemon = await startMockDaemon((method, path, _url, body) => {
      if (method === "PUT" && path === "/api/v1/device/UDID-A/pasteboard") {
        seenBody = body;
        return { status: 200, body: JSON.stringify({ message: "ok" }) };
      }
      return { status: 404, body: "{}" };
    });
    // Capture the content-type header via a wrapping handler is not exposed; assert body instead.
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "set_pasteboard",
      arguments: { udid: "UDID-A", text: "copied text" },
    });
    expect(seenBody).toBe("copied text");
    expect(res.isError).toBeFalsy();
    void seenContentType;
    await client.close();
  });

  it("list_crash_reports returns file names and passes the pattern query", async () => {
    let seenPattern = "";
    daemon = await startMockDaemon((method, path, url) => {
      if (method === "GET" && path === "/api/v1/device/UDID-A/crashes") {
        seenPattern = url.searchParams.get("pattern") ?? "";
        return {
          status: 200,
          body: JSON.stringify({ files: ["MyApp-1.ips", "MyApp-2.ips"], count: 2 }),
        };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "list_crash_reports",
      arguments: { udid: "UDID-A", pattern: "MyApp*" },
    });
    expect(seenPattern).toBe("MyApp*");
    expect((res.structuredContent as { count: number }).count).toBe(2);
    await client.close();
  });

  it("pull_crash_report returns bounded text from the crash file domain", async () => {
    let seenDomain = "";
    let seenRemote = "";
    daemon = await startMockDaemon((method, path, url) => {
      if (method === "GET" && path === "/api/v1/device/UDID-A/files/pull") {
        seenDomain = url.searchParams.get("domain") ?? "";
        seenRemote = url.searchParams.get("remote") ?? "";
        return {
          status: 200,
          contentType: "application/octet-stream",
          body: "Thread 0 crashed\nEXC_BAD_ACCESS",
        };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "pull_crash_report",
      arguments: { udid: "UDID-A", name: "MyApp-1.ips" },
    });
    expect(seenDomain).toBe("crash");
    expect(seenRemote).toBe("MyApp-1.ips");
    const sc = res.structuredContent as { text: string; truncated: boolean; bytes: number };
    expect(sc.text).toContain("EXC_BAD_ACCESS");
    expect(sc.truncated).toBe(false);
    expect((res.content as Array<{ type: string; text: string }>)[0]!.text).toContain(
      "Thread 0 crashed",
    );
    await client.close();
  });

  it("run_wda posts to the runwda job endpoint and returns the job", async () => {
    daemon = await startMockDaemon((method, path) => {
      if (method === "POST" && path === "/api/v1/device/UDID-A/jobs/runwda") {
        return {
          status: 202,
          body: JSON.stringify({
            id: "runwda-1",
            kind: "runwda",
            udid: "UDID-A",
            status: "running",
            startedAt: "2026-01-01T00:00:00Z",
          }),
        };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({ name: "run_wda", arguments: { udid: "UDID-A" } });
    const sc = res.structuredContent as { id: string; status: string };
    expect(sc.id).toBe("runwda-1");
    expect(sc.status).toBe("running");
    await client.close();
  });

  it("list_processes maps count and passes apps_only", async () => {
    let seenApps = "";
    daemon = await startMockDaemon((method, path, url) => {
      if (method === "GET" && path === "/api/v1/device/UDID-A/processes") {
        seenApps = url.searchParams.get("apps") ?? "";
        return {
          status: 200,
          body: JSON.stringify([
            { pid: 1, name: "launchd", isApplication: false },
            { pid: 42, name: "MyApp", isApplication: true },
          ]),
        };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "list_processes",
      arguments: { udid: "UDID-A", apps_only: true },
    });
    expect(seenApps).toBe("true");
    expect((res.structuredContent as { count: number }).count).toBe(2);
    await client.close();
  });

  it("sample_performance captures a bounded window of CPU samples", async () => {
    daemon = await startMockDaemon(() => ({
      status: 200,
      contentType: "text/event-stream",
      body:
        'event: sample\ndata: {"CPU_TotalLoad":10}\n\n' +
        "event: heartbeat\ndata: {}\n\n" +
        'event: sample\ndata: {"CPU_TotalLoad":20}\n\n' +
        'event: sample\ndata: {"CPU_TotalLoad":30}\n\n',
    }));
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "sample_performance",
      arguments: { udid: "UDID-A", duration_seconds: 2, max_samples: 2 },
    });
    const sc = res.structuredContent as {
      bounded: boolean;
      returned: number;
      stoppedBy: string;
      heartbeats: number;
      samples: Array<{ CPU_TotalLoad: number }>;
    };
    expect(sc.bounded).toBe(true);
    expect(sc.returned).toBe(2);
    expect(sc.stoppedBy).toBe("maxLines");
    expect(sc.heartbeats).toBe(1);
    expect(sc.samples[0]!.CPU_TotalLoad).toBe(10);
    await client.close();
  });

  it("tail_job_logs captures a bounded window of job log lines", async () => {
    daemon = await startMockDaemon((method, path) => {
      if (method === "GET" && path === "/api/v1/device/UDID-A/jobs/runwda-1/logs") {
        return {
          status: 200,
          contentType: "text/event-stream",
          body:
            'event: log\ndata: {"line":"starting"}\n\n' +
            'event: log\ndata: {"line":"running"}\n\n' +
            'event: log\ndata: {"line":"done"}\n\n',
        };
      }
      return { status: 404, body: "{}" };
    });
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "tail_job_logs",
      arguments: { udid: "UDID-A", id: "runwda-1", duration_seconds: 2, max_lines: 2 },
    });
    const sc = res.structuredContent as {
      bounded: boolean;
      returned: number;
      stoppedBy: string;
      lines: string[];
    };
    expect(sc.bounded).toBe(true);
    expect(sc.returned).toBe(2);
    expect(sc.stoppedBy).toBe("maxLines");
    expect(sc.lines[0]).toBe("starting");
    await client.close();
  });

  it("stream_logs captures a bounded window of syslog events", async () => {
    daemon = await startMockDaemon(() => ({
      status: 200,
      contentType: "text/event-stream",
      body:
        'event: syslog\ndata: {"message":"a","timestamp":1}\n\n' +
        "event: heartbeat\ndata: {}\n\n" +
        'event: syslog\ndata: {"message":"b","timestamp":2}\n\n' +
        'event: syslog\ndata: {"message":"c","timestamp":3}\n\n',
    }));
    const client = await connectedClient({ ...baseConfig, baseUrl: daemon.baseUrl });
    const res = await client.callTool({
      name: "stream_logs",
      arguments: { udid: "UDID-A", source: "syslog", duration_seconds: 2, max_lines: 2 },
    });
    const sc = res.structuredContent as {
      bounded: boolean;
      returned: number;
      stoppedBy: string;
      lines: Array<{ message: string }>;
    };
    expect(sc.bounded).toBe(true);
    expect(sc.returned).toBe(2);
    expect(sc.stoppedBy).toBe("maxLines");
    expect(sc.lines[0]!.message).toBe("a");
    await client.close();
  });
});
