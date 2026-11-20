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
 *
 * <p>{@code baseUrl} is optional. When it is not set explicitly, the builder
 * resolves the daemon endpoint in this order: an explicit
 * {@link Builder#baseUrl(String)}; the {@code GO_IOS_BASE_URL} env var; then
 * discovery of a locally running daemon via {@code <home>/rest-api.json} (see
 * {@link Discovery}). If none is available, {@link #build()} throws an
 * {@link IosDiscoveryException} pointing at the expected discovery path.
 */
public final class IosClient implements AutoCloseable {

    private final RawHttp http;
    private final Devices devices;
    private final Tunnels tunnels;
    private final Sign sign;
    private final Prepare prepare;

    private IosClient(Builder b) {
        HttpClient client = b.httpClient != null ? b.httpClient
                : HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(10)).build();
        String baseUrl = b.baseUrl != null && !b.baseUrl.isBlank()
                ? b.baseUrl
                : b.discovery.resolveBaseUrl();
        this.http = new RawHttp(baseUrl, b.apiKey, client, b.timeout);
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
        private String baseUrl;
        private String apiKey;
        private Duration timeout = Duration.ofSeconds(30);
        private HttpClient httpClient;
        private Discovery discovery = Discovery.system();

        /**
         * Pin the daemon origin (e.g. {@code http://localhost:8080}); {@code /api/v1}
         * is appended automatically. Optional: when unset the builder falls back to
         * the {@code GO_IOS_BASE_URL} env var and then to {@link Discovery discovery}
         * of a local daemon.
         */
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

        /** Override the discovery seam (test-only). */
        Builder discovery(Discovery discovery) {
            this.discovery = discovery;
            return this;
        }

        /**
         * Build the client, resolving {@code baseUrl} if it was not set explicitly.
         *
         * @throws IosDiscoveryException if no {@code baseUrl} is set, no
         *                               {@code GO_IOS_BASE_URL} env var is present, and
         *                               no local daemon discovery file can be read.
         */
        public IosClient build() {
            return new IosClient(this);
        }
    }
}
