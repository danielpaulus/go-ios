import { defineConfig } from "@hey-api/openapi-ts";

// Generates the low-level typed client from the canonical OpenAPI 3.1 spec.
// The ergonomic facade in src/ wraps this generated client. SSE endpoints are
// modeled as bare `text/event-stream` in 3.1, so the facade hand-writes typed
// SSE dispatch from the `x-sse-events` extension (see src/streaming.ts).
export default defineConfig({
  input: "../../spec/openapi/openapi.yaml",
  output: {
    path: "src/generated",
    format: false,
    lint: false,
  },
  plugins: [
    {
      name: "@hey-api/client-fetch",
      runtimeConfigPath: "./src/client-config.ts",
    },
  ],
});
