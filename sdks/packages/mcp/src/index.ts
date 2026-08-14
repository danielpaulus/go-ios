/**
 * go-ios MCP server entry point.
 *
 * Transports:
 *  - stdio (default): for local agent clients (Claude Desktop, etc.).
 *  - Streamable HTTP: for remote clients per the current MCP spec (SSE is the
 *    response-streaming mode within Streamable HTTP).
 *
 * Selection: `--http` / `--stdio` flag, or GO_IOS_MCP_TRANSPORT=http|stdio.
 */
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import { randomUUID } from "node:crypto";
import { createServer as createHttpServer } from "node:http";
import { configFromEnv, type ServerConfig } from "./config.js";
import { createServer, SERVER_NAME, SERVER_VERSION } from "./server.js";

function parseArgs(argv: string[], base: ServerConfig): ServerConfig {
  const cfg = { ...base };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    switch (a) {
      case "--http":
        cfg.transport = "http";
        break;
      case "--stdio":
        cfg.transport = "stdio";
        break;
      case "--port":
        cfg.httpPort = Number.parseInt(argv[++i] ?? "", 10) || cfg.httpPort;
        break;
      case "--host":
        cfg.httpHost = argv[++i] ?? cfg.httpHost;
        break;
      case "--base-url":
        cfg.baseUrl = (argv[++i] ?? cfg.baseUrl).replace(/\/+$/, "");
        break;
      case "--help":
      case "-h":
        printHelp();
        process.exit(0);
    }
  }
  return cfg;
}

function printHelp(): void {
  process.stderr.write(
    `${SERVER_NAME} v${SERVER_VERSION} — MCP server for go-ios\n\n` +
      `Usage: go-ios-mcp [--stdio | --http] [options]\n\n` +
      `Transports:\n` +
      `  --stdio           Serve over stdio (default; for local agent clients)\n` +
      `  --http            Serve over Streamable HTTP\n\n` +
      `Options:\n` +
      `  --port <n>        HTTP port (default 3000; env GO_IOS_MCP_HTTP_PORT)\n` +
      `  --host <h>        HTTP host (default 127.0.0.1; env GO_IOS_MCP_HTTP_HOST)\n` +
      `  --base-url <url>  go-ios daemon base URL (env GO_IOS_BASE_URL)\n` +
      `  -h, --help        Show this help\n\n` +
      `Environment:\n` +
      `  GO_IOS_BASE_URL   go-ios daemon base URL (default http://localhost:8080)\n` +
      `  GO_IOS_API_KEY    Bearer token for the daemon\n`,
  );
}

async function runStdio(config: ServerConfig): Promise<void> {
  const server = createServer(config);
  const transport = new StdioServerTransport();
  await server.connect(transport);
  process.stderr.write(
    `${SERVER_NAME} listening on stdio (daemon: ${config.baseUrl})\n`,
  );
}

async function runHttp(config: ServerConfig): Promise<void> {
  // One MCP server + transport per initialized session (stateful mode).
  const transports = new Map<string, StreamableHTTPServerTransport>();

  const http = createHttpServer(async (req, res) => {
    if (!req.url) {
      res.writeHead(400).end();
      return;
    }
    const url = new URL(req.url, `http://${req.headers.host ?? "localhost"}`);
    if (url.pathname !== "/mcp") {
      res.writeHead(404, { "Content-Type": "application/json" }).end(
        JSON.stringify({ error: "not found; MCP endpoint is /mcp" }),
      );
      return;
    }

    const sessionId = req.headers["mcp-session-id"] as string | undefined;
    let transport = sessionId ? transports.get(sessionId) : undefined;

    // Read the body (SDK expects the parsed body passed in for POST).
    let body: unknown;
    if (req.method === "POST") {
      const chunks: Buffer[] = [];
      for await (const c of req) chunks.push(c as Buffer);
      const raw = Buffer.concat(chunks).toString("utf8");
      body = raw ? JSON.parse(raw) : undefined;
    }

    if (!transport) {
      // New session: create a fresh server + transport.
      transport = new StreamableHTTPServerTransport({
        sessionIdGenerator: () => randomUUID(),
        onsessioninitialized: (id) => {
          transports.set(id, transport!);
        },
      });
      transport.onclose = () => {
        if (transport!.sessionId) transports.delete(transport!.sessionId);
      };
      const server = createServer(config);
      await server.connect(transport);
    }

    await transport.handleRequest(req, res, body);
  });

  await new Promise<void>((resolve) => {
    http.listen(config.httpPort, config.httpHost, resolve);
  });
  process.stderr.write(
    `${SERVER_NAME} listening on http://${config.httpHost}:${config.httpPort}/mcp ` +
      `(daemon: ${config.baseUrl})\n`,
  );
}

async function main(): Promise<void> {
  const config = parseArgs(process.argv.slice(2), configFromEnv());
  if (config.transport === "http") {
    await runHttp(config);
  } else {
    await runStdio(config);
  }
}

main().catch((err) => {
  process.stderr.write(`fatal: ${err instanceof Error ? err.stack : String(err)}\n`);
  process.exit(1);
});
