package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.StatusOk;
import com.github.danielpaulus.goios.generated.model.UnlockToken;

/**
 * MDM operations for a single device ({@code /mdm/*}). All take a supervision
 * identity ({@code p12}) uploaded via {@code multipart/form-data}.
 */
public final class Mdm {

    private final Device d;

    Mdm(Device d) {
        this.d = d;
    }

    /** Query MDM security info ({@code POST /mdm/security-info}). */
    public Object securityInfo(byte[] p12, String password) {
        return d.http().multipart("POST", d.devicePath("/mdm/security-info"), null, RawHttp.parts(
                RawHttp.Part.file("p12", "identity.p12", p12),
                password == null ? null : RawHttp.Part.field("password", password)),
                Object.class);
    }

    /** Fetch the escrow unlock token ({@code POST /mdm/fetch-unlock-token}). */
    public UnlockToken fetchUnlockToken(byte[] p12, String password) {
        return d.http().multipart("POST", d.devicePath("/mdm/fetch-unlock-token"), null, RawHttp.parts(
                RawHttp.Part.file("p12", "identity.p12", p12),
                password == null ? null : RawHttp.Part.field("password", password)),
                UnlockToken.class);
    }

    /** Clear the device passcode via MDM ({@code POST /mdm/clear-passcode}). */
    public StatusOk clearPasscode(byte[] p12, String password, String token) {
        return d.http().multipart("POST", d.devicePath("/mdm/clear-passcode"), null, RawHttp.parts(
                RawHttp.Part.file("p12", "identity.p12", p12),
                RawHttp.Part.field("token", token),
                password == null ? null : RawHttp.Part.field("password", password)),
                StatusOk.class);
    }

    /** Clear the Screen Time password via MDM ({@code POST /mdm/clear-screen-time-password}). */
    public StatusOk clearScreenTimePassword(byte[] p12, String password) {
        return d.http().multipart("POST", d.devicePath("/mdm/clear-screen-time-password"), null, RawHttp.parts(
                RawHttp.Part.file("p12", "identity.p12", p12),
                password == null ? null : RawHttp.Part.field("password", password)),
                StatusOk.class);
    }
}
