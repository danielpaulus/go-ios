package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Devices;
import com.github.danielpaulus.goios.IosClient;
import com.github.danielpaulus.goios.generated.model.DeviceEntry;
import com.github.danielpaulus.goios.generated.model.DeviceProperties;

import java.util.List;

/**
 * Example 1 — list the attached devices.
 *
 * <p>The smallest useful program against the daemon: build an {@link IosClient}
 * from the environment and enumerate every device the daemon can see via
 * {@code GET /list}. This is also the fleet-level entry point used by the other
 * examples to auto-select a device when {@code GO_IOS_UDID} is unset.
 *
 * <p>Run:
 * <pre>{@code
 *   export GO_IOS_BASE_URL=http://localhost:8080   # optional (this is the default)
 *   export GO_IOS_API_KEY=...                       # required
 *   java -cp <classpath> com.github.danielpaulus.goios.examples.ListDevicesExample
 * }</pre>
 */
public final class ListDevicesExample {

    private ListDevicesExample() {
    }

    public static void main(String[] args) {
        // Fail fast with a helpful message if the token is missing.
        Env.requireApiKey();

        // try-with-resources guarantees the underlying HTTP client is released.
        try (IosClient client = Env.client()) {
            String baseUrl = Env.baseUrl();
            System.out.println("Listing devices from "
                    + (baseUrl != null ? baseUrl : "(auto-discovered local daemon)") + " ...");

            // GET /list -> the typed device envelope, unwrapped to a List.
            List<DeviceEntry> devices = client.devices().list();

            if (devices.isEmpty()) {
                System.out.println("No devices attached.");
                return;
            }

            System.out.println("Found " + devices.size() + " device(s):");
            for (DeviceEntry d : devices) {
                // Devices.udid(d) is a null-safe accessor for properties.serialNumber.
                String udid = Devices.udid(d);
                DeviceProperties props = d.getProperties();
                String connType = props == null ? "?" : String.valueOf(props.getConnectionType());
                System.out.printf("  - udid=%s  deviceID=%s  connection=%s%n",
                        udid, d.getDeviceID(), connType);
            }
        }
    }
}
