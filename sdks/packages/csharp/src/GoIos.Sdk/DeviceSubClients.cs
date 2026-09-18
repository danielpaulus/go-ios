using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>Crash-report operations for a single device.</summary>
public sealed class CrashesClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    internal CrashesClient(string udid, Gen.DefaultApi api) { _udid = udid; _api = api; }

    /// <summary>List crash-report file names, optionally filtered by glob <paramref name="pattern"/> (<c>GET /crashes</c>).</summary>
    public Task<GenModel.CrashListing> ListAsync(string? pattern = null, CancellationToken cancellationToken = default)
        => _api.DevicesListCrashesAsync(_udid, pattern, cancellationToken);

    /// <summary>
    /// Remove (copy-out then delete) crash reports matching <paramref name="pattern"/>
    /// under working directory <paramref name="cwd"/> (defaults to <c>"."</c>) (<c>DELETE /crashes</c>).
    /// </summary>
    public Task<GenModel.GenericResponse> RemoveAsync(string pattern, string cwd = ".", CancellationToken cancellationToken = default)
        => _api.DevicesRemoveCrashesAsync(_udid, cwd, pattern, cancellationToken);
}

/// <summary>
/// Media / presentation operations: wallpaper, SpringBoard icon layout and the
/// pasteboard. Binary and multipart transfers use the raw HTTP pipeline.
/// </summary>
public sealed class MediaClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;
    internal MediaClient(string udid, Gen.DefaultApi api, RawHttp raw) { _udid = udid; _api = api; _raw = raw; }

    /// <summary>Get the current wallpaper as PNG bytes (<c>GET /wallpaper</c>).</summary>
    public Task<byte[]> WallpaperAsync(CancellationToken cancellationToken = default)
        => _raw.GetBytesAsync($"api/v1/device/{Esc(_udid)}/wallpaper", "image/png", cancellationToken);

    /// <summary>
    /// Set the wallpaper (<c>PUT /wallpaper</c>, supervised). Requires a supervision
    /// identity <paramref name="p12"/>; <paramref name="screen"/> is <c>home</c>,
    /// <c>lock</c> or <c>both</c>.
    /// </summary>
    public Task<GenModel.GenericResponse> SetWallpaperAsync(
        byte[] image, byte[] p12, string? password = null, string? screen = null,
        string imageFileName = "wallpaper.png", CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent
        {
            { Octet(image), "image", imageFileName },
            { Octet(p12), "p12", "supervision.p12" },
        };
        if (!string.IsNullOrEmpty(password)) form.Add(new StringContent(password), "password");
        if (!string.IsNullOrEmpty(screen)) form.Add(new StringContent(screen), "screen");
        var req = _raw.NewRequest(HttpMethod.Put, $"api/v1/device/{Esc(_udid)}/wallpaper");
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    /// <summary>Get the SpringBoard icon layout (<c>GET /icon-layout</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> IconLayoutAsync(CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.DevicesGetIconLayoutAsync(_udid, cancellationToken).ConfigureAwait(false));

    /// <summary>Set the SpringBoard icon layout (<c>PUT /icon-layout</c>).</summary>
    public Task<GenModel.GenericResponse> SetIconLayoutAsync(object layout, CancellationToken cancellationToken = default)
        => _api.DevicesSetIconLayoutAsync(_udid, layout, cancellationToken);

    /// <summary>Read the device pasteboard (<c>GET /pasteboard</c>).</summary>
    public Task<GenModel.PasteboardContent> PasteboardAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetPasteboardAsync(_udid, cancellationToken);

    /// <summary>Set the device pasteboard text (<c>PUT /pasteboard</c>, <c>text/plain</c> body).</summary>
    public Task<GenModel.GenericResponse> SetPasteboardAsync(string text, CancellationToken cancellationToken = default)
    {
        var req = _raw.NewRequest(HttpMethod.Put, $"api/v1/device/{Esc(_udid)}/pasteboard");
        req.Content = new StringContent(text, Encoding.UTF8, "text/plain");
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    private static ByteArrayContent Octet(byte[] bytes)
    {
        var c = new ByteArrayContent(bytes);
        c.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        return c;
    }

    private static string Esc(string s) => Uri.EscapeDataString(s);
}

/// <summary>Device settings toggles: AssistiveTouch, time format and Wi-Fi.</summary>
public sealed class SettingsClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    internal SettingsClient(string udid, Gen.DefaultApi api) { _udid = udid; _api = api; }

    /// <summary>Get the AssistiveTouch enabled state (<c>GET /assistivetouch</c>).</summary>
    public Task<GenModel.AssistiveTouchState> AssistiveTouchAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetAssistiveTouchAsync(_udid, cancellationToken);

    /// <summary>Enable or disable AssistiveTouch (<c>PUT /assistivetouch</c>).</summary>
    public Task<GenModel.AssistiveTouchState> SetAssistiveTouchAsync(bool enabled, CancellationToken cancellationToken = default)
        => _api.DevicesSetAssistiveTouchAsync(_udid, new GenModel.EnabledRequest(enabled), cancellationToken);

    /// <summary>Get the 24-hour time-format state (<c>GET /timeformat</c>).</summary>
    public Task<GenModel.TimeFormatState> TimeFormatAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetTimeFormatAsync(_udid, cancellationToken);

    /// <summary>Set whether the device uses a 24-hour clock (<c>PUT /timeformat</c>).</summary>
    public Task<GenModel.TimeFormatState> SetTimeFormatAsync(bool uses24Hour, CancellationToken cancellationToken = default)
        => _api.DevicesSetTimeFormatAsync(_udid, new GenModel.TimeFormatRequest(uses24Hour), cancellationToken);

    /// <summary>Join a Wi-Fi network (<c>PUT /wifi</c>).</summary>
    public Task<GenModel.GenericResponse> SetWifiAsync(
        string ssid, string? password = null, string? encType = null, CancellationToken cancellationToken = default)
        => _api.DevicesSetWifiAsync(
            _udid,
            new GenModel.WifiRequest(ssid) { Password = password!, EncType = encType! },
            cancellationToken);

    /// <summary>Forget / remove a Wi-Fi network by <paramref name="ssid"/> (<c>DELETE /wifi</c>).</summary>
    public Task<GenModel.GenericResponse> RemoveWifiAsync(string ssid, CancellationToken cancellationToken = default)
        => _api.DevicesRemoveWifiAsync(_udid, ssid, cancellationToken);
}

