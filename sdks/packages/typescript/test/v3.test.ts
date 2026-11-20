import { describe, expect, it, vi } from "vitest";

import { IosClient, openBinaryStream, IosApiError } from "../src/index";

/** Extract the Request passed to a mocked fetch. */
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

function makeClient(fetchMock: unknown): IosClient {
  return new IosClient({
    baseUrl: "http://ios.test",
    apiKey: "k",
    fetch: fetchMock as typeof fetch,
  });
}

describe("diagnostics / network group", () => {
  it("gets disk space", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ TotalDiskCapacity: 128 }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").diskSpace();
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/diskspace",
    );
  });

  it("gets ip / rsd / battery-registry under the right paths", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({}));
    const client = makeClient(fetchMock);
    await client.device("udid-1").ip();
    await client.device("udid-1").rsd();
    await client.device("udid-1").batteryRegistry();
    const paths = (fetchMock.mock.calls as unknown[][]).map(
      (c) => new URL((c[0] as Request).url).pathname,
    );
    expect(paths).toEqual([
      "/api/v1/device/udid-1/ip",
      "/api/v1/device/udid-1/rsd",
      "/api/v1/device/udid-1/battery/registry",
    ]);
  });
});

describe("lockdown domain", () => {
  it("passes the optional domain as a query param", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ DeviceName: "x" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").lockdown("com.apple.mobile.battery");
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/lockdown");
    expect(url.searchParams.get("domain")).toBe("com.apple.mobile.battery");
  });

  it("omits the domain query when not given", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({}));
    const client = makeClient(fetchMock);
    await client.device("udid-1").lockdown();
    expect(new URL(firstRequest(fetchMock).url).searchParams.has("domain")).toBe(
      false,
    );
  });
});

describe("accessibility group", () => {
  it("toggles voiceOver via the enabled query param", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ VoiceOverEnabled: true }));
    const client = makeClient(fetchMock);
    const state = await client.device("udid-1").setVoiceOver(true);
    expect(state.VoiceOverEnabled).toBe(true);
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("PUT");
    const url = new URL(req.url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/voiceover");
    expect(url.searchParams.get("enabled")).toBe("true");
  });

  it("toggles zoom via the enabled query param", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ enabled: false }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").setZoom(false);
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/zoom");
    expect(url.searchParams.get("enabled")).toBe("false");
  });

  it("runs an ax audit with a timeout and returns the issues array", async () => {
    const fetchMock = vi.fn(async () => jsonResponse([{ type: "contrast" }]));
    const client = makeClient(fetchMock);
    const issues = await client.device("udid-1").axAudit({ timeout: 30 });
    expect(issues).toHaveLength(1);
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/ax/audit");
    expect(url.searchParams.get("timeout")).toBe("30");
  });

  it("uploads a gpx file as multipart on setLocationGpx", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "ok" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").setLocationGpx(new Uint8Array([1, 2, 3]));
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("PUT");
    expect(new URL(req.url).pathname).toBe(
      "/api/v1/device/udid-1/setlocation/gpx",
    );
    const form = await req.formData();
    expect(form.get("gpx")).toBeInstanceOf(Blob);
  });
});

describe("fsync group", () => {
  it("lists a path scoped to a bundle id", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ entries: [] }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").fsync.ls("/Documents", { bundleId: "com.x.y" });
    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/fsync/ls");
    expect(url.searchParams.get("path")).toBe("/Documents");
    expect(url.searchParams.get("bundleID")).toBe("com.x.y");
  });

  it("pulls a file as raw bytes (Uint8Array)", async () => {
    const bytes = new Uint8Array([5, 6, 7, 8]);
    const fetchMock = vi.fn(
      async () =>
        new Response(bytes, {
          status: 200,
          headers: { "content-type": "application/octet-stream" },
        }),
    );
    const client = makeClient(fetchMock);
    const out = await client.device("udid-1").fsync.pull("/tmp/a.bin");
    expect(out).toBeInstanceOf(Uint8Array);
    expect(Array.from(out)).toEqual([5, 6, 7, 8]);
    expect(new URL(firstRequest(fetchMock).url).searchParams.get("path")).toBe(
      "/tmp/a.bin",
    );
  });

  it("pushes bytes with the raw body", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ path: "/tmp/x", size: 3 }));
    const client = makeClient(fetchMock);
    const res = await client
      .device("udid-1")
      .fsync.push("/tmp/x", new Uint8Array([1, 2, 3]));
    expect(res.size).toBe(3);
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    expect(new URL(req.url).pathname).toBe("/api/v1/device/udid-1/fsync/push");
  });

  it("removes recursively via the recursive query param", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "removed" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").fsync.rm("/tmp/dir", { recursive: true });
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("DELETE");
    const url = new URL(req.url);
    expect(url.searchParams.get("path")).toBe("/tmp/dir");
    expect(url.searchParams.get("recursive")).toBe("true");
  });

  it("makes a directory", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ message: "created" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").fsync.mkdir("/tmp/new");
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/fsync/mkdir",
    );
  });

  it("reads cloud config at the device level", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ IsSupervised: true }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").cloudConfig();
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/cloudconfig",
    );
  });
});

