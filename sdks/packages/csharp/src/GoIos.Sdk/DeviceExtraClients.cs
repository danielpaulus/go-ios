using System.Net.Http;
using System.Net.Http.Headers;
using System.Text.Json;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>
/// AFC file-sync operations (<c>ios fsync ...</c>) for a single device. Every
/// call takes a device <c>path</c> and an optional app <c>bundleId</c> that
/// scopes the operation to that app's container (otherwise the media dir).
/// Obtained via <see cref="DeviceClient.Fsync"/>.
/// </summary>
public sealed class FsyncClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    internal FsyncClient(string udid, Gen.DefaultApi api, RawHttp raw)
    {
        _udid = udid;
        _api = api;
        _raw = raw;
    }

    /// <summary>List the immediate entries under <paramref name="path"/> (<c>GET /fsync/ls</c>).</summary>
    public Task<GenModel.FsyncListing> LsAsync(
        string? path = null, string? bundleId = null, CancellationToken cancellationToken = default)
        => _api.FsyncFsyncLsAsync(_udid, bundleId, path, cancellationToken);

    /// <summary>Recursively list the tree under <paramref name="path"/> (<c>GET /fsync/tree</c>).</summary>
    public Task<GenModel.FsyncTreeListing> TreeAsync(
        string? path = null, string? bundleId = null, CancellationToken cancellationToken = default)
        => _api.FsyncFsyncTreeAsync(_udid, bundleId, path, cancellationToken);

    /// <summary>Download a file over AFC and return its raw bytes (<c>GET /fsync/pull</c>).</summary>
    public Task<byte[]> PullAsync(
        string path, string? bundleId = null, CancellationToken cancellationToken = default)
    {
        var qs = BuildQuery(("bundleID", bundleId), ("path", path));
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/fsync/pull{qs}");
        req.Headers.Accept.ParseAdd("application/octet-stream");
        return _raw.SendBytesAsync(req, cancellationToken);
    }

    /// <summary>Upload raw bytes to <paramref name="path"/> over AFC (<c>POST /fsync/push</c>, octet-stream body).</summary>
    public Task<GenModel.FsyncPushResult> PushAsync(
        string path, byte[] content, string? bundleId = null, CancellationToken cancellationToken = default)
    {
        var qs = BuildQuery(("bundleID", bundleId), ("path", path));
        var req = _raw.NewRequest(HttpMethod.Post, $"api/v1/device/{Esc(_udid)}/fsync/push{qs}");
        var body = new ByteArrayContent(content);
        body.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        req.Content = body;
        return _raw.SendJsonAsync<GenModel.FsyncPushResult>(req, cancellationToken);
    }

    /// <summary>Upload a local file's contents to <paramref name="path"/> over AFC (<c>POST /fsync/push</c>).</summary>
    public async Task<GenModel.FsyncPushResult> PushAsync(
        string path, string localPath, string? bundleId = null, CancellationToken cancellationToken = default)
    {
        var bytes = await File.ReadAllBytesAsync(localPath, cancellationToken).ConfigureAwait(false);
        return await PushAsync(path, bytes, bundleId, cancellationToken).ConfigureAwait(false);
    }

    /// <summary>Remove a file or directory at <paramref name="path"/> (<c>DELETE /fsync/rm</c>).</summary>
    public Task<GenModel.FsyncMessage> RmAsync(
        string path, bool recursive = false, string? bundleId = null, CancellationToken cancellationToken = default)
        => _api.FsyncFsyncRmAsync(_udid, path, bundleId, recursive, cancellationToken);

    /// <summary>Create a directory at <paramref name="path"/> (<c>POST /fsync/mkdir</c>).</summary>
    public Task<GenModel.FsyncMessage> MkdirAsync(
        string path, string? bundleId = null, CancellationToken cancellationToken = default)
        => _api.FsyncFsyncMkdirAsync(_udid, path, bundleId, cancellationToken);

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

