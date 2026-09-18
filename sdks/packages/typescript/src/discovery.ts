import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";

/**
 * Shape of `<home>/rest-api.json`, written by the go-ios REST daemon after it
 * binds. `baseUrl` is authoritative (scheme + host + port); the rest is
 * informational. See DISCOVERY-CONTRACT.md.
 */
export interface RestApiDescriptor {
  baseUrl: string;
  host?: string;
  port?: number;
  pid?: number;
  startedAt?: string;
  tls?: boolean;
}

/**
 * Resolve the go-ios home directory: `GO_IOS_HOME` if set and non-empty,
 * otherwise `<userHome>/.go-ios` (which is `%USERPROFILE%\.go-ios` on Windows).
 */
export function goIosHome(): string {
  const fromEnv = process.env.GO_IOS_HOME;
  if (fromEnv && fromEnv.length > 0) {
    return fromEnv;
  }
  return join(homedir(), ".go-ios");
}

/** Absolute path to the discovery file `<home>/rest-api.json`. */
export function discoveryFilePath(): string {
  return join(goIosHome(), "rest-api.json");
}

/**
 * Read and parse `<home>/rest-api.json`, returning its `baseUrl`.
 *
 * @throws if the file is missing/unreadable/unparseable or lacks a `baseUrl`,
 *   with a clear message telling the caller to start the go-ios REST daemon or
 *   pass an explicit `baseUrl`.
 */
export function discoverBaseUrl(): string {
  const path = discoveryFilePath();
  const notFound = new Error(
    `no local go-ios REST daemon found at ${path}; start the go-ios REST API or pass a baseUrl`,
  );

  let raw: string;
  try {
    raw = readFileSync(path, "utf8");
  } catch {
    throw notFound;
  }

  let parsed: RestApiDescriptor;
  try {
    parsed = JSON.parse(raw) as RestApiDescriptor;
  } catch {
    throw notFound;
  }

  if (
    !parsed ||
    typeof parsed.baseUrl !== "string" ||
    parsed.baseUrl.length === 0
  ) {
    throw notFound;
  }

  return parsed.baseUrl;
}

/**
 * Resolve the base URL for an {@link IosClient} in the contract's precedence:
 *   1. explicit `baseUrl` option → verbatim (skips discovery, for remote daemons)
 *   2. `GO_IOS_BASE_URL` env
 *   3. discovery file (`<home>/rest-api.json` → `baseUrl`)
 *   4. none → clear throw
 */
export function resolveBaseUrl(explicit?: string): string {
  if (explicit && explicit.length > 0) {
    return explicit;
  }
  const fromEnv = process.env.GO_IOS_BASE_URL;
  if (fromEnv && fromEnv.length > 0) {
    return fromEnv;
  }
  return discoverBaseUrl();
}
