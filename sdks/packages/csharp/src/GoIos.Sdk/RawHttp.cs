using System.Net.Http;
using System.Net.Http.Headers;
using System.Text.Json;
using GoIos;

namespace GoIos.Sdk;

/// <summary>
/// Internal helper wrapping the <see cref="HttpClient"/> used for endpoints the
/// generated JSON client cannot handle well: binary bodies (screenshot, image
/// mount), multipart uploads (app install, supervised pairing), and long-lived
/// SSE streams. Applies bearer auth and normalizes error responses.
/// </summary>
internal sealed class RawHttp
{
    private readonly HttpClient _http;
    private readonly Uri _baseUrl;
    private readonly string? _apiKey;

    public RawHttp(HttpClient http, Uri baseUrl, string? apiKey)
    {
        _http = http;
        _baseUrl = baseUrl;
        _apiKey = apiKey;
    }

    public HttpClient Http => _http;

    public HttpRequestMessage NewRequest(HttpMethod method, string relativePath)
    {
        var req = new HttpRequestMessage(method, new Uri(_baseUrl, relativePath));
        if (!string.IsNullOrEmpty(_apiKey))
            req.Headers.Authorization = new AuthenticationHeaderValue("Bearer", _apiKey);
        return req;
    }

    public async Task<byte[]> GetBytesAsync(string path, string accept, CancellationToken ct)
    {
        using var req = NewRequest(HttpMethod.Get, path);
        req.Headers.Accept.ParseAdd(accept);
        using var resp = await _http.SendAsync(req, HttpCompletionOption.ResponseHeadersRead, ct).ConfigureAwait(false);
        await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
        return await resp.Content.ReadAsByteArrayAsync(ct).ConfigureAwait(false);
    }

    /// <summary>
    /// Open a raw binary stream (NOT SSE): sends the request with
    /// <see cref="HttpCompletionOption.ResponseHeadersRead"/> and returns the live
    /// response <see cref="Stream"/> of bytes. The returned <see cref="BinaryStream"/>
    /// owns the request/response and must be disposed to release the connection.
    /// </summary>
    public async Task<BinaryStream> OpenBinaryStreamAsync(HttpRequestMessage req, string? accept, CancellationToken ct)
    {
        if (!string.IsNullOrEmpty(accept)) req.Headers.Accept.ParseAdd(accept);
        var resp = await _http.SendAsync(req, HttpCompletionOption.ResponseHeadersRead, ct).ConfigureAwait(false);
        try
        {
            await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
#if NET5_0_OR_GREATER
            var stream = await resp.Content.ReadAsStreamAsync(ct).ConfigureAwait(false);
#else
            var stream = await resp.Content.ReadAsStreamAsync().ConfigureAwait(false);
#endif
            return new BinaryStream(stream, resp, req, resp.Content.Headers.ContentType?.MediaType);
        }
        catch
        {
            resp.Dispose();
            req.Dispose();
            throw;
        }
    }

    /// <summary>Send a request whose response body is raw bytes (e.g. octet-stream file pull).</summary>
    public async Task<byte[]> SendBytesAsync(HttpRequestMessage req, CancellationToken ct)
    {
        using (req)
        {
            using var resp = await _http.SendAsync(req, HttpCompletionOption.ResponseHeadersRead, ct).ConfigureAwait(false);
            await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
            return await resp.Content.ReadAsByteArrayAsync(ct).ConfigureAwait(false);
        }
    }

    /// <summary>Send a request expecting no meaningful body; returns the (possibly empty) text.</summary>
    public async Task<string> SendTextAsync(HttpRequestMessage req, CancellationToken ct)
    {
        using (req)
        {
            using var resp = await _http.SendAsync(req, ct).ConfigureAwait(false);
            await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
            return await resp.Content.ReadAsStringAsync(ct).ConfigureAwait(false);
        }
    }

    public async Task<T> SendJsonAsync<T>(HttpRequestMessage req, CancellationToken ct)
    {
        using (req)
        {
            using var resp = await _http.SendAsync(req, ct).ConfigureAwait(false);
            await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
#if NET5_0_OR_GREATER
            await using var s = await resp.Content.ReadAsStreamAsync(ct).ConfigureAwait(false);
#else
            using var s = await resp.Content.ReadAsStreamAsync().ConfigureAwait(false);
#endif
            var value = await JsonSerializer.DeserializeAsync<T>(s, JsonOptions.Default, ct).ConfigureAwait(false);
            return value!;
        }
    }

    public static async Task EnsureSuccessAsync(HttpResponseMessage resp, CancellationToken ct)
    {
        if (resp.IsSuccessStatusCode) return;
        string body = "";
        try { body = await resp.Content.ReadAsStringAsync(ct).ConfigureAwait(false); } catch { /* ignore */ }
        throw new IosApiException((int)resp.StatusCode, resp.ReasonPhrase, body);
    }
}

/// <summary>
/// A live read-only <see cref="Stream"/> of raw response bytes returned by a
/// binary streaming endpoint (pcap, UI video, MJPEG screenshot stream). It is a
/// pass-through over the HTTP response body: reads pull bytes off the socket as
/// they arrive and honor the <see cref="CancellationToken"/> passed to the
/// originating call. Disposing it releases the underlying HTTP connection.
/// </summary>
public sealed class BinaryStream : Stream
{
    private readonly Stream _inner;
    private readonly HttpResponseMessage _response;
    private readonly HttpRequestMessage _request;

    /// <summary>The response <c>Content-Type</c> (e.g. <c>application/vnd.tcpdump.pcap</c>, <c>multipart/x-mixed-replace</c>), if any.</summary>
    public string? ContentType { get; }

    internal BinaryStream(Stream inner, HttpResponseMessage response, HttpRequestMessage request, string? contentType)
    {
        _inner = inner;
        _response = response;
        _request = request;
        ContentType = contentType;
    }

    /// <inheritdoc/>
    public override bool CanRead => true;
    /// <inheritdoc/>
    public override bool CanSeek => false;
    /// <inheritdoc/>
    public override bool CanWrite => false;
    /// <inheritdoc/>
    public override long Length => throw new NotSupportedException();
    /// <inheritdoc/>
    public override long Position { get => throw new NotSupportedException(); set => throw new NotSupportedException(); }

    /// <inheritdoc/>
    public override int Read(byte[] buffer, int offset, int count) => _inner.Read(buffer, offset, count);
    /// <inheritdoc/>
    public override Task<int> ReadAsync(byte[] buffer, int offset, int count, CancellationToken cancellationToken)
        => _inner.ReadAsync(buffer, offset, count, cancellationToken);
    /// <inheritdoc/>
    public override ValueTask<int> ReadAsync(Memory<byte> buffer, CancellationToken cancellationToken = default)
        => _inner.ReadAsync(buffer, cancellationToken);

    /// <inheritdoc/>
    public override void Flush() { }
    /// <inheritdoc/>
    public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();
    /// <inheritdoc/>
    public override void SetLength(long value) => throw new NotSupportedException();
    /// <inheritdoc/>
    public override void Write(byte[] buffer, int offset, int count) => throw new NotSupportedException();

    /// <inheritdoc/>
    protected override void Dispose(bool disposing)
    {
        if (disposing)
        {
            _inner.Dispose();
            _response.Dispose();
            _request.Dispose();
        }
        base.Dispose(disposing);
    }

    /// <inheritdoc/>
    public override async ValueTask DisposeAsync()
    {
        await _inner.DisposeAsync().ConfigureAwait(false);
        _response.Dispose();
        _request.Dispose();
        await base.DisposeAsync().ConfigureAwait(false);
    }
}
