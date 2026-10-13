using System.Net;
using System.Net.Http;
using GoIos;
using Xunit;

namespace GoIos.Sdk.Tests;

/// <summary>Tests for the endpoints added to reach the full 80-op daemon surface.</summary>
public class ExtendedFacadeTests
{
    private static IosClient ClientWith(HttpMessageHandler handler, string? apiKey = "secret")
    {
        var http = new HttpClient(handler) { Timeout = Timeout.InfiniteTimeSpan };
        return new IosClient(new IosClientOptions
        {
            BaseUrl = "http://localhost:60105",
            ApiKey = apiKey,
            HttpClient = http,
        });
    }

    // --- Device information -------------------------------------------------

    [Fact]
    public async Task Battery_Deserializes()
    {
        var handler = StubHttpMessageHandler.Json(
            "{\"CurrentCapacity\":83,\"ExternalConnected\":true,\"IsCharging\":true,\"Temperature\":2980}");
        using var client = ClientWith(handler);

        var battery = await client.Device("udid1").BatteryAsync();

        Assert.Equal(83, battery.CurrentCapacity);
        Assert.True(battery.IsCharging);
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/battery", req.RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task MobileGestalt_Sends_Comma_Joined_Keys()
    {
        var handler = StubHttpMessageHandler.Json("{\"ProductType\":\"iPhone14,2\",\"UniqueDeviceID\":\"abc\"}");
        using var client = ClientWith(handler);

        var map = await client.Device("udid1").MobileGestaltAsync(new[] { "ProductType", "UniqueDeviceID" });

        Assert.Equal("iPhone14,2", map["ProductType"]?.ToString());
        var req = Assert.Single(handler.Requests);
        // explode=false array -> single "key=" param with comma-joined values.
        Assert.Contains("key=ProductType%2CUniqueDeviceID", req.RequestUri!.Query);
    }

    // --- Management ---------------------------------------------------------

    [Fact]
    public async Task Erase_Requires_Confirm_Query()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"erasing\"}");
        using var client = ClientWith(handler);

