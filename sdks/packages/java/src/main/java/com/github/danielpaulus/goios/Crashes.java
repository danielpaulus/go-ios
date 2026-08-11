package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.CrashListing;
import com.github.danielpaulus.goios.generated.model.GenericResponse;

import java.util.Map;

/** Crash-report operations for a single device. */
public final class Crashes {

    private final Device d;

    Crashes(Device d) {
        this.d = d;
    }

    /** List crash reports ({@code GET /crashes}); matches all reports. */
    public CrashListing list() {
        return list("*");
    }

    /** List crash reports matching {@code pattern} ({@code GET /crashes}). */
    public CrashListing list(String pattern) {
        Map<String, String> q = RawHttp.query();
        q.put("pattern", pattern);
        return d.http().getJson(d.devicePath("/crashes"), q, CrashListing.class);
    }

    /** Remove crash reports matching {@code pattern} under the current dir ({@code DELETE /crashes}). */
    public GenericResponse remove(String pattern) {
        return remove(pattern, ".");
    }

    /** Remove crash reports matching {@code pattern} under {@code cwd} ({@code DELETE /crashes}). */
    public GenericResponse remove(String pattern, String cwd) {
        Map<String, String> q = RawHttp.query();
        q.put("cwd", cwd);
        q.put("pattern", pattern);
        return d.http().deleteJson(d.devicePath("/crashes"), q, GenericResponse.class);
    }
}
