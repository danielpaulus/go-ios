using GoIos;

namespace GoIos.Examples;

/// <summary>
/// Shared, environment-driven configuration for every example.
///
/// The examples are configured entirely through environment variables so they
/// can run unchanged in a shell, a container, or CI:
///
///   GO_IOS_BASE_URL   Base URL of the go-ios REST daemon. Optional — when unset
///                     the examples pass no BaseUrl, so the SDK auto-discovers
///                     the local daemon via ~/.go-ios/rest-api.json. Set it only
///                     to target a pinned or remote daemon.
///   GO_IOS_API_KEY    Bearer token the daemon was started with. REQUIRED
///                     unless the daemon runs with --disable-auth (in which
///                     case set it to any non-empty placeholder, or export it
///                     empty and start the daemon with --disable-auth).
///   GO_IOS_UDID       Target device udid. Optional: when unset the examples
///                     pick the first device the daemon reports.
///   RUN_UI            Set to "1" to also run the (mutating) UI-automation
///                     example, which needs a forwarded WebDriverAgent.
///
/// This type resolves those variables once, builds a single reusable
/// <see cref="IosClient"/>, and offers a small helper to resolve the target
/// device udid (explicit env var, else first attached device).
/// </summary>
public sealed class ExampleContext : IDisposable
{
    /// <summary>Base URL of the go-ios daemon (from GO_IOS_BASE_URL), or null to auto-discover the local daemon.</summary>
    public string? BaseUrl { get; }

    /// <summary>Bearer token (from GO_IOS_API_KEY). May be empty when the daemon runs with --disable-auth.</summary>
    public string ApiKey { get; }

    /// <summary>Explicit target udid (from GO_IOS_UDID), or null to auto-pick the first device.</summary>
    public string? PreferredUdid { get; }

    /// <summary>The single, reusable, thread-safe SDK client. Construct once, share everywhere.</summary>
    public IosClient Client { get; }

    private ExampleContext(string? baseUrl, string apiKey, string? preferredUdid)
    {
        BaseUrl = baseUrl;
        ApiKey = apiKey;
        PreferredUdid = preferredUdid;

        // The IosClient is thread-safe and intended to be created once and
        // reused. It owns its own long-lived HttpClient (streaming endpoints
        // are long-lived, so there is no request timeout).
        Client = new IosClient(new IosClientOptions
        {
            BaseUrl = baseUrl,
            // Sent as "Authorization: Bearer <ApiKey>" on every request. When
            // empty the SDK simply omits the header (works with --disable-auth).
            ApiKey = string.IsNullOrEmpty(apiKey) ? null : apiKey,
        });
    }

    /// <summary>
    /// Read the environment and build a context. Returns null (after printing a
    /// helpful message) when GO_IOS_API_KEY is missing AND the daemon is not
    /// obviously in --disable-auth mode — the caller should then exit non-zero.
    /// </summary>
    public static ExampleContext? FromEnvironment()
    {
        // Leave baseUrl null when GO_IOS_BASE_URL is unset so the SDK falls
        // through to local-daemon discovery; we no longer hardcode a default port.
        var baseUrl = Environment.GetEnvironmentVariable("GO_IOS_BASE_URL");
        if (string.IsNullOrWhiteSpace(baseUrl))
            baseUrl = null;

        var apiKey = Environment.GetEnvironmentVariable("GO_IOS_API_KEY") ?? "";
        var udid = Environment.GetEnvironmentVariable("GO_IOS_UDID");
        if (string.IsNullOrWhiteSpace(udid))
            udid = null;

        // A missing API key is the single most common misconfiguration, and the
        // daemon rejects every /api/v1 route without a bearer token unless it
        // was started with --disable-auth. We refuse to guess: if the key is
        // absent we explain exactly what to set and let the caller exit 2.
        if (string.IsNullOrEmpty(apiKey))
        {
            Console.Error.WriteLine(
                "GO_IOS_API_KEY is not set.\n" +
                "\n" +
                "The go-ios daemon protects every /api/v1 route with a bearer token.\n" +
                "Start it (it prints the key on startup) and export the key, e.g.:\n" +
                "\n" +
                "    ios api --udid <udid>        # note the API key it prints\n" +
                "    export GO_IOS_API_KEY=<the-key>\n" +
                "    export GO_IOS_BASE_URL=http://localhost:8080\n" +
                "\n" +
                "If you deliberately started the daemon with --disable-auth, set\n" +
                "GO_IOS_API_KEY to any non-empty placeholder to acknowledge that.\n");
            return null;
        }

        return new ExampleContext(baseUrl, apiKey, udid);
    }

    /// <summary>
    /// Resolve the udid to operate on: the explicit GO_IOS_UDID if set,
    /// otherwise the first device the daemon reports. Returns null (no device
    /// attached) so examples can print SKIP instead of failing.
    /// </summary>
    public async Task<string?> ResolveUdidAsync(CancellationToken ct = default)
    {
        if (PreferredUdid is not null)
            return PreferredUdid;

        var devices = await Client.Devices.ListAsync(ct).ConfigureAwait(false);
        var first = devices.VarDeviceList?.FirstOrDefault();
        return first?.Properties?.SerialNumber;
    }

    public void Dispose() => Client.Dispose();
}
