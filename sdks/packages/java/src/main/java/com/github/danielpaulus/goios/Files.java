package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.FileListing;
import com.github.danielpaulus.goios.generated.model.FilePushResult;

import java.util.Map;

/**
 * On-device house-arrest file-service operations for a single device
 * ({@code /files*}). The {@code domain} is one of {@code app}, {@code app-group},
 * {@code crash} or {@code temp}; {@code identifier} is the bundle/group id for
 * the app domains.
 */
public final class Files {

    private final Device d;

    Files(Device d) {
        this.d = d;
    }

    /** List files in a house-arrest domain ({@code GET /files}). */
    public FileListing ls(String domain, String identifier, String path) {
        Map<String, String> q = RawHttp.query();
        q.put("domain", domain);
        if (identifier != null) {
            q.put("identifier", identifier);
        }
        if (path != null) {
            q.put("path", path);
        }
        return d.http().getJson(d.devicePath("/files"), q, FileListing.class);
    }

    /** Pull a file's raw bytes off the device ({@code GET /files/pull}). */
    public byte[] pull(String domain, String identifier, String remote) {
        Map<String, String> q = RawHttp.query();
        q.put("domain", domain);
        q.put("remote", remote);
        if (identifier != null) {
            q.put("identifier", identifier);
        }
        return d.http().getBytes(d.devicePath("/files/pull"), q);
    }

    /** Push raw bytes to a file on the device ({@code POST /files/push}), octet-stream body. */
    public FilePushResult push(String domain, String identifier, String remote, byte[] data) {
        Map<String, String> q = RawHttp.query();
        q.put("domain", domain);
        q.put("remote", remote);
        if (identifier != null) {
            q.put("identifier", identifier);
        }
        return d.http().requestJson("POST", d.devicePath("/files/push"), q, data,
                "application/octet-stream", FilePushResult.class);
    }
}
