using System.Net.Http;
using System.Net.Http.Headers;
using System.Runtime.CompilerServices;
using System.Text.Json;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>
/// Device-scoped operations. Obtained via <see cref="IosClient.Device(string)"/>.
/// </summary>
public sealed class DeviceClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    /// <summary>The device udid this handle is scoped to (its <c>properties.serialNumber</c>).</summary>
    public string Udid => _udid;

    /// <summary>App lifecycle operations for this device.</summary>
    public AppsClient Apps { get; }

    /// <summary>WebDriverAgent (XCUITest) session operations for this device.</summary>
    public WdaClient Wda { get; }

    /// <summary>AFC file-system operations for this device.</summary>
    public FilesClient Files { get; }

    /// <summary>Crash-report operations for this device.</summary>
    public CrashesClient Crashes { get; }

    /// <summary>Wallpaper / icon-layout / pasteboard operations for this device.</summary>
    public MediaClient Media { get; }

    /// <summary>Settings toggles (AssistiveTouch / time format / Wi-Fi) for this device.</summary>
    public SettingsClient Settings { get; }

    /// <summary>MDM (supervised) operations for this device.</summary>
    public MdmClient Mdm { get; }

    /// <summary>Global HTTP-proxy configuration for this device.</summary>
    public ProxyClient Proxy { get; }

    /// <summary>Background-job operations (runtest / runwda / forward) for this device.</summary>
    public JobsClient Jobs { get; }

    /// <summary>AFC file-sync operations (<c>ios fsync</c>) for this device.</summary>
    public FsyncClient Fsync { get; }

    /// <summary>WebInspector (remote web debugging) operations for this device.</summary>
    public WebInspectorClient WebInspector { get; }

    /// <summary>UI automation (WDA / DeviceKit) operations for this device.</summary>
    public UiClient Ui { get; }

    internal DeviceClient(string udid, Gen.DefaultApi api, RawHttp raw)
    {
        _udid = udid;
        _api = api;
        _raw = raw;
        Apps = new AppsClient(udid, api, raw);
        Wda = new WdaClient(udid, api);
        Files = new FilesClient(udid, api, raw);
        Crashes = new CrashesClient(udid, api);
        Media = new MediaClient(udid, api, raw);
        Settings = new SettingsClient(udid, api);
        Mdm = new MdmClient(udid, raw);
        Proxy = new ProxyClient(udid, api, raw);
        Jobs = new JobsClient(udid, api, raw);
        Fsync = new FsyncClient(udid, api, raw);
        WebInspector = new WebInspectorClient(udid, api);
        Ui = new UiClient(udid, api, raw);
    }

    // --- Info / lifecycle --------------------------------------------------

    /// <summary>Get lockdown values plus <c>instruments:*</c> keys for the device.</summary>
    public async Task<IReadOnlyDictionary<string, object?>> InfoAsync(CancellationToken cancellationToken = default)
    {
        var raw = await _api.DevicesGetInfoAsync(_udid, cancellationToken).ConfigureAwait(false);
        return ToDictionary(raw);
    }

    /// <summary>Activate the device.</summary>
    public Task<GenModel.GenericResponse> ActivateAsync(CancellationToken cancellationToken = default)
        => _api.DevicesActivateAsync(_udid, cancellationToken);

    /// <summary>Capture a screenshot and return the raw PNG bytes.</summary>
    public Task<byte[]> ScreenshotAsync(CancellationToken cancellationToken = default)
        => _raw.GetBytesAsync($"api/v1/device/{Esc(_udid)}/screenshot", "image/png", cancellationToken);

    // --- Pairing -----------------------------------------------------------

    /// <summary>
    /// Pair the device. For supervised pairing supply the supervision identity
    /// (<paramref name="p12File"/>) and passphrase (<paramref name="supervisionPassword"/>).
    /// </summary>
    public Task<GenModel.GenericResponse> PairAsync(
        bool supervised = false,
        byte[]? p12File = null,
        string? supervisionPassword = null,
        CancellationToken cancellationToken = default)
    {
        var url = $"api/v1/device/{Esc(_udid)}/pair?supervised={(supervised ? "true" : "false")}";
        var req = _raw.NewRequest(HttpMethod.Post, url);
        if (!string.IsNullOrEmpty(supervisionPassword))
            req.Headers.TryAddWithoutValidation("Supervision-Password", supervisionPassword);

        if (p12File is not null)
        {
            var form = new MultipartFormDataContent();
            var part = new ByteArrayContent(p12File);
            part.Headers.ContentType = new MediaTypeHeaderValue("application/x-pkcs12");
            form.Add(part, "p12file", "supervision.p12");
            req.Content = form;
        }

        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    // --- Conditions --------------------------------------------------------

    /// <summary>List available condition inducer profile types.</summary>
    public Task<List<GenModel.ProfileType>> ConditionsAsync(CancellationToken cancellationToken = default)
        => _api.DevicesListConditionsAsync(_udid, cancellationToken);

    /// <summary>Enable a condition inducer profile.</summary>
    public Task<GenModel.GenericResponse> EnableConditionAsync(
        string profileTypeId, string profileId, CancellationToken cancellationToken = default)
        => _api.DevicesEnableConditionAsync(_udid, profileTypeId, profileId, cancellationToken);

    /// <summary>Disable the active condition inducer profile.</summary>
    public Task<GenModel.GenericResponse> DisableConditionAsync(CancellationToken cancellationToken = default)
        => _api.DevicesDisableConditionAsync(_udid, cancellationToken);

    // --- Developer disk images --------------------------------------------

    /// <summary>List the developer disk images mounted on / known to the device.</summary>
    public Task<List<string>> ImagesAsync(CancellationToken cancellationToken = default)
        => _api.DevicesListImagesAsync(_udid, cancellationToken);

    /// <summary>
    /// Mount a Developer Disk Image. Either let the server auto-resolve and
    /// download the correct image (<paramref name="auto"/> = true, optionally with
    /// <paramref name="baseDir"/>), or stream the raw image bytes
    /// (<paramref name="imageBytes"/>).
    /// </summary>
    public Task<GenModel.GenericResponse> InstallImageAsync(
        bool auto = false,
        string? baseDir = null,
        byte[]? imageBytes = null,
        CancellationToken cancellationToken = default)
    {
        var query = new List<string>();
        if (auto) query.Add("auto=true");
        if (!string.IsNullOrEmpty(baseDir)) query.Add("basedir=" + Uri.EscapeDataString(baseDir));
        var qs = query.Count > 0 ? "?" + string.Join("&", query) : "";

        var req = _raw.NewRequest(HttpMethod.Put, $"api/v1/device/{Esc(_udid)}/image{qs}");
        if (imageBytes is not null)
        {
            var body = new ByteArrayContent(imageBytes);
            body.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
            req.Content = body;
        }
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    // --- Profiles / resets / location -------------------------------------

    /// <summary>List installed configuration profiles.</summary>
    public async Task<IReadOnlyDictionary<string, object?>> ProfilesAsync(CancellationToken cancellationToken = default)
    {
        var raw = await _api.DevicesGetProfilesAsync(_udid, cancellationToken).ConfigureAwait(false);
        return ToDictionary(raw);
    }

    /// <summary>Reset accessibility settings on the device.</summary>
    public Task<GenModel.GenericResponse> ResetAccessibilityAsync(CancellationToken cancellationToken = default)
        => _api.DevicesResetAccessibilityAsync(_udid, cancellationToken);

    /// <summary>Reset the simulated location back to the device's real GPS.</summary>
    public Task<GenModel.GenericResponse> ResetLocationAsync(CancellationToken cancellationToken = default)
        => _api.DevicesResetLocationAsync(_udid, cancellationToken);

    /// <summary>Set a simulated GPS location.</summary>
    public Task<GenModel.GenericResponse> SetLocationAsync(
        double latitude, double longitude, CancellationToken cancellationToken = default)
        => _api.DevicesSetLocationAsync(
            _udid,
            latitude.ToString(System.Globalization.CultureInfo.InvariantCulture),
            longitude.ToString(System.Globalization.CultureInfo.InvariantCulture),
            cancellationToken);

    // --- Device information ------------------------------------------------

    /// <summary>Get the device name (<c>GET /devicename</c>).</summary>
    public Task<GenModel.DeviceName> DeviceNameAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetDeviceNameAsync(_udid, cancellationToken);

    /// <summary>Get the device's current date/time (<c>GET /date</c>).</summary>
    public Task<GenModel.DeviceDate> DateAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetDeviceDateAsync(_udid, cancellationToken);

    /// <summary>Get battery status (<c>GET /battery</c>).</summary>
    public Task<GenModel.BatteryInfo> BatteryAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetBatteryAsync(_udid, cancellationToken);

    /// <summary>Get IORegistry diagnostics (<c>GET /diagnostics</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> DiagnosticsAsync(CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.DevicesGetDiagnosticsAsync(_udid, cancellationToken).ConfigureAwait(false));

    /// <summary>Query one or more MobileGestalt keys (<c>GET /mobilegestalt</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> MobileGestaltAsync(
        IEnumerable<string> keys, CancellationToken cancellationToken = default)
    {
        var list = keys as List<string> ?? new List<string>(keys);
        return JsonHelpers.ToDictionary(await _api.DevicesGetMobileGestaltAsync(_udid, list, cancellationToken).ConfigureAwait(false));
    }

    /// <summary>List running processes (<c>GET /processes</c>). Set <paramref name="apps"/> to include app metadata.</summary>
    public Task<List<GenModel.ProcessInfo>> ProcessesAsync(bool? apps = null, CancellationToken cancellationToken = default)
        => _api.DevicesGetProcessesAsync(_udid, apps, cancellationToken);

    /// <summary>
    /// Get lockdown values (<c>GET /lockdown</c>). Pass <paramref name="domain"/> to
    /// read a specific lockdown domain (e.g. <c>com.apple.mobile.battery</c>);
    /// omit it for the default domain.
    /// </summary>
    public async Task<IReadOnlyDictionary<string, object?>> LockdownAsync(
        string? domain = null, CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.DevicesGetLockdownValuesAsync(_udid, domain, cancellationToken).ConfigureAwait(false));

    // --- Diagnostics / network --------------------------------------------

    /// <summary>Get the device's data-partition disk usage (<c>GET /diskspace</c>).</summary>
    public Task<GenModel.DiskSpaceInfo> DiskSpaceAsync(CancellationToken cancellationToken = default)
        => _api.DiagnosticsNetGetDiskSpaceAsync(_udid, cancellationToken);

    /// <summary>Get the device's MAC / IPv4 / IPv6 addresses (<c>GET /ip</c>).</summary>
    public Task<GenModel.NetworkInfo> IpAsync(CancellationToken cancellationToken = default)
        => _api.DiagnosticsNetGetDeviceIpAsync(_udid, cancellationToken);

    /// <summary>Get the RemoteServiceDiscovery (RSD) service map for the device (<c>GET /rsd</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> RsdAsync(CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.DiagnosticsNetGetRsdServicesAsync(_udid, cancellationToken).ConfigureAwait(false));

    /// <summary>Get the detailed IOKit battery registry snapshot (<c>GET /battery/registry</c>).</summary>
    public Task<GenModel.BatteryRegistry> BatteryRegistryAsync(CancellationToken cancellationToken = default)
        => _api.DiagnosticsNetGetBatteryRegistryAsync(_udid, cancellationToken);

    // --- Accessibility -----------------------------------------------------

    /// <summary>Get the VoiceOver enabled state (<c>GET /voiceover</c>).</summary>
    public Task<GenModel.VoiceOverState> VoiceOverAsync(CancellationToken cancellationToken = default)
        => _api.AccessibilityGetVoiceOverAsync(_udid, cancellationToken);

    /// <summary>Enable or disable VoiceOver (<c>PUT /voiceover</c>).</summary>
    public Task<GenModel.VoiceOverState> SetVoiceOverAsync(bool enabled, CancellationToken cancellationToken = default)
        => _api.AccessibilitySetVoiceOverAsync(_udid, null, new GenModel.AXEnabledRequest(enabled), cancellationToken);

    /// <summary>Get the Zoom (accessibility) enabled state (<c>GET /zoom</c>).</summary>
    public Task<GenModel.ZoomTouchState> ZoomAsync(CancellationToken cancellationToken = default)
        => _api.AccessibilityGetZoomTouchAsync(_udid, cancellationToken);

    /// <summary>Enable or disable Zoom (<c>PUT /zoom</c>).</summary>
    public Task<GenModel.ZoomTouchState> SetZoomAsync(bool enabled, CancellationToken cancellationToken = default)
        => _api.AccessibilitySetZoomTouchAsync(_udid, null, new GenModel.AXEnabledRequest(enabled), cancellationToken);

    /// <summary>Run the accessibility audit against the focused app (<c>POST /ax/audit</c>). Bounded by <paramref name="timeout"/> seconds.</summary>
    public async Task<IReadOnlyList<IReadOnlyDictionary<string, object?>>> AxAuditAsync(
        int? timeout = null, CancellationToken cancellationToken = default)
    {
        var raw = await _api.AccessibilityRunAxAuditAsync(_udid, timeout, cancellationToken).ConfigureAwait(false);
        return (raw ?? new List<object>()).Select(JsonHelpers.ToDictionary).ToList();
    }

    /// <summary>Get the current accessibility (AX) element snapshot/hierarchy (<c>GET /ax</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> AxAsync(CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.AccessibilityGetAxSnapshotAsync(_udid, cancellationToken).ConfigureAwait(false));

    /// <summary>Simulate live location tracking from an uploaded GPX file's bytes (<c>PUT /setlocation/gpx</c>, multipart).</summary>
    public Task<GenModel.GenericResponse> SetLocationGpxAsync(byte[] gpx, CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent { { Octet(gpx), "gpx", "route.gpx" } };
        var req = _raw.NewRequest(HttpMethod.Put, $"api/v1/device/{Esc(_udid)}/setlocation/gpx");
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    /// <summary>Read the device's cloud-configuration (supervision) payload (<c>GET /cloudconfig</c>).</summary>
    public async Task<IReadOnlyDictionary<string, object?>> CloudConfigAsync(CancellationToken cancellationToken = default)
        => JsonHelpers.ToDictionary(await _api.FsyncGetCloudConfigAsync(_udid, cancellationToken).ConfigureAwait(false));

    // --- Preparation -------------------------------------------------------

    /// <summary>
    /// Run the device preparation/provisioning flow (<c>POST /prepare</c>, multipart).
    /// To supervise the device supply a <paramref name="cert"/> (DER/PEM/P12 identity)
    /// and optional <paramref name="p12Password"/>; <paramref name="skip"/> lists setup
    /// panes to skip (see <see cref="PrepareClient.SkipOptionsAsync"/>).
    /// </summary>
    public Task<GenModel.PrepareResult> PrepareAsync(
        byte[]? cert = null, string? p12Password = null, IEnumerable<string>? skip = null,
        string? orgName = null, string? locale = null, string? lang = null,
        CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent();
        if (cert is not null) form.Add(Octet(cert), "cert", "supervision.p12");
        if (!string.IsNullOrEmpty(p12Password)) form.Add(new StringContent(p12Password), "p12password");
        if (skip is not null) foreach (var s in skip) form.Add(new StringContent(s), "skip");
        if (!string.IsNullOrEmpty(orgName)) form.Add(new StringContent(orgName), "orgname");
        if (!string.IsNullOrEmpty(locale)) form.Add(new StringContent(locale), "locale");
        if (!string.IsNullOrEmpty(lang)) form.Add(new StringContent(lang), "lang");
        var req = _raw.NewRequest(HttpMethod.Post, $"api/v1/device/{Esc(_udid)}/prepare");
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.PrepareResult>(req, cancellationToken);
    }

    // --- Binary streams (raw bytes, NOT SSE) ------------------------------

    /// <summary>
    /// Open an MJPEG (multipart/x-mixed-replace) stream of device screenshots
    /// (<c>GET /screenshot/stream</c>). Returns a raw <see cref="BinaryStream"/>;
    /// dispose it to stop. <paramref name="quality"/> is the JPEG quality (1–100).
    /// </summary>
    public Task<BinaryStream> ScreenshotStreamAsync(int? quality = null, CancellationToken cancellationToken = default)
    {
        var qs = quality.HasValue ? $"?quality={quality.Value}" : "";
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/screenshot/stream{qs}");
        return _raw.OpenBinaryStreamAsync(req, "image/jpeg", cancellationToken);
    }

    /// <summary>
    /// Open a live pcap capture as a libpcap byte stream (<c>GET /pcap</c>). Returns
    /// a raw <see cref="BinaryStream"/>; dispose it to stop. Runs until
    /// <paramref name="timeout"/> seconds elapse or the client disconnects.
    /// </summary>
    public Task<BinaryStream> PcapAsync(int? timeout = null, CancellationToken cancellationToken = default)
    {
        var qs = timeout.HasValue ? $"?timeout={timeout.Value}" : "";
        var req = _raw.NewRequest(HttpMethod.Get, $"api/v1/device/{Esc(_udid)}/pcap{qs}");
        return _raw.OpenBinaryStreamAsync(req, "application/vnd.tcpdump.pcap", cancellationToken);
    }

    // --- Management --------------------------------------------------------

    /// <summary>Reboot the device (<c>POST /reboot</c>).</summary>
    public Task<GenModel.GenericResponse> RebootAsync(CancellationToken cancellationToken = default)
        => _api.DevicesRebootAsync(_udid, cancellationToken);

    /// <summary>Shut the device down (<c>POST /shutdown</c>).</summary>
    public Task<GenModel.GenericResponse> ShutdownAsync(CancellationToken cancellationToken = default)
        => _api.DevicesShutdownAsync(_udid, cancellationToken);

    /// <summary>Erase all content and settings (<c>POST /erase</c>). Destructive — <paramref name="confirm"/> must be true.</summary>
    public Task<GenModel.GenericResponse> EraseAsync(bool confirm, CancellationToken cancellationToken = default)
        => _api.DevicesEraseAsync(_udid, confirm, cancellationToken);

    /// <summary>Get developer-mode state (<c>GET /devmode</c>).</summary>
    public Task<GenModel.DevModeState> DevmodeAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetDevModeAsync(_udid, cancellationToken);

    /// <summary>
    /// Set developer mode (<c>POST /devmode</c>). <paramref name="action"/> is
    /// <c>enable</c> or <c>reveal</c>; <paramref name="enablePostRestart"/> arms it across the next reboot.
    /// </summary>
    public Task<GenModel.GenericResponse> SetDevmodeAsync(
        string action, bool enablePostRestart = false, CancellationToken cancellationToken = default)
        => _api.DevicesSetDevModeAsync(_udid, new GenModel.DevModeRequest(action, enablePostRestart), cancellationToken);

    /// <summary>Get the language/locale configuration (<c>GET /lang</c>).</summary>
    public Task<GenModel.LanguageConfiguration> LangAsync(CancellationToken cancellationToken = default)
        => _api.DevicesGetLanguageAsync(_udid, cancellationToken);

    /// <summary>Set the language and/or locale (<c>PUT /lang</c>).</summary>
    public Task<GenModel.LanguageConfiguration> SetLangAsync(
        string? language = null, string? locale = null, CancellationToken cancellationToken = default)
        => _api.DevicesSetLanguageAsync(
            _udid,
            new GenModel.SetLanguageRequest { Language = language!, Locale = locale! },
            cancellationToken);

    /// <summary>Waive the memory limit for a process (<c>POST /memlimitoff</c>).</summary>
    public Task<GenModel.MemLimitResult> MemLimitOffAsync(string process, CancellationToken cancellationToken = default)
        => _api.DevicesMemLimitOffAsync(_udid, process, new GenModel.MemLimitRequest(process), cancellationToken);

    // --- Images / profiles -------------------------------------------------

    /// <summary>List the signatures of mounted developer disk images (<c>GET /image/list</c>).</summary>
    public Task<GenModel.MountedImages> MountedImagesAsync(CancellationToken cancellationToken = default)
        => _api.DevicesListMountedImagesAsync(_udid, cancellationToken);

    /// <summary>Unmount the developer disk image (<c>DELETE /image</c>).</summary>
    public Task<GenModel.GenericResponse> UnmountImageAsync(CancellationToken cancellationToken = default)
        => _api.DevicesUnmountImageAsync(_udid, cancellationToken);

    /// <summary>
    /// Install a configuration profile (<c>POST /profiles</c>). Supply the
    /// <c>.mobileconfig</c> bytes; for a supervised install add a <paramref name="p12"/>
    /// identity (and its <paramref name="password"/>).
    /// </summary>
    public Task<GenModel.GenericResponse> AddProfileAsync(
        byte[] profile, byte[]? p12 = null, string? password = null,
        CancellationToken cancellationToken = default)
    {
        var form = new MultipartFormDataContent { { Octet(profile), "profile", "profile.mobileconfig" } };
        if (p12 is not null) form.Add(Octet(p12), "p12", "supervision.p12");
        if (!string.IsNullOrEmpty(password)) form.Add(new StringContent(password), "password");
        var req = _raw.NewRequest(HttpMethod.Post, $"api/v1/device/{Esc(_udid)}/profiles");
        req.Content = form;
        return _raw.SendJsonAsync<GenModel.GenericResponse>(req, cancellationToken);
    }

    /// <summary>Remove an installed configuration profile by name/identifier (<c>DELETE /profiles/{name}</c>).</summary>
    public Task<GenModel.GenericResponse> RemoveProfileAsync(string name, CancellationToken cancellationToken = default)
        => _api.DevicesRemoveProfileAsync(_udid, name, cancellationToken);

    // --- Streaming (Server-Sent Events) -----------------------------------

    /// <summary>Stream syslog lines as they arrive. Heartbeats are surfaced as <see cref="HeartbeatEvent"/>.</summary>
    public IAsyncEnumerable<SseEvent> SyslogAsync(CancellationToken cancellationToken = default)
        => StreamAsync($"api/v1/device/{Esc(_udid)}/syslog", SyslogFactory, cancellationToken);

    /// <summary>Stream app foreground/background/lifecycle notifications.</summary>
    public IAsyncEnumerable<SseEvent> NotificationsAsync(CancellationToken cancellationToken = default)
        => StreamAsync($"api/v1/device/{Esc(_udid)}/notifications", NotificationsFactory, cancellationToken);

    /// <summary>
    /// Stream structured os_log trace entries, optionally filtered (AND-combined).
    /// </summary>
    public IAsyncEnumerable<SseEvent> OsTraceAsync(
        OsTraceFilters? filters = null, CancellationToken cancellationToken = default)
    {
        var path = $"api/v1/device/{Esc(_udid)}/ostrace" + (filters?.ToQueryString() ?? "");
        return StreamAsync(path, OsTraceFactory, cancellationToken);
    }

    /// <summary>Stream device attach/detach/pair events from the host.</summary>
    public IAsyncEnumerable<SseEvent> ListenAsync(CancellationToken cancellationToken = default)
        => StreamAsync($"api/v1/device/{Esc(_udid)}/listen", ListenFactory, cancellationToken);

    /// <summary>
    /// Stream sysmontap CPU-usage samples (<c>GET /sysmontap</c>). Samples are
    /// surfaced as <see cref="CpuUsageSampleEvent"/>; keep-alives as <see cref="HeartbeatEvent"/>.
    /// </summary>
    public IAsyncEnumerable<SseEvent> SysmontapAsync(CancellationToken cancellationToken = default)
        => StreamAsync($"api/v1/device/{Esc(_udid)}/sysmontap", SysmontapFactory, cancellationToken);

    // --- SSE dispatch factories -------------------------------------------

    private static SseEvent? SyslogFactory(string name, string data) => name switch
    {
        "syslog" => SseReader.Deserialize<SyslogMessageEvent>(data),
        _ => null,
    };

    private static SseEvent? NotificationsFactory(string name, string data) => name switch
    {
        "appstate" => SseReader.Deserialize<AppStateNotificationEvent>(data),
        _ => null,
    };

    private static SseEvent? OsTraceFactory(string name, string data) => name switch
    {
        "ostrace" => SseReader.Deserialize<OsTraceEntryEvent>(data),
        _ => null,
    };

    private static SseEvent? ListenFactory(string name, string data) => name switch
    {
        "attachdetach" => SseReader.Deserialize<AttachDetachEventEvent>(data),
        _ => null,
    };

    private static SseEvent? SysmontapFactory(string name, string data) => name switch
    {
        "sample" => SseReader.Deserialize<CpuUsageSampleEvent>(data),
        _ => null,
    };

    private async IAsyncEnumerable<SseEvent> StreamAsync(
        string path, SseEventFactory factory,
        [EnumeratorCancellation] CancellationToken cancellationToken)
    {
        using var req = _raw.NewRequest(HttpMethod.Get, path);
        req.Headers.Accept.ParseAdd("text/event-stream");
        using var resp = await _raw.Http
            .SendAsync(req, HttpCompletionOption.ResponseHeadersRead, cancellationToken)
            .ConfigureAwait(false);
        await RawHttp.EnsureSuccessAsync(resp, cancellationToken).ConfigureAwait(false);

#if NET5_0_OR_GREATER
        await using var stream = await resp.Content.ReadAsStreamAsync(cancellationToken).ConfigureAwait(false);
#else
        using var stream = await resp.Content.ReadAsStreamAsync().ConfigureAwait(false);
#endif
        await foreach (var e in SseReader.ReadAsync(stream, factory, cancellationToken).ConfigureAwait(false))
            yield return e;
    }

    // --- helpers -----------------------------------------------------------

    private static IReadOnlyDictionary<string, object?> ToDictionary(object? raw)
        => JsonHelpers.ToDictionary(raw);

    private static ByteArrayContent Octet(byte[] bytes)
    {
        var c = new ByteArrayContent(bytes);
        c.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
        return c;
    }

    private static string Esc(string s) => Uri.EscapeDataString(s);
}

/// <summary>Optional AND-combined filters for <see cref="DeviceClient.OsTraceAsync"/>.</summary>
public sealed class OsTraceFilters
{
    /// <summary>Only include entries from this process id.</summary>
    public int? Pid { get; set; }
    /// <summary>Minimum log level (e.g. <c>info</c>, <c>debug</c>, <c>error</c>).</summary>
    public string? Level { get; set; }
    /// <summary>Only include entries from this subsystem.</summary>
    public string? Subsystem { get; set; }
    /// <summary>Only include entries whose message matches this substring/pattern.</summary>
    public string? Match { get; set; }
    /// <summary>Exclude entries whose message matches this substring/pattern.</summary>
    public string? Exclude { get; set; }

    internal string ToQueryString()
    {
        var parts = new List<string>();
        if (Pid.HasValue) parts.Add("pid=" + Pid.Value);
        if (!string.IsNullOrEmpty(Level)) parts.Add("level=" + Uri.EscapeDataString(Level));
        if (!string.IsNullOrEmpty(Subsystem)) parts.Add("subsystem=" + Uri.EscapeDataString(Subsystem));
        if (!string.IsNullOrEmpty(Match)) parts.Add("match=" + Uri.EscapeDataString(Match));
        if (!string.IsNullOrEmpty(Exclude)) parts.Add("exclude=" + Uri.EscapeDataString(Exclude));
        return parts.Count > 0 ? "?" + string.Join("&", parts) : "";
    }
}

/// <summary>Thrown when a raw (streaming / binary / multipart) request returns a non-success status.</summary>
public sealed class IosApiException : Exception
{
    /// <summary>HTTP status code.</summary>
    public int StatusCode { get; }
    /// <summary>Response body, if any.</summary>
    public string? ResponseBody { get; }

    internal IosApiException(int statusCode, string? reason, string? body)
        : base($"go-ios API request failed: {statusCode} {reason}")
    {
        StatusCode = statusCode;
        ResponseBody = body;
    }
}
