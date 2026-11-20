package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Device;
import com.github.danielpaulus.goios.IosClient;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Example 4 — capture a PNG screenshot and write it to disk.
 *
 * <p>{@code device.screenshot()} performs {@code GET
 * /device/{udid}/screenshot} and returns the raw {@code image/png} bytes. We
 * write them to {@code ./screenshot.png} and print the file size.
 *
 * <p>Skips gracefully when no device is attached.
 */
public final class ScreenshotExample {

    /** Where the captured screenshot is written, relative to the working directory. */
    private static final Path OUTPUT = Path.of("screenshot.png");

    private ScreenshotExample() {
    }

    public static void main(String[] args) throws IOException {
        Env.requireApiKey();

        try (IosClient client = Env.client()) {
            String udid = Env.resolveUdid(client);
            if (udid == null) {
                System.out.println("SKIP ScreenshotExample: no device attached.");
                return;
            }

            Device device = client.device(udid);
            System.out.println("Capturing screenshot from " + udid + " ...");

            // GET /device/{udid}/screenshot -> raw PNG bytes.
            byte[] png = device.screenshot();

            Files.write(OUTPUT, png);
            System.out.printf("Wrote %d bytes to %s%n", png.length, OUTPUT.toAbsolutePath());
        }
    }
}
