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
