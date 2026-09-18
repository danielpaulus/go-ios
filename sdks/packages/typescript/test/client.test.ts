import { describe, expect, it, vi } from "vitest";

import { IosClient, IosApiError, isSseEvent, deviceUdid } from "../src/index";

const enc = new TextEncoder();

/** Extract the Request passed to a mocked fetch (typed loosely for test use). */
function firstRequest(fetchMock: { mock: { calls: unknown[][] } }): Request {
  const call = fetchMock.mock.calls[0];
  if (!call) throw new Error("fetch was not called");
  return call[0] as Request;
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

describe("IosClient facade", () => {
  it("sends the bearer token and calls the list endpoint", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ deviceList: [{ deviceID: 1, properties: { serialNumber: "udid-1" } }] }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      apiKey: "secret-key",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const result = await client.devices.list();
    expect(result.deviceList[0]!.properties.serialNumber).toBe("udid-1");

    const req = firstRequest(fetchMock);
    expect(req.url).toBe("http://ios.test/api/v1/list");
    expect(req.headers.get("authorization")).toBe("Bearer secret-key");
  });

  it("sends bundleID as a query param when launching an app", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "ok" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      apiKey: "k",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const res = await client.device("udid-1").apps.launch("com.apple.Preferences");
    expect(res.message).toBe("ok");

    const req = firstRequest(fetchMock);
    const url = new URL(req.url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/apps/launch");
    expect(url.searchParams.get("bundleID")).toBe("com.apple.Preferences");
    expect(req.method).toBe("POST");
  });

  it("returns screenshot bytes as a Blob", async () => {
    const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47]);
    const fetchMock = vi.fn(
      async () =>
        new Response(png, { status: 200, headers: { "content-type": "image/png" } }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const blob = await client.device("udid-1").screenshot();
    expect(blob).toBeInstanceOf(Blob);
    const bytes = new Uint8Array(await blob.arrayBuffer());
    expect(Array.from(bytes)).toEqual([0x89, 0x50, 0x4e, 0x47]);
  });

  it("throws IosApiError with status and message on error responses", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ error: "device not found" }, 404));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await expect(client.device("missing").info()).rejects.toMatchObject({
      name: "IosApiError",
      status: 404,
      message: "device not found",
    });
    await expect(client.device("missing").info()).rejects.toBeInstanceOf(IosApiError);
  });

  it("maps setLocation numbers to string query params (longitude spelled correctly)", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "ok" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.device("udid-1").setLocation(52.52, 13.405);
    const req = firstRequest(fetchMock);
    const url = new URL(req.url);
    expect(url.searchParams.get("latitude")).toBe("52.52");
    expect(url.searchParams.get("longitude")).toBe("13.405");
    expect(req.method).toBe("PUT");
  });

  it("streams typed SSE syslog events over a mock fetch", async () => {
    const body = new ReadableStream<Uint8Array>({
      start(c) {
        c.enqueue(enc.encode("event: heartbeat\ndata: {}\n\n"));
        c.enqueue(enc.encode('event: syslog\ndata: {"message":"line-1"}\n\n'));
        c.enqueue(enc.encode('event: syslog\ndata: {"message":"line-2"}\n\n'));
        c.close();
      },
    });
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "text/event-stream" },
        }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const messages: string[] = [];
    for await (const ev of client.device("udid-1").syslog()) {
      if (isSseEvent(ev, "syslog")) messages.push(ev.data.message);
    }
    expect(messages).toEqual(["line-1", "line-2"]);
  });

  it("throws before streaming when the SSE endpoint returns an error status", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ error: "no device" }, 404));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await expect(async () => {
      for await (const _ of client.device("x").syslog()) {
        void _;
      }
    }).rejects.toMatchObject({ name: "IosApiError", status: 404 });
  });
});

