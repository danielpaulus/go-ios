using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using GoIos;
using Xunit;

namespace GoIos.Sdk.Tests;

/// <summary>
/// Tests for the endpoints added to reach the full 125-op daemon surface:
/// diagnostics/network, accessibility, fsync, webinspector, ui, host signing,
/// and the raw binary streams.
/// </summary>
public class V3FacadeTests
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

    private static StubHttpMessageHandler Capturing(string json, out Func<(HttpMethod?, string?, string?)> last)
    {
        HttpRequestMessage? seen = null;
        string? body = null;
        var handler = new StubHttpMessageHandler(req =>
        {
            seen = req;
            body = req.Content?.ReadAsStringAsync().GetAwaiter().GetResult();
            return new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(json, Encoding.UTF8, "application/json"),
            };
        });
        last = () => (seen?.Method, seen?.RequestUri?.AbsolutePath, body);
        return handler;
    }

    // --- Diagnostics / network ---------------------------------------------

    [Fact]
    public async Task DiskSpace_Deserializes()
    {
        var handler = StubHttpMessageHandler.Json(
            "{\"FSTotalBytes\":128000000000,\"FSFreeBytes\":64000000000,\"FSBlockSize\":4096,\"Model\":\"APPLE SSD\"}");
        using var client = ClientWith(handler);

        var d = await client.Device("udid1").DiskSpaceAsync();

        Assert.Equal(128000000000L, d.FSTotalBytes);
        Assert.Equal(64000000000L, d.FSFreeBytes);
        Assert.EndsWith("/diskspace", Assert.Single(handler.Requests).RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task Ip_Deserializes()
    {
        var handler = StubHttpMessageHandler.Json(
            "{\"MacAddress\":\"aa:bb:cc:dd:ee:ff\",\"IPv4\":\"192.168.0.5\",\"IPv6\":\"fe80::1\"}");
        using var client = ClientWith(handler);

        var ip = await client.Device("udid1").IpAsync();

        Assert.Equal("192.168.0.5", ip.IPv4);
        Assert.EndsWith("/ip", Assert.Single(handler.Requests).RequestUri!.AbsolutePath);
    }

    // --- Accessibility ------------------------------------------------------

    [Fact]
    public async Task SetVoiceOver_Puts_Enabled_Body()
    {
        var handler = Capturing("{\"voiceOverEnabled\":true}", out var last);
        using var client = ClientWith(handler);

        var state = await client.Device("udid1").SetVoiceOverAsync(true);

        Assert.True(state.VoiceOverEnabled);
        var (method, path, body) = last();
        Assert.Equal(HttpMethod.Put, method);
        Assert.EndsWith("/voiceover", path);
        Assert.Contains("true", body);
    }

    [Fact]
    public async Task AxAudit_Sends_Timeout_And_Returns_List()
    {
        var handler = StubHttpMessageHandler.Json("[{\"type\":\"contrast\",\"element\":\"Button\"}]");
        using var client = ClientWith(handler);

        var issues = await client.Device("udid1").AxAuditAsync(timeout: 30);

        var issue = Assert.Single(issues);
        Assert.Equal("contrast", issue["type"]?.ToString());
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/ax/audit", req.RequestUri!.AbsolutePath);
        Assert.Contains("timeout=30", req.RequestUri!.Query);
    }

    // --- Fsync --------------------------------------------------------------

    [Fact]
    public async Task Fsync_Ls_Sends_BundleId_And_Path()
    {
        var handler = StubHttpMessageHandler.Json("{\"path\":\"/Documents\",\"files\":[\"a.txt\",\"b.txt\"],\"count\":2}");
        using var client = ClientWith(handler);

        var listing = await client.Device("udid1").Fsync.LsAsync(path: "/Documents", bundleId: "com.example.app");

        Assert.Equal(2, listing.Count);
        Assert.Equal(new[] { "a.txt", "b.txt" }, listing.Files);
        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/fsync/ls", req.RequestUri!.AbsolutePath);
        Assert.Contains("bundleID=com.example.app", req.RequestUri!.Query);
        Assert.Contains("path=%2FDocuments", req.RequestUri!.Query);
    }

    [Fact]
    public async Task Fsync_Pull_Returns_Raw_Bytes()
    {
        var payload = new byte[] { 9, 8, 7 };
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new ByteArrayContent(payload)
                {
                    Headers = { ContentType = new MediaTypeHeaderValue("application/octet-stream") },
                },
            });
        using var client = ClientWith(handler);

        var bytes = await client.Device("udid1").Fsync.PullAsync(path: "/var/log.txt");

        Assert.Equal(payload, bytes);
        Assert.EndsWith("/fsync/pull", Assert.Single(handler.Requests).RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task Fsync_Rm_Sends_Recursive_Query()
    {
        var handler = StubHttpMessageHandler.Json("{\"message\":\"removed\",\"path\":\"/tmp/x\"}");
        using var client = ClientWith(handler);

        var res = await client.Device("udid1").Fsync.RmAsync("/tmp/x", recursive: true);

        Assert.Equal("removed", res.Message);
        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Delete, req.Method);
        Assert.Contains("recursive=true", req.RequestUri!.Query);
    }

    // --- WebInspector -------------------------------------------------------

    [Fact]
    public async Task WebInspector_Eval_Posts_Script()
    {
        var handler = Capturing("{\"page\":\"1\",\"result\":42}", out var last);
        using var client = ClientWith(handler);

        var res = await client.Device("udid1").WebInspector.EvalAsync("1+1", page: "1");

        Assert.Equal("1", res.Page);
        var (method, path, body) = last();
        Assert.Equal(HttpMethod.Post, method);
        Assert.EndsWith("/webinspector/eval", path);
        Assert.Contains("1+1", body);
    }

    // --- UI -----------------------------------------------------------------

    [Fact]
    public async Task Ui_Tap_Posts_Coordinates_With_Backend_Option()
    {
        var handler = Capturing("{\"status\":\"ok\"}", out var last);
        using var client = ClientWith(handler);

        var res = await client.Device("udid1").Ui.TapAsync(
            10, 20, new UiClient.Options { Backend = "wda", Timeout = 15 });

        Assert.Equal("ok", res["status"]?.ToString());
        var (method, path, body) = last();
        Assert.Equal(HttpMethod.Post, method);
        Assert.EndsWith("/ui/tap", path);
        Assert.Contains("\"x\":10", body!.Replace(" ", ""));
        var req = Assert.Single(handler.Requests);
        Assert.Contains("backend=wda", req.RequestUri!.Query);
        Assert.Contains("timeout=15", req.RequestUri!.Query);
    }

    [Fact]
    public async Task Ui_Screenshot_Returns_Png_Bytes()
    {
        var png = new byte[] { 0x89, 0x50, 0x4E, 0x47 };
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new ByteArrayContent(png)
                {
                    Headers = { ContentType = new MediaTypeHeaderValue("image/png") },
                },
            });
        using var client = ClientWith(handler);

        var bytes = await client.Device("udid1").Ui.ScreenshotAsync();

        Assert.Equal(png, bytes);
        Assert.EndsWith("/ui/screenshot", Assert.Single(handler.Requests).RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task Ui_Source_Returns_Xml_Text()
    {
        var xml = "<AppiumAUT><Application/></AppiumAUT>";
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(xml, Encoding.UTF8, "application/xml"),
            });
        using var client = ClientWith(handler);

        var source = await client.Device("udid1").Ui.SourceAsync();

        Assert.Equal(xml, source);
        Assert.EndsWith("/ui/source", Assert.Single(handler.Requests).RequestUri!.AbsolutePath);
    }

    // --- Host: signing / prepare -------------------------------------------

    [Fact]
    public async Task Sign_Certificate_Posts_Multipart_And_Returns_P12_Bytes()
    {
        var p12 = new byte[] { 1, 2, 3, 4 };
        string? contentType = null;
        var handler = new StubHttpMessageHandler(req =>
        {
            contentType = req.Content?.Headers.ContentType?.MediaType;
            return new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new ByteArrayContent(p12)
                {
                    Headers = { ContentType = new MediaTypeHeaderValue("application/x-pkcs12") },
                },
            };
        });
        using var client = ClientWith(handler);

        var bytes = await client.Sign.CertificateAsync(
            ascPrivateKey: new byte[] { 5, 6 }, ascKeyId: "KEYID", ascIssuerId: "ISSUER");

        Assert.Equal(p12, bytes);
        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Post, req.Method);
        Assert.EndsWith("/sign/certificate", req.RequestUri!.AbsolutePath);
        Assert.StartsWith("multipart/form-data", contentType);
    }

    [Fact]
    public async Task Prepare_SkipOptions_Deserializes()
    {
        var handler = StubHttpMessageHandler.Json("{\"options\":[\"Passcode\",\"Siri\"],\"count\":2}");
        using var client = ClientWith(handler);

        var opts = await client.Prepare.SkipOptionsAsync();

        Assert.Equal(2, opts.Count);
        Assert.Contains("Siri", opts.Options);
        Assert.EndsWith("/prepare/skip-options", Assert.Single(handler.Requests).RequestUri!.AbsolutePath);
    }

    [Fact]
    public async Task Device_Prepare_Posts_Multipart_With_Skip_Fields()
    {
        string? contentType = null;
        string? body = null;
        var handler = new StubHttpMessageHandler(req =>
        {
            contentType = req.Content?.Headers.ContentType?.MediaType;
            body = req.Content?.ReadAsStringAsync().GetAwaiter().GetResult();
            return new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"status\":\"prepared\",\"supervised\":true}", Encoding.UTF8, "application/json"),
            };
        });
        using var client = ClientWith(handler);

        var res = await client.Device("udid1").PrepareAsync(
            cert: new byte[] { 1 }, p12Password: "pw", skip: new[] { "Siri" }, orgName: "Acme");

        Assert.Equal("prepared", res.Status);
        Assert.True(res.Supervised);
        var req = Assert.Single(handler.Requests);
        Assert.Equal(HttpMethod.Post, req.Method);
        Assert.EndsWith("/prepare", req.RequestUri!.AbsolutePath);
        Assert.StartsWith("multipart/form-data", contentType);
        Assert.Contains("Acme", body);
        Assert.Contains("Siri", body);
    }

    // --- Lockdown domain (regenerated signature) ---------------------------

    [Fact]
    public async Task Lockdown_Sends_Domain_Query_When_Provided()
    {
        var handler = StubHttpMessageHandler.Json("{\"BatteryCurrentCapacity\":80}");
        using var client = ClientWith(handler);

        var map = await client.Device("udid1").LockdownAsync(domain: "com.apple.mobile.battery");

        Assert.Equal("80", map["BatteryCurrentCapacity"]?.ToString());
        var req = Assert.Single(handler.Requests);
        Assert.Contains("domain=com.apple.mobile.battery", req.RequestUri!.Query);
    }

    // --- Binary stream: chunked read --------------------------------------

    [Fact]
    public async Task PcapStream_Reads_Chunks_As_A_Raw_Stream()
    {
        // A finite binary body delivered in two chunks: the facade must expose it
        // as a live byte Stream (via ResponseHeadersRead), readable incrementally.
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new BinaryChunkedContent(
                    new[] { new byte[] { 1, 2, 3 }, new byte[] { 4, 5, 6 } },
                    "application/vnd.tcpdump.pcap"),
            });
        using var client = ClientWith(handler);

        await using var stream = await client.Device("udid1").PcapAsync(timeout: 5);

        Assert.Equal("application/vnd.tcpdump.pcap", stream.ContentType);

        var all = new MemoryStream();
        await stream.CopyToAsync(all);
        Assert.Equal(new byte[] { 1, 2, 3, 4, 5, 6 }, all.ToArray());

        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/pcap", req.RequestUri!.AbsolutePath);
        Assert.Contains("timeout=5", req.RequestUri!.Query);
    }

    // --- Binary stream: cancellation --------------------------------------

    [Fact]
    public async Task BinaryStream_Open_Honors_Cancellation()
    {
        // The CancellationToken must flow through the whole binary-stream open path
        // (SendAsync / ResponseHeadersRead). The handler observes the token: if the
        // facade did not forward it, no exception would surface.
        var handler = new CancellationObservingHandler();
        using var client = ClientWith(handler);

        using var cts = new CancellationTokenSource();
        cts.Cancel();
        await Assert.ThrowsAnyAsync<OperationCanceledException>(
            async () => await client.Device("udid1").PcapAsync(cancellationToken: cts.Token));
    }

    [Fact]
    public async Task BinaryStream_Read_Passes_Token_To_Underlying_Stream()
    {
        // After the stream is open, a canceled read must not silently succeed:
        // cancellation is forwarded to the underlying response stream.
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new BinaryChunkedContent(new[] { new byte[] { 1, 2, 3 } }, "application/vnd.tcpdump.pcap"),
            });
        using var client = ClientWith(handler);

        await using var stream = await client.Device("udid1").PcapAsync();
        using var cts = new CancellationTokenSource();
        cts.Cancel();
        await Assert.ThrowsAnyAsync<OperationCanceledException>(
            async () => await stream.ReadAsync(new byte[16].AsMemory(), cts.Token));
    }

    [Fact]
    public async Task ScreenshotStream_Sends_Quality_And_Exposes_ContentType()
    {
        var handler = new StubHttpMessageHandler(_ =>
            new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new BinaryChunkedContent(
                    new[] { new byte[] { 0xFF, 0xD8, 0xFF } }, "image/jpeg"),
            });
        using var client = ClientWith(handler);

        await using var stream = await client.Device("udid1").ScreenshotStreamAsync(quality: 70);
        Assert.Equal("image/jpeg", stream.ContentType);
        var buf = new byte[3];
        int n = await stream.ReadAsync(buf.AsMemory());
        Assert.Equal(3, n);

        var req = Assert.Single(handler.Requests);
        Assert.EndsWith("/screenshot/stream", req.RequestUri!.AbsolutePath);
        Assert.Contains("quality=70", req.RequestUri!.Query);
    }
}

