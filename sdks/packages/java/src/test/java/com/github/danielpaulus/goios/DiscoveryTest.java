package com.github.danielpaulus.goios;

import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Function;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Tests for ephemeral-daemon discovery: home-dir resolution, the discovery file,
 * env precedence, and the {@link IosClient.Builder} resolution order.
 *
 * <p>Java can't mutate the process environment at runtime, so instead of setting
 * {@code GO_IOS_HOME} as a real env var these tests drive {@link Discovery}
 * through its injectable env / system-property lookups (the same seam the builder
 * uses) with a {@link TempDir temp} home.
 */
class DiscoveryTest {

    /** Build a Discovery whose GO_IOS_HOME points at {@code home}, plus extra env. */
    private static Discovery discoveryWithHome(Path home, Map<String, String> extraEnv) {
        Map<String, String> env = new HashMap<>();
        env.put("GO_IOS_HOME", home.toString());
        if (extraEnv != null) {
            env.putAll(extraEnv);
        }
        Function<String, String> props = k -> "user.home".equals(k) ? "/nonexistent-home" : null;
        return Discovery.of(env::get, props);
    }

    private static void writeDiscoveryFile(Path home, String baseUrl) throws IOException {
        Files.createDirectories(home);
        Files.writeString(home.resolve(Discovery.DISCOVERY_FILE),
                "{\"baseUrl\":\"" + baseUrl + "\",\"host\":\"127.0.0.1\",\"port\":54321,"
                        + "\"pid\":12345,\"startedAt\":\"2026-08-11T15:00:00Z\",\"tls\":false}");
    }

    // -- home-dir resolution ----------------------------------------------

    @Test
    void homeUsesGoIosHomeEnvWhenSet(@TempDir Path tmp) {
        Discovery d = discoveryWithHome(tmp, null);
        assertEquals(tmp, d.home());
        assertEquals(tmp.resolve("rest-api.json"), d.discoveryFile());
    }

    @Test
    void homeFallsBackToUserHomeDotGoIos() {
        Discovery d = Discovery.of(k -> null, k -> "user.home".equals(k) ? "/home/tester" : null);
        assertEquals(Path.of("/home/tester", ".go-ios"), d.home());
    }

    @Test
    void blankGoIosHomeFallsBackToUserHome() {
        Discovery d = Discovery.of(
                k -> "GO_IOS_HOME".equals(k) ? "  " : null,
                k -> "user.home".equals(k) ? "/home/tester" : null);
        assertEquals(Path.of("/home/tester", ".go-ios"), d.home());
    }

    // -- discovery file ----------------------------------------------------

    @Test
    void readsBaseUrlFromDiscoveryFile(@TempDir Path tmp) throws IOException {
        writeDiscoveryFile(tmp, "http://127.0.0.1:54321");
        Discovery d = discoveryWithHome(tmp, null);
        assertEquals("http://127.0.0.1:54321", d.resolveBaseUrl());
    }

    @Test
    void missingDiscoveryFileThrowsClearException(@TempDir Path tmp) {
        Discovery d = discoveryWithHome(tmp, null);
        IosDiscoveryException ex = assertThrows(IosDiscoveryException.class, d::resolveBaseUrl);
        assertTrue(ex.getMessage().contains(tmp.resolve("rest-api.json").toString()),
                "message must name the expected path: " + ex.getMessage());
        assertTrue(ex.getMessage().contains("no local go-ios REST daemon found"), ex.getMessage());
    }

    @Test
    void malformedDiscoveryFileThrowsClearException(@TempDir Path tmp) throws IOException {
        Files.createDirectories(tmp);
        Files.writeString(tmp.resolve(Discovery.DISCOVERY_FILE), "not json at all");
        Discovery d = discoveryWithHome(tmp, null);
        IosDiscoveryException ex = assertThrows(IosDiscoveryException.class, d::resolveBaseUrl);
        assertTrue(ex.getMessage().contains("no local go-ios REST daemon found"), ex.getMessage());
    }

