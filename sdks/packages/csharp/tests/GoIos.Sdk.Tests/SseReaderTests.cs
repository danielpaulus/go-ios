using System.Text;
using Xunit;

namespace GoIos.Sdk.Tests;

public class SseReaderTests
{
    private static SseEvent? SyslogFactory(string name, string data) => name switch
    {
        "syslog" => SseReader.Deserialize<SyslogMessageEvent>(data),
        _ => null,
    };

    private static async Task<List<SseEvent>> ReadAll(string wire, CancellationToken ct = default)
    {
        using var stream = new MemoryStream(Encoding.UTF8.GetBytes(wire));
        var events = new List<SseEvent>();
        await foreach (var e in SseReader.ReadAsync(stream, SyslogFactory, ct))
            events.Add(e);
        return events;
    }

    [Fact]
    public async Task Parses_Multiple_Frames()
    {
        var wire =
            "event: syslog\ndata: {\"message\":\"one\",\"timestamp\":1}\n\n" +
            "event: syslog\ndata: {\"message\":\"two\"}\n\n";

        var events = await ReadAll(wire);

        Assert.Equal(2, events.Count);
        var first = Assert.IsType<SyslogMessageEvent>(events[0]);
        Assert.Equal("one", first.Message);
        Assert.Equal(1, first.Timestamp);
        Assert.Equal("syslog", first.EventName);
        var second = Assert.IsType<SyslogMessageEvent>(events[1]);
        Assert.Equal("two", second.Message);
        Assert.Null(second.Timestamp);
    }

    [Fact]
    public async Task Surfaces_Heartbeat()
    {
        var wire = "event: heartbeat\ndata: {}\n\n";
        var events = await ReadAll(wire);
        var e = Assert.Single(events);
        Assert.IsType<HeartbeatEvent>(e);
        Assert.Equal("heartbeat", e.EventName);
    }

    [Fact]
    public async Task Surfaces_Unknown_Event_With_RawData()
    {
        var wire = "event: somethingnew\ndata: {\"x\":1}\n\n";
        var events = await ReadAll(wire);
        var e = Assert.Single(events);
        var unknown = Assert.IsType<UnknownEvent>(e);
        Assert.Equal("somethingnew", unknown.EventName);
        Assert.Equal("{\"x\":1}", unknown.RawData);
    }

    [Fact]
    public async Task Handles_Frame_Split_Across_Reads()
    {
        // Deliver the same logical stream but split mid-frame across chunks.
        var chunks = new[]
        {
            "event: sys",
            "log\ndata: {\"mess",
            "age\":\"split\"}\n",
            "\nevent: heartbeat\ndata: {}\n\n",
        };
        using var content = new ChunkedContent(chunks);
        using var stream = await content.ReadAsStreamAsync();

        var events = new List<SseEvent>();
        await foreach (var e in SseReader.ReadAsync(stream, SyslogFactory))
            events.Add(e);

        Assert.Equal(2, events.Count);
        Assert.Equal("split", Assert.IsType<SyslogMessageEvent>(events[0]).Message);
        Assert.IsType<HeartbeatEvent>(events[1]);
    }

    [Fact]
    public async Task Handles_CRLF_And_Multiline_Data()
    {
        // CRLF line endings and a data field spanning two "data:" lines.
        var wire = "event: syslog\r\ndata: {\"message\":\r\ndata: \"joined\"}\r\n\r\n";
        var events = await ReadAll(wire);
        var e = Assert.Single(events);
        Assert.Equal("joined", Assert.IsType<SyslogMessageEvent>(e).Message);
    }

    [Fact]
    public async Task Ignores_Comment_Lines()
    {
        var wire = ": this is a keep-alive comment\nevent: syslog\ndata: {\"message\":\"ok\"}\n\n";
        var events = await ReadAll(wire);
        var e = Assert.Single(events);
        Assert.Equal("ok", Assert.IsType<SyslogMessageEvent>(e).Message);
    }

    [Fact]
    public async Task Flushes_Trailing_Frame_Without_Blank_Line()
    {
        var wire = "event: syslog\ndata: {\"message\":\"last\"}";
        var events = await ReadAll(wire);
        var e = Assert.Single(events);
        Assert.Equal("last", Assert.IsType<SyslogMessageEvent>(e).Message);
    }

    [Fact]
    public async Task Honors_Cancellation()
    {
        // A frame followed by an unterminated one; cancel after the first.
        var wire = "event: syslog\ndata: {\"message\":\"one\"}\n\n";
        using var stream = new MemoryStream(Encoding.UTF8.GetBytes(wire));
        using var cts = new CancellationTokenSource();

        var events = new List<SseEvent>();
        await foreach (var e in SseReader.ReadAsync(stream, SyslogFactory, cts.Token))
        {
            events.Add(e);
            cts.Cancel(); // request stop after first event
        }

        Assert.Single(events);
    }
}
