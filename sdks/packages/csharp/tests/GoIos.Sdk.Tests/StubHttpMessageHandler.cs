using System.Net;
using System.Net.Http;

namespace GoIos.Sdk.Tests;

/// <summary>
/// A stub <see cref="HttpMessageHandler"/> that returns a caller-supplied
/// response (built lazily so streaming bodies can be provided). Records the
/// requests it saw for assertions.
/// </summary>
internal sealed class StubHttpMessageHandler : HttpMessageHandler
{
    private readonly Func<HttpRequestMessage, HttpResponseMessage> _responder;
    public List<HttpRequestMessage> Requests { get; } = new();

    public StubHttpMessageHandler(Func<HttpRequestMessage, HttpResponseMessage> responder)
        => _responder = responder;

    public static StubHttpMessageHandler Json(string json, HttpStatusCode status = HttpStatusCode.OK)
        => new(_ => new HttpResponseMessage(status)
        {
            Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json"),
        });

    protected override Task<HttpResponseMessage> SendAsync(
        HttpRequestMessage request, CancellationToken cancellationToken)
    {
        Requests.Add(request);
        return Task.FromResult(_responder(request));
    }
}

/// <summary>
/// An HttpContent whose stream releases bytes in caller-controlled chunks, so we
/// can exercise the SSE reader against a frame that arrives split across reads.
/// </summary>
internal sealed class ChunkedContent : HttpContent
{
    private readonly IReadOnlyList<byte[]> _chunks;

    public ChunkedContent(IEnumerable<string> chunks)
    {
        _chunks = chunks.Select(c => System.Text.Encoding.UTF8.GetBytes(c)).ToList();
        Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("text/event-stream");
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
