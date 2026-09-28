See @AGENTS.md for guidance on working on the go-ios codebase (build, test, and
the dispatch-only release process).

## Agent-friendly CLI behavior

go-ios started as a CLI for humans, but agents now drive it heavily. When adding
or changing commands, preserve the human experience while making the command easy
to call from scripts and coding agents.

- Keep every operation available through flags and arguments. Do not require
  interactive prompts, terminal UIs, pagers, or keypresses for normal automation.
  If confirmation is needed, provide an explicit flag such as `--yes` or
  `--force`, and fail fast in non-interactive contexts when it is missing.
- Prefer structured output for machine consumption. Where a command returns
  data, support stable JSON output and keep progress, diagnostics, and logs off
  stdout so `ios ... --json | jq ...` remains parseable.
- Make stdout and stderr predictable: stdout is for the requested result, stderr
  is for diagnostics, warnings, and human progress. Avoid banners, spinners, and
  mixed log lines in machine-readable modes.
- Use clear exit semantics. Return zero only when the requested operation was
  completed or the desired state already existed. Return non-zero for validation
  errors, device connection failures, timeouts, and command failures; include a
  concise, actionable stderr message.
- Make retries safe. Commands should be idempotent where practical, and errors
  should distinguish "already done", "not found", "device unavailable",
  "permission denied", and "timed out" instead of collapsing them into vague
  failures.
- Make commands discoverable from the terminal. Keep `ios help`, subcommand
  help, examples, and flag descriptions accurate enough that an agent can learn
  the command without reading source code.
- Keep output bounded and useful. Avoid unbounded dumps by default; provide
  filtering, selectors, or explicit verbose/debug flags for large payloads.
- Do not break existing human-readable output casually. If a command already has
  established text output, add machine-readable behavior behind an explicit flag
  unless the command already has a project convention for JSON output.
