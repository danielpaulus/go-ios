import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { IosClient } from "../src/index";
import { discoverBaseUrl, resolveBaseUrl } from "../src/discovery";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

function firstRequest(fetchMock: { mock: { calls: unknown[][] } }): Request {
  const call = fetchMock.mock.calls[0];
  if (!call) throw new Error("fetch was not called");
  return call[0] as Request;
}

/** Write a rest-api.json into `home` with the given baseUrl. */
function writeDescriptor(home: string, baseUrl: string): void {
  writeFileSync(
    join(home, "rest-api.json"),
    JSON.stringify({
      baseUrl,
      host: "127.0.0.1",
      port: 54321,
      pid: 12345,
      startedAt: "2026-08-11T15:00:00Z",
      tls: false,
    }),
  );
}

describe("discovery", () => {
  let home: string;
  const savedHome = process.env.GO_IOS_HOME;
  const savedBaseUrl = process.env.GO_IOS_BASE_URL;

  beforeEach(() => {
    home = mkdtempSync(join(tmpdir(), "go-ios-home-"));
    process.env.GO_IOS_HOME = home;
    delete process.env.GO_IOS_BASE_URL;
  });

  afterEach(() => {
    rmSync(home, { recursive: true, force: true });
    if (savedHome === undefined) delete process.env.GO_IOS_HOME;
    else process.env.GO_IOS_HOME = savedHome;
    if (savedBaseUrl === undefined) delete process.env.GO_IOS_BASE_URL;
    else process.env.GO_IOS_BASE_URL = savedBaseUrl;
  });

  it("reads baseUrl from <home>/rest-api.json", () => {
    writeDescriptor(home, "http://127.0.0.1:54321");
    expect(discoverBaseUrl()).toBe("http://127.0.0.1:54321");
  });

  it("IosClient() with no baseUrl uses the discovered daemon", async () => {
    writeDescriptor(home, "http://127.0.0.1:54321");
    const fetchMock = vi.fn(async () =>
      jsonResponse({ deviceList: [{ deviceID: 1, properties: { serialNumber: "u" } }] }),
    );
    const client = new IosClient({ fetch: fetchMock as unknown as typeof fetch });
    await client.devices.list();
    expect(firstRequest(fetchMock).url).toBe("http://127.0.0.1:54321/api/v1/list");
  });

  it("explicit baseUrl overrides discovery", async () => {
    writeDescriptor(home, "http://127.0.0.1:54321");
    const fetchMock = vi.fn(async () =>
      jsonResponse({ deviceList: [] }),
    );
    const client = new IosClient({
      baseUrl: "http://explicit.test",
      fetch: fetchMock as unknown as typeof fetch,
    });
    await client.devices.list();
    expect(firstRequest(fetchMock).url).toBe("http://explicit.test/api/v1/list");
  });

  it("GO_IOS_BASE_URL is used when no explicit option (over discovery)", async () => {
    writeDescriptor(home, "http://127.0.0.1:54321");
    process.env.GO_IOS_BASE_URL = "http://from-env.test";
    const fetchMock = vi.fn(async () => jsonResponse({ deviceList: [] }));
    const client = new IosClient({ fetch: fetchMock as unknown as typeof fetch });
    await client.devices.list();
    expect(firstRequest(fetchMock).url).toBe("http://from-env.test/api/v1/list");
  });

  it("explicit option beats GO_IOS_BASE_URL", () => {
    process.env.GO_IOS_BASE_URL = "http://from-env.test";
    expect(resolveBaseUrl("http://explicit.test")).toBe("http://explicit.test");
  });

  it("missing discovery file throws a clear error", () => {
    // temp home exists but has no rest-api.json
    expect(() => new IosClient()).toThrow(
      /no local go-ios REST daemon found at .*rest-api\.json; start the go-ios REST API or pass a baseUrl/,
    );
  });

  it("unparseable discovery file throws a clear error", () => {
    writeFileSync(join(home, "rest-api.json"), "{ not json");
    expect(() => discoverBaseUrl()).toThrow(/no local go-ios REST daemon found/);
  });

  it("discovery file without baseUrl throws a clear error", () => {
    writeFileSync(join(home, "rest-api.json"), JSON.stringify({ port: 1 }));
    expect(() => discoverBaseUrl()).toThrow(/no local go-ios REST daemon found/);
  });
});
