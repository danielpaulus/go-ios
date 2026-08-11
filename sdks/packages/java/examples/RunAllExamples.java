import com.github.danielpaulus.goios.examples.DeviceInfoExample;
import com.github.danielpaulus.goios.examples.Env;
import com.github.danielpaulus.goios.examples.ListAppsExample;
import com.github.danielpaulus.goios.examples.ListDevicesExample;
import com.github.danielpaulus.goios.examples.ScreenshotExample;
import com.github.danielpaulus.goios.examples.StreamSyslogExample;
import com.github.danielpaulus.goios.examples.UiAutomationExample;

/**
 * Runs the go-ios Java SDK examples end-to-end as a single program. This doubles
 * as a <b>pre-release smoke test</b>: it exercises the real public API against a
 * live daemon and fails loudly (non-zero exit) if any core example throws.
 *
 * <p>Order and semantics:
 * <ol>
 *   <li>{@code ListDevicesExample}</li>
 *   <li>{@code DeviceInfoExample}</li>
 *   <li>{@code ListAppsExample}</li>
 *   <li>{@code ScreenshotExample}</li>
 *   <li>{@code StreamSyslogExample}</li>
 *   <li>{@code UiAutomationExample} — only when {@code RUN_UI=1}</li>
 * </ol>
 *
 * <p>Steps 1–5 must complete without throwing. Steps that need a device print a
 * {@code SKIP} line and return normally when no device is attached, so the suite
 * still passes against a device-less daemon (a genuine transport/auth failure
 * still throws and fails the run). The UI example runs only when {@code RUN_UI=1}
 * and never fails the suite for an environmental reason.
 *
 * <p>{@code Env.requireApiKey()} short-circuits with a helpful message and exit
 * code {@code 1} if {@code GO_IOS_API_KEY} is unset, mirroring each standalone
 * example.
 */
public final class RunAllExamples {

    private RunAllExamples() {
    }

    @FunctionalInterface
    private interface Example {
        void run() throws Exception;
    }

    public static void main(String[] args) {
        // Fail fast (exit 1) with a helpful message before running anything.
        Env.requireApiKey();

        System.out.println("== go-ios Java SDK examples ==");
        String baseUrl = Env.baseUrl();
        System.out.println("baseUrl = " + (baseUrl != null ? baseUrl : "(auto-discovered local daemon)"));
        System.out.println();

        // Core examples (1-5). Any exception here fails the whole run.
        runOrExit("1. ListDevicesExample", () -> ListDevicesExample.main(args));
        runOrExit("2. DeviceInfoExample", () -> DeviceInfoExample.main(args));
        runOrExit("3. ListAppsExample", () -> ListAppsExample.main(args));
        runOrExit("4. ScreenshotExample", () -> ScreenshotExample.main(args));
        runOrExit("5. StreamSyslogExample", () -> StreamSyslogExample.main(args));

        // Optional UI example (6). It self-skips unless RUN_UI=1; a failure there
        // should still surface, so we run it through the same guard.
        runOrExit("6. UiAutomationExample", () -> UiAutomationExample.main(args));

        System.out.println();
        System.out.println("All examples completed successfully.");
    }

    /** Run one example, printing a header; on any exception print it and exit non-zero. */
    private static void runOrExit(String title, Example example) {
        System.out.println("--- " + title + " ---");
        try {
            example.run();
        } catch (Throwable t) {
            System.err.println();
            System.err.println("FAILED: " + title);
            t.printStackTrace();
            System.exit(1);
        }
        System.out.println();
    }
}