/// <summary>
/// Binary <see cref="HttpContent"/> that releases the given byte chunks then
/// completes, so a test can read a finite raw stream incrementally.
/// </summary>
internal sealed class BinaryChunkedContent : HttpContent
{
    private readonly IReadOnlyList<byte[]> _chunks;

    public BinaryChunkedContent(IEnumerable<byte[]> chunks, string mediaType)
    {
        _chunks = chunks.ToList();
        Headers.ContentType = new MediaTypeHeaderValue(mediaType);
    }

    protected override async Task SerializeToStreamAsync(Stream stream, System.Net.TransportContext? context)
    {
        foreach (var chunk in _chunks)
        {
            await stream.WriteAsync(chunk);
            await stream.FlushAsync();
        }
    }

    protected override bool TryComputeLength(out long length)
    {
        length = 0;
        return false;
    }
}

/// <summary>
/// A handler that honors the <see cref="CancellationToken"/> it is given, so we
/// can prove the facade forwards it into the binary-stream open path.
/// </summary>
internal sealed class CancellationObservingHandler : HttpMessageHandler
{
    protected override Task<HttpResponseMessage> SendAsync(
        HttpRequestMessage request, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new BinaryChunkedContent(new[] { new byte[] { 1, 2, 3 } }, "application/vnd.tcpdump.pcap"),
        });
    }
}
