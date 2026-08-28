package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.DiskSpaceInfo;
import com.github.danielpaulus.goios.generated.model.FsyncListing;
import com.github.danielpaulus.goios.generated.model.FsyncPushResult;
import com.github.danielpaulus.goios.generated.model.NetworkInfo;
import com.github.danielpaulus.goios.generated.model.PrepareSkipOptions;
import com.github.danielpaulus.goios.generated.model.SupervisionCert;
import com.github.danielpaulus.goios.generated.model.VoiceOverState;
import com.github.danielpaulus.goios.generated.model.WebInspectorEvalResult;
import com.github.danielpaulus.goios.stream.BinaryStream;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Exercises the v3 facade groups added on top of the base surface: accessibility,
 * network/disk diagnostics, fsync (AFC), web-inspector, host-scoped sign/prepare,
 * UI automation, and the binary (non-SSE) streaming endpoints.
 */
class V3SurfaceHttpTest {

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

        // ---- diagnostics-net
        ctx(U + "diskspace", ex -> json(ex, 200,
                "{\"Model\":\"disk0\",\"FSTotalBytes\":128000000000,\"FSFreeBytes\":64000000000,\"FSBlockSize\":4096}"));
        ctx(U + "ip", ex -> json(ex, 200, "{\"MacAddress\":\"aa:bb:cc:dd:ee:ff\"}"));

        // ---- accessibility
        ctx(U + "voiceover", ex -> {
            put(lastMethod, "voiceover", ex.getRequestMethod());
            if ("PUT".equals(ex.getRequestMethod())) {
                put(lastBody, "voiceover", ex.getRequestBody().readAllBytes());
                json(ex, 200, "{\"VoiceOverEnabled\":true}");
            } else {
                json(ex, 200, "{\"VoiceOverEnabled\":false}");
            }
        });
        ctx(U + "setlocation/gpx", ex -> {
            put(lastContentType, "gpx", first(ex, "Content-Type"));
            put(lastBody, "gpx", ex.getRequestBody().readAllBytes());
            json(ex, 200, "{\"message\":\"replaying\"}");
        });

        // ---- fsync
        ctx(U + "fsync/ls", ex -> {
            put(lastQuery, "fsync/ls", ex.getRequestURI().getRawQuery());
            json(ex, 200, "{\"path\":\"/Documents\",\"count\":2,\"files\":[\"a.txt\",\"b.txt\"]}");
        });
        ctx(U + "fsync/pull", ex -> {
            byte[] b = "afc-bytes".getBytes(StandardCharsets.UTF_8);
            ex.getResponseHeaders().set("Content-Type", "application/octet-stream");
            ex.sendResponseHeaders(200, b.length);
            try (OutputStream os = ex.getResponseBody()) { os.write(b); }
        });
        ctx(U + "fsync/push", ex -> {
            put(lastMethod, "fsync/push", ex.getRequestMethod());
            put(lastContentType, "fsync/push", first(ex, "Content-Type"));
            put(lastBody, "fsync/push", ex.getRequestBody().readAllBytes());
            json(ex, 200, "{\"path\":\"/Documents/x.bin\",\"size\":4}");
        });

        // ---- webinspector
        ctx(U + "webinspector/eval", ex -> {
            put(lastBody, "webinspector/eval", ex.getRequestBody().readAllBytes());
            json(ex, 200, "{\"page\":\"1\",\"result\":42}");
        });

        // ---- ui
        ctx(U + "ui/tap", ex -> {
            put(lastQuery, "ui/tap", ex.getRequestURI().getRawQuery());
            put(lastBody, "ui/tap", ex.getRequestBody().readAllBytes());
            json(ex, 200, "{\"ok\":true}");
        });
        ctx(U + "ui/screenshot", ex -> {
            byte[] png = {(byte) 0x89, 'P', 'N', 'G'};
            ex.getResponseHeaders().set("Content-Type", "image/png");
            ex.sendResponseHeaders(200, png.length);
            try (OutputStream os = ex.getResponseBody()) { os.write(png); }
        });
        // Binary UI video stream: emit chunked raw bytes.
        ctx(U + "ui/stream", ex -> {
            ex.getResponseHeaders().set("Content-Type", "multipart/x-mixed-replace; boundary=frame");
            ex.sendResponseHeaders(200, 0); // chunked
            try (OutputStream os = ex.getResponseBody()) {
                for (int i = 0; i < 4; i++) {
                    os.write(("chunk-" + i + ";").getBytes(StandardCharsets.UTF_8));
                    os.flush();
                }
            }
        });

