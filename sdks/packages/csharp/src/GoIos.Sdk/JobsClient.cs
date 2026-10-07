using System.Net.Http;
using System.Runtime.CompilerServices;
using GoIos.Sdk;
using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>
/// Background-job operations (runtest / runwda / port-forward) for a single
/// device. Jobs are device-scoped in the daemon
/// (<c>/api/v1/device/{udid}/jobs/...</c>); obtain this via
/// <see cref="DeviceClient.Jobs"/>.
/// </summary>
public sealed class JobsClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;
    private readonly RawHttp _raw;

    internal JobsClient(string udid, Gen.DefaultApi api, RawHttp raw)
    {
        _udid = udid;
        _api = api;
        _raw = raw;
    }

    /// <summary>Start an XCUITest run as a background job (<c>POST /jobs/runtest</c>).</summary>
    public Task<GenModel.Job> RuntestAsync(
        GenModel.RunTestRequest request, CancellationToken cancellationToken = default)
        => _api.DevicesStartRunTestAsync(_udid, request, cancellationToken);

    /// <summary>Start WebDriverAgent as a background job (<c>POST /jobs/runwda</c>).</summary>
    public Task<GenModel.Job> RunwdaAsync(
        GenModel.RunTestRequest? request = null, CancellationToken cancellationToken = default)
        => _api.DevicesStartRunWdaAsync(_udid, request, cancellationToken);

    /// <summary>Start a host↔device TCP port-forward as a background job (<c>POST /jobs/forward</c>).</summary>
    public Task<GenModel.Job> ForwardAsync(
        int hostPort, int targetPort, CancellationToken cancellationToken = default)
        => _api.DevicesStartForwardAsync(_udid, new GenModel.ForwardRequest(hostPort, targetPort), cancellationToken);

    /// <summary>List background jobs for this device (<c>GET /jobs</c>).</summary>
    public Task<List<GenModel.Job>> ListAsync(CancellationToken cancellationToken = default)
        => _api.DevicesListJobsAsync(_udid, cancellationToken);

    /// <summary>Get a single job by id (<c>GET /jobs/{id}</c>).</summary>
    public Task<GenModel.Job> GetAsync(string id, CancellationToken cancellationToken = default)
        => _api.DevicesGetJobAsync(_udid, id, cancellationToken);

    /// <summary>Stop / delete a job by id (<c>DELETE /jobs/{id}</c>).</summary>
    public Task<GenModel.GenericResponse> DeleteAsync(string id, CancellationToken cancellationToken = default)
        => _api.DevicesStopJobAsync(_udid, id, cancellationToken);

    /// <summary>
    /// Stream a job's log output (<c>GET /jobs/{id}/logs</c>) as typed events.
    /// Log lines are surfaced as <see cref="JobLogLineEvent"/>; keep-alives as
    /// <see cref="HeartbeatEvent"/>.
    /// </summary>
    public IAsyncEnumerable<SseEvent> LogsAsync(string id, CancellationToken cancellationToken = default)
        => StreamAsync($"api/v1/device/{Esc(_udid)}/jobs/{Esc(id)}/logs", JobLogsFactory, cancellationToken);

    private static SseEvent? JobLogsFactory(string name, string data) => name switch
    {
        "log" => SseReader.Deserialize<JobLogLineEvent>(data),
        _ => null,
    };

    private async IAsyncEnumerable<SseEvent> StreamAsync(
        string path, SseEventFactory factory,
        [EnumeratorCancellation] CancellationToken cancellationToken)
    {
        using var req = _raw.NewRequest(HttpMethod.Get, path);
        req.Headers.Accept.ParseAdd("text/event-stream");
        using var resp = await _raw.Http
            .SendAsync(req, HttpCompletionOption.ResponseHeadersRead, cancellationToken)
            .ConfigureAwait(false);
        await RawHttp.EnsureSuccessAsync(resp, cancellationToken).ConfigureAwait(false);

        await using var stream = await resp.Content.ReadAsStreamAsync(cancellationToken).ConfigureAwait(false);
        await foreach (var e in SseReader.ReadAsync(stream, factory, cancellationToken).ConfigureAwait(false))
            yield return e;
    }

    private static string Esc(string s) => Uri.EscapeDataString(s);
}

/// <summary>Global CoreDevice / RemoteXPC tunnel-agent operations. Obtained via <see cref="IosClient.Tunnels"/>.</summary>
public sealed class TunnelsClient
{
    private readonly Gen.DefaultApi _api;
    internal TunnelsClient(Gen.DefaultApi api) => _api = api;

    /// <summary>List active tunnels (<c>GET /tunnels</c>).</summary>
    public Task<List<GenModel.Tunnel>> ListAsync(CancellationToken cancellationToken = default)
        => _api.ListTunnelsAsync(cancellationToken);

    /// <summary>Stop a tunnel for a device (<c>DELETE /tunnels/{udid}</c>).</summary>
    public Task<GenModel.TunnelStopped> DeleteAsync(string udid, CancellationToken cancellationToken = default)
        => _api.StopTunnelAsync(udid, cancellationToken);

    /// <summary>Refresh (re-establish) a device's tunnel (<c>POST /tunnels/{udid}/refresh</c>).</summary>
    public Task<GenModel.Tunnel> RefreshAsync(string udid, CancellationToken cancellationToken = default)
        => _api.RefreshTunnelAsync(udid, cancellationToken);

    /// <summary>Shut down the whole tunnel agent (<c>POST /tunnel-agent/shutdown</c>).</summary>
    public Task<GenModel.AgentShutdown> ShutdownAgentAsync(CancellationToken cancellationToken = default)
        => _api.ShutdownTunnelAgentAsync(cancellationToken);
}
