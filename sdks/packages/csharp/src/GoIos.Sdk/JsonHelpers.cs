using System.Text.Json;

namespace GoIos.Sdk;

/// <summary>
/// Shared normalization helpers for endpoints whose response is an open,
/// schema-less JSON object (the generated client surfaces these as
/// <see cref="object"/>). Returns a plain string-keyed dictionary.
/// </summary>
internal static class JsonHelpers
{
    /// <summary>Normalize an arbitrary generated/Newtonsoft-parsed value into a dictionary.</summary>
    public static IReadOnlyDictionary<string, object?> ToDictionary(object? raw)
    {
        if (raw is null) return new Dictionary<string, object?>();
        var json = raw is string s ? s : Newtonsoft.Json.JsonConvert.SerializeObject(raw);
        if (string.IsNullOrWhiteSpace(json)) return new Dictionary<string, object?>();
        var dict = JsonSerializer.Deserialize<Dictionary<string, object?>>(json, JsonOptions.Default);
        return dict ?? new Dictionary<string, object?>();
    }
}
