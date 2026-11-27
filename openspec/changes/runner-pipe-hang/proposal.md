## Why

The ripwire runner (`internal/ripwireadapter/runner.go`) can block forever.
It reads the subprocess's stdout and stderr pipes until EOF before it calls
`cmd.Wait`, and EOF only arrives once every process holding a pipe's write
end has closed it. A process ripwire starts inherits both pipes. The runner's
two shutdown paths kill only the direct process:

- **Oversized output**: the stdout reader crosses the capture bound, kills
  the process, and stops reading. A surviving child stays blocked writing to
  the unread stdout pipe, so the stderr reader never sees EOF.
- **Context cancellation**: `exec.CommandContext` kills the direct process,
  but a surviving child keeps both pipes open, so `run` never returns. That
  breaks the `ripwire-adapter` scenario "Context cancellation terminates the
  subprocess": a timeout or Ctrl-C would hang the analysis.

This already fails CI-equivalent runs. `TestRunner_OversizedOutput` drives
the runner with a `/bin/sh` script whose `cat` child is exactly such a
process. It hung for the full 10 minute test timeout once on 2026-11-25, and
it hung again within 300 repeated runs (`go test -race -count=300 -run
TestRunner_OversizedOutput`), stuck in the stderr reader at
`runner.go:96`. Whether it hangs depends on whether `cat` fits the rest of the
128 KiB blob into the pipe buffer before the kill.

## What Changes

- **Capture through bounded writers and `cmd.WaitDelay`**, the standard
  library's fix for golang/go#23019, instead of reading the pipes by hand.
  The runner assigns `cmd.Stdout` and `cmd.Stderr` to per-stream bounded
  writers. A writer that crosses the bound kills the process and fails the
  write, and `os/exec` then closes that pipe, so a writing child gets EPIPE.
  With `WaitDelay` set, once the process has exited or the context is done,
  `Wait` closes any pipe a child still holds after the delay and returns.
  This works the same on Linux, macOS and Windows.
- **Fail closed on abandoned pipes**: if ripwire exits successfully but a
  child still holds its pipes after the delay (`exec.ErrWaitDelay`), the run
  is an error, not a success with possibly incomplete output.
- **Tests that hang deterministically on the old runner**: the oversize test
  writes enough output that the child can never finish, plus new cases for
  mid-flight cancellation and a successful exit that leaves a pipe-holding
  child, plus precedence pins for co-occurring failures (cancellation over
  oversize, oversize over abandoned pipes). Each is bounded by a test-level
  deadline so a regression fails fast instead of eating the package timeout.

## Capabilities

### New Capabilities

(None.)

### Modified Capabilities

- `ripwire-adapter`: the "Subprocess hygiene and data-source constraints"
  requirement gains two scenarios (shutdown with a pipe-holding child, and a
  successful exit with an abandoned pipe) and states that the runner returns
  in bounded time.

### Removed Capabilities

(None.)

## Impact

- **Modified code**: `internal/ripwireadapter/runner.go`,
  `runner_test.go`, and `version.go` — the version pin-check's runner is
  built through the defaulted constructor so it inherits the bounded pipe
  wait, and a zero `waitDelay` falls back to the default at the use site, so
  no construction path can lose the bound. No change to the runner's
  signature.
- **Unchanged**: `metrics/` (Layer 1), the schema, the output bound, error
  messages for oversize/nonzero exit/cancellation, the Go adapter, and the
  CLI. No new dependencies.
- **Behavior**: the only visible change is that `run` now returns where it
  used to hang, and that a ripwire which leaves a child holding its output
  fails the analysis instead of hanging. The pinned ripwire 0.7.0 does not
  leave children behind (its real-binary integration test returns promptly).

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS

No output format change.

### II. Composability First

**Assessment**: PASS

ripwire stays optional and PATH-discovered. Standard library only.

### III. Observable Quality

**Assessment**: PASS

A hang becomes a named error, which a caller can see and report.

### IV. Testability

**Assessment**: PASS — improves

Removes an intermittent hang from the suite and pins the shutdown paths with
deterministic, deadline-bounded tests.

### V. Security by Default

**Assessment**: PASS — improves

A hostile or broken binary can no longer hold the analysis hostage by
leaving a child on the output pipes. Every run returns in bounded time, and
abandoned pipes fail closed. No new dependencies.

### VI. Metric Fidelity

**Assessment**: N/A

No metric computation changes. A run whose output may be incomplete now
fails instead of feeding partial data to the metrics.

### VII. Language Agnosticism

**Assessment**: PASS

The change is confined to the ripwire adapter's runner. The core engine and
the adapter interface are untouched.
