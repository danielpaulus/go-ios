package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.FsyncListing;
import com.github.danielpaulus.goios.generated.model.FsyncMessage;
import com.github.danielpaulus.goios.generated.model.FsyncPushResult;
import com.github.danielpaulus.goios.generated.model.FsyncTreeListing;

import java.util.Map;

/**
 * AFC file-transfer operations for a single device ({@code /fsync/*}). When a
 * {@code bundleId} is supplied the operation is scoped to that app's container;
 * otherwise it targets the device media directory.
 */
public final class Fsync {

    private final Device d;

    Fsync(Device d) {
        this.d = d;
    }

    private Map<String, String> scoped(String path, String bundleId) {
        Map<String, String> q = RawHttp.query();
        if (bundleId != null) {
            q.put("bundleID", bundleId);
        }
        if (path != null) {
            q.put("path", path);
        }
        return q;
    }

    /** List a device directory over AFC ({@code GET /fsync/ls}). */
    public FsyncListing ls(String path, String bundleId) {
        return d.http().getJson(d.devicePath("/fsync/ls"), scoped(path, bundleId), FsyncListing.class);
    }

    /** Recursively list a device directory over AFC ({@code GET /fsync/tree}). */
    public FsyncTreeListing tree(String path, String bundleId) {
        return d.http().getJson(d.devicePath("/fsync/tree"), scoped(path, bundleId), FsyncTreeListing.class);
    }

    /** Download a file over AFC ({@code GET /fsync/pull}); returns raw bytes. */
    public byte[] pull(String path, String bundleId) {
        return d.http().getBytes(d.devicePath("/fsync/pull"), scoped(path, bundleId));
    }

    /** Upload a file over AFC ({@code POST /fsync/push}), octet-stream body. */
    public FsyncPushResult push(String path, byte[] data, String bundleId) {
        return d.http().requestJson("POST", d.devicePath("/fsync/push"), scoped(path, bundleId),
                data, "application/octet-stream", FsyncPushResult.class);
    }

    /** Remove a file or directory over AFC ({@code DELETE /fsync/rm}). */
    public FsyncMessage rm(String path, String bundleId, boolean recursive) {
        Map<String, String> q = scoped(path, bundleId);
        if (recursive) {
            q.put("recursive", "true");
        }
        return d.http().deleteJson(d.devicePath("/fsync/rm"), q, FsyncMessage.class);
    }

    public FsyncMessage rm(String path, String bundleId) {
        return rm(path, bundleId, false);
    }

    /** Create a directory over AFC ({@code POST /fsync/mkdir}). */
    public FsyncMessage mkdir(String path, String bundleId) {
        return d.http().postJson(d.devicePath("/fsync/mkdir"), scoped(path, bundleId), null, FsyncMessage.class);
    }
}
