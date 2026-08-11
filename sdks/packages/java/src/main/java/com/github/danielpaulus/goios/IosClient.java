package com.github.danielpaulus.goios;

import java.net.http.HttpClient;
import java.time.Duration;

/**
 * Ergonomic synchronous client for the go-ios REST API.
 *
 * <pre>{@code
 * try (IosClient client = IosClient.builder()
 *         .baseUrl("http://localhost:60105")
 *         .apiKey("secret")
 *         .build()) {
 *     for (DeviceEntry d : client.devices().list()) { ... }
 *     Device dev = client.device(udid);
 *     BatteryInfo b = dev.battery();
 *     byte[] png = dev.screenshot();
 *     try (SseReader syslog = dev.syslog()) { for (var ev : syslog) { ... } }
 * }
 * }</pre>
 *
 * <p>Mirrors the public shape of the TypeScript/Python/C# SDKs. When an
 * {@code apiKey} is set it is sent as {@code Authorization: Bearer <key>};
 * a server started with {@code --disable-auth} needs none.
 */
public final class IosClient implements AutoCloseable {

    /** Default local daemon endpoint. */
    public static final String DEFAULT_BASE_URL = "http://localhost:60105";

    private final RawHttp http;
    private final Devices devices;
    private final Tunnels tunnels;
    private final Sign sign;
    private final Prepare prepare;

    private IosClient(Builder b) {
        HttpClient client = b.httpClient != null ? b.httpClient
                : HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(10)).build();
        this.http = new RawHttp(b.baseUrl, b.apiKey, client, b.timeout);
        this.devices = new Devices(http);
        this.tunnels = new Tunnels(http);
        this.sign = new Sign(http);
        this.prepare = new Prepare(http);
    }

    public static Builder builder() {
        return new Builder();
    }

    /** Fleet-level device operations ({@code GET /list}). */
    public Devices devices() {
        return devices;
    }

    /** userspace-tunnel (RemoteXPC) management (iOS 17+). */
    public Tunnels tunnels() {
        return tunnels;
    }

    /** Host-scoped app-signing operations ({@code /sign/*}). */
    public Sign sign() {
        return sign;
    }

    /** Host-scoped device-preparation helpers ({@code /prepare/*}). */
    public Prepare prepare() {
        return prepare;
    }

    /** Return a {@link Device} handle scoped to {@code udid}. */
    public Device device(String udid) {
        return new Device(http, udid);
    }

    @Override
    public void close() {
        http.close();
    }

    /** Builder for {@link IosClient}. */
    public static final class Builder {
        private String baseUrl = DEFAULT_BASE_URL;
        private String apiKey;
        private Duration timeout = Duration.ofSeconds(30);
        private HttpClient httpClient;

        public Builder baseUrl(String baseUrl) {
            this.baseUrl = baseUrl;
            return this;
        }

        public Builder apiKey(String apiKey) {
            this.apiKey = apiKey;
            return this;
        }

        /** Per-request timeout for non-streaming calls (streams are exempt). */
        public Builder timeout(Duration timeout) {
            this.timeout = timeout;
            return this;
        }

        /** Bring your own configured {@link HttpClient} (e.g. custom TLS). */
        public Builder httpClient(HttpClient httpClient) {
            this.httpClient = httpClient;
            return this;
        }

        public IosClient build() {
            return new IosClient(this);
        }
    }
}
