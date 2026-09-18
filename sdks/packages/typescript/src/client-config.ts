import type { CreateClientConfig } from "./generated/client.gen";

// Runtime config hook consumed by the generated client. The facade calls
// `createClient` with its own baseUrl/auth, so this only supplies inert
// defaults; keeping it here keeps the generated client self-contained.
export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
});
