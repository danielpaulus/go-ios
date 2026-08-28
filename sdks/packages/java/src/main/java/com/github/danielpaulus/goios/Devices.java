package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.DeviceEntry;
import com.github.danielpaulus.goios.generated.model.DeviceList;

import java.util.ArrayList;
import java.util.List;

/** Fleet-level device operations. */
public final class Devices {

    private final RawHttp http;

    Devices(RawHttp http) {
        this.http = http;
    }

    /** List attached devices ({@code GET /list}). */
    public List<DeviceEntry> list() {
        DeviceList envelope = http.getJson("/list", null, DeviceList.class);
        return envelope == null || envelope.getDeviceList() == null
                ? List.of() : envelope.getDeviceList();
    }

    /** Convenience: the udids ({@code properties.serialNumber}) of attached devices. */
    public List<String> udids() {
        List<String> out = new ArrayList<>();
        for (DeviceEntry e : list()) {
            String u = udid(e);
            if (u != null) {
                out.add(u);
            }
        }
        return out;
    }

    /** Null-safe accessor for a {@link DeviceEntry}'s udid ({@code properties.serialNumber}). */
    public static String udid(DeviceEntry entry) {
        if (entry == null || entry.getProperties() == null) {
            return null;
        }
        return entry.getProperties().getSerialNumber();
    }
}
