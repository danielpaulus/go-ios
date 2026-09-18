package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Device;
import com.github.danielpaulus.goios.IosClient;

/**
 * Example 2 — fetch device info for the target device.
 *
 * <p>Resolves a device (either {@code GO_IOS_UDID} or the first attached
 * device), then reads its lockdown + instruments values via {@code GET /info}.
 * The facade returns {@code info()} as a loosely-typed {@code Object} (a decoded
 * JSON map) because the surface is large and device-dependent; this example
 * simply prints it.
 *
 * <p>When no device is attached it prints {@code SKIP} and returns normally so
 * the pre-release runner does not fail on a device-less daemon.
 */
public final class DeviceInfoExample {

    private DeviceInfoExample() {
    }

    public static void main(String[] args) {
        Env.requireApiKey();

        try (IosClient client = Env.client()) {
            String udid = Env.resolveUdid(client);
            if (udid == null) {
                System.out.println("SKIP DeviceInfoExample: no device attached.");
                return;
            }

            System.out.println("Fetching info for device " + udid + " ...");

            // Device is a lightweight handle scoped to one udid.
            Device device = client.device(udid);

            // GET /device/{udid}/info
            Object info = device.info();
            System.out.println(info);
        }
    }
}
