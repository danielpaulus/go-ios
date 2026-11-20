package com.github.danielpaulus.goios;

import com.github.danielpaulus.goios.generated.model.ProvisioningResult;

/**
 * Host-scoped (device-free) app-signing operations ({@code /sign/*}).
 *
 * <p>{@link #app} and {@link #certificate} return raw binary artifacts (a signed
 * IPA and an {@code application/x-pkcs12} P12 respectively); {@link #provision}
 * returns a JSON envelope with the base64 artifacts.
 */
public final class Sign {

    private final RawHttp http;

    Sign(RawHttp http) {
        this.http = http;
    }

    /**
     * Resign an app/IPA with a signing identity + provisioning profile
     * ({@code POST /sign/app}); returns the signed IPA bytes.
     */
    public byte[] app(byte[] ipa, byte[] p12file, byte[] profile,
                      String p12password, String bundleId) {
        return http.multipartBytes("POST", "/sign/app", null, RawHttp.parts(
                RawHttp.Part.file("ipa", "app.ipa", ipa),
                RawHttp.Part.file("p12file", "identity.p12", p12file),
                RawHttp.Part.file("profile", "profile.mobileprovision", profile),
                p12password == null ? null : RawHttp.Part.field("p12password", p12password),
                bundleId == null ? null : RawHttp.Part.field("bundleid", bundleId)));
    }

    /**
     * Create one App Store Connect signing certificate
     * ({@code POST /sign/certificate}); returns the P12 (cert + private key) bytes.
     */
    public byte[] certificate(byte[] ascPrivateKey, String ascKeyId, String ascIssuerId,
                              boolean revokeExisting, String p12password) {
        return http.multipartBytes("POST", "/sign/certificate", null, RawHttp.parts(
                RawHttp.Part.file("asc-private-key", "AuthKey.p8", ascPrivateKey),
                RawHttp.Part.field("asc-key-id", ascKeyId),
                RawHttp.Part.field("asc-issuer-id", ascIssuerId),
                revokeExisting ? RawHttp.Part.field("revoke-existing", "true") : null,
                p12password == null ? null : RawHttp.Part.field("p12password", p12password)));
    }

    /**
     * Create a bundle id, development certificate and provisioning profile
     * ({@code POST /sign/provision}); returns a JSON envelope with the artifacts.
     */
    public ProvisioningResult provision(byte[] ascPrivateKey, String ascKeyId, String ascIssuerId,
                                        String bundleId, String udid, String bundleName,
                                        String profileName, String deviceName, String certificateId,
                                        boolean revokeExisting, String p12password) {
        return http.multipart("POST", "/sign/provision", null, RawHttp.parts(
                RawHttp.Part.file("asc-private-key", "AuthKey.p8", ascPrivateKey),
                RawHttp.Part.field("asc-key-id", ascKeyId),
                RawHttp.Part.field("asc-issuer-id", ascIssuerId),
                RawHttp.Part.field("bundleid", bundleId),
                RawHttp.Part.field("udid", udid),
                bundleName == null ? null : RawHttp.Part.field("bundlename", bundleName),
                profileName == null ? null : RawHttp.Part.field("profilename", profileName),
                deviceName == null ? null : RawHttp.Part.field("devicename", deviceName),
                certificateId == null ? null : RawHttp.Part.field("certificate-id", certificateId),
                revokeExisting ? RawHttp.Part.field("revoke-existing", "true") : null,
                p12password == null ? null : RawHttp.Part.field("p12password", p12password)),
                ProvisioningResult.class);
    }
}
