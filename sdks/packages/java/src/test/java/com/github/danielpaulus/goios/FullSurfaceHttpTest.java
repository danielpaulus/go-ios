package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.BatteryInfo;
import com.github.danielpaulus.goios.generated.model.CrashListing;
import com.github.danielpaulus.goios.generated.model.GenericResponse;
import com.github.danielpaulus.goios.generated.model.DeviceEntry;
import com.github.danielpaulus.goios.generated.model.FileListing;
import com.github.danielpaulus.goios.generated.model.FilePushResult;
import com.github.danielpaulus.goios.generated.model.Job;
import com.github.danielpaulus.goios.generated.model.MemLimitResult;
import com.github.danielpaulus.goios.generated.model.StatusOk;
import com.github.danielpaulus.goios.generated.model.Tunnel;
import com.github.danielpaulus.goios.generated.model.TunnelStopped;
import com.github.danielpaulus.goios.stream.JobLogEvent;
import com.github.danielpaulus.goios.stream.SseReader;
import com.github.danielpaulus.goios.stream.SysmontapEvent;
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

import static org.junit.jupiter.api.Assertions.*;

/**
 * Exercises the full-surface facade groups added in the 80-endpoint extension
 * against an in-process {@link HttpServer} stub: device management, files,
 * settings, media, MDM (multipart), crashes, jobs, tunnels, and the two new SSE
 * streams (sysmontap, job logs).
 */
class FullSurfaceHttpTest {

    private HttpServer server;
    private String baseUrl;
    private final ConcurrentHashMap<String, String> lastQuery = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, String> lastMethod = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, byte[]> lastBody = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, String> lastContentType = new ConcurrentHashMap<>();