/// <summary>MDM (supervised) operations for a single device. All require a supervision identity.</summary>
public sealed class MdmClient
{
    private readonly string _udid;
    private readonly RawHttp _raw;
    internal MdmClient(string udid, RawHttp raw) { _udid = udid; _raw = raw; }

    /// <summary>Fetch device security info (<c>POST /mdm/security-info</c>).</summary>
    public Task<IReadOnlyDictionary<string, object?>> SecurityInfoAsync(
        byte[] p12, string? password = null, CancellationToken cancellationToken = default)
        => PostP12DictAsync("security-info", p12, password, null, cancellationToken);

    /// <summary>Fetch the escrow unlock token (<c>POST /mdm/fetch-unlock-token</c>).</summary>
    public async Task<GenModel.UnlockToken> FetchUnlockTokenAsync(
        byte[] p12, string? password = null, CancellationToken cancellationToken = default)
    {
        var req = BuildP12Request("fetch-unlock-token", p12, password, null);
        return await _raw.SendJsonAsync<GenModel.UnlockToken>(req, cancellationToken).ConfigureAwait(false);
    }

    /// <summary>Clear the device passcode using an escrow <paramref name="token"/> (<c>POST /mdm/clear-passcode</c>).</summary>
    public Task<GenModel.StatusOk> ClearPasscodeAsync(
        byte[] p12, string token, string? password = null, CancellationToken cancellationToken = default)
    {
        var req = BuildP12Request("clear-passcode", p12, password, form =>
            form.Add(new StringContent(token), "token"));
        return _raw.SendJsonAsync<GenModel.StatusOk>(req, cancellationToken);
    }

    /// <summary>Clear the Screen Time passcode (<c>POST /mdm/clear-screen-time-password</c>).</summary>
    public Task<GenModel.StatusOk> ClearScreenTimePasswordAsync(
        byte[] p12, string? password = null, CancellationToken cancellationToken = default)
    {
        var req = BuildP12Request("clear-screen-time-password", p12, password, null);
        return _raw.SendJsonAsync<GenModel.StatusOk>(req, cancellationToken);
    }

    private async Task<IReadOnlyDictionary<string, object?>> PostP12DictAsync(
        string leaf, byte[] p12, string? password, Action<MultipartFormDataContent>? extra, CancellationToken ct)
    {
        var req = BuildP12Request(leaf, p12, password, extra);
        var text = await _raw.SendTextAsync(req, ct).ConfigureAwait(false);
        return JsonHelpers.ToDictionary(text);
    }

    private HttpRequestMessage BuildP12Request(
        string leaf, byte[] p12, string? password, Action<MultipartFormDataContent>? extra)
    {
        var p12Content = new ByteArrayContent(p12);
        p12Content.Headers.ContentType = new MediaTypeHeaderValue("application/x-pkcs12");
        var form = new MultipartFormDataContent { { p12Content, "p12", "supervision.p12" } };
        if (!string.IsNullOrEmpty(password)) form.Add(new StringContent(password), "password");
        extra?.Invoke(form);
        var req = _raw.NewRequest(HttpMethod.Post, $"api/v1/device/{Uri.EscapeDataString(_udid)}/mdm/{leaf}");
        req.Content = form;
        return req;
    }
}

/// <summary>Global HTTP-proxy configuration for a single device.</summary>
public sealed class ProxyClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;
    internal ProxyClient(string udid, Gen.DefaultApi api, RawHttp raw) { _udid = udid; _api = api; _raw = raw; }

    /// <summary>
    /// Configure the global HTTP proxy (<c>PUT /httpproxy</c>, supervised). Requires a
    /// supervision identity <paramref name="p12"/>.
    /// </summary>
    public Task<GenModel.GenericResponse> SetHttpProxyAsync(
        string host, string port, byte[] p12,
        string? user = null, string? pass = null, string? password = null,
        CancellationToken cancellationToken = default)
    {
        var p12Content = new ByteArrayContent(p12);
        p12Content.Headers.ContentType = new MediaTypeHeaderValue("application/x-pkcs12");
        var form = new MultipartFormDataContent
        {
            { new StringContent(host), "host" },
            { new StringContent(port), "port" },
            { p12Content, "p12", "supervision.p12" },
        };
        if (!string.IsNullOrEmpty(user)) form.Add(new StringContent(user), "user");
        if (!string.IsNullOrEmpty(pass)) form.Add(new StringContent(pass), "pass");
        if (!string.IsNullOrEmpty(password)) form.Add(new StringContent(password), "password");
        var req = _raw.NewRequest(HttpMethod.Put, $"api/v1/device/{Uri.EscapeDataString(_udid)}/httpproxy");
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    /// <summary>Remove the global HTTP proxy (<c>DELETE /httpproxy</c>).</summary>
    public Task<GenModel.GenericResponse> RemoveHttpProxyAsync(CancellationToken cancellationToken = default)
        => _api.DevicesRemoveHttpProxyAsync(_udid, cancellationToken);
}