describe("extended device facade", () => {
  it("unwraps the device name from the { devicename } envelope", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ devicename: "Daniel's iPhone" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const name = await client.device("udid-1").deviceName();
    expect(name).toBe("Daniel's iPhone");
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/devicename");
  });

  it("repeats the key query param for mobileGestalt", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ ProductType: "iPhone15,2" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.device("udid-1").mobileGestalt(["ProductType", "BuildVersion"]);
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/mobilegestalt");
    // hey-api serializes array query params comma-joined (explode: false).
    expect(url.searchParams.get("key")).toBe("ProductType,BuildVersion");
  });

  it("requires confirm=true on the destructive erase", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "erasing" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.device("udid-1").erase(true);
    const req = firstRequest(fetchMock);
    const url = new URL(req.url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/erase");
    expect(url.searchParams.get("confirm")).toBe("true");
    expect(req.method).toBe("POST");
  });

  it("passes domain and path as query params for files.ls", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ path: "/Documents", files: [], count: 0 }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client
      .device("udid-1")
      .files.ls({ domain: "app", identifier: "com.x.y" }, "/Documents");
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/files");
    expect(url.searchParams.get("domain")).toBe("app");
    expect(url.searchParams.get("identifier")).toBe("com.x.y");
    expect(url.searchParams.get("path")).toBe("/Documents");
  });

  it("sends the raw body and query for files.push", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ remote: "/tmp/x", size: 4 }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const res = await client
      .device("udid-1")
      .files.push({ domain: "temp" }, "/tmp/x", new Uint8Array([1, 2, 3, 4]));
    expect(res.size).toBe(4);
    const req = firstRequest(fetchMock);
    const url = new URL(req.url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/files/push");
    expect(url.searchParams.get("remote")).toBe("/tmp/x");
    expect(req.method).toBe("POST");
  });

  it("returns pulled file bytes as a Blob", async () => {
    const bytes = new Uint8Array([9, 8, 7]);
    const fetchMock = vi.fn(
      async () =>
        new Response(bytes, {
          status: 200,
          headers: { "content-type": "application/octet-stream" },
        }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const blob = await client
      .device("udid-1")
      .files.pull({ domain: "crash" }, "/report.ips");
    expect(blob).toBeInstanceOf(Blob);
    expect(Array.from(new Uint8Array(await blob.arrayBuffer()))).toEqual([9, 8, 7]);
  });

  it("sends pasteboard text as a text/plain PUT body", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "ok" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.device("udid-1").media.setPasteboard("hello clipboard");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("PUT");
    expect(new URL(req.url).pathname).toBe("/api/v1/device/udid-1/pasteboard");
    expect(await req.text()).toBe("hello clipboard");
  });

  it("returns the AssistiveTouch state from setAssistiveTouch", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ AssistiveTouchEnabled: true }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const state = await client.device("udid-1").settings.setAssistiveTouch(true);
    expect(state.AssistiveTouchEnabled).toBe(true);
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("PUT");
    expect(await req.json()).toEqual({ enabled: true });
  });

  it("removes wifi by ssid via query param", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "ok" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.device("udid-1").settings.removeWifi("MyNet");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("DELETE");
    expect(new URL(req.url).searchParams.get("ssid")).toBe("MyNet");
  });

  it("posts the unlock token with the p12 identity on clearPasscode", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ status: "ok" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const res = await client
      .device("udid-1")
      .mdm.clearPasscode({ p12: new Uint8Array([1]), password: "pw" }, "base64token");
    expect(res.status).toBe("ok");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    expect(new URL(req.url).pathname).toBe(
      "/api/v1/device/udid-1/mdm/clear-passcode",
    );
    const form = await req.formData();
    expect(form.get("token")).toBe("base64token");
    expect(form.get("password")).toBe("pw");
  });
});

describe("jobs facade", () => {
  it("starts a test run and returns the created job", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse(
        { id: "runtest-1", kind: "runtest", udid: "udid-1", status: "running", startedAt: "t" },
        202,
      ),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const job = await client.device("udid-1").jobs.runtest({ bundleId: "com.x" });
    expect(job.id).toBe("runtest-1");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    expect(new URL(req.url).pathname).toBe("/api/v1/device/udid-1/jobs/runtest");
    expect(await req.json()).toEqual({ bundleId: "com.x" });
  });

  it("gets a job by id under the device path", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ id: "j1", kind: "forward", udid: "udid-1", status: "succeeded", startedAt: "t" }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const job = await client.device("udid-1").jobs.get("j1");
    expect(job.status).toBe("succeeded");
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/jobs/j1",
    );
  });

  it("deletes a job", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "stopped" }));
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.device("udid-1").jobs.delete("j1");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("DELETE");
    expect(new URL(req.url).pathname).toBe("/api/v1/device/udid-1/jobs/j1");
  });

  it("streams typed job-log SSE events", async () => {
    const body = new ReadableStream<Uint8Array>({
      start(c) {
        c.enqueue(enc.encode("event: heartbeat\ndata: {}\n\n"));
        c.enqueue(enc.encode('event: log\ndata: {"line":"build started"}\n\n'));
        c.enqueue(enc.encode('event: log\ndata: {"line":"test passed"}\n\n'));
        c.close();
      },
    });
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "text/event-stream" },
        }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const lines: string[] = [];
    for await (const ev of client.device("udid-1").jobs.logs("j1")) {
      if (isSseEvent(ev, "log")) lines.push(ev.data.line);
    }
    expect(lines).toEqual(["build started", "test passed"]);
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/jobs/j1/logs",
    );
  });
});

describe("tunnels facade", () => {
  it("lists tunnels", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse([{ Udid: "udid-1", Address: "fd00::1", RsdPort: 5000 }]),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const tunnels = await client.tunnels.list();
    expect(tunnels[0]!.Udid).toBe("udid-1");
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe("/api/v1/tunnels");
  });

  it("refreshes a tunnel by udid", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ Udid: "udid-1", Address: "fd00::1", RsdPort: 5000 }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    await client.tunnels.refresh("udid-1");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    expect(new URL(req.url).pathname).toBe("/api/v1/tunnels/udid-1/refresh");
  });
});

describe("sysmontap SSE stream", () => {
  it("streams typed CPU-usage samples", async () => {
    const body = new ReadableStream<Uint8Array>({
      start(c) {
        c.enqueue(enc.encode('event: sample\ndata: {"CPU_TotalLoad":42}\n\n'));
        c.enqueue(enc.encode("event: heartbeat\ndata: {}\n\n"));
        c.enqueue(enc.encode('event: sample\ndata: {"CPU_TotalLoad":7}\n\n'));
        c.close();
      },
    });
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "text/event-stream" },
        }),
    );
    const client = new IosClient({
      baseUrl: "http://ios.test",
      fetch: fetchMock as unknown as typeof fetch,
    });

    const loads: number[] = [];
    for await (const ev of client.device("udid-1").sysmontap()) {
      if (isSseEvent(ev, "sample")) loads.push(ev.data.CPU_TotalLoad ?? -1);
    }
    expect(loads).toEqual([42, 7]);
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/sysmontap",
    );
  });
});

describe("deviceUdid accessor", () => {
  it("reads properties.serialNumber", () => {
    expect(deviceUdid({ properties: { serialNumber: "abc123" } })).toBe("abc123");
  });
});
