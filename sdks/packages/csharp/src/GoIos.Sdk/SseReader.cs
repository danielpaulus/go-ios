using System.Runtime.CompilerServices;
using System.Text;
using System.Text.Json;

namespace GoIos.Sdk;

/// <summary>
/// Maps an SSE <c>event:</c> name plus its raw <c>data:</c> JSON to a typed
/// <see cref="SseEvent"/>. Return <c>null</c> to fall through to the reader's
/// unknown-event handling.
/// </summary>
public delegate SseEvent? SseEventFactory(string eventName, string data);

/// <summary>
/// A reusable, allocation-light parser for the go-ios Server-Sent Events wire
/// contract. Each frame is:
/// <code>
/// event: &lt;name&gt;\n
/// data: &lt;compact-json&gt;\n
/// \n
/// </code>
/// The reader tolerates split reads (a frame arriving across multiple chunks),
/// multiple frames per chunk, CRLF or LF line endings, and multi-line
/// <c>data:</c> fields (joined with <c>\n</c> per the SSE spec). Comment lines
/// (starting with <c>:</c>) are ignored.
/// </summary>
public static class SseReader
{
    /// <summary>
    /// Reads an event stream and yields typed events. Heartbeats are surfaced as
    /// <see cref="HeartbeatEvent"/>; frames whose <c>event:</c> name the
    /// <paramref name="factory"/> does not recognize are surfaced as
    /// <see cref="UnknownEvent"/>.
    /// </summary>
    public static async IAsyncEnumerable<SseEvent> ReadAsync(
        Stream stream,
        SseEventFactory factory,
        [EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        using var reader = new StreamReader(stream, Encoding.UTF8, detectEncodingFromByteOrderMarks: false);

        string? eventName = null;
        var data = new StringBuilder();
        bool sawData = false;

        while (!cancellationToken.IsCancellationRequested)
        {
            string? line;
            try
            {
                line = await reader.ReadLineAsync(cancellationToken).ConfigureAwait(false);
            }
            catch (OperationCanceledException)
            {
                yield break;
            }

            if (line is null)
            {
                // End of stream. Flush a trailing frame that had no blank-line terminator.
                if (eventName is not null || sawData)
                {
                    var evt = Dispatch(factory, eventName, data.ToString());
                    if (evt is not null) yield return evt;
                }
                yield break;
            }

            if (line.Length == 0)
            {
                // Blank line: dispatch the accumulated frame (if any).
                if (eventName is not null || sawData)
                {
                    var evt = Dispatch(factory, eventName, data.ToString());
                    if (evt is not null) yield return evt;
                }
                eventName = null;
                data.Clear();
                sawData = false;
                continue;
            }

            if (line[0] == ':')
            {
                // SSE comment / keep-alive colon line — ignore.
                continue;
            }

            var colon = line.IndexOf(':');
            string field, value;
            if (colon < 0)
            {
                field = line;
                value = "";
            }
            else
            {
                field = line.Substring(0, colon);
                value = line.Substring(colon + 1);
                if (value.StartsWith(' ')) value = value.Substring(1); // strip one leading space
            }

            switch (field)
            {
                case "event":
                    eventName = value;
                    break;
                case "data":
                    if (sawData) data.Append('\n');
                    data.Append(value);
                    sawData = true;
                    break;
                // "id" and "retry" are part of SSE but unused by this contract.
                default:
                    break;
            }
        }
    }

    private static SseEvent? Dispatch(SseEventFactory factory, string? eventName, string data)
    {
        var name = eventName ?? "message";

        if (name == "heartbeat")
            return new HeartbeatEvent { EventName = name };

        var typed = factory(name, data);
        if (typed is not null)
            return typed with { EventName = name };

        return new UnknownEvent { EventName = name, RawData = data };
    }

    /// <summary>
    /// Helper for factories: deserialize <paramref name="data"/> to
    /// <typeparamref name="T"/> using the SDK's JSON options, tolerating empty
    /// payloads (returns a default-constructed instance).
    /// </summary>
    public static T Deserialize<T>(string data) where T : new()
    {
        if (string.IsNullOrWhiteSpace(data))
            return new T();
        return JsonSerializer.Deserialize<T>(data, JsonOptions.Default) ?? new T();
    }
}

/// <summary>Shared System.Text.Json options for SSE payload deserialization.</summary>
internal static class JsonOptions
{
    public static readonly JsonSerializerOptions Default = new()
    {
        PropertyNameCaseInsensitive = true,
        NumberHandling = System.Text.Json.Serialization.JsonNumberHandling.AllowReadingFromString,
    };
}
