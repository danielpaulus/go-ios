using System.Net.Http;
using System.Net.Http.Headers;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>
/// Host-scoped (device-free) code-signing operations backed by App Store Connect
/// (<c>ios sign ...</c>). Obtained via <see cref="IosClient.Sign"/>. Uploads are
/// multipart; the certificate/app results are returned as raw bytes.
/// </summary>
public sealed class SignClient
{
    private readonly RawHttp _raw;
    internal SignClient(RawHttp raw) => _raw = raw;

    /// <summary>
    /// Create one App Store Connect signing certificate and return its P12
    /// (certificate + private key) bytes (<c>POST /sign/certificate</c>). The
    /// generated P12 password is echoed back in the <c>X-P12-Password</c> response
    /// header.
    /// </summary>
    public Task<byte[]> CertificateAsync(
        byte[] ascPrivateKey, string ascKeyId, string ascIssuerId,
        bool revokeExisting = false, string? p12Password = null,
        CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent
        {
            { Octet(ascPrivateKey), "asc-private-key", "AuthKey.p8" },
            { new StringContent(ascKeyId), "asc-key-id" },
            { new StringContent(ascIssuerId), "asc-issuer-id" },
        };
        if (revokeExisting) form.Add(new StringContent("true"), "revoke-existing");
        if (!string.IsNullOrEmpty(p12Password)) form.Add(new StringContent(p12Password), "p12password");

        var req = _raw.NewRequest(HttpMethod.Post, "api/v1/sign/certificate");
        req.Content = form;
        req.Headers.Accept.ParseAdd("application/x-pkcs12");
        return _raw.SendBytesAsync(req, cancellationToken);
    }

    /// <summary>
    /// Create a bundle id, development certificate and provisioning profile for a
    /// device and return both artifacts base64-encoded (<c>POST /sign/provision</c>).
    /// </summary>
    public Task<GenModel.ProvisioningResult> ProvisionAsync(
        byte[] ascPrivateKey, string ascKeyId, string ascIssuerId,
        string bundleId, string udid,
        string? bundleName = null, string? profileName = null, string? deviceName = null,
        string? certificateId = null, bool revokeExisting = false, string? p12Password = null,
        CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent
        {
            { Octet(ascPrivateKey), "asc-private-key", "AuthKey.p8" },
            { new StringContent(ascKeyId), "asc-key-id" },
            { new StringContent(ascIssuerId), "asc-issuer-id" },
            { new StringContent(bundleId), "bundleid" },
            { new StringContent(udid), "udid" },
        };
        if (!string.IsNullOrEmpty(bundleName)) form.Add(new StringContent(bundleName), "bundlename");
        if (!string.IsNullOrEmpty(profileName)) form.Add(new StringContent(profileName), "profilename");
        if (!string.IsNullOrEmpty(deviceName)) form.Add(new StringContent(deviceName), "devicename");
        if (!string.IsNullOrEmpty(certificateId)) form.Add(new StringContent(certificateId), "certificateId");
        if (revokeExisting) form.Add(new StringContent("true"), "revoke-existing");
        if (!string.IsNullOrEmpty(p12Password)) form.Add(new StringContent(p12Password), "p12password");

        var req = _raw.NewRequest(HttpMethod.Post, "api/v1/sign/provision");
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.ProvisioningResult>(req, cancellationToken);
    }

    /// <summary>
    /// Resign an uploaded app/IPA with a P12 identity and provisioning profile and
    /// return the signed IPA bytes (<c>POST /sign/app</c>).
    /// </summary>
    public Task<byte[]> AppAsync(
        byte[] ipa, byte[] p12File, byte[] profile,
        string? p12Password = null, string? bundleId = null,
        CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent
        {
            { Octet(ipa), "ipa", "app.ipa" },
            { P12(p12File), "p12file", "signing.p12" },
            { Octet(profile), "profile", "profile.mobileprovision" },
        };
        if (!string.IsNullOrEmpty(p12Password)) form.Add(new StringContent(p12Password), "p12password");
        if (!string.IsNullOrEmpty(bundleId)) form.Add(new StringContent(bundleId), "bundleid");

        var req = _raw.NewRequest(HttpMethod.Post, "api/v1/sign/app");
        req.Content = form;
        req.Headers.Accept.ParseAdd("application/octet-stream");
        return _raw.SendBytesAsync(req, cancellationToken);
    }

    private static ByteArrayContent Octet(byte[] bytes)
    {
        var c = new ByteArrayContent(bytes);
        c.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        return c;
    }

    private static ByteArrayContent P12(byte[] bytes)
    {
        var c = new ByteArrayContent(bytes);
        c.Headers.ContentType = new MediaTypeHeaderValue("application/x-pkcs12");
        return c;
    }
}

/// <summary>
/// Host-scoped device-preparation helpers (<c>ios prepare ...</c>). Obtained via
/// <see cref="IosClient.Prepare"/>. The device-scoped preparation flow itself is
/// <see cref="DeviceClient.PrepareAsync"/>.
/// </summary>
public sealed class PrepareClient
{
    private readonly Gen.DefaultApi _api;
    internal PrepareClient(Gen.DefaultApi api) => _api = api;

    /// <summary>
    /// Generate a self-signed supervision identity and return the DER (base64) and
    /// PEM for the certificate and private key (<c>POST /prepare/create-cert</c>).
    /// </summary>
    public Task<GenModel.SupervisionCert> CreateCertAsync(CancellationToken cancellationToken = default)
        => _api.PrepareCreateCertAsync(cancellationToken);

    /// <summary>List the setup-pane skip options usable when preparing a device (<c>GET /prepare/skip-options</c>).</summary>
    public Task<GenModel.PrepareSkipOptions> SkipOptionsAsync(CancellationToken cancellationToken = default)
        => _api.GetPrepareSkipOptionsAsync(cancellationToken);
}
