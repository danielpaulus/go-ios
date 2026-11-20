namespace GoIos.Examples;

/// <summary>
/// Outcome of a single example. An example either succeeds (<see cref="Ok"/>) or
/// is skipped for a documented, non-fatal reason (<see cref="Skip"/>, e.g. no
/// device attached, or a WDA prerequisite is not met). A thrown exception is a
/// hard failure and is handled by the runner (non-zero exit).
/// </summary>
public readonly record struct ExampleResult(bool Skipped, string? Reason)
{
    /// <summary>The example ran to completion.</summary>
    public static ExampleResult Ok(string? note = null) => new(false, note);

    /// <summary>The example was skipped for the given (non-fatal) reason.</summary>
    public static ExampleResult Skip(string reason) => new(true, reason);
}