/// <summary>
/// WebInspector (Safari / WKWebView remote debugging) operations for a single
/// device. Obtained via <see cref="DeviceClient.WebInspector"/>.
/// </summary>
public sealed class WebInspectorClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;

    internal WebInspectorClient(string udid, Gen.DefaultApi api)
    {
        _udid = udid;
        _api = api;
    }

    /// <summary>List the inspectable pages/targets on the device (<c>GET /webinspector/pages</c>).</summary>
    public async Task<IReadOnlyList<IReadOnlyDictionary<string, object?>>> PagesAsync(
        CancellationToken cancellationToken = default)
    {
        var raw = await _api.WebInspectorWebInspectorPagesAsync(_udid, cancellationToken).ConfigureAwait(false);
        return (raw ?? new List<object>()).Select(JsonHelpers.ToDictionary).ToList();
    }

    /// <summary>Open a URL (or a bundle id's default page) for inspection (<c>POST /webinspector/launch</c>).</summary>
    public Task<GenModel.WebInspectorLaunchResult> LaunchAsync(
        string? url = null, string? bundleId = null, CancellationToken cancellationToken = default)
        => _api.WebInspectorWebInspectorLaunchAsync(
            _udid, null, new GenModel.WebInspectorLaunchRequest(url: url!, bundleId: bundleId!), cancellationToken);

    /// <summary>Evaluate a JavaScript <paramref name="script"/> against a page (<c>POST /webinspector/eval</c>).</summary>
    public Task<GenModel.WebInspectorEvalResult> EvalAsync(
        string script, string? page = null, string? bundleId = null, CancellationToken cancellationToken = default)
        => _api.WebInspectorWebInspectorEvalAsync(
            _udid, new GenModel.WebInspectorEvalRequest(page: page!, bundleId: bundleId!, script: script),
            cancellationToken);
}

