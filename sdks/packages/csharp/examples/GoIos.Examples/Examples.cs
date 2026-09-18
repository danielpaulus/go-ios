using GoIos;
using GoIos.Sdk;

namespace GoIos.Examples;

/// <summary>
/// Every example is a small, heavily-commented static method that takes the
/// shared <see cref="ExampleContext"/> and returns an <see cref="ExampleResult"/>.
///
/// The methods are intentionally independent so each one reads as standalone
/// documentation for one slice of the SDK. The dispatcher in Program.cs runs
/// them by name (or all of them in sequence).
/// </summary>
public static class Examples
{
    // ---------------------------------------------------------------------
    // 1. list-devices — build the client and enumerate attached devices.
    // ---------------------------------------------------------------------
    public static async Task<ExampleResult> ListDevicesAsync(ExampleContext ctx, CancellationToken ct)
    {
        // Devices.ListAsync() returns the daemon's device list. The generated
        // model exposes the JSON "deviceList" array as `VarDeviceList`.
        var devices = await ctx.Client.Devices.ListAsync(ct);
        var list = devices.VarDeviceList ?? new();

        Console.WriteLine($"Found {list.Count} device(s):");
        foreach (var d in list)
        {
            // `Properties.SerialNumber` is the udid used to scope every
            // device-specific call. ConnectionType is "USB" or "Network".
            Console.WriteLine(
                $"  - udid={d.Properties?.SerialNumber} " +
                $"connection={d.Properties?.ConnectionType} " +
                $"deviceId={d.DeviceID}");
        }

        if (list.Count == 0)
            return ExampleResult.Skip("no devices attached to the daemon");

        return ExampleResult.Ok();
    }

    // ---------------------------------------------------------------------
    // 2. device-info — lockdown + instruments:* values for one device.
    // ---------------------------------------------------------------------
    public static async Task<ExampleResult> DeviceInfoAsync(ExampleContext ctx, CancellationToken ct)
    {
        var udid = await ctx.ResolveUdidAsync(ct);
        if (udid is null)
            return ExampleResult.Skip("no target device (set GO_IOS_UDID or attach a device)");

        // Scope operations to one device by udid. `Device(udid)` is cheap;
        // it just wraps the shared HTTP pipeline.
        var device = ctx.Client.Device(udid);

        // InfoAsync() returns an open map of lockdown values (ProductType,
        // ProductVersion, DeviceName, ...) plus instruments:* keys. It is
        // surfaced as a dictionary so nothing is lost to a fixed DTO.
        var info = await device.InfoAsync(ct);

        Console.WriteLine($"Device info for {udid} ({info.Count} keys):");
        foreach (var key in new[] { "ProductType", "ProductVersion", "DeviceName", "BuildVersion" })
        {
            if (info.TryGetValue(key, out var value))
                Console.WriteLine($"  {key,-16}= {value}");
        }

        return ExampleResult.Ok();
    }

    // ---------------------------------------------------------------------
    // 3. list-apps — installed applications on the device.
    // ---------------------------------------------------------------------
    public static async Task<ExampleResult> ListAppsAsync(ExampleContext ctx, CancellationToken ct)
    {
        var udid = await ctx.ResolveUdidAsync(ct);
        if (udid is null)
            return ExampleResult.Skip("no target device (set GO_IOS_UDID or attach a device)");

        var device = ctx.Client.Device(udid);

        // Apps.ListAsync() returns installed apps. Each AppInfo is an Info.plist
        // map; the well-known keys are surfaced strongly-typed.
        var apps = await device.Apps.ListAsync(ct);

        Console.WriteLine($"{apps.Count} installed app(s) on {udid}. First few:");
        foreach (var app in apps.Take(10))
        {
            Console.WriteLine(
                $"  - {app.CFBundleIdentifier} " +
                $"({app.CFBundleName} {app.CFBundleShortVersionString})");
        }

        return ExampleResult.Ok();
    }

    // ---------------------------------------------------------------------
    // 4. screenshot — capture a PNG to ./screenshot.png.
    // ---------------------------------------------------------------------
    public static async Task<ExampleResult> ScreenshotAsync(ExampleContext ctx, CancellationToken ct)
    {
        var udid = await ctx.ResolveUdidAsync(ct);
        if (udid is null)
            return ExampleResult.Skip("no target device (set GO_IOS_UDID or attach a device)");

        var device = ctx.Client.Device(udid);

        // ScreenshotAsync() returns the raw PNG bytes straight off the wire.
        byte[] png = await device.ScreenshotAsync(ct);

        var path = Path.Combine(Directory.GetCurrentDirectory(), "screenshot.png");
        await File.WriteAllBytesAsync(path, png, ct);

        Console.WriteLine($"Wrote {png.Length:N0} bytes to {path}");
        return ExampleResult.Ok();
    }

