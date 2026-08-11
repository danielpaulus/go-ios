package com.github.danielpaulus.goios;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.danielpaulus.goios.generated.invoker.JSON;
import com.github.danielpaulus.goios.generated.model.GenericResponse;
import com.github.danielpaulus.goios.stream.BinaryStream;
import com.github.danielpaulus.goios.stream.EventDecoder;
import com.github.danielpaulus.goios.stream.SseReader;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.UncheckedIOException;
import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;

/**
 * Thin transport helper over {@link java.net.http.HttpClient} shared by the
 * facade. Handles bearer auth, query building, JSON (de)serialization via the
 * generated {@link JSON} mapper, raw-byte and {@code application/octet-stream}
 * bodies, {@code multipart/form-data} uploads, SSE line streaming, and raw
 * binary streaming. All non-2xx responses raise {@link IosApiException}.
 *
 * <p>The facade deliberately drives HTTP directly (rather than through the
 * generated {@code DefaultApi}) so that query-parameter spelling, multipart
 * boundaries and octet-stream bodies match the wire format the go-ios server
 * expects byte-for-byte.
 */
final class RawHttp implements AutoCloseable {

    static final ObjectMapper MAPPER = new JSON().getMapper();
    private static final String API_PREFIX = "/api/v1";

    private final String baseUrl;
    private final String apiKey;
    private final HttpClient client;
    private final Duration timeout;

    RawHttp(String baseUrl, String apiKey, HttpClient client, Duration timeout) {
        this.baseUrl = stripTrailingSlash(baseUrl);
        this.apiKey = apiKey;
        this.client = client;
        this.timeout = timeout;
    }

    private static String stripTrailingSlash(String s) {
        return s.endsWith("/") ? s.substring(0, s.length() - 1) : s;
    }

    // -- URL / query -------------------------------------------------------

    /** Build an absolute URI from an API-relative suffix (already {@code /api/v1}-prefixed by callers). */
    URI uri(String suffix, Map<String, String> query) {
        StringBuilder sb = new StringBuilder(baseUrl).append(API_PREFIX).append(suffix);
        if (query != null && !query.isEmpty()) {
            sb.append('?');
            boolean first = true;
            for (Map.Entry<String, String> e : query.entrySet()) {
                if (e.getValue() == null) {
                    continue;
                }
                if (!first) {
                    sb.append('&');
                }
                first = false;
                sb.append(enc(e.getKey())).append('=').append(enc(e.getValue()));
            }
        }
        return URI.create(sb.toString());
    }

    private static String enc(String s) {
        return URLEncoder.encode(s, StandardCharsets.UTF_8);
    }

    private HttpRequest.Builder base(URI uri) {
        HttpRequest.Builder b = HttpRequest.newBuilder(uri).timeout(timeout);
        if (apiKey != null && !apiKey.isBlank()) {
            b.header("Authorization", "Bearer " + apiKey);
        }
        return b;
    }

    // -- request helpers ---------------------------------------------------

    private HttpResponse<byte[]> send(HttpRequest req) {
        try {
            HttpResponse<byte[]> resp = client.send(req, HttpResponse.BodyHandlers.ofByteArray());
            if (resp.statusCode() >= 300) {
                throw error(resp);
            }
            return resp;
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new RuntimeException("request interrupted", e);
        }
    }

    private IosApiException error(HttpResponse<byte[]> resp) {
        String body = resp.body() == null ? "" : new String(resp.body(), StandardCharsets.UTF_8);
        GenericResponse envelope = null;
        try {
            if (!body.isBlank() && body.trim().startsWith("{")) {
                envelope = MAPPER.readValue(body, GenericResponse.class);
            }
        } catch (Exception ignore) {
            // Non-JSON error body; keep the raw text only.
        }
        return new IosApiException(resp.statusCode(), body, envelope);
    }

    // -- JSON reads/writes -------------------------------------------------

    <T> T getJson(String suffix, Map<String, String> query, Class<T> type) {
        return decode(send(base(uri(suffix, query)).GET().build()).body(), type);
    }

    <T> T getJson(String suffix, Map<String, String> query, TypeReference<T> type) {
        return decode(send(base(uri(suffix, query)).GET().build()).body(), type);
    }