        // ---- host-scoped
        ctx("/api/v1/prepare/skip-options", ex -> json(ex, 200,
                "{\"count\":2,\"options\":[\"Passcode\",\"Siri\"]}"));
        ctx("/api/v1/prepare/create-cert", ex -> {
            put(lastMethod, "create-cert", ex.getRequestMethod());
            json(ex, 200, "{\"certPem\":\"-----BEGIN CERT-----\",\"privateKeyPem\":\"-----BEGIN KEY-----\"}");
        });
        ctx("/api/v1/sign/certificate", ex -> {
            put(lastContentType, "sign/certificate", first(ex, "Content-Type"));
            put(lastBody, "sign/certificate", ex.getRequestBody().readAllBytes());
            byte[] p12 = {0x30, (byte) 0x82, 0x01, 0x02}; // pkcs12 magic-ish
            ex.getResponseHeaders().set("Content-Type", "application/x-pkcs12");
            ex.sendResponseHeaders(200, p12.length);
            try (OutputStream os = ex.getResponseBody()) { os.write(p12); }
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
    void diskSpaceAndIp() {
        try (IosClient c = client()) {
            DiskSpaceInfo disk = c.device("UDID-A").diskSpace();
            assertEquals("disk0", disk.getModel());
            assertEquals(64000000000L, disk.getFsFreeBytes());
            NetworkInfo net = c.device("UDID-A").ip();
            assertEquals("aa:bb:cc:dd:ee:ff", net.getMacAddress());
        }
    }

    @Test
    void voiceOverGetAndSet() {
        try (IosClient c = client()) {
            assertFalse(c.device("UDID-A").voiceOver().getVoiceOverEnabled());
            VoiceOverState set = c.device("UDID-A").setVoiceOver(true);
            assertTrue(set.getVoiceOverEnabled());
        }
        assertEquals("PUT", lastMethod.get("voiceover"));
        assertTrue(new String(lastBody.get("voiceover"), StandardCharsets.UTF_8).contains("enabled"),
                new String(lastBody.get("voiceover"), StandardCharsets.UTF_8));
    }

    @Test
    void setLocationGpxIsMultipart() {
        try (IosClient c = client()) {
            assertEquals("replaying",
                    c.device("UDID-A").setLocationGpx("<gpx/>".getBytes(StandardCharsets.UTF_8)).getMessage());
        }
        assertTrue(lastContentType.get("gpx").startsWith("multipart/form-data"), lastContentType.get("gpx"));
        String mp = new String(lastBody.get("gpx"), StandardCharsets.UTF_8);
        assertTrue(mp.contains("name=\"gpx\""), mp);
        assertTrue(mp.contains("<gpx/>"), mp);
    }

    @Test
    void fsyncLsPullPush() {
        try (IosClient c = client()) {
            FsyncListing ls = c.device("UDID-A").fsync().ls("/Documents", "com.x");
            assertEquals(2, ls.getCount());
            assertTrue(lastQuery.get("fsync/ls").contains("bundleID=com.x"), lastQuery.get("fsync/ls"));
            assertTrue(lastQuery.get("fsync/ls").contains("path=%2FDocuments"), lastQuery.get("fsync/ls"));

            byte[] pulled = c.device("UDID-A").fsync().pull("/Documents/a.txt", null);
            assertEquals("afc-bytes", new String(pulled, StandardCharsets.UTF_8));

            FsyncPushResult push = c.device("UDID-A").fsync()
                    .push("/Documents/x.bin", "data".getBytes(StandardCharsets.UTF_8), null);
            assertEquals(4L, push.getSize());
            assertEquals("POST", lastMethod.get("fsync/push"));
            assertEquals("application/octet-stream", lastContentType.get("fsync/push"));
            assertArrayEquals("data".getBytes(StandardCharsets.UTF_8), lastBody.get("fsync/push"));
        }
    }

    @Test
    void webInspectorEvalSendsScript() {
        try (IosClient c = client()) {
            WebInspectorEvalResult r = c.device("UDID-A").webinspector().eval("1+1", "1", null);
            assertEquals("1", r.getPage());
            assertEquals(42, ((Number) r.getResult()).intValue());
        }
        String body = new String(lastBody.get("webinspector/eval"), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"script\":\"1+1\""), body);
    }

    @Test
    void uiTapSendsJsonBodyAndBackendQuery() {
        try (IosClient c = client()) {
            c.device("UDID-A").ui().tap(10, 20, new Ui.Options("devicekit", null, 30));
        }
        String q = lastQuery.get("ui/tap");
        assertTrue(q.contains("backend=devicekit"), q);
        assertTrue(q.contains("timeout=30"), q);
        String body = new String(lastBody.get("ui/tap"), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"x\":10"), body);
        assertTrue(body.contains("\"y\":20"), body);
    }

    @Test
    void uiScreenshotReturnsPngBytes() {
        try (IosClient c = client()) {
            byte[] png = c.device("UDID-A").ui().screenshot();
            assertEquals((byte) 0x89, png[0]);
            assertEquals('P', png[1]);
        }
    }

    @Test
    void hostSignPrepareGroups() {
        try (IosClient c = client()) {
            PrepareSkipOptions opts = c.prepare().skipOptions();
            assertEquals(2, opts.getCount());
            assertEquals(List.of("Passcode", "Siri"), opts.getOptions());

            SupervisionCert cert = c.prepare().createCert();
            assertTrue(cert.getCertPem().startsWith("-----BEGIN CERT"));
            assertEquals("POST", lastMethod.get("create-cert"));

            byte[] p12 = c.sign().certificate(
                    "p8key".getBytes(StandardCharsets.UTF_8), "KEYID", "ISSUER", false, "pw");
            assertEquals(4, p12.length);
            assertEquals(0x30, p12[0]);
            assertTrue(lastContentType.get("sign/certificate").startsWith("multipart/form-data"));
            String mp = new String(lastBody.get("sign/certificate"), StandardCharsets.UTF_8);
            assertTrue(mp.contains("name=\"asc-private-key\""), mp);
            assertTrue(mp.contains("name=\"asc-key-id\""), mp);
        }
    }

    @Test
    void uiStreamReadsRawChunkedBytesAndCloses() throws IOException {
        // The binary stream (x-stream: binary) is a plain InputStream of raw bytes,
        // distinct from the typed SSE reader.
        try (IosClient c = client()) {
            BinaryStream stream = c.device("UDID-A").ui().stream();
            assertNotNull(stream.contentType());
            assertTrue(stream.contentType().startsWith("multipart/x-mixed-replace"));
            byte[] all = stream.readAllBytes();
            String s = new String(all, StandardCharsets.UTF_8);
            assertEquals("chunk-0;chunk-1;chunk-2;chunk-3;", s);
            stream.close(); // idempotent, releases the connection
            stream.close();
        }
    }
}
