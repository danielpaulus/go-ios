using Gen = GoIos.Sdk.Generated.Api;
using GenModel = GoIos.Sdk.Generated.Model;

namespace GoIos;

/// <summary>WebDriverAgent (XCUITest) session operations for a single device.</summary>
public sealed class WdaClient
{
    private readonly string _udid;
    private readonly Gen.DefaultApi _api;

    internal WdaClient(string udid, Gen.DefaultApi api)
    {
        _udid = udid;
        _api = api;
    }

    /// <summary>Start a new WebDriverAgent (XCUITest) runner session.</summary>
    public Task<GenModel.WdaSession> CreateSessionAsync(
        GenModel.WdaConfig config, CancellationToken cancellationToken = default)
        => _api.DevicesCreateWdaSessionAsync(_udid, config, cancellationToken);

    /// <summary>Read a running WebDriverAgent session by id.</summary>
    public Task<GenModel.WdaSession> ReadSessionAsync(
        string sessionId, CancellationToken cancellationToken = default)
        => _api.DevicesGetWdaSessionAsync(_udid, sessionId, cancellationToken);

    /// <summary>Stop and delete a running WebDriverAgent session by id.</summary>
    public Task<GenModel.WdaSession> DeleteSessionAsync(
        string sessionId, CancellationToken cancellationToken = default)
        => _api.DevicesDeleteWdaSessionAsync(_udid, sessionId, cancellationToken);
}