    byte[] getBytes(String suffix, Map<String, String> query) {
        return send(base(uri(suffix, query)).GET().build()).body();
    }

    <T> T requestJson(String method, String suffix, Map<String, String> query,
                      byte[] body, String contentType, Class<T> type) {
        HttpRequest.Builder b = base(uri(suffix, query));
        HttpRequest.BodyPublisher pub = body == null
                ? HttpRequest.BodyPublishers.noBody()
                : HttpRequest.BodyPublishers.ofByteArray(body);
        if (contentType != null && body != null) {
            b.header("Content-Type", contentType);
        }
        b.method(method, pub);
        return decode(send(b.build()).body(), type);
    }

    <T> T postJson(String suffix, Map<String, String> query, Object jsonBody, Class<T> type) {
        byte[] body = jsonBody == null ? null : encode(jsonBody);
        return requestJson("POST", suffix, query, body, body == null ? null : "application/json", type);
    }

    <T> T putJson(String suffix, Map<String, String> query, Object jsonBody, Class<T> type) {
        byte[] body = jsonBody == null ? null : encode(jsonBody);
        return requestJson("PUT", suffix, query, body, body == null ? null : "application/json", type);
    }

    <T> T deleteJson(String suffix, Map<String, String> query, Class<T> type) {
        return requestJson("DELETE", suffix, query, null, null, type);
    }

    // -- multipart ---------------------------------------------------------

    /** A single multipart part: either a file (bytes + filename) or a plain text field. */
    record Part(String name, String filename, byte[] content, String textValue) {
        static Part file(String name, String filename, byte[] content) {
            return new Part(name, filename, content, null);
        }

        static Part field(String name, String value) {
            return new Part(name, null, null, value);
        }
    }

    <T> T multipart(String method, String suffix, Map<String, String> query,
                    List<Part> parts, Class<T> type) {
        String boundary = "----goios" + Long.toHexString(new Random().nextLong());
        byte[] body = buildMultipart(parts, boundary);
        HttpRequest.Builder b = base(uri(suffix, query))
                .header("Content-Type", "multipart/form-data; boundary=" + boundary)
                .method(method, HttpRequest.BodyPublishers.ofByteArray(body));
        return decode(send(b.build()).body(), type);
    }

    /** Multipart returning the raw response bytes (e.g. sign endpoints returning a P12/IPA). */
    byte[] multipartBytes(String method, String suffix, Map<String, String> query, List<Part> parts) {
        String boundary = "----goios" + Long.toHexString(new Random().nextLong());
        byte[] body = buildMultipart(parts, boundary);
        HttpRequest.Builder b = base(uri(suffix, query))
                .header("Content-Type", "multipart/form-data; boundary=" + boundary)
                .method(method, HttpRequest.BodyPublishers.ofByteArray(body));
        return send(b.build()).body();
    }

    private static byte[] buildMultipart(List<Part> parts, String boundary) {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        try {
            for (Part p : parts) {
                if (p == null) {
                    continue;
                }
                out.write(("--" + boundary + "\r\n").getBytes(StandardCharsets.UTF_8));
                if (p.filename() != null) {
                    out.write(("Content-Disposition: form-data; name=\"" + p.name()
                            + "\"; filename=\"" + p.filename() + "\"\r\n").getBytes(StandardCharsets.UTF_8));
                    out.write("Content-Type: application/octet-stream\r\n\r\n".getBytes(StandardCharsets.UTF_8));
                    out.write(p.content() == null ? new byte[0] : p.content());
                } else {
                    out.write(("Content-Disposition: form-data; name=\"" + p.name() + "\"\r\n\r\n")
                            .getBytes(StandardCharsets.UTF_8));
                    out.write((p.textValue() == null ? "" : p.textValue()).getBytes(StandardCharsets.UTF_8));
                }
                out.write("\r\n".getBytes(StandardCharsets.UTF_8));
            }
            out.write(("--" + boundary + "--\r\n").getBytes(StandardCharsets.UTF_8));
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }
        return out.toByteArray();
    }

    // -- streaming (SSE) ---------------------------------------------------

