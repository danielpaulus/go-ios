package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Device;
import com.github.danielpaulus.goios.IosClient;
import com.github.danielpaulus.goios.generated.model.AppInfo;

import java.util.List;

/**
 * Example 3 — list installed apps on the target device.
 *
 * <p>Reads the installed application list via {@code GET
 * /device/{udid}/apps/}, which returns strongly-typed {@link AppInfo} records.
 * We print each app's bundle identifier, display name and version.
 *
 * <p>Skips gracefully when no device is attached.
 */
public final class ListAppsExample {

    private ListAppsExample() {
    }

    public static void main(String[] args) {
        Env.requireApiKey();

        try (IosClient client = Env.client()) {
            String udid = Env.resolveUdid(client);
            if (udid == null) {
                System.out.println("SKIP ListAppsExample: no device attached.");
                return;
            }

            Device device = client.device(udid);
            System.out.println("Listing installed apps on " + udid + " ...");

            // GET /device/{udid}/apps/
            List<AppInfo> apps = device.apps().list();
            System.out.println("Found " + apps.size() + " app(s):");
            for (AppInfo app : apps) {
                System.out.printf("  - %s  (%s %s)%n",
                        app.getCfBundleIdentifier(),
                        app.getCfBundleName(),
                        app.getCfBundleShortVersionString());
            }
        }
    }
}
