package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.stream.BinaryStream;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * UI-automation operations for a single device ({@code /ui/*}), backed by a
 * WebDriverAgent (default) or DeviceKit backend.
 *
 * <p>Every call accepts optional {@link Options} selecting the backend, a
 * forwarded backend URL, and a per-request timeout. Convenience overloads use
 * the backend defaults.
 */
public final class Ui {

    private final Device d;

    Ui(Device d) {
        this.d = d;
    }

    /** Common per-request UI backend options. {@code null} fields fall back to server defaults. */
    public record Options(String backend, String wdaUrl, Integer timeoutSeconds) {
        public static Options defaults() {
            return new Options(null, null, null);
        }
    }

    private Map<String, String> query(Options o) {
        Map<String, String> q = RawHttp.query();
        if (o != null) {
            if (o.backend() != null) {
                q.put("backend", o.backend());
            }
            if (o.wdaUrl() != null) {
                q.put("wdaUrl", o.wdaUrl());
            }
            if (o.timeoutSeconds() != null) {
                q.put("timeout", Integer.toString(o.timeoutSeconds()));
            }
        }
        return q;
    }

    private Object post(String suffix, Options o, Object body) {
        return d.http().postJson(d.devicePath(suffix), query(o), body, Object.class);
    }

    private Object get(String suffix, Options o) {
        return d.http().getJson(d.devicePath(suffix), query(o), Object.class);
    }

    // -- gestures ----------------------------------------------------------

    /** Tap at (x, y) ({@code POST /ui/tap}). */
    public Object tap(int x, int y, Options o) {
        return post("/ui/tap", o, Map.of("x", x, "y", y));
    }

    public Object tap(int x, int y) {
        return tap(x, y, null);
    }

    /** Drag from (x1, y1) to (x2, y2) over {@code duration} seconds ({@code POST /ui/swipe}). */
    public Object swipe(int x1, int y1, int x2, int y2, Double duration, Options o) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("x1", x1);
        body.put("y1", y1);
        body.put("x2", x2);
        body.put("y2", y2);
        if (duration != null) {
            body.put("duration", duration);
        }
        return post("/ui/swipe", o, body);
    }

    public Object swipe(int x1, int y1, int x2, int y2) {
        return swipe(x1, y1, x2, y2, null, null);
    }

    /** Long-press at (x, y) ({@code POST /ui/longpress}). */
    public Object longPress(int x, int y, Double duration, Options o) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("x", x);
        body.put("y", y);
        if (duration != null) {
            body.put("duration", duration);
        }
        return post("/ui/longpress", o, body);
    }

    public Object longPress(int x, int y) {
        return longPress(x, y, null, null);
    }

    /** Type text ({@code POST /ui/type}). */
    public Object type(String text, Options o) {
        return post("/ui/type", o, Map.of("text", text));
    }

    public Object type(String text) {
        return type(text, null);
    }

    /** Press a hardware/software button by name ({@code POST /ui/button}). */
    public Object button(String name, Options o) {
        return post("/ui/button", o, Map.of("name", name));
    }

    public Object button(String name) {
        return button(name, null);
    }

    // -- introspection -----------------------------------------------------

    /** Capture a PNG screenshot via the UI backend ({@code GET /ui/screenshot}); returns raw bytes. */
    public byte[] screenshot(Options o) {
        return d.http().getBytes(d.devicePath("/ui/screenshot"), query(o));
    }

    public byte[] screenshot() {
        return screenshot(null);
    }

    /** Get the UI element source tree ({@code GET /ui/source}). */
    public Object source(Options o) {
        return get("/ui/source", o);
    }

    public Object source() {
        return source(null);
    }

    /** Get the window size ({@code GET /ui/size}). */
    public Object size(Options o) {
        return get("/ui/size", o);
    }

    public Object size() {
        return size(null);
    }

    /** Get the device orientation ({@code GET /ui/orientation}). */
    public Object orientation(Options o) {
        return get("/ui/orientation", o);
    }

    public Object orientation() {
        return orientation(null);
    }

    /** Set the device orientation ({@code POST /ui/orientation}). */
    public Object setOrientation(String orientation, Options o) {
        return post("/ui/orientation", o, Map.of("orientation", orientation));
    }

    public Object setOrientation(String orientation) {
        return setOrientation(orientation, null);
    }

    /** Get the UI backend status ({@code GET /ui/status}). */
    public Object status(Options o) {
        return get("/ui/status", o);
    }

    public Object status() {
        return status(null);
    }

    // -- app control -------------------------------------------------------

    /** Launch an app via the UI backend ({@code POST /ui/app/launch}). */
    public Object appLaunch(String bundleId, Options o) {
        return post("/ui/app/launch", o, Map.of("bundleId", bundleId));
    }

    public Object appLaunch(String bundleId) {
        return appLaunch(bundleId, null);
    }

    /** Terminate an app via the UI backend ({@code POST /ui/app/terminate}). */
    public Object appTerminate(String bundleId, Options o) {
        return post("/ui/app/terminate", o, Map.of("bundleId", bundleId));
    }

    public Object appTerminate(String bundleId) {
        return appTerminate(bundleId, null);
    }

    /** Foreground the backgrounded app ({@code POST /ui/app/foreground}); devicekit only. */
    public Object appForeground(Options o) {
        return post("/ui/app/foreground", o, null);
    }

    public Object appForeground() {
        return appForeground(null);
    }

    // -- raw passthrough ---------------------------------------------------

    /** Raw backend passthrough ({@code POST /ui/api}); {@code body} is the backend request payload. */
    public Object api(Object body, Options o) {
        return post("/ui/api", o, body);
    }

    public Object api(Object body) {
        return api(body, null);
    }

    // -- binary UI video stream --------------------------------------------

    /** Open a live UI video stream ({@code GET /ui/stream}); returns raw bytes (MJPEG or H.264). */
    public BinaryStream stream(Options o, StreamOptions video) {
        Map<String, String> q = query(o);
        if (video != null) {
            video.apply(q);
        }
        return d.http().binaryStream(d.devicePath("/ui/stream"), q);
    }

    public BinaryStream stream() {
        return stream(null, null);
    }

    /** Video-encoding options for {@link #stream}. All fields optional. */
    public record StreamOptions(String codec, String fps, String quality, String scale, String bitrate) {
        void apply(Map<String, String> q) {
            if (codec != null) {
                q.put("codec", codec);
            }
            if (fps != null) {
                q.put("fps", fps);
            }
            if (quality != null) {
                q.put("quality", quality);
            }
            if (scale != null) {
                q.put("scale", scale);
            }
            if (bitrate != null) {
                q.put("bitrate", bitrate);
            }
        }
    }
}
