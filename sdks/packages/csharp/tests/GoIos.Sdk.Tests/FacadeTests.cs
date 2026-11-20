using System.Net;
using System.Net.Http;
using GoIos;
using Xunit;

namespace GoIos.Sdk.Tests;

public class FacadeTests
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

    [Fact]
    public async Task Screenshot_Returns_Raw_Bytes_And_Sends_Bearer()
    {
        var png = new byte[] { 0x89, 0x50, 0x4E, 0x47, 1, 2, 3 };
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new ByteArrayContent(png)
                {
                    Headers = { ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("image/png") },
                },
            });
        using var client = ClientWith(handler);

        var bytes = await client.Device("00008110-ABC").ScreenshotAsync();

        Assert.Equal(png, bytes);
        var req = Assert.Single(handler.Requests);
        Assert.Equal("Bearer", req.Headers.Authorization?.Scheme);
        Assert.Equal("secret", req.Headers.Authorization?.Parameter);
        Assert.EndsWith("/screenshot", req.RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task SetLocation_Sends_Longitude_Query_Param()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"ok\"}");
        using var client = ClientWith(handler);

        await client.Device("udid1").SetLocationAsync(52.5, 13.4);

        var req = Assert.Single(handler.Requests);
        var q = req.RequestUri!.Query;
        Assert.Contains("longitude=13.4", q);
        Assert.Contains("latitude=52.5", q);
        Assert.DoesNotContain("longtitude", q);
    }

    [Fact]
    public async Task Devices_List_Deserializes()
    {
        var json = "{\"deviceList\":[{\"deviceID\":5,\"properties\":{\"serialNumber\":\"00008110-XYZ\"}}]}";
        var handler = StubHttpMessageHandler.Json(json);
        using var client = ClientWith(handler);

        var list = await client.Devices.ListAsync();

        Assert.Single(list.VarDeviceList);
        Assert.Equal("00008110-XYZ", list.VarDeviceList[0].Properties.SerialNumber);
    }

    [Fact]
    public async Task Syslog_Streams_Typed_Events_EndToEnd()
    {
        var wire =
            "event: syslog\ndata: {\"message\":\"hello\"}\n\n" +
            "event: heartbeat\ndata: {}\n\n" +
            "event: syslog\ndata: {\"message\":\"world\"}\n\n";
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK) { Content = new ChunkedContent(new[] { wire }) });
        using var client = ClientWith(handler);

        var messages = new List<string>();
        var heartbeats = 0;
        await foreach (var e in client.Device("udid1").SyslogAsync())
        {
            switch (e)
            {
                case SyslogMessageEvent s: messages.Add(s.Message); break;
                case HeartbeatEvent: heartbeats++; break;
            }
        }

        Assert.Equal(new[] { "hello", "world" }, messages);
        Assert.Equal(1, heartbeats);
        var req = Assert.Single(handler.Requests);
        Assert.Contains("text/event-stream", req.Headers.Accept.ToString());
    }

    [Fact]
    public async Task OsTrace_Applies_Filters_To_Query()
    {
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK) { Content = new ChunkedContent(new[] { "" }) });
        using var client = ClientWith(handler);

        await foreach (var _ in client.Device("udid1")
            .OsTraceAsync(new OsTraceFilters { Pid = 123, Level = "error", Subsystem = "com.apple.network" }))
        {
            // drain
        }

        var req = Assert.Single(handler.Requests);
        var q = req.RequestUri!.Query;
        Assert.Contains("pid=123", q);
        Assert.Contains("level=error", q);
        Assert.Contains("subsystem=com.apple.network", q);
    }

    [Fact]
    public async Task Non_Success_Stream_Throws_IosApiException()
    {
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.NotFound)
            {
                Content = new StringContent("{\"error\":\"device not found\"}"),
            });
        using var client = ClientWith(handler);

        var ex = await Assert.ThrowsAsync<IosApiException>(async () =>
        {
            await foreach (var _ in client.Device("nope").SyslogAsync()) { }
        });
        Assert.Equal(404, ex.StatusCode);
    }
}
