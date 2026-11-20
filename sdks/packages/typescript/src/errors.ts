import type { GenericResponse } from "./generated/types.gen";

/**
 * Thrown for any non-2xx response from the go-ios REST server. Carries the HTTP
 * status and, when the body was the `GenericResponse` error envelope, the parsed
 * `message`/`error` fields.
 */
export class IosApiError extends Error {
  readonly status: number;
  readonly body?: GenericResponse | unknown;

  constructor(status: number, message: string, body?: GenericResponse | unknown) {
    super(message);
    this.name = "IosApiError";
    this.status = status;
    this.body = body;
  }
}

/**
 * Normalizes the generated client's `{ data, error, response }` result into
 * either the data (on success) or a thrown {@link IosApiError}. The generated
 * client is configured with `throwOnError: false`, so we do the mapping here to
 * keep a consistent error surface across every facade method.
 */
export function unwrap<T>(result: {
  data?: T;
  error?: unknown;
  response: Response;
}): T {
  const { data, error, response } = result;
  if (response.ok && error === undefined) {
    return data as T;
  }
  const envelope = (error ?? data) as Partial<GenericResponse> | undefined;
  const message =
    (envelope && typeof envelope === "object" && "error" in envelope && envelope.error) ||
    (envelope && typeof envelope === "object" && "message" in envelope && envelope.message) ||
    `go-ios request failed with status ${response.status}`;
  throw new IosApiError(response.status, String(message), error ?? data);
}