describe("webinspector group", () => {
  it("lists pages", async () => {
    const fetchMock = vi.fn(async () => jsonResponse([{ pageId: "1" }]));
    const client = makeClient(fetchMock);
    const pages = await client.device("udid-1").webinspector.pages();
    expect(pages).toHaveLength(1);
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/webinspector/pages",
    );
  });

  it("evaluates a script with a page + bundle id in the body", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ result: "42" }));
    const client = makeClient(fetchMock);
    await client
      .device("udid-1")
      .webinspector.eval("1+1", { page: "p1", bundleId: "com.apple.mobilesafari" });
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    expect(await req.json()).toEqual({
      script: "1+1",
      page: "p1",
      bundleId: "com.apple.mobilesafari",
    });
  });

  it("launches a url", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ ok: true }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").webinspector.launch("https://example.com");
    const req = firstRequest(fetchMock);
    expect(new URL(req.url).pathname).toBe(
      "/api/v1/device/udid-1/webinspector/launch",
    );
    expect(await req.json()).toEqual({ url: "https://example.com" });
  });
});

describe("ui group", () => {
  it("taps with x/y in the body and backend/timeout query", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ status: "ok" }));
    const client = makeClient(fetchMock);
    await client
      .device("udid-1")
      .ui.tap(100, 200, { backend: "wda", timeout: 30 });
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    const url = new URL(req.url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/ui/tap");
    expect(url.searchParams.get("backend")).toBe("wda");
    expect(url.searchParams.get("timeout")).toBe("30");
    expect(await req.json()).toEqual({ x: 100, y: 200 });
  });

  it("swipes with a duration", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ status: "ok" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").ui.swipe(1, 2, 3, 4, 0.5);
    expect(await firstRequest(fetchMock).json()).toEqual({
      x1: 1,
      y1: 2,
      x2: 3,
      y2: 4,
      duration: 0.5,
    });
  });

  it("returns ui screenshot bytes as Uint8Array", async () => {
    const png = new Uint8Array([0x89, 0x50]);
    const fetchMock = vi.fn(
      async () =>
        new Response(png, { status: 200, headers: { "content-type": "image/png" } }),
    );
    const client = makeClient(fetchMock);
    const out = await client.device("udid-1").ui.screenshot();
    expect(Array.from(out)).toEqual([0x89, 0x50]);
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/ui/screenshot",
    );
  });

  it("returns the ui source as an XML string", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response("<Appium><XCUIElement/></Appium>", {
          status: 200,
          headers: { "content-type": "application/xml" },
        }),
    );
    const client = makeClient(fetchMock);
    const xml = await client.device("udid-1").ui.source();
    expect(xml).toContain("XCUIElement");
  });

  it("sets orientation via the body", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ status: "ok" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").ui.setOrientation("landscapeLeft");
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("PUT");
    expect(await req.json()).toEqual({ orientation: "landscapeLeft" });
  });

  it("launches an app through ui.app.launch", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ status: "ok" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").ui.app.launch("com.apple.Preferences");
    const req = firstRequest(fetchMock);
    expect(new URL(req.url).pathname).toBe(
      "/api/v1/device/udid-1/ui/app/launch",
    );
    expect(await req.json()).toEqual({ bundleId: "com.apple.Preferences" });
  });

  it("foregrounds without a body", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ status: "ok" }));
    const client = makeClient(fetchMock);
    await client.device("udid-1").ui.app.foreground();
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/device/udid-1/ui/app/foreground",
    );
  });
});

describe("host-level sign / prepare groups", () => {
  it("returns a pkcs12 certificate as bytes", async () => {
    const p12 = new Uint8Array([0x30, 0x82]);
    const fetchMock = vi.fn(
      async () =>
        new Response(p12, {
          status: 200,
          headers: { "content-type": "application/x-pkcs12" },
        }),
    );
    const client = makeClient(fetchMock);
    const out = await client.sign.certificate({
      ascPrivateKey: new Uint8Array([1]),
      ascKeyId: "KID",
      ascIssuerId: "IID",
    });
    expect(Array.from(out)).toEqual([0x30, 0x82]);
    const req = firstRequest(fetchMock);
    expect(new URL(req.url).pathname).toBe("/api/v1/sign/certificate");
    const form = await req.formData();
    expect(form.get("asc-key-id")).toBe("KID");
    expect(form.get("asc-issuer-id")).toBe("IID");
  });

  it("resigns an app and returns ipa bytes", async () => {
    const ipa = new Uint8Array([0x50, 0x4b]);
    const fetchMock = vi.fn(
      async () =>
        new Response(ipa, {
          status: 200,
          headers: { "content-type": "application/octet-stream" },
        }),
    );
    const client = makeClient(fetchMock);
    const out = await client.sign.app({
      ipa: new Uint8Array([1]),
      p12: new Uint8Array([2]),
      profile: new Uint8Array([3]),
      bundleId: "com.x.y",
    });
    expect(Array.from(out)).toEqual([0x50, 0x4b]);
    const form = await firstRequest(fetchMock).formData();
    expect(form.get("bundleid")).toBe("com.x.y");
    expect(form.get("ipa")).toBeInstanceOf(Blob);
  });

  it("creates a supervision cert (host-level, device-free)", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ certPem: "-----BEGIN" }));
    const client = makeClient(fetchMock);
    await client.prepare.createCert();
    const req = firstRequest(fetchMock);
    expect(req.method).toBe("POST");
    expect(new URL(req.url).pathname).toBe("/api/v1/prepare/create-cert");
  });

  it("lists skip options", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ options: ["Passcode"] }));
    const client = makeClient(fetchMock);
    await client.prepare.skipOptions();
    expect(new URL(firstRequest(fetchMock).url).pathname).toBe(
      "/api/v1/prepare/skip-options",
    );
  });
});

