using System.Net.Http;

namespace GoIos;

/// <summary>
/// Configuration for <see cref="IosClient"/>.
/// </summary>
public sealed class IosClientOptions
{
    /// <summary>
    /// Base URL of the go-ios REST server, e.g. <c>http://127.0.0.1:54321</c>.
    /// Optional: when unset (null/empty), the client resolves the address in this
    /// order — the <c>GO_IOS_BASE_URL</c> environment variable, then discovery of a
    /// local daemon via <c>&lt;home&gt;/rest-api.json</c> (see <see cref="Discovery"/>).
    /// Set this explicitly to target a remote daemon; it is then used verbatim and
    /// discovery is skipped.
    /// </summary>
    public string? BaseUrl { get; set; }

    /// <summary>
    /// Bearer token sent as <c>Authorization: Bearer &lt;ApiKey&gt;</c> on every
    /// request. Optional (the server may be launched with <c>--disable-auth</c>),
    /// but strongly encouraged and sent whenever set.
    /// </summary>
    public string? ApiKey { get; set; }

    /// <summary>
    /// Optional caller-supplied <see cref="HttpClient"/> used for both unary and
    /// streaming (SSE) calls. When null, the SDK creates and owns one.
    /// Note: streaming endpoints are long-lived; if you supply your own client,
    /// do not set a short <see cref="HttpClient.Timeout"/>.
    /// </summary>
    public HttpClient? HttpClient { get; set; }
}