    @BeforeEach
    void start() throws IOException {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        String U = "/api/v1/device/UDID-A/";

        // ---- device info / management
        ctx(U + "battery", ex -> json(ex, 200,
                "{\"CurrentCapacity\":83,\"IsCharging\":true,\"FullyCharged\":false}"));
        ctx(U + "reboot", ex -> json(ex, 200, "{\"message\":\"rebooting\"}"));
        ctx(U + "erase", ex -> {
            put(lastQuery, "erase", ex.getRequestURI().getRawQuery());
            json(ex, 200, "{\"message\":\"erased\"}");
        });
        ctx(U + "memlimitoff", ex -> json(ex, 200,
                "{\"process\":\"backboardd\",\"pid\":42,\"disabled\":true}"));
        ctx(U + "mobilegestalt", ex -> {
            put(lastQuery, "mobilegestalt", ex.getRequestURI().getRawQuery());
            json(ex, 200, "{\"ProductType\":\"iPhone14,2\"}");
        });

        // ---- files
        ctx(U + "files", ex -> {
            put(lastQuery, "files", ex.getRequestURI().getRawQuery());
            json(ex, 200, "{\"path\":\"/Documents\",\"count\":1," +
                    "\"files\":[{\"name\":\"log.txt\",\"isDir\":false,\"size\":12}]}");
        });
        ctx(U + "files/pull", ex -> {
            byte[] b = "hello-file".getBytes(StandardCharsets.UTF_8);
            ex.getResponseHeaders().set("Content-Type", "application/octet-stream");
            ex.sendResponseHeaders(200, b.length);
            try (OutputStream os = ex.getResponseBody()) { os.write(b); }
        });
        ctx(U + "files/push", ex -> {
            put(lastMethod, "files/push", ex.getRequestMethod());
            put(lastBody, "files/push", ex.getRequestBody().readAllBytes());
            json(ex, 200, "{\"remote\":\"/Documents/out.txt\",\"size\":5}");
        });

        // ---- settings
        ctx(U + "assistivetouch", ex -> json(ex, 200, "{\"AssistiveTouchEnabled\":true}"));
        ctx(U + "wifi", ex -> {
            put(lastMethod, "wifi", ex.getRequestMethod());
            put(lastQuery, "wifi", ex.getRequestURI().getRawQuery());
            json(ex, 200, "{\"message\":\"ok\"}");
        });

        // ---- media
        ctx(U + "wallpaper", ex -> {
            put(lastMethod, "wallpaper", ex.getRequestMethod());
            put(lastContentType, "wallpaper", first(ex, "Content-Type"));
            if ("PUT".equals(ex.getRequestMethod())) {
                put(lastBody, "wallpaper", ex.getRequestBody().readAllBytes());
                json(ex, 200, "{\"message\":\"set\"}");
            } else {
                byte[] png = {(byte) 0x89, 'P', 'N', 'G'};
                ex.getResponseHeaders().set("Content-Type", "image/png");
                ex.sendResponseHeaders(200, png.length);
                try (OutputStream os = ex.getResponseBody()) { os.write(png); }
            }
        });
        ctx(U + "pasteboard", ex -> {
            put(lastMethod, "pasteboard", ex.getRequestMethod());
            if ("PUT".equals(ex.getRequestMethod())) {
                put(lastBody, "pasteboard", ex.getRequestBody().readAllBytes());
                put(lastContentType, "pasteboard", first(ex, "Content-Type"));
                json(ex, 200, "{\"message\":\"ok\"}");
            } else {
                json(ex, 200, "{\"present\":true,\"text\":\"clip\"}");
            }
        });

        // ---- mdm (multipart)
        ctx(U + "mdm/clear-passcode", ex -> {
            put(lastContentType, "mdm", first(ex, "Content-Type"));
            put(lastBody, "mdm", ex.getRequestBody().readAllBytes());
            json(ex, 200, "{\"status\":\"ok\"}");
        });

        // ---- crashes
        ctx(U + "crashes", ex -> {
            put(lastMethod, "crashes", ex.getRequestMethod());
            put(lastQuery, "crashes", ex.getRequestURI().getRawQuery());
            if ("DELETE".equals(ex.getRequestMethod())) {
                json(ex, 200, "{\"message\":\"removed\"}");
            } else {
                json(ex, 200, "{\"files\":[\"a.crash\",\"b.crash\"],\"count\":2}");
            }
        });

        // ---- profiles (multipart POST)
        ctx(U + "profiles", ex -> {
            put(lastMethod, "profiles", ex.getRequestMethod());
            put(lastContentType, "profiles", first(ex, "Content-Type"));
            json(ex, 200, "{\"message\":\"installed\"}");
        });

        // ---- jobs
        ctx(U + "jobs/forward", ex -> {
            put(lastBody, "jobs/forward", ex.getRequestBody().readAllBytes());
            json(ex, 202, "{\"id\":\"forward-1\",\"kind\":\"forward\",\"udid\":\"UDID-A\"," +
                    "\"status\":\"running\",\"startedAt\":\"2024-01-01T00:00:00Z\"}");
        });
        ctx(U + "jobs/job-1/logs", ex -> {
            ex.getResponseHeaders().set("Content-Type", "text/event-stream");
            byte[] body = ("event: log\ndata: {\"line\":\"starting\"}\n\n"
                    + "event: heartbeat\ndata: {}\n\n"
                    + "event: log\ndata: {\"line\":\"done\"}\n\n").getBytes(StandardCharsets.UTF_8);
            ex.sendResponseHeaders(200, body.length);
            try (OutputStream os = ex.getResponseBody()) { os.write(body); }
        });

        // ---- sysmontap SSE
        ctx(U + "sysmontap", ex -> {
            ex.getResponseHeaders().set("Content-Type", "text/event-stream");
            byte[] body = ("event: sample\ndata: {\"CPU_TotalLoad\":42.5}\n\n"
                    + "event: heartbeat\ndata: {}\n\n"
                    + "event: sample\ndata: {\"CPU_TotalLoad\":7.0}\n\n").getBytes(StandardCharsets.UTF_8);
            ex.sendResponseHeaders(200, body.length);
            try (OutputStream os = ex.getResponseBody()) { os.write(body); }
        });

        // ---- tunnels (fleet-level)
        ctx("/api/v1/tunnels", ex -> json(ex, 200,
                "[{\"Udid\":\"UDID-A\",\"Address\":\"fd00::1\",\"RsdPort\":50000}]"));
        ctx("/api/v1/tunnels/UDID-A/refresh", ex -> json(ex, 200,
                "{\"Udid\":\"UDID-A\",\"Address\":\"fd00::2\",\"RsdPort\":50001}"));
        ctx("/api/v1/tunnels/UDID-A", ex -> {
            // DELETE handler (the /refresh context wins for that suffix).
            if ("DELETE".equals(ex.getRequestMethod())) {
                json(ex, 200, "{\"udid\":\"UDID-A\",\"status\":\"stopped\"}");
            } else {
                json(ex, 404, "{\"error\":\"nope\"}");
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

    private void ctx(String path, com.sun.net.httpserver.HttpHandler h) {
        server.createContext(path, h);
    }

    /** Null-safe record into a map ({@link ConcurrentHashMap} forbids null values). */
    private static <V> void put(ConcurrentHashMap<String, V> map, String key, V value) {
        if (value != null) {
            map.put(key, value);
        }
    }

    private static String first(HttpExchange ex, String header) {
        List<String> v = ex.getRequestHeaders().get(header);
        return v == null || v.isEmpty() ? null : v.get(0);
    }

    private static void json(HttpExchange ex, int status, String body) throws IOException {
        byte[] b = body.getBytes(StandardCharsets.UTF_8);
        ex.getResponseHeaders().set("Content-Type", "application/json");
        ex.sendResponseHeaders(status, b.length);
        try (OutputStream os = ex.getResponseBody()) { os.write(b); }
    }

    private IosClient client() {
        return IosClient.builder().baseUrl(baseUrl).apiKey("secret").build();
    }

    // ------------------------------------------------------------------ tests

    @Test
    void udidConvenienceAccessor() {
        DeviceEntry e = new DeviceEntry();
        assertNull(Devices.udid(e)); // no properties
    }

    @Test
    void battery() {
        try (IosClient c = client()) {
            BatteryInfo b = c.device("UDID-A").battery();
            assertEquals(83, b.getCurrentCapacity());
            assertTrue(b.getIsCharging());
        }
    }

    @Test
    void rebootAndEraseConfirm() {
        try (IosClient c = client()) {
            assertEquals("rebooting", c.device("UDID-A").reboot().getMessage());
            c.device("UDID-A").erase(true);
        }
        assertTrue(lastQuery.get("erase").contains("confirm=true"), lastQuery.get("erase"));
    }

    @Test
    void memlimitoff() {
        try (IosClient c = client()) {
            MemLimitResult r = c.device("UDID-A").memlimitoff("backboardd");
            assertEquals("backboardd", r.getProcess());
            assertTrue(r.getDisabled());
        }
    }

    @Test
    void mobileGestaltPassesKeys() {
        try (IosClient c = client()) {
            Object g = c.device("UDID-A").mobileGestalt(List.of("ProductType", "BuildVersion"));
            assertTrue(g.toString().contains("iPhone14,2"));
        }
        // The spec marks `key` as explode:false, so the generated client sends a
        // single comma-joined value (URL-encoded comma: %2C).
        String q = lastQuery.get("mobilegestalt");
        assertTrue(q.startsWith("key="), q);
        assertTrue(q.contains("ProductType"), q);
        assertTrue(q.contains("BuildVersion"), q);
    }

    @Test
    void filesLsPullPush() {
        try (IosClient c = client()) {
            FileListing ls = c.device("UDID-A").files().ls("app", "com.x", "/Documents");
            assertEquals(1, ls.getCount());
            assertEquals("log.txt", ls.getFiles().get(0).getName());
            assertTrue(lastQuery.get("files").contains("domain=app"), lastQuery.get("files"));
            assertTrue(lastQuery.get("files").contains("identifier=com.x"));

            byte[] pulled = c.device("UDID-A").files().pull("temp", null, "/x");
            assertEquals("hello-file", new String(pulled, StandardCharsets.UTF_8));

            FilePushResult push = c.device("UDID-A").files()
                    .push("temp", null, "/Documents/out.txt", "hello".getBytes(StandardCharsets.UTF_8));
            assertEquals(5, push.getSize());
            assertEquals("POST", lastMethod.get("files/push"));
            assertArrayEquals("hello".getBytes(StandardCharsets.UTF_8), lastBody.get("files/push"));
        }
    }

    @Test
    void settingsAssistiveTouchAndWifi() {
        try (IosClient c = client()) {
            assertTrue(c.device("UDID-A").settings().assistiveTouch().getAssistiveTouchEnabled());
            c.device("UDID-A").settings().setWifi("net", "pw", "WPA2");
            c.device("UDID-A").settings().removeWifi("net");
        }
        assertTrue(lastQuery.get("wifi").contains("ssid=net"), lastQuery.get("wifi"));
    }

    @Test
    void mediaWallpaperMultipartAndPasteboardTextPlain() {
        try (IosClient c = client()) {
            byte[] png = c.device("UDID-A").media().wallpaper();
            assertEquals((byte) 0x89, png[0]);

            c.device("UDID-A").media().setWallpaper(
                    "img".getBytes(StandardCharsets.UTF_8),
                    "p12".getBytes(StandardCharsets.UTF_8), "pass", "home");
            assertEquals("PUT", lastMethod.get("wallpaper"));
            assertTrue(lastContentType.get("wallpaper").startsWith("multipart/form-data"),
                    lastContentType.get("wallpaper"));
            String mp = new String(lastBody.get("wallpaper"), StandardCharsets.UTF_8);
            assertTrue(mp.contains("name=\"image\""), mp);
            assertTrue(mp.contains("name=\"p12\""), mp);
            assertTrue(mp.contains("name=\"screen\""), mp);

            assertEquals("clip", c.device("UDID-A").media().pasteboard().getText());
            c.device("UDID-A").media().setPasteboard("copied");
            assertEquals("PUT", lastMethod.get("pasteboard"));
            assertEquals("copied", new String(lastBody.get("pasteboard"), StandardCharsets.UTF_8));
        }
    }

    @Test
    void mdmClearPasscodeSendsMultipartWithTokenAndP12() {
        try (IosClient c = client()) {
            StatusOk ok = c.device("UDID-A").mdm().clearPasscode(
                    "p12bytes".getBytes(StandardCharsets.UTF_8), "pw", "TOKEN123");
            assertEquals("ok", ok.getStatus());
        }
        assertTrue(lastContentType.get("mdm").startsWith("multipart/form-data"));
        String mp = new String(lastBody.get("mdm"), StandardCharsets.UTF_8);
        assertTrue(mp.contains("name=\"p12\""), mp);
        assertTrue(mp.contains("name=\"token\""), mp);
        assertTrue(mp.contains("TOKEN123"), mp);
    }

    @Test
    void crashesListAndProfilesMultipart() {
        try (IosClient c = client()) {
            CrashListing cr = c.device("UDID-A").crashes().list();
            assertEquals(2, cr.getCount());
            // remove(pattern) — pattern is the primary arg; cwd defaults.
            GenericResponse rm = c.device("UDID-A").crashes().remove("*.crash");
            assertEquals("removed", rm.getMessage());
            assertEquals("DELETE", lastMethod.get("crashes"));
            String cq = lastQuery.get("crashes");
            assertTrue(cq != null && cq.contains("pattern=*.crash"), cq);
            assertTrue(cq.contains("cwd=."), cq);
            // remove(pattern, cwd) — explicit working directory.
            c.device("UDID-A").crashes().remove("*.ips", "/tmp/crashes");
            cq = lastQuery.get("crashes");
            assertTrue(cq.contains("pattern=*.ips"), cq);
            assertTrue(cq.contains("cwd=%2Ftmp%2Fcrashes"), cq);
            c.device("UDID-A").addProfile("cfg".getBytes(StandardCharsets.UTF_8), null, null);
        }
        assertEquals("POST", lastMethod.get("profiles"));
        assertTrue(lastContentType.get("profiles").startsWith("multipart/form-data"));
    }

    @Test
    void jobsForwardReturns202Job() {
        try (IosClient c = client()) {
            Job job = c.device("UDID-A").jobs().forward(8080, 9090);
            assertEquals("forward-1", job.getId());
            assertEquals("forward", job.getKind());
        }
        String body = new String(lastBody.get("jobs/forward"), StandardCharsets.UTF_8);
        assertTrue(body.contains("8080"), body);
        assertTrue(body.contains("9090"), body);
    }

    @Test
    void tunnelsListRefreshDelete() {
        try (IosClient c = client()) {
            List<Tunnel> tunnels = c.tunnels().list();
            assertEquals(1, tunnels.size());
            assertEquals("UDID-A", tunnels.get(0).getUdid());

            Tunnel refreshed = c.tunnels().refresh("UDID-A");
            assertEquals(50001, refreshed.getRsdPort());

            TunnelStopped stopped = c.tunnels().delete("UDID-A");
            assertEquals("stopped", stopped.getStatus());
        }
    }

    @Test
    void sysmontapStreamSkipsHeartbeats() {
        try (IosClient c = client()) {
            List<Object> samples = new ArrayList<>();
            try (SseReader stream = c.device("UDID-A").sysmontap()) {
                for (var ev : stream) {
                    if (ev instanceof SysmontapEvent s) {
                        samples.add(s.payload());
                    }
                }
            }
            assertEquals(2, samples.size(), "two sample events, heartbeat skipped");
        }
    }

    @Test
    void jobLogsStreamDecodesLines() {
        try (IosClient c = client()) {
            List<String> lines = new ArrayList<>();
            try (SseReader stream = c.device("UDID-A").jobs().logs("job-1")) {
                for (var ev : stream) {
                    if (ev instanceof JobLogEvent l) {
                        lines.add(l.payload().getLine());
                    }
                }
            }
            assertEquals(List.of("starting", "done"), lines);
        }
    }
}
