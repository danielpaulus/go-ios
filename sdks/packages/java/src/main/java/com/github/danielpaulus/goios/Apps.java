package com.github.danielpaulus.goios;

import com.fasterxml.jackson.core.type.TypeReference;
import com.github.danielpaulus.goios.generated.model.AppInfo;
import com.github.danielpaulus.goios.generated.model.GenericResponse;

import java.util.List;
import java.util.Map;

/** App-management operations for a single device. */
public final class Apps {

    private final Device d;

    Apps(Device d) {
        this.d = d;
    }

    /** List installed apps ({@code GET /apps/}). */
    public List<AppInfo> list() {
        List<AppInfo> apps = d.http().getJson(d.devicePath("/apps/"), null,
                new TypeReference<List<AppInfo>>() { });
        return apps == null ? List.of() : apps;
    }

    /** Launch an app by bundle id ({@code POST /apps/launch}). */
    public GenericResponse launch(String bundleId) {
        Map<String, String> q = RawHttp.query();
        q.put("bundleID", bundleId);
        return d.http().postJson(d.devicePath("/apps/launch"), q, null, GenericResponse.class);
    }

    /** Kill a running app by bundle id ({@code POST /apps/kill}). */
    public GenericResponse kill(String bundleId) {
        Map<String, String> q = RawHttp.query();
        q.put("bundleID", bundleId);
        return d.http().postJson(d.devicePath("/apps/kill"), q, null, GenericResponse.class);
    }

    /** Install an {@code .ipa}/{@code .app} archive ({@code POST /apps/install}), multipart. */
    public GenericResponse install(byte[] ipa) {
        return d.http().multipart("POST", d.devicePath("/apps/install"), null,
                RawHttp.parts(RawHttp.Part.file("file", "app.ipa", ipa)), GenericResponse.class);
    }

    /** Uninstall an app by bundle id ({@code POST /apps/uninstall}). */
    public GenericResponse uninstall(String bundleId) {
        Map<String, String> q = RawHttp.query();
        q.put("bundleID", bundleId);
        return d.http().postJson(d.devicePath("/apps/uninstall"), q, null, GenericResponse.class);
    }
}
