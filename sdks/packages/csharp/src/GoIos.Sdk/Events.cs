using System.Text.Json;
using System.Text.Json.Serialization;

namespace GoIos.Sdk;

/// <summary>
/// Base type for every event surfaced by a streaming (Server-Sent Events)
/// endpoint. The <see cref="EventName"/> is the SSE <c>event:</c> field.
/// </summary>
public abstract record SseEvent
{
    /// <summary>The SSE <c>event:</c> name this frame carried.</summary>
    [JsonIgnore]
    public string EventName { get; init; } = "";
}

/// <summary>
/// Periodic keep-alive frame emitted on every stream (<c>event: heartbeat</c>,
/// empty <c>{}</c> payload). Lets a client tell "live but idle" from "dropped".
/// </summary>
public sealed record HeartbeatEvent : SseEvent;

/// <summary>
/// An event whose <c>event:</c> name did not match any known type for the stream.
/// Surfaced (not dropped) for forward-compatibility. The raw JSON payload is kept.
/// </summary>
public sealed record UnknownEvent : SseEvent
{
    /// <summary>The raw, undeserialized <c>data:</c> JSON payload (may be empty).</summary>
    public string RawData { get; init; } = "";
}

// --- /syslog : SyslogEvents ------------------------------------------------

/// <summary>A single syslog line from the device (<c>event: syslog</c>).</summary>
public sealed record SyslogMessageEvent : SseEvent
{
    [JsonPropertyName("message")] public string Message { get; init; } = "";
    [JsonPropertyName("timestamp")] public long? Timestamp { get; init; }
}

// --- /notifications : NotificationEvents -----------------------------------

/// <summary>An app foreground/background/lifecycle state change (<c>event: appstate</c>).</summary>
public sealed record AppStateNotificationEvent : SseEvent
{
    [JsonPropertyName("bundleId")] public string BundleId { get; init; } = "";

    /// <summary>
    /// New application state. Typical values: <c>foreground</c>, <c>background</c>,
    /// <c>suspended</c>, <c>terminated</c>, <c>unknown</c>.
    /// </summary>
    [JsonPropertyName("state")] public string State { get; init; } = "";
    [JsonPropertyName("timestamp")] public long? Timestamp { get; init; }
}

// --- /ostrace : OsTraceEvents ----------------------------------------------

/// <summary>A structured os_log trace entry (<c>event: ostrace</c>).</summary>
public sealed record OsTraceEntryEvent : SseEvent
{
    [JsonPropertyName("pid")] public int? Pid { get; init; }
    [JsonPropertyName("processName")] public string? ProcessName { get; init; }

    /// <summary>Log level, e.g. <c>default</c>, <c>info</c>, <c>debug</c>, <c>error</c>, <c>fault</c>.</summary>
    [JsonPropertyName("level")] public string? Level { get; init; }
    [JsonPropertyName("subsystem")] public string? Subsystem { get; init; }
    [JsonPropertyName("category")] public string? Category { get; init; }
    [JsonPropertyName("message")] public string Message { get; init; } = "";
    [JsonPropertyName("timestamp")] public long? Timestamp { get; init; }
}

// --- /listen : ListenEvents ------------------------------------------------

/// <summary>Device properties reported by usbmuxd / lockdown (payload of an attach event).</summary>
public sealed record DevicePropertiesData
{
    [JsonPropertyName("connectionSpeed")] public int? ConnectionSpeed { get; init; }
    [JsonPropertyName("connectionType")] public string? ConnectionType { get; init; }
    [JsonPropertyName("deviceID")] public int? DeviceID { get; init; }
    [JsonPropertyName("locationID")] public int? LocationID { get; init; }
    [JsonPropertyName("productID")] public int? ProductID { get; init; }
    [JsonPropertyName("serialNumber")] public string SerialNumber { get; init; } = "";
}

/// <summary>A device was attached to or detached from the host (<c>event: attachdetach</c>).</summary>
public sealed record AttachDetachEventEvent : SseEvent
{
    /// <summary>Event kind: <c>attached</c>, <c>detached</c>, or <c>paired</c>.</summary>
    [JsonPropertyName("event")] public string Event { get; init; } = "";
    [JsonPropertyName("deviceID")] public int? DeviceID { get; init; }
    [JsonPropertyName("udid")] public string? Udid { get; init; }

    /// <summary>Full device properties, present on <c>attached</c>.</summary>
    [JsonPropertyName("properties")] public DevicePropertiesData? Properties { get; init; }
}

// --- /sysmontap : SysmontapEvents ------------------------------------------

/// <summary>
/// A single sysmontap CPU-usage sample (<c>event: sample</c>). This is an open
/// map — samplers report additional keys depending on the OS — so the well-known
/// load fields are surfaced strongly-typed and the rest is kept in <see cref="Extra"/>.
/// </summary>
public sealed record CpuUsageSampleEvent : SseEvent
{
    /// <summary>Total CPU load across all cores (0–100).</summary>
    [JsonPropertyName("CPU_TotalLoad")] public double? CpuTotalLoad { get; init; }

    /// <summary>System (kernel) CPU load.</summary>
    [JsonPropertyName("SystemLoad")] public double? SystemLoad { get; init; }

    /// <summary>User CPU load.</summary>
    [JsonPropertyName("UserLoad")] public double? UserLoad { get; init; }

    /// <summary>Any extra sampler keys not modelled above (OS-dependent).</summary>
    [JsonExtensionData] public Dictionary<string, object?>? Extra { get; init; }
}

// --- /jobs/{id}/logs : JobLogEvents ----------------------------------------

/// <summary>A single line of a job's log output (<c>event: log</c>).</summary>
public sealed record JobLogLineEvent : SseEvent
{
    /// <summary>The raw log line (already newline-terminated in the buffer).</summary>
    [JsonPropertyName("line")] public string Line { get; init; } = "";
}