describe("binary streams", () => {
  /** Build a chunked ReadableStream from byte arrays. */
  function chunkedStream(
    chunks: number[][],
    onCancel?: () => void,
  ): ReadableStream<Uint8Array> {
    let i = 0;
    return new ReadableStream<Uint8Array>({
      pull(controller) {
        if (i < chunks.length) {
          controller.enqueue(new Uint8Array(chunks[i]!));
          i++;
        } else {
          controller.close();
        }
      },
      cancel() {
        onCancel?.();
      },
    });
  }

  it("iterates a mock chunked body as a byte stream (ui.stream)", async () => {
    const body = chunkedStream([
      [0xff, 0xd8],
      [0xff, 0xd9],
    ]);
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "application/octet-stream" },
        }),
    );
    const client = makeClient(fetchMock);
    const stream = await client.device("udid-1").ui.stream({ codec: "mjpeg" });

    const received: number[] = [];
    for await (const chunk of stream) received.push(...chunk);
    expect(received).toEqual([0xff, 0xd8, 0xff, 0xd9]);

    const url = new URL(firstRequest(fetchMock).url);
    expect(url.pathname).toBe("/api/v1/device/udid-1/ui/stream");
    expect(url.searchParams.get("codec")).toBe("mjpeg");
  });

  it("exposes the raw body and content type (screenshotStream)", async () => {
    const body = chunkedStream([[1, 2, 3]]);
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "image/jpeg" },
        }),
    );
    const client = makeClient(fetchMock);
    const stream = await client
      .device("udid-1")
      .screenshotStream({ quality: 90 });
    expect(stream.contentType).toBe("image/jpeg");
    expect(stream.body).toBeInstanceOf(ReadableStream);
    expect(
      new URL(firstRequest(fetchMock).url).searchParams.get("quality"),
    ).toBe("90");
  });

  it("cancels the underlying reader when aborted mid-stream (pcap)", async () => {
    let cancelled = false;
    // A stream that would emit forever; abort should stop it.
    const body = new ReadableStream<Uint8Array>({
      pull(controller) {
        controller.enqueue(new Uint8Array([0]));
      },
      cancel() {
        cancelled = true;
      },
    });
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "application/vnd.tcpdump.pcap" },
        }),
    );
    const client = makeClient(fetchMock);
    const controller = new AbortController();
    const stream = await client
      .device("udid-1")
      .pcap({ timeout: 5, signal: controller.signal });

    let count = 0;
    for await (const _chunk of stream) {
      void _chunk;
      count++;
      if (count >= 3) {
        controller.abort();
        break;
      }
    }
    expect(count).toBe(3);
    expect(cancelled).toBe(true);
  });

  it("stops immediately when the signal is already aborted", async () => {
    let cancelled = false;
    const body = chunkedStream([[1], [2]], () => {
      cancelled = true;
    });
    const fetchMock = vi.fn(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "content-type": "application/octet-stream" },
        }),
    );
    const client = makeClient(fetchMock);
    const controller = new AbortController();
    controller.abort();
    const stream = await client
      .device("udid-1")
      .screenshotStream({ signal: controller.signal });

    const received: number[] = [];
    for await (const chunk of stream) received.push(...chunk);
    expect(received).toEqual([]);
    expect(cancelled).toBe(true);
  });

  it("surfaces a non-2xx binary-stream response as IosApiError", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ error: "no device" }, 404));
    const client = makeClient(fetchMock);
    await expect(client.device("x").pcap()).rejects.toMatchObject({
      name: "IosApiError",
      status: 404,
      message: "no device",
    });
    await expect(client.device("x").pcap()).rejects.toBeInstanceOf(IosApiError);
  });

  it("openBinaryStream helper is exported and usable directly", async () => {
    const body = new ReadableStream<Uint8Array>({
      start(c) {
        c.enqueue(new Uint8Array([9]));
        c.close();
      },
    });
    const stream = await openBinaryStream(
      Promise.resolve({
        response: new Response(body, {
          status: 200,
          headers: { "content-type": "application/octet-stream" },
        }),
      }),
    );
    const received: number[] = [];
    for await (const chunk of stream) received.push(...chunk);
    expect(received).toEqual([9]);
  });
});
