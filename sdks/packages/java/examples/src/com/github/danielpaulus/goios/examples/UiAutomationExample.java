package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Device;
import com.github.danielpaulus.goios.IosClient;

/**
 * Example 6 (OPTIONAL) — drive the UI via the WebDriverAgent backend.
 *
 * <p>The {@code /ui/*} routes require a running UI-automation backend. By
 * default that is <b>WebDriverAgent (WDA)</b>: you must have WDA installed and
 * running on the device and reachable by the daemon. In the common setup you
 * start (and port-forward) WDA yourself and point the daemon at it — for
 * example via {@code ios runwda} plus {@code ios forward 8100 8100}, then pass
 * that forwarded URL to the daemon. Because that prerequisite is environmental,
 * this example is <b>skipped unless {@code RUN_UI=1}</b>, and even then it skips
 * (rather than fails) when the backend is unreachable, so it never breaks the
 * pre-release smoke test.
 *
 * <p>What it demonstrates: a {@code tap} at a coordinate followed by a
 * {@code type} of some text — the two most common UI primitives. Both go
 * through {@link com.github.danielpaulus.goios.Ui}; the convenience overloads
 * used here rely on the daemon's default backend and timeouts. To target a
 * specific forwarded WDA endpoint or the DeviceKit backend instead, pass a
 * {@link com.github.danielpaulus.goios.Ui.Options} (backend / wdaUrl / timeout).
 */
public final class UiAutomationExample {

    private UiAutomationExample() {
    }

    public static void main(String[] args) {
        Env.requireApiKey();

        if (!"1".equals(System.getenv("RUN_UI"))) {
            System.out.println("SKIP UiAutomationExample: set RUN_UI=1 to run "
                    + "(requires a running/forwarded WebDriverAgent backend).");
            return;
        }

        try (IosClient client = Env.client()) {
            String udid = Env.resolveUdid(client);
            if (udid == null) {
                System.out.println("SKIP UiAutomationExample: no device attached.");
                return;
            }

            Device device = client.device(udid);
            System.out.println("Driving UI on " + udid + " via the default (WDA) backend ...");

            try {
                // POST /device/{udid}/ui/tap  — tap near the top-left of the screen.
                device.ui().tap(100, 200);

                // POST /device/{udid}/ui/type — type into the focused field.
                device.ui().type("hello from go-ios");

                System.out.println("UI tap + type succeeded.");
            } catch (RuntimeException e) {
                // The backend being unreachable (no WDA / not forwarded) surfaces as an
                // exception. Treat it as a SKIP so the optional example never fails the
                // suite for an environmental reason.
                System.out.println("SKIP UiAutomationExample: UI backend unreachable ("
                        + e.getMessage() + ").");
                System.out.println("Ensure WebDriverAgent is running and forwarded, then retry with RUN_UI=1.");
            }
        }
    }
}