    // ---------------------------------------------------------------------
    // 5. stream-syslog — consume the syslog SSE stream for a bounded window.
    // ---------------------------------------------------------------------
    public static async Task<ExampleResult> StreamSyslogAsync(ExampleContext ctx, CancellationToken ct)
    {
        var udid = await ctx.ResolveUdidAsync(ct);
        if (udid is null)
            return ExampleResult.Skip("no target device (set GO_IOS_UDID or attach a device)");

        var device = ctx.Client.Device(udid);

        // Streaming endpoints are long-lived IAsyncEnumerables. We bound this
        // example two ways so it always terminates: a ~5s time budget AND a
        // cap of ~20 syslog events. Linking the caller's token means Ctrl-C
        // still cancels promptly.
        const int maxEvents = 20;
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(5));
        using var linked = CancellationTokenSource.CreateLinkedTokenSource(ct, timeout.Token);

        Console.WriteLine($"Streaming syslog from {udid} (up to {maxEvents} events / ~5s)...");
        int count = 0;

        try
        {
            await foreach (var e in device.SyslogAsync(linked.Token))
            {
                switch (e)
                {
                    case SyslogMessageEvent s:
                        count++;
                        // Trim noisy long lines for the console.
                        var msg = s.Message.Length > 100 ? s.Message[..100] + "..." : s.Message;
                        Console.WriteLine($"  [{count:D2}] {msg}");
                        if (count >= maxEvents)
                            return ExampleResult.Ok($"captured {count} syslog event(s)");
                        break;

                    case HeartbeatEvent:
                        // Live-but-idle keep-alive: the stream is up, nothing to log yet.
                        break;

                    case UnknownEvent u:
                        Console.WriteLine($"  (unknown event '{u.EventName}': {u.RawData})");
                        break;
                }
            }
        }
        catch (OperationCanceledException) when (timeout.IsCancellationRequested && !ct.IsCancellationRequested)
        {
            // Expected: our 5s budget elapsed. That is a clean stop, not a failure.
        }

        return ExampleResult.Ok($"captured {count} syslog event(s) before the window closed");
    }

    // ---------------------------------------------------------------------
    // 6. ui-automation — OPTIONAL. Tap + type via the UI (WDA) backend.
    //
    //    PREREQUISITE: a running / forwarded WebDriverAgent. Start it with
    //    `ios runwda` (or the daemon's runwda job) and, if needed, forward its
    //    port so the daemon can reach it. Without a reachable WDA the calls
    //    fail; this example catches that and SKIPs rather than failing the run.
    //
    //    This example MUTATES the device (it injects a tap and keystrokes), so
    //    the runner only executes it when RUN_UI=1.
    // ---------------------------------------------------------------------
    public static async Task<ExampleResult> UiAutomationAsync(ExampleContext ctx, CancellationToken ct)
    {
        var udid = await ctx.ResolveUdidAsync(ct);
        if (udid is null)
            return ExampleResult.Skip("no target device (set GO_IOS_UDID or attach a device)");

        var device = ctx.Client.Device(udid);
        var ui = device.Ui;

        // Give UI calls a short timeout so an unreachable WDA fails fast rather
        // than hanging. Backend defaults to "wda".
        var options = new UiClient.Options { Timeout = 15 };

        try
        {
            // Query the backend health first — the cheapest way to detect that
            // WDA is actually forwarded and reachable.
            var status = await ui.StatusAsync(options, ct);
            Console.WriteLine($"UI backend reachable ({status.Count} status keys). Injecting a tap + type...");

            // A tap near the top-centre, then type some text into whatever is
            // focused. Coordinates are illustrative; adjust for your screen.
            await ui.TapAsync(150, 150, options, ct);
            await ui.TypeAsync("hello from GoIos.Sdk", options, ct);

            Console.WriteLine("Tap + type dispatched.");
            return ExampleResult.Ok();
        }
        catch (IosApiException ex)
        {
            // The daemon reachable but the UI backend/WDA is not wired up.
            return ExampleResult.Skip($"UI backend unavailable (HTTP {ex.StatusCode}) — is WDA forwarded? Run `ios runwda`.");
        }
        catch (HttpRequestException ex)
        {
            return ExampleResult.Skip($"UI backend unreachable ({ex.Message}) — is WDA forwarded? Run `ios runwda`.");
        }
    }
}