        await client.Device("udid1").EraseAsync(confirm: true);

        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Post, req.Method);
        Assert.EndsWith("/erase", req.RequestUri!.AbsolutePath);
        Assert.Contains("confirm=true", req.RequestUri!.Query);
    }

    [Fact]
    public async Task Reboot_Posts_To_Reboot()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"ok\"}");
        using var client = ClientWith(handler);

        await client.Device("udid1").RebootAsync();

        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Post, req.Method);
        Assert.EndsWith("/reboot", req.RequestUri!.AbsolutePath);
    }

    // --- Settings -----------------------------------------------------------

    [Fact]
    public async Task Settings_SetWifi_Puts_Json_Body()
    {
        string? body = null;
        var handler = new StubHttpMessageHandler(req =>
        {
            body = req.Content?.ReadAsStringAsync().GetAwaiter().GetResult();
            return new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"message\":\"ok\"}", System.Text.Encoding.UTF8, "application/json"),
            };
        });
        using var client = ClientWith(handler);

        await client.Device("udid1").Settings.SetWifiAsync("MyNet", "hunter2", "WPA2");

        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Put, req.Method);
        Assert.EndsWith("/wifi", req.RequestUri!.AbsolutePath);
        Assert.Contains("MyNet", body);
        Assert.Contains("hunter2", body);
    }

    [Fact]
    public async Task Settings_RemoveWifi_Sends_Ssid_Query()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"ok\"}");
        using var client = ClientWith(handler);

        await client.Device("udid1").Settings.RemoveWifiAsync("MyNet");

        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Delete, req.Method);
        Assert.Contains("ssid=MyNet", req.RequestUri!.Query);
    }

    // --- Media (pasteboard text/plain + multipart wallpaper) ---------------

    [Fact]
    public async Task Media_SetPasteboard_Sends_Text_Plain()
    {
        string? contentType = null;
        string? body = null;
        var handler = new StubHttpMessageHandler(req =>
        {
            contentType = req.Content?.Headers.ContentType?.MediaType;
            body = req.Content?.ReadAsStringAsync().GetAwaiter().GetResult();
            return new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"message\":\"ok\"}", System.Text.Encoding.UTF8, "application/json"),
            };
        });
        using var client = ClientWith(handler);

        await client.Device("udid1").Media.SetPasteboardAsync("copied text");

        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Put, req.Method);
        Assert.EndsWith("/pasteboard", req.RequestUri!.AbsolutePath);
        Assert.Equal("text/plain", contentType);
        Assert.Equal("copied text", body);
    }

    // --- Files (raw binary pull) -------------------------------------------

    [Fact]
    public async Task Files_Pull_Returns_Raw_Bytes_With_Domain_Query()
    {
        var payload = new byte[] { 1, 2, 3, 4, 5 };
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new ByteArrayContent(payload)
                {
                    Headers = { ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("application/octet-stream") },
                },
            });
        using var client = ClientWith(handler);

        var bytes = await client.Device("udid1").Files.PullAsync(
            domain: "appDocuments", remote: "Documents/log.txt", identifier: "com.example.app");

        Assert.Equal(payload, bytes);
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/files/pull", req.RequestUri!.AbsolutePath);
        var q = req.RequestUri!.Query;
        Assert.Contains("domain=appDocuments", q);
        Assert.Contains("identifier=com.example.app", q);
        Assert.Contains("remote=Documents%2Flog.txt", q);
    }

    // --- Crashes ------------------------------------------------------------

    [Fact]
    public async Task Crashes_List_Deserializes_And_Sends_Pattern()
    {
        var handler = StubHttpMessageHandler.Json("{\"files\":[\"a.ips\",\"b.ips\"],\"count\":2}");
        using var client = ClientWith(handler);

        var listing = await client.Device("udid1").Crashes.ListAsync("*.ips");

        Assert.Equal(2, listing.Count);
        Assert.Equal(new[] { "a.ips", "b.ips" }, listing.Files);
        var req = Assert.Single(handler.Requests);
        Assert.Contains("pattern=", req.RequestUri!.Query);
    }

    [Fact]
    public async Task Crashes_Remove_Sends_Pattern_And_Defaults_Cwd_To_Dot()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"removed\"}");
        using var client = ClientWith(handler);

        // pattern is primary/required; cwd defaults to ".".
        await client.Device("udid1").Crashes.RemoveAsync("*.ips");

        var req = Assert.Single(handler.Requests);
        Assert.Contains("pattern=", req.RequestUri!.Query);
        Assert.Contains("cwd=.", req.RequestUri!.Query);
    }

    [Fact]
    public async Task Crashes_Remove_Sends_Explicit_Cwd()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"removed\"}");
        using var client = ClientWith(handler);

        await client.Device("udid1").Crashes.RemoveAsync("*.ips", cwd: "/tmp/crashes");

        var req = Assert.Single(handler.Requests);
        Assert.Contains("pattern=", req.RequestUri!.Query);
        Assert.Contains("cwd=", req.RequestUri!.Query);
    }

    // --- Jobs (unary) -------------------------------------------------------

    [Fact]
    public async Task Jobs_Forward_Posts_Job_And_Deserializes()
    {
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.Accepted)
            {
                Content = new StringContent(
                    "{\"id\":\"forward-1\",\"kind\":\"forward\",\"udid\":\"udid1\",\"status\":\"running\",\"startedAt\":\"2024-01-01T00:00:00Z\"}",
                    System.Text.Encoding.UTF8, "application/json"),
            });
        using var client = ClientWith(handler);

        var job = await client.Device("udid1").Jobs.ForwardAsync(hostPort: 8100, targetPort: 8100);

        Assert.Equal("forward-1", job.Id);
        Assert.Equal("forward", job.Kind);
        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Post, req.Method);
        Assert.EndsWith("/jobs/forward", req.RequestUri!.AbsolutePath);
    }

    // --- Tunnels ------------------------------------------------------------

    [Fact]
    public async Task Tunnels_List_Deserializes()
    {
        var handler = StubHttpMessageHandler.Json(
            "[{\"Udid\":\"udid1\",\"Address\":\"fd00::1\",\"RsdPort\":50000,\"UserspaceTUN\":false,\"UserspaceTUNPort\":0}]");
        using var client = ClientWith(handler);

        var tunnels = await client.Tunnels.ListAsync();

        var t = Assert.Single(tunnels);
        Assert.Equal("udid1", t.Udid);
        Assert.Equal(50000, t.RsdPort);
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/tunnels", req.RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task Tunnels_ShutdownAgent_Posts()
    {
        var handler = StubHttpMessageHandler.Json("{\"status\":\"shutting-down\"}");
        using var client = ClientWith(handler);

        var res = await client.Tunnels.ShutdownAgentAsync();

        Assert.Equal("shutting-down", res.Status);
        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Post, req.Method);
        Assert.EndsWith("/tunnel-agent/shutdown", req.RequestUri!.AbsolutePath);
    }

    // --- New SSE stream #1: sysmontap --------------------------------------

    [Fact]
    public async Task Sysmontap_Streams_Typed_Samples_And_Heartbeats()
    {
        var wire =
            "event: sample\ndata: {\"CPU_TotalLoad\":42.5,\"SystemLoad\":10.0,\"UserLoad\":32.5,\"nCPU\":6}\n\n" +
            "event: heartbeat\ndata: {}\n\n" +
            "event: sample\ndata: {\"CPU_TotalLoad\":11.0}\n\n";
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK) { Content = new ChunkedContent(new[] { wire }) });
        using var client = ClientWith(handler);

        var loads = new List<double?>();
        var heartbeats = 0;
        object? extraValue = null;
        await foreach (var e in client.Device("udid1").SysmontapAsync())
        {
            switch (e)
            {
                case CpuUsageSampleEvent s:
                    loads.Add(s.CpuTotalLoad);
                    extraValue ??= s.Extra != null && s.Extra.TryGetValue("nCPU", out var v) ? v : null;
                    break;
                case HeartbeatEvent: heartbeats++; break;
            }
        }

        Assert.Equal(new double?[] { 42.5, 11.0 }, loads);
        Assert.Equal(1, heartbeats);
        Assert.NotNull(extraValue); // open-map extension keys preserved
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/sysmontap", req.RequestUri!.AbsolutePath);
        Assert.Contains("text/event-stream", req.Headers.Accept.ToString());
    }

    // --- New SSE stream #2: job logs ---------------------------------------

    [Fact]
    public async Task Jobs_Logs_Streams_Typed_Lines_And_Heartbeats()
    {
        var wire =
            "event: log\ndata: {\"line\":\"Test suite started\"}\n\n" +
            "event: heartbeat\ndata: {}\n\n" +
            "event: log\ndata: {\"line\":\"Test suite passed\"}\n\n";
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK) { Content = new ChunkedContent(new[] { wire }) });
        using var client = ClientWith(handler);

        var lines = new List<string>();
        var heartbeats = 0;
        await foreach (var e in client.Device("udid1").Jobs.LogsAsync("runtest-3"))
        {
            switch (e)
            {
                case JobLogLineEvent l: lines.Add(l.Line); break;
                case HeartbeatEvent: heartbeats++; break;
            }
        }

        Assert.Equal(new[] { "Test suite started", "Test suite passed" }, lines);
        Assert.Equal(1, heartbeats);
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/jobs/runtest-3/logs", req.RequestUri!.AbsolutePath);
    }

    // --- udid convenience accessor -----------------------------------------

    [Fact]
    public void Device_Exposes_Udid_Accessor()
    {
        using var client = ClientWith(StubHttpMessageHandler.Json("{}"));
        Assert.Equal("00008110-ABC", client.Device("00008110-ABC").Udid);
    }
}
