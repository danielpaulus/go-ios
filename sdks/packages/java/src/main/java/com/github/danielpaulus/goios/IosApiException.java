package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.GenericResponse;

/**
 * Thrown when the go-ios REST API returns a non-2xx status. Carries the HTTP
 * status code, the raw response body, and — when the body is a JSON error
 * envelope — a decoded {@link GenericResponse} accessible via {@link #errorBody()}.
 */
public class IosApiException extends RuntimeException {

    private final int statusCode;
    private final String rawBody;
    private final transient GenericResponse errorBody;

    public IosApiException(int statusCode, String rawBody, GenericResponse errorBody) {
        super(buildMessage(statusCode, rawBody, errorBody));
        this.statusCode = statusCode;
        this.rawBody = rawBody;
        this.errorBody = errorBody;
    }

    private static String buildMessage(int statusCode, String rawBody, GenericResponse errorBody) {
        String detail = null;
        if (errorBody != null) {
            detail = errorBody.getError() != null ? errorBody.getError() : errorBody.getMessage();
        }
        if (detail == null) {
            detail = rawBody;
        }
        return "go-ios API error " + statusCode + (detail == null || detail.isBlank() ? "" : ": " + detail);
    }

    /** The HTTP status code. */
    public int statusCode() {
        return statusCode;
    }

    /** The raw (undecoded) response body, if any. */
    public String rawBody() {
        return rawBody;
    }

    /**
     * The decoded error envelope ({@code {"error": ...}} / {@code {"message": ...}}),
     * or {@code null} if the body was not a JSON error object.
     */
    public GenericResponse errorBody() {
        return errorBody;
    }
}
