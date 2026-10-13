using System.Net.Http;
using System.Net.Http.Headers;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>App lifecycle operations for a single device.</summary>
public sealed class AppsClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    internal AppsClient(string udid, Gen.DefaultApi api, RawHttp raw)
    {
        _udid = udid;
        _api = api;
        _raw = raw;
    }

    /// <summary>List installed applications. Each entry is an open Info.plist map.</summary>
    public Task<List<GenModel.AppInfo>> ListAsync(CancellationToken cancellationToken = default)
        => _api.DevicesListAppsAsync(_udid, cancellationToken);

    /// <summary>Launch an application by bundle id.</summary>
    public Task<GenModel.GenericResponse> LaunchAsync(string bundleId, CancellationToken cancellationToken = default)
        => _api.DevicesLaunchAppAsync(_udid, bundleId, cancellationToken);

    /// <summary>Kill a running application by bundle id.</summary>
    public Task<GenModel.GenericResponse> KillAsync(string bundleId, CancellationToken cancellationToken = default)
        => _api.DevicesKillAppAsync(_udid, bundleId, cancellationToken);

    /// <summary>Uninstall an application by bundle id.</summary>
    public Task<GenModel.GenericResponse> UninstallAsync(string bundleId, CancellationToken cancellationToken = default)
        => _api.DevicesUninstallAppAsync(_udid, bundleId, cancellationToken);

    /// <summary>Install an application from a local <c>.ipa</c>/<c>.app</c> archive path.</summary>
    public async Task<GenModel.GenericResponse> InstallAsync(string filePath, CancellationToken cancellationToken = default)
    {
        var bytes = await File.ReadAllBytesAsync(filePath, cancellationToken).ConfigureAwait(false);
        return await InstallAsync(bytes, Path.GetFileName(filePath), cancellationToken).ConfigureAwait(false);
    }

    /// <summary>
    /// Install an application from in-memory archive bytes. Uploaded as multipart
    /// <c>file</c> per the spec (the part must be 1 byte–200 MB).
    /// </summary>
    public Task<GenModel.GenericResponse> InstallAsync(
        byte[] archive, string fileName = "app.ipa", CancellationToken cancellationToken = default)
    {
        var req = _raw.NewRequest(HttpMethod.Post, $"api/v1/device/{Uri.EscapeDataString(_udid)}/apps/install");
        var form = new MultipartFormDataContent();
        var part = new ByteArrayContent(archive);
        part.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        form.Add(part, "file", fileName);
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }
}
