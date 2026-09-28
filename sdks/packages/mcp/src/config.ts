/**
 * Runtime configuration for the go-ios MCP server, resolved from environment
 * variables (with CLI-flag overrides applied by the caller before use).
 */
export interface ServerConfig {
  /** Base URL of the go-ios REST daemon, e.g. http://localhost:8080. No trailing slash. */
  baseUrl: string;
  /** Bearer token sent as `Authorization: Bearer <token>` when set. */
  apiKey?: string;
  /** Transport the MCP server listens on. */
  transport: "stdio" | "http";
  /** Port for the Streamable HTTP transport (ignored for stdio). */
  httpPort: number;
  /** Host/interface for the Streamable HTTP transport (ignored for stdio). */
  httpHost: string;
}

const DEFAULT_BASE_URL = "http://localhost:8080";
const DEFAULT_HTTP_PORT = 3000;
const DEFAULT_HTTP_HOST = "127.0.0.1";

/**
 * Build a ServerConfig from environment variables. Recognized vars:
 *  - GO_IOS_BASE_URL   (default http://localhost:8080)
 *  - GO_IOS_API_KEY    (bearer token; optional but strongly encouraged)
 *  - GO_IOS_MCP_TRANSPORT   ("stdio" | "http", default "stdio")
 *  - GO_IOS_MCP_HTTP_PORT   (default 3000)
 *  - GO_IOS_MCP_HTTP_HOST   (default 127.0.0.1)
 */
export function configFromEnv(env: NodeJS.ProcessEnv = process.env): ServerConfig {
  const baseUrl = (env.GO_IOS_BASE_URL ?? DEFAULT_BASE_URL).replace(/\/+$/, "");
  const apiKey = env.GO_IOS_API_KEY?.trim() || undefined;
  const transport = env.GO_IOS_MCP_TRANSPORT === "http" ? "http" : "stdio";
  const httpPort = env.GO_IOS_MCP_HTTP_PORT
    ? Number.parseInt(env.GO_IOS_MCP_HTTP_PORT, 10)
    : DEFAULT_HTTP_PORT;
  const httpHost = env.GO_IOS_MCP_HTTP_HOST ?? DEFAULT_HTTP_HOST;

  return { baseUrl, apiKey, transport, httpPort, httpHost };
}
