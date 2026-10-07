using System.Net.Http;
using System.Net.Http.Headers;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenClient = GoIos.Sdk.Generated.Client;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>
/// Entry point for the go-ios SDK. Construct once and reuse; it is thread-safe.
/// </summary>
/// <example>
/// <code>
/// var client = new IosClient(new IosClientOptions { BaseUrl = "http://localhost:60105", ApiKey = "secret" });
/// var devices = await client.Devices.ListAsync();
/// var info = await client.Device(udid).InfoAsync();
/// await foreach (var e in client.Device(udid).SyslogAsync(ct)) { /* ... */ }
/// </code>
/// </example>
public sealed class IosClient : IDisposable
{
    private readonly HttpClient _http;
    private readonly bool _ownsHttp;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    /// <summary>Device-collection operations (list all devices).</summary>
    public DevicesClient Devices { get; }

    /// <summary>Global tunnel-agent operations.</summary>
    public TunnelsClient Tunnels { get; }

    /// <summary>Host-scoped code-signing operations (App Store Connect).</summary>
    public SignClient Sign { get; }

    /// <summary>Host-scoped device-preparation helpers (supervision cert / skip options).</summary>
    public PrepareClient Prepare { get; }

    /// <summary>
    /// Create a client that auto-resolves the go-ios REST daemon address via
    /// <c>GO_IOS_BASE_URL</c> or local discovery (see <see cref="IosClient(IosClientOptions)"/>).
    /// </summary>
    public IosClient() : this(new IosClientOptions())
    {
    }

    /// <summary>
    /// Create a new client with the given options.
    /// </summary>
    /// <remarks>
    /// When <see cref="IosClientOptions.BaseUrl"/> is not set, the base URL is
    /// resolved in this order:
    /// <list type="number">
    /// <item>explicit <see cref="IosClientOptions.BaseUrl"/> (targets remote daemons; skips discovery);</item>
    /// <item>the <c>GO_IOS_BASE_URL</c> environment variable;</item>
    /// <item>discovery of a local daemon via <c>&lt;home&gt;/rest-api.json</c>.</item>
    /// </list>
    /// If none resolve, a <see cref="DaemonNotFoundException"/> is thrown.
    /// </remarks>
    public IosClient(IosClientOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);

        var baseUrl = new Uri(ResolveBaseUrl(options).TrimEnd('/') + "/", UriKind.Absolute);
        var apiKey = options.ApiKey;

        if (options.HttpClient is not null)
        {
            _http = options.HttpClient;
            _ownsHttp = false;
        }
        else
        {
            // No default timeout: streaming endpoints are long-lived.
            _http = new HttpClient { Timeout = Timeout.InfiniteTimeSpan };
            _ownsHttp = true;
        }

        var config = new GenClient.Configuration { BasePath = baseUrl.ToString().TrimEnd('/') };
        if (!string.IsNullOrEmpty(apiKey))
        {
            // Send the bearer token explicitly so it does not depend on the
            // generator's (nonstandard "Bearer" scheme) auth wiring.
            config.DefaultHeaders["Authorization"] = "Bearer " + apiKey;
        }
        // Share the (possibly caller-supplied) HttpClient with the generated
        // client so unary and streaming/binary calls go through one pipeline.
        _api = new Gen.DefaultApi(_http, config);
        _raw = new RawHttp(_http, baseUrl, apiKey);

        Devices = new DevicesClient(_api);
        Tunnels = new TunnelsClient(_api);
        Sign = new SignClient(_raw);
        Prepare = new PrepareClient(_api);
    }

    /// <summary>Scope subsequent operations to a single device by udid.</summary>
    public DeviceClient Device(string udid)
    {
        if (string.IsNullOrWhiteSpace(udid))
            throw new ArgumentException("udid must be set", nameof(udid));
        return new DeviceClient(udid, _api, _raw);
    }

    /// <summary>
    /// Resolve the base URL: explicit option &gt; <c>GO_IOS_BASE_URL</c> env &gt;
    /// local discovery. Throws <see cref="DaemonNotFoundException"/> when nothing
    /// resolves.
    /// </summary>
    private static string ResolveBaseUrl(IosClientOptions options)
    {
        if (!string.IsNullOrWhiteSpace(options.BaseUrl))
            return options.BaseUrl!;

        var env = Environment.GetEnvironmentVariable("GO_IOS_BASE_URL");
        if (!string.IsNullOrWhiteSpace(env))
            return env;

        return Discovery.DiscoverBaseUrl();
    }

    /// <inheritdoc/>
    public void Dispose()
    {
        if (_ownsHttp) _http.Dispose();
    }
}

/// <summary>Operations over the device collection.</summary>
public sealed class DevicesClient
{
    private readonly Gen.DefaultApi _api;
    internal DevicesClient(Gen.DefaultApi api) => _api = api;

    /// <summary>List all attached / reachable devices.</summary>
    public async Task<GenModel.DeviceList> ListAsync(CancellationToken cancellationToken = default)
        => await _api.ListDevicesAsync(cancellationToken).ConfigureAwait(false);
}
