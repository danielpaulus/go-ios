using System.Net.Http;
using System.Net.Http.Headers;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>
/// AFC file-system operations for a single device, scoped to an application
/// sandbox <c>domain</c> (e.g. <c>appDocuments</c>, <c>appContainer</c>,
/// <c>appGroupContainer</c>, <c>media</c>, <c>root</c>). Binary transfers go
/// through the raw HTTP pipeline.
/// </summary>
public sealed class FilesClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    internal FilesClient(string udid, Gen.DefaultApi api, RawHttp raw)
    {
        _udid = udid;
        _api = api;
        _raw = raw;
    }

    /// <summary>
    /// List files under <paramref name="path"/> in the given <paramref name="domain"/>
    /// (<c>GET /files</c>). <paramref name="identifier"/> is the app bundle id when
    /// the domain is app-scoped.
    /// </summary>
    public Task<GenModel.FileListing> LsAsync(
        string domain, string? path = null, string? identifier = null,
        CancellationToken cancellationToken = default)
    {
        var qs = BuildQuery(("domain", domain), ("identifier", identifier), ("path", path));
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/files{qs}");
        return _raw.SendJsonAsync<GenModel.FileListing>(req, cancellationToken);
    }

    /// <summary>Pull a single file's raw bytes (<c>GET /files/pull</c>).</summary>
    public Task<byte[]> PullAsync(
        string domain, string remote, string? identifier = null,
        CancellationToken cancellationToken = default)
    {
        var qs = BuildQuery(("domain", domain), ("identifier", identifier), ("remote", remote));
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/files/pull{qs}");
        req.Headers.Accept.ParseAdd("application/octet-stream");
        return _raw.SendBytesAsync(req, cancellationToken);
    }

    /// <summary>Push bytes to <paramref name="remote"/> as an octet-stream body (<c>POST /files/push</c>).</summary>
    public Task<GenModel.FilePushResult> PushAsync(
        string domain, string remote, byte[] content, string? identifier = null,
        CancellationToken cancellationToken = default)
    {
        var qs = BuildQuery(("domain", domain), ("identifier", identifier), ("remote", remote));
        var req = _raw.NewRequest(HttpMethod.Post, $"api/v1/device/{Esc(_udid)}/files/push{qs}");
        var body = new ByteArrayContent(content);
        body.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        req.Content = body;
        return _raw.SendJsonAsync<GenModel.FilePushResult>(req, cancellationToken);
    }

    /// <summary>Push a local file's contents to <paramref name="remote"/> (<c>POST /files/push</c>).</summary>
    public async Task<GenModel.FilePushResult> PushAsync(
        string domain, string remote, string localPath, string? identifier = null,
        CancellationToken cancellationToken = default)
    {
        var bytes = await File.ReadAllBytesAsync(localPath, cancellationToken).ConfigureAwait(false);
        return await PushAsync(domain, remote, bytes, identifier, cancellationToken).ConfigureAwait(false);
    }

    private static string BuildQuery(params (string Key, string? Value)[] pairs)
    {
        var parts = new List<string>();
        foreach (var (k, v) in pairs)
            if (!string.IsNullOrEmpty(v))
                parts.Add(k + "=" + Uri.EscapeDataString(v));
        return parts.Count > 0 ? "?" + string.Join("&", parts) : "";
    }

    private static string Esc(string s) => Uri.EscapeDataString(s);
}
