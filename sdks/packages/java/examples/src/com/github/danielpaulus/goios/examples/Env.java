package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Devices;
import com.github.danielpaulus.goios.IosClient;
import com.github.danielpaulus.goios.generated.model.DeviceEntry;

import java.util.List;

/**
 * Shared environment / configuration helper for the examples.
 *
 * <p>Every example is configured purely through environment variables so it can
 * be run against any go-ios daemon without editing code:
 *
 * <ul>
 *   <li>{@code GO_IOS_BASE_URL} — base URL of the daemon. The SDK appends
 *       {@code /api/v1} automatically, so pass just the origin. Defaults to
 *       {@code http://localhost:8080}.</li>
 *   <li>{@code GO_IOS_API_KEY} — bearer token. The daemon refuses to start
 *       without an API key unless launched with {@code --disable-auth}. Every
 *       example treats a missing key as a fatal misconfiguration and exits with
 *       a helpful message (a daemon started with {@code --disable-auth} still
 *       accepts a bogus key, so set it to any non-empty value in that case).</li>
 *   <li>{@code GO_IOS_UDID} — optional. The device to target. When unset, the
 *       examples fall back to the first device reported by {@code GET /list}.</li>
 * </ul>
 *
 * <p>This class is intentionally tiny and dependency-free so the example
 * "programs" below can stay focused on demonstrating one SDK feature each.
 */
public final class Env {

    /** Default daemon endpoint when {@code GO_IOS_BASE_URL} is unset. */
    public static final String DEFAULT_BASE_URL = "http://localhost:8080";

    private Env() {
    }

    /** The configured base URL, or {@link #DEFAULT_BASE_URL}. */
    public static String baseUrl() {
        String v = System.getenv("GO_IOS_BASE_URL");
        return (v == null || v.isBlank()) ? DEFAULT_BASE_URL : v;
    }

    /** The configured API key, or {@code null} when unset. */
    public static String apiKey() {
        String v = System.getenv("GO_IOS_API_KEY");
        return (v == null || v.isBlank()) ? null : v;
    }

    /** The explicitly configured udid, or {@code null} to auto-select. */
    public static String udid() {
        String v = System.getenv("GO_IOS_UDID");
        return (v == null || v.isBlank()) ? null : v;
    }

    /**
     * Enforce that {@code GO_IOS_API_KEY} is set, printing a helpful message and
     * calling {@link System#exit(int)} with status {@code 1} when it is not.
     * Each example calls this first so a missing key fails fast and clearly
     * rather than surfacing as an opaque {@code 401} later.
     */
    public static void requireApiKey() {
        if (apiKey() == null) {
            System.err.println("ERROR: GO_IOS_API_KEY is not set.");
            System.err.println();
            System.err.println("The go-ios daemon requires a bearer token on every /api/v1 route.");
            System.err.println("Export it before running the examples, for example:");
            System.err.println();
            System.err.println("    export GO_IOS_API_KEY=\"$(cat ~/.go-ios-api-key)\"");
            System.err.println();
            System.err.println("If your daemon was started with --disable-auth, any non-empty");
            System.err.println("value works: export GO_IOS_API_KEY=dev");
            System.exit(1);
        }
    }

    /**
     * Build an {@link IosClient} from the environment. The caller owns the
     * returned client and must {@link IosClient#close() close} it (ideally via
     * try-with-resources).
     */
    public static IosClient client() {
        return IosClient.builder()
                .baseUrl(baseUrl())
                .apiKey(apiKey())
                .build();
    }

    /**
     * Resolve the target device udid: {@code GO_IOS_UDID} when set, otherwise the
     * first device from {@code GET /list}. Returns {@code null} when no device is
     * attached — callers should treat that as a "SKIP" rather than a failure so
     * the suite still passes on a daemon with no devices.
     */
    public static String resolveUdid(IosClient client) {
        String explicit = udid();
        if (explicit != null) {
            return explicit;
        }
        List<DeviceEntry> devices = client.devices().list();
        for (DeviceEntry d : devices) {
            String u = Devices.udid(d);
            if (u != null) {
                return u;
            }
        }
        return null;
    }
}
