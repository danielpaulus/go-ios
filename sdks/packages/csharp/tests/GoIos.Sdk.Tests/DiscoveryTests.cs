using System.Net;
using System.Net.Http;
using GoIos;
using Xunit;

namespace GoIos.Sdk.Tests;

/// <summary>
/// Tests for ephemeral-daemon discovery. These mutate process-wide environment
/// variables (<c>GO_IOS_HOME</c> / <c>GO_IOS_BASE_URL</c>), so they run in a
/// dedicated non-parallel collection and restore the environment afterwards.
/// </summary>
[Collection("discovery-env")]
public sealed class DiscoveryTests : IDisposable
{
    private readonly string? _origHome = Environment.GetEnvironmentVariable("GO_IOS_HOME");
    private readonly string? _origBaseUrl = Environment.GetEnvironmentVariable("GO_IOS_BASE_URL");
    private readonly string _tempHome;

    public DiscoveryTests()
    {
        _tempHome = Path.Combine(Path.GetTempPath(), "goios-discovery-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(_tempHome);
        Environment.SetEnvironmentVariable("GO_IOS_HOME", _tempHome);
        Environment.SetEnvironmentVariable("GO_IOS_BASE_URL", null);
    }

    public void Dispose()
    {
        Environment.SetEnvironmentVariable("GO_IOS_HOME", _origHome);
        Environment.SetEnvironmentVariable("GO_IOS_BASE_URL", _origBaseUrl);
        try { Directory.Delete(_tempHome, recursive: true); } catch { /* best effort */ }
    }

    private void WriteDiscoveryFile(string baseUrl)
    {
        var json = $"{{\"baseUrl\":\"{baseUrl}\",\"host\":\"127.0.0.1\",\"port\":54321,\"pid\":12345,\"startedAt\":\"2026-08-11T15:00:00Z\",\"tls\":false}}";
        File.WriteAllText(Path.Combine(_tempHome, Discovery.DiscoveryFileName), json);
    }

    private static (IosClient client, StubHttpMessageHandler handler) ClientWith(IosClientOptions options)
    {
        var handler = StubHttpMessageHandler.Json("{\"deviceList\":[]}");
        options.HttpClient = new HttpClient(handler) { Timeout = Timeout.InfiniteTimeSpan };
        return (new IosClient(options), handler);
    }

    [Fact]
    public void Discovery_HomeDirectory_Uses_GoIosHome_Env()
    {
        Assert.Equal(_tempHome, Discovery.HomeDirectory());
        Assert.Equal(Path.Combine(_tempHome, "rest-api.json"), Discovery.DiscoveryFilePath());
    }

    [Fact]
    public async Task NoBaseUrl_Uses_Discovered_BaseUrl()
    {
        WriteDiscoveryFile("http://127.0.0.1:54321");
        var (client, handler) = ClientWith(new IosClientOptions());
        using (client)
        {
            await client.Devices.ListAsync();
        }

        var req = Assert.Single(handler.Requests);
        Assert.Equal("127.0.0.1", req.RequestUri!.Host);
        Assert.Equal(54321, req.RequestUri!.Port);
    }

    [Fact]
    public void Parameterless_Constructor_Uses_Discovery()
    {
        WriteDiscoveryFile("http://127.0.0.1:54321");
        // The parameterless ctor owns its HttpClient; constructing without an
        // exception proves the discovered address resolved.
        using var client = new IosClient();
        Assert.NotNull(client);
    }

    [Fact]
    public async Task Explicit_BaseUrl_Overrides_Discovery_And_Env()
    {
        WriteDiscoveryFile("http://127.0.0.1:54321");
        Environment.SetEnvironmentVariable("GO_IOS_BASE_URL", "http://127.0.0.1:11111");

        var (client, handler) = ClientWith(new IosClientOptions { BaseUrl = "http://127.0.0.1:22222" });
        using (client)
        {
            await client.Devices.ListAsync();
        }

        var req = Assert.Single(handler.Requests);
        Assert.Equal(22222, req.RequestUri!.Port);
    }

    [Fact]
    public async Task GoIosBaseUrl_Env_Overrides_Discovery()
    {
        WriteDiscoveryFile("http://127.0.0.1:54321");
        Environment.SetEnvironmentVariable("GO_IOS_BASE_URL", "http://127.0.0.1:33333");

        var (client, handler) = ClientWith(new IosClientOptions());
        using (client)
        {
            await client.Devices.ListAsync();
        }

        var req = Assert.Single(handler.Requests);
        Assert.Equal(33333, req.RequestUri!.Port);
    }

    [Fact]
    public void Missing_Discovery_File_Throws_Clear_Exception()
    {
        // No file written, no env set.
        var ex = Assert.Throws<DaemonNotFoundException>(() => new IosClient(new IosClientOptions()));
        Assert.Contains("no local go-ios REST daemon found", ex.Message);
        Assert.Contains(Discovery.DiscoveryFilePath(), ex.Message);
        Assert.Contains("BaseUrl", ex.Message);
    }

    [Fact]
    public void Malformed_Discovery_File_Throws_Clear_Exception()
    {
        File.WriteAllText(Path.Combine(_tempHome, Discovery.DiscoveryFileName), "{ not json");
        var ex = Assert.Throws<DaemonNotFoundException>(() => Discovery.DiscoverBaseUrl());
        Assert.Contains("no local go-ios REST daemon found", ex.Message);
    }

    [Fact]
    public void Discovery_File_Without_BaseUrl_Throws()
    {
        File.WriteAllText(Path.Combine(_tempHome, Discovery.DiscoveryFileName), "{\"port\":54321}");
        Assert.Throws<DaemonNotFoundException>(() => Discovery.DiscoverBaseUrl());
    }
}

/// <summary>
/// Serializes discovery tests (which mutate process env) so they don't race the
/// rest of the suite.
/// </summary>
[CollectionDefinition("discovery-env", DisableParallelization = true)]
public sealed class DiscoveryEnvCollection
{
}
