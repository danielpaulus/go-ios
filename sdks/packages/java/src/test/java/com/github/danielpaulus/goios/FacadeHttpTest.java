package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.AppInfo;
import com.github.danielpaulus.goios.generated.model.DeviceEntry;
import com.github.danielpaulus.goios.stream.SseReader;
import com.github.danielpaulus.goios.stream.SyslogEvent;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.*;

/** Facade tests against an in-process {@link HttpServer} stub. */
class FacadeHttpTest {

    private HttpServer server;
    private String baseUrl;
    private final List<String> seenAuthHeaders = new CopyOnWriteArrayList<>();
    private final ConcurrentHashMap<String, String> lastQuery = new ConcurrentHashMap<>();

    @BeforeEach
    void start() throws IOException {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);

        server.createContext("/api/v1/list", ex -> {
            record(ex);
            respondJson(ex, 200, "{\"deviceList\":[" +
                    "{\"deviceID\":1,\"properties\":{\"serialNumber\":\"UDID-A\"}}," +
                    "{\"deviceID\":2,\"properties\":{\"serialNumber\":\"UDID-B\"}}]}");
        });

        server.createContext("/api/v1/device/UDID-A/apps/", ex -> {
            record(ex);
            respondJson(ex, 200, "[{\"CFBundleIdentifier\":\"com.apple.Preferences\"," +
                    "\"CFBundleName\":\"Settings\"}]");
        });

        server.createContext("/api/v1/device/UDID-A/screenshot", ex -> {
            record(ex);
            byte[] png = new byte[]{(byte) 0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A};
            ex.getResponseHeaders().set("Content-Type", "image/png");
            ex.sendResponseHeaders(200, png.length);
            try (OutputStream os = ex.getResponseBody()) {
                os.write(png);
            }
        });

        server.createContext("/api/v1/device/UDID-A/setlocation", ex -> {
            record(ex);
            lastQuery.put("setlocation", ex.getRequestURI().getRawQuery());
            respondJson(ex, 200, "{\"message\":\"ok\"}");
        });

        server.createContext("/api/v1/device/MISSING/info", ex -> {
            record(ex);
            respondJson(ex, 404, "{\"error\":\"device not found\"}");
        });

        server.createContext("/api/v1/device/UDID-A/syslog", ex -> {
            record(ex);
            ex.getResponseHeaders().set("Content-Type", "text/event-stream");
            byte[] body = ("event: syslog\ndata: {\"message\":\"boot\"}\n\n"
                    + "event: heartbeat\ndata: {}\n\n"
                    + "event: syslog\ndata: {\"message\":\"ready\"}\n\n").getBytes(StandardCharsets.UTF_8);
            ex.sendResponseHeaders(200, body.length);
            try (OutputStream os = ex.getResponseBody()) {
                os.write(body);
            }
        });

        server.setExecutor(null);
        server.start();
        baseUrl = "http://127.0.0.1:" + server.getAddress().getPort();
    }

    @AfterEach
    void stop() {
        server.stop(0);
    }

    private void record(HttpExchange ex) {
        List<String> auth = ex.getRequestHeaders().get("Authorization");
        seenAuthHeaders.add(auth == null ? "<none>" : String.join(",", auth));
    }

    private static void respondJson(HttpExchange ex, int status, String json) throws IOException {
        byte[] body = json.getBytes(StandardCharsets.UTF_8);
        ex.getResponseHeaders().set("Content-Type", "application/json");
        ex.sendResponseHeaders(status, body.length);
        try (OutputStream os = ex.getResponseBody()) {
            os.write(body);
        }
    }

    private IosClient client() {
        return IosClient.builder().baseUrl(baseUrl).apiKey("secret-token").build();
    }

    @Test
    void listsDevicesAndSendsBearerAuth() {
        try (IosClient c = client()) {
            List<DeviceEntry> devices = c.devices().list();
            assertEquals(2, devices.size());
            assertEquals("UDID-A", devices.get(0).getProperties().getSerialNumber());
        }
        assertTrue(seenAuthHeaders.contains("Bearer secret-token"),
                "Authorization: Bearer header must be sent; saw " + seenAuthHeaders);
    }

    @Test
    void omitsAuthHeaderWhenNoApiKey() {
        try (IosClient c = IosClient.builder().baseUrl(baseUrl).build()) {
            c.devices().list();
        }
        assertEquals("<none>", seenAuthHeaders.get(seenAuthHeaders.size() - 1));
    }

    @Test
    void listsApps() {
        try (IosClient c = client()) {
            List<AppInfo> apps = c.device("UDID-A").apps().list();
            assertEquals(1, apps.size());
            assertEquals("com.apple.Preferences", apps.get(0).getCfBundleIdentifier());
        }
    }

    @Test
    void screenshotReturnsRawPngBytes() {
        try (IosClient c = client()) {
            byte[] png = c.device("UDID-A").screenshot();
            assertEquals(8, png.length);
            assertEquals((byte) 0x89, png[0]);
            assertEquals('P', png[1]);
            assertEquals('N', png[2]);
            assertEquals('G', png[3]);
        }
    }

    @Test
    void setLocationUsesCorrectlySpelledLongitude() {
        try (IosClient c = client()) {
            c.device("UDID-A").setLocation(37.3349, -122.009);
        }
        String q = lastQuery.get("setlocation");
        assertNotNull(q);
        assertTrue(q.contains("longitude="), "query must use 'longitude': " + q);
        assertFalse(q.contains("longtitude"), "must not use the legacy misspelling: " + q);
        assertTrue(q.contains("latitude="), q);
    }

    @Test
    void notFoundRaisesIosApiExceptionWithEnvelope() {
        try (IosClient c = client()) {
            IosApiException ex = assertThrows(IosApiException.class, () -> c.device("MISSING").info());
            assertEquals(404, ex.statusCode());
            assertNotNull(ex.errorBody());
            assertEquals("device not found", ex.errorBody().getError());
        }
    }

    @Test
    void streamsSyslogOverRealHttpSkippingHeartbeats() {
        try (IosClient c = client()) {
            List<String> messages = new ArrayList<>();
            AtomicReference<SseReader> ref = new AtomicReference<>();
            try (SseReader stream = c.device("UDID-A").syslog()) {
                ref.set(stream);
                for (var ev : stream) {
                    if (ev instanceof SyslogEvent s) {
                        messages.add(s.payload().getMessage());
                    }
                }
            }
            assertEquals(List.of("boot", "ready"), messages);
        }
    }

    @Test
    void streamIsCancellableMidflight() {
        try (IosClient c = client()) {
            SseReader stream = c.device("UDID-A").syslog();
            assertTrue(stream.hasNext());
            stream.next(); // read one event then bail
            stream.close(); // should not throw
            assertFalse(stream.hasNext());
        }
    }
}
