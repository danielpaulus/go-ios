using System.Text.Json;
using System.Text.Json.Serialization;

namespace GoIos;

/// <summary>
/// Locates a locally running go-ios REST daemon by reading the discovery file
/// (<c>&lt;home&gt;/rest-api.json</c>) that the daemon writes after it binds a
/// (by default ephemeral, loopback-only) port.
/// </summary>
/// <remarks>
/// Home directory resolution matches the cross-language discovery contract:
/// the <c>GO_IOS_HOME</c> environment variable when set and non-empty, otherwise
/// <c>~/.go-ios</c> (the user profile directory).
/// </remarks>
public static class Discovery
{
    /// <summary>Name of the discovery file written by the daemon.</summary>
    public const string DiscoveryFileName = "rest-api.json";

    /// <summary>
    /// Resolve the go-ios home directory: <c>GO_IOS_HOME</c> env if set and
    /// non-empty, otherwise <c>~/.go-ios</c>.
    /// </summary>
    public static string HomeDirectory()
    {
        var home = Environment.GetEnvironmentVariable("GO_IOS_HOME");
        if (!string.IsNullOrWhiteSpace(home))
            return home;

        var profile = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
        if (string.IsNullOrEmpty(profile))
            profile = Environment.GetEnvironmentVariable("HOME") ?? ".";
        return Path.Combine(profile, ".go-ios");
    }

    /// <summary>Full path to the discovery file (<c>&lt;home&gt;/rest-api.json</c>).</summary>
    public static string DiscoveryFilePath() => Path.Combine(HomeDirectory(), DiscoveryFileName);

    /// <summary>
    /// Discover the base URL of the local go-ios REST daemon by reading the
    /// discovery file. Throws <see cref="DaemonNotFoundException"/> with a clear,
    /// actionable message when the file is missing, unreadable, or malformed.
    /// </summary>
    public static string DiscoverBaseUrl()
    {
        var path = DiscoveryFilePath();

        string json;
        try
        {
            json = File.ReadAllText(path);
        }
        catch (Exception ex) when (ex is FileNotFoundException or DirectoryNotFoundException or IOException or UnauthorizedAccessException)
        {
            throw NotFound(path, ex);
        }

        DiscoveryFile? info;
        try
        {
            info = JsonSerializer.Deserialize(json, DiscoveryJsonContext.Default.DiscoveryFile);
        }
        catch (JsonException ex)
        {
            throw NotFound(path, ex);
        }

        if (info is null || string.IsNullOrWhiteSpace(info.BaseUrl))
            throw NotFound(path, null);

        return info.BaseUrl;
    }

    private static DaemonNotFoundException NotFound(string path, Exception? inner) =>
        new(
            $"no local go-ios REST daemon found at {path}; start it (run the go-ios REST API) or pass an explicit BaseUrl",
            inner);

    /// <summary>Shape of <c>rest-api.json</c>. Only <c>baseUrl</c> is authoritative.</summary>
    internal sealed class DiscoveryFile
    {
        [JsonPropertyName("baseUrl")]
        public string? BaseUrl { get; set; }

        [JsonPropertyName("host")]
        public string? Host { get; set; }

        [JsonPropertyName("port")]
        public int Port { get; set; }

        [JsonPropertyName("pid")]
        public int Pid { get; set; }

        [JsonPropertyName("startedAt")]
        public string? StartedAt { get; set; }

        [JsonPropertyName("tls")]
        public bool Tls { get; set; }
    }
}

[JsonSourceGenerationOptions(PropertyNameCaseInsensitive = true)]
[JsonSerializable(typeof(Discovery.DiscoveryFile))]
internal sealed partial class DiscoveryJsonContext : JsonSerializerContext
{
}

/// <summary>
/// Thrown when no local go-ios REST daemon can be discovered and no explicit
/// <see cref="IosClientOptions.BaseUrl"/> or <c>GO_IOS_BASE_URL</c> was provided.
/// </summary>
public sealed class DaemonNotFoundException : Exception
{
    /// <summary>Create a new <see cref="DaemonNotFoundException"/>.</summary>
    public DaemonNotFoundException(string message, Exception? innerException = null)
        : base(message, innerException)
    {
    }
}