/// <summary>
/// UI automation operations backed by WebDriverAgent or DeviceKit
/// (<c>ios ui ...</c>). Every call accepts optional <paramref name="backend"/>
/// (<c>wda</c> | <c>devicekit</c>), <c>wdaUrl</c> and <c>timeout</c> (seconds)
/// selectors. Obtained via <see cref="DeviceClient.Ui"/>.
/// </summary>
public sealed class UiClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    internal UiClient(string udid, Gen.DefaultApi api, RawHttp raw)
    {
        _udid = udid;
        _api = api;
        _raw = raw;
    }

    /// <summary>Optional backend selectors shared by every UI call.</summary>
    public sealed class Options
    {
        /// <summary>Backend to target: <c>wda</c> (default) or <c>devicekit</c>.</summary>
        public string? Backend { get; set; }
        /// <summary>Forwarded backend base URL (defaults per backend).</summary>
        public string? WdaUrl { get; set; }
        /// <summary>Per-request HTTP timeout in seconds (default 60).</summary>
        public int? Timeout { get; set; }
    }

    // --- Gestures ----------------------------------------------------------

    /// <summary>Tap at (<paramref name="x"/>, <paramref name="y"/>) (<c>POST /ui/tap</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> TapAsync(
        int x, int y, Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiTapAsync(
            _udid, new GenModel.UITapRequest(x, y), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Swipe from (<paramref name="x1"/>, <paramref name="y1"/>) to (<paramref name="x2"/>, <paramref name="y2"/>) (<c>POST /ui/swipe</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> SwipeAsync(
        int x1, int y1, int x2, int y2, double duration = 0,
        Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiSwipeAsync(
            _udid, new GenModel.UISwipeRequest(x1, y1, x2, y2, duration), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Long-press at (<paramref name="x"/>, <paramref name="y"/>) for <paramref name="duration"/> seconds (<c>POST /ui/longpress</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> LongPressAsync(
        int x, int y, double duration = 1.0,
        Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiLongPressAsync(
            _udid, new GenModel.UILongPressRequest(x, y, duration), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Type <paramref name="text"/> into the focused element (<c>POST /ui/type</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> TypeAsync(
        string text, Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiTypeAsync(
            _udid, new GenModel.UITypeRequest(text), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Press a hardware/system button by <paramref name="name"/> (e.g. <c>home</c>) (<c>POST /ui/button</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> ButtonAsync(
        string name, Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiButtonAsync(
            _udid, new GenModel.UIButtonRequest(name), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    // --- Introspection -----------------------------------------------------

    /// <summary>Capture the screen and return raw PNG bytes (<c>GET /ui/screenshot</c>).</summary>
    public Task<byte[]> ScreenshotAsync(Options? options = null, CancellationToken cancellationToken = default)
    {
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/ui/screenshot{Query(options)}");
        req.Headers.Accept.ParseAdd("image/png");
        return _raw.SendBytesAsync(req, cancellationToken);
    }

    /// <summary>Return the current view hierarchy (XML for WDA) (<c>GET /ui/source</c>).</summary>
    public Task<string> SourceAsync(Options? options = null, CancellationToken cancellationToken = default)
    {
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/ui/source{Query(options)}");
        req.Headers.Accept.ParseAdd("application/xml");
        return _raw.SendTextAsync(req, cancellationToken);
    }

    /// <summary>Get the window/screen size (<c>GET /ui/size</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> SizeAsync(
        Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiWindowSizeAsync(
            _udid, B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Get the current device orientation (<c>GET /ui/orientation</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> OrientationAsync(
        Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiGetOrientationAsync(
            _udid, B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Set the device orientation (<c>PUT /ui/orientation</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> SetOrientationAsync(
        string orientation, Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiSetOrientationAsync(
            _udid, new GenModel.UIOrientationRequest(orientation), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Get the backend's status/health payload (<c>GET /ui/status</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> StatusAsync(
        Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiStatusAsync(
            _udid, B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    // --- App control -------------------------------------------------------

    /// <summary>Launch an app by bundle id via the UI backend (<c>POST /ui/app/launch</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> AppLaunchAsync(
        string bundleId, Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiAppLaunchAsync(
            _udid, new GenModel.UIAppRequest(bundleId), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Terminate an app by bundle id via the UI backend (<c>POST /ui/app/terminate</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> AppTerminateAsync(
        string bundleId, Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiAppTerminateAsync(
            _udid, new GenModel.UIAppRequest(bundleId), B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    /// <summary>Get the currently foregrounded app (<c>POST /ui/app/foreground</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> AppForegroundAsync(
        Options? options = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.UIUiAppForegroundAsync(
            _udid, B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));

    // --- Passthrough / streaming ------------------------------------------

    /// <summary>
    /// Send a raw request through to the backend (<c>POST /ui/api</c>): either an
    /// HTTP passthrough (<paramref name="method"/> + <paramref name="path"/> [+ <paramref name="body"/>])
    /// or a DeviceKit RPC (<paramref name="rpcMethod"/> + <paramref name="rpcParams"/>).
    /// </summary>
    public async Task<IReadOnlyDictionary<string, object?>> ApiAsync(
        string? method = null, string? path = null, string? body = null,
        string? rpcMethod = null, object? rpcParams = null,
        Options? options = null, CancellationToken cancellationToken = default)
    {
        var request = new GenModel.UIAPIRequest
        {
            Method = method!, Path = path!, Body = body!, RpcMethod = rpcMethod!, RpcParams = rpcParams!,
        };
        return JsonHelpers.ToDictionary(await _api.UIUiApiAsync(
            _udid, request, B(options), U(options), T(options), cancellationToken).ConfigureAwait(false));
    }

    /// <summary>
    /// Open a live UI video stream (<c>GET /ui/stream</c>). Returns a raw
    /// <see cref="BinaryStream"/> of MJPEG (default) or H.264 bytes; dispose it to
    /// stop. Honors the <see cref="CancellationToken"/>.
    /// </summary>
    public Task<BinaryStream> StreamAsync(
        Options? options = null, string? codec = null, string? fps = null, string? quality = null,
        string? scale = null, string? bitrate = null, CancellationToken cancellationToken = default)
    {
        var qs = BuildQuery(
            ("backend", options?.Backend), ("wdaUrl", options?.WdaUrl),
            ("timeout", options?.Timeout?.ToString()),
            ("codec", codec), ("fps", fps), ("quality", quality), ("scale", scale), ("bitrate", bitrate));
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/ui/stream{qs}");
        return _raw.OpenBinaryStreamAsync(req, "application/octet-stream", cancellationToken);
    }

    // --- helpers -----------------------------------------------------------

    private static string? B(Options? o) => string.IsNullOrEmpty(o?.Backend) ? null : o!.Backend;
    private static string? U(Options? o) => string.IsNullOrEmpty(o?.WdaUrl) ? null : o!.WdaUrl;
    private static int? T(Options? o) => o?.Timeout;

    private static string Query(Options? o)
        => BuildQuery(("backend", o?.Backend), ("wdaUrl", o?.WdaUrl), ("timeout", o?.Timeout?.ToString()));

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