    /** Open an SSE stream at {@code suffix}, decoding each frame with {@code decoder}. */
    SseReader sseStream(String suffix, Map<String, String> query,
                        EventDecoder decoder, boolean includeHeartbeats) {
        HttpRequest req = base(uri(suffix, query))
                .timeout(Duration.ofDays(3650)) // effectively no read timeout for long-lived streams
                .GET().build();
        HttpResponse<InputStream> resp;
        try {
            resp = client.send(req, HttpResponse.BodyHandlers.ofInputStream());
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new RuntimeException("request interrupted", e);
        }
        if (resp.statusCode() >= 300) {
            throw drainError(resp);
        }
        InputStream in = resp.body();
        Iterator<String> lines = new java.io.BufferedReader(
                new java.io.InputStreamReader(in, StandardCharsets.UTF_8)).lines().iterator();
        return new SseReader(lines, decoder, includeHeartbeats, () -> {
            try {
                in.close();
            } catch (IOException ignore) {
                // best-effort abort of the underlying connection
            }
        });
    }

    // -- streaming (binary) -----------------------------------------------

    /** Open a raw binary stream at {@code suffix} (UI video, MJPEG screenshots, pcap). */
    BinaryStream binaryStream(String suffix, Map<String, String> query) {
        HttpRequest req = base(uri(suffix, query))
                .timeout(Duration.ofDays(3650))
                .GET().build();
        HttpResponse<InputStream> resp;
        try {
            resp = client.send(req, HttpResponse.BodyHandlers.ofInputStream());
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new RuntimeException("request interrupted", e);
        }
        if (resp.statusCode() >= 300) {
            throw drainError(resp);
        }
        InputStream in = resp.body();
        String ct = resp.headers().firstValue("Content-Type").orElse(null);
        return new BinaryStream(in, ct, () -> { });
    }

    private IosApiException drainError(HttpResponse<InputStream> resp) {
        String body = "";
        try (InputStream in = resp.body()) {
            body = new String(in.readAllBytes(), StandardCharsets.UTF_8);
        } catch (IOException ignore) {
            // ignore
        }
        GenericResponse envelope = null;
        try {
            if (!body.isBlank() && body.trim().startsWith("{")) {
                envelope = MAPPER.readValue(body, GenericResponse.class);
            }
        } catch (Exception ignore) {
            // non-JSON
        }
        return new IosApiException(resp.statusCode(), body, envelope);
    }

    // -- (de)serialization -------------------------------------------------

    static byte[] encode(Object value) {
        try {
            return MAPPER.writeValueAsBytes(value);
        } catch (Exception e) {
            throw new IllegalArgumentException("failed to serialize request body", e);
        }
    }

    static <T> T decode(byte[] body, Class<T> type) {
        if (type == Void.class) {
            return null;
        }
        try {
            if (body == null || body.length == 0) {
                return type == String.class ? type.cast("") : null;
            }
            if (type == String.class) {
                return type.cast(new String(body, StandardCharsets.UTF_8));
            }
            if (type == byte[].class) {
                return type.cast(body);
            }
            return MAPPER.readValue(body, type);
        } catch (Exception e) {
            throw new IllegalStateException("failed to decode response as " + type.getSimpleName()
                    + ": " + e.getMessage(), e);
        }
    }

    static <T> T decode(byte[] body, TypeReference<T> type) {
        try {
            if (body == null || body.length == 0) {
                return null;
            }
            return MAPPER.readValue(body, type);
        } catch (Exception e) {
            throw new IllegalStateException("failed to decode response: " + e.getMessage(), e);
        }
    }

    /** Ordered mutable query map that skips null values on build. */
    static Map<String, String> query() {
        return new LinkedHashMap<>();
    }

    /** Immutable list helper for parts, skipping nulls. */
    static List<RawHttp.Part> parts(RawHttp.Part... items) {
        List<RawHttp.Part> list = new ArrayList<>();
        for (RawHttp.Part p : items) {
            if (p != null) {
                list.add(p);
            }
        }
        return list;
    }

    @Override
    public void close() {
        // java.net.http.HttpClient (JDK 17) has no explicit close; GC/keepalive handles it.
    }
}
