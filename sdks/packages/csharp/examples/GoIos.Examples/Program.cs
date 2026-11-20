using GoIos.Examples;

// =============================================================================
// GoIos.Examples — runnable docs + pre-release smoke test for GoIos.Sdk.
//
// Usage:
//   dotnet run --project examples/GoIos.Examples -- run-all
//   dotnet run --project examples/GoIos.Examples -- list-devices
//   dotnet run --project examples/GoIos.Examples -- device-info
//   dotnet run --project examples/GoIos.Examples -- list-apps
//   dotnet run --project examples/GoIos.Examples -- screenshot
//   dotnet run --project examples/GoIos.Examples -- stream-syslog
//   dotnet run --project examples/GoIos.Examples -- ui-automation
//
// `run-all` runs examples 1-5 (all read-only) in sequence. The mutating
// ui-automation example is only included in run-all when RUN_UI=1.
//
// Exit codes:
//   0  every selected example either succeeded or SKIPped cleanly
//   1  an example threw (hard failure)
//   2  configuration error (missing GO_IOS_API_KEY, unknown command)
//
// Configuration is via environment variables — see ExampleContext for details.
// =============================================================================

// Cancel promptly on Ctrl-C and let the streaming example unwind cleanly.
using var cts = new CancellationTokenSource();
Console.CancelKeyPress += (_, e) =>
{
    e.Cancel = true;      // don't hard-kill; let cooperative cancellation run.
    cts.Cancel();
};

// One command word selects which example(s) to run; default is run-all.
var command = args.Length > 0 ? args[0].Trim().ToLowerInvariant() : "run-all";

if (command is "-h" or "--help" or "help")
{
    PrintUsage();
    return 0;
}

// Build the environment-driven context (client + config). A missing API key is
// a configuration error: the helper already printed how to fix it.
using var ctx = ExampleContext.FromEnvironment();
if (ctx is null)
    return 2;

Console.WriteLine($"go-ios examples -> {ctx.BaseUrl ?? "(auto-discovered local daemon)"}");
Console.WriteLine();

// Named example registry. Order matters for run-all.
var registry = new (string Name, Func<ExampleContext, CancellationToken, Task<ExampleResult>> Run)[]
{
    ("list-devices",  Examples.ListDevicesAsync),
    ("device-info",   Examples.DeviceInfoAsync),
    ("list-apps",     Examples.ListAppsAsync),
    ("screenshot",    Examples.ScreenshotAsync),
    ("stream-syslog", Examples.StreamSyslogAsync),
    ("ui-automation", Examples.UiAutomationAsync),
};

// Decide the set to run.
List<(string Name, Func<ExampleContext, CancellationToken, Task<ExampleResult>> Run)> toRun;
if (command == "run-all")
{
    var runUi = Environment.GetEnvironmentVariable("RUN_UI") == "1";
    // Examples 1-5 always; ui-automation only when explicitly opted in.
    toRun = registry.Where(e => e.Name != "ui-automation" || runUi).ToList();
}
else
{
    var match = registry.FirstOrDefault(e => e.Name == command);
    if (match.Run is null)
    {
        Console.Error.WriteLine($"Unknown command: '{command}'.");
        Console.Error.WriteLine();
        PrintUsage();
        return 2;
    }
    toRun = new() { match };
}

// Run selected examples. A thrown exception is a hard failure that fails the
// whole run (exit 1). A SKIP is reported but does not fail.
int failures = 0;
int skipped = 0;

foreach (var (name, run) in toRun)
{
    Console.WriteLine($"=== {name} ===");
    try
    {
        var result = await run(ctx, cts.Token);
        if (result.Skipped)
        {
            skipped++;
            Console.WriteLine($"SKIP: {result.Reason}");
        }
        else
        {
            Console.WriteLine(result.Reason is { } note ? $"OK: {note}" : "OK");
        }
    }
    catch (OperationCanceledException) when (cts.IsCancellationRequested)
    {
        Console.Error.WriteLine("CANCELLED (Ctrl-C)");
        return 1;
    }
    catch (Exception ex)
    {
        failures++;
        Console.Error.WriteLine($"FAIL: {ex.GetType().Name}: {ex.Message}");
    }
    Console.WriteLine();
}

// Summary line, then the exit code the smoke test keys off of.
Console.WriteLine($"Summary: {toRun.Count - failures - skipped} ok, {skipped} skipped, {failures} failed.");
return failures == 0 ? 0 : 1;

static void PrintUsage()
{
    Console.WriteLine(
        "GoIos.Examples — runnable examples for the go-ios C#/.NET SDK.\n" +
        "\n" +
        "Usage: dotnet run --project examples/GoIos.Examples -- <command>\n" +
        "\n" +
        "Commands:\n" +
        "  run-all         Run examples 1-5 in sequence (ui-automation too if RUN_UI=1).\n" +
        "  list-devices    List attached devices.\n" +
        "  device-info     Print lockdown/info values for the target device.\n" +
        "  list-apps       List installed applications.\n" +
        "  screenshot      Capture a PNG to ./screenshot.png.\n" +
        "  stream-syslog   Stream ~20 syslog events (~5s) then stop.\n" +
        "  ui-automation   Tap + type via WDA (needs a forwarded WebDriverAgent).\n" +
        "\n" +
        "Environment:\n" +
        "  GO_IOS_BASE_URL  Daemon base URL (default http://localhost:8080).\n" +
        "  GO_IOS_API_KEY   Bearer token (required unless daemon runs --disable-auth).\n" +
        "  GO_IOS_UDID      Target device udid (optional; first device if unset).\n" +
        "  RUN_UI=1         Include ui-automation in run-all.\n");
}