    @Test
    void discoveryFileWithoutBaseUrlThrows(@TempDir Path tmp) throws IOException {
        Files.createDirectories(tmp);
        Files.writeString(tmp.resolve(Discovery.DISCOVERY_FILE), "{\"host\":\"127.0.0.1\",\"port\":1}");
        Discovery d = discoveryWithHome(tmp, null);
        assertThrows(IosDiscoveryException.class, d::resolveBaseUrl);
    }

    // -- env precedence ----------------------------------------------------

    @Test
    void goIosBaseUrlEnvTakesPrecedenceOverDiscoveryFile(@TempDir Path tmp) throws IOException {
        writeDiscoveryFile(tmp, "http://127.0.0.1:54321");
        Discovery d = discoveryWithHome(tmp, Map.of("GO_IOS_BASE_URL", "http://10.0.0.5:9000"));
        assertEquals("http://10.0.0.5:9000", d.resolveBaseUrl());
    }

    @Test
    void goIosBaseUrlEnvUsedWhenNoDiscoveryFile(@TempDir Path tmp) {
        Discovery d = discoveryWithHome(tmp, Map.of("GO_IOS_BASE_URL", "http://10.0.0.5:9000"));
        assertEquals("http://10.0.0.5:9000", d.resolveBaseUrl());
    }

    // -- builder resolution order (end to end over a real HttpServer) ------

    @Test
    void builderWithoutBaseUrlUsesDiscoveredBaseUrl(@TempDir Path tmp) throws IOException {
        try (Stub stub = new Stub()) {
            writeDiscoveryFile(tmp, stub.baseUrl());
            try (IosClient c = IosClient.builder()
                    .discovery(discoveryWithHome(tmp, null))
                    .build()) {
                assertEquals(1, c.devices().list().size());
            }
        }
    }

    @Test
    void explicitBaseUrlOverridesDiscovery(@TempDir Path tmp) throws IOException {
        try (Stub stub = new Stub()) {
            // Discovery file points somewhere unreachable; explicit baseUrl must win.
            writeDiscoveryFile(tmp, "http://127.0.0.1:1");
            try (IosClient c = IosClient.builder()
                    .baseUrl(stub.baseUrl())
                    .discovery(discoveryWithHome(tmp, null))
                    .build()) {
                assertEquals(1, c.devices().list().size());
            }
        }
    }

    @Test
    void goIosBaseUrlEnvUsedByBuilder(@TempDir Path tmp) throws IOException {
        try (Stub stub = new Stub()) {
            // No discovery file; only the env var is set.
            try (IosClient c = IosClient.builder()
                    .discovery(discoveryWithHome(tmp, Map.of("GO_IOS_BASE_URL", stub.baseUrl())))
                    .build()) {
                assertEquals(1, c.devices().list().size());
            }
        }
    }

    @Test
    void builderWithoutBaseUrlAndNoDaemonThrowsClearException(@TempDir Path tmp) {
        IosDiscoveryException ex = assertThrows(IosDiscoveryException.class, () ->
                IosClient.builder().discovery(discoveryWithHome(tmp, null)).build());
        assertTrue(ex.getMessage().contains(tmp.resolve("rest-api.json").toString()), ex.getMessage());
    }

    /** Minimal in-process daemon stub serving {@code GET /api/v1/list}. */
    private static final class Stub implements AutoCloseable {
        private final HttpServer server;

        Stub() throws IOException {
            server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
            server.createContext("/api/v1/list", ex -> {
                byte[] body = ("{\"deviceList\":[{\"deviceID\":1,"
                        + "\"properties\":{\"serialNumber\":\"UDID-A\"}}]}")
                        .getBytes(StandardCharsets.UTF_8);
                ex.getResponseHeaders().set("Content-Type", "application/json");
                ex.sendResponseHeaders(200, body.length);
                try (OutputStream os = ex.getResponseBody()) {
                    os.write(body);
                }
            });
            server.setExecutor(null);
            server.start();
        }

        String baseUrl() {
            return "http://127.0.0.1:" + server.getAddress().getPort();
        }

        @Override
        public void close() {
            server.stop(0);
        }
    }
}
