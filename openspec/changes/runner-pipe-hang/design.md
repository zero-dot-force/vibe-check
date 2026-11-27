## Context

`ripwireRunner.run` starts ripwire with `StdoutPipe`/`StderrPipe`, drains both
in goroutines with a bounded reader, and calls `cmd.Wait` only after both
readers hit EOF. See `proposal.md` for how a child process that inherits the
pipes turns the oversize and cancellation paths into hangs, and for the
reproduction.

## Goals / Non-Goals

**Goals**

- `run` returns in bounded time on every path: normal exit, nonzero exit,
  oversize, and cancellation, whatever the subprocess's children do.
- Keep the existing contract: sanitized env, working directory, per-stream
  16 MiB bound with no truncation, cancellation classified with
  `errors.Is(err, context.Canceled)`, bounded stderr preview.

**Non-Goals**

- Killing ripwire's descendants. That needs process groups on Unix and job
  objects on Windows. Closing the pipes is enough to unblock `run`, and a
  child that writes afterwards gets EPIPE. A child that never writes again
  can outlive the run, but it no longer holds the analysis hostage.
- Changing the runner's signature, error text, or the output bound.

## Decisions

### D1: Bounded writers plus `cmd.WaitDelay`, not hand-rolled pipe reads

`os/exec` already solves this (golang/go#23019, Go 1.20). When `cmd.Stdout`
is a non-`*os.File` writer, `os/exec` copies from its own pipe in a
goroutine, and closes that pipe if the writer returns an error. With
`WaitDelay` set, the delay starts when the context is done or `Wait` sees
the process exit. When it elapses, `Wait` kills the process if it is still
running, closes any pipe still open, and returns.

The runner sets `cmd.Stdout`/`cmd.Stderr` to a `boundedWriter` per stream.
Once a write would push the buffer past the bound, the writer marks itself
oversized, kills the process (shared `sync.Once`), and returns
`errOutputOversized` without buffering the excess.

Rejected: closing our own read ends on kill. Closing an `*os.File` interrupts
a blocked `Read` on Unix pipes, which use the poller, but not reliably on
Windows. vibe-check ships Windows binaries.

Rejected: killing the process group (`Setpgid` + `kill(-pgid)`). It is Unix
only, so it needs build tags and a separate Windows design. It also isn't
needed for the goal, which is that `run` returns.

### D2: `WaitDelay` is a runner field, default 2 seconds

Two seconds is enough for an exiting ripwire to flush its pipes, and it is
small next to a typical analysis. It is a field next to `maxOutputBytes`, so
tests can shrink it. Production uses the default. A zero or negative value
falls back to the default at the use site, so a runner built without the
constructor — a struct literal — cannot silently lose the bounded wait.

### D3: Error precedence is unchanged, and `ErrWaitDelay` is an error

The runner classifies results in the same order as today: cancellation,
then oversize, then exit status. A new last rule is added: a successful
exit that still returns `exec.ErrWaitDelay` becomes an error naming
abandoned pipes. Output collected before the delay may be incomplete, and
the adapter fails closed on incomplete input everywhere else.

When a writer kills the process, `Wait` reports the kill signal as the exit
error. That case stays classified as oversize, as it is today.

## Risks / Trade-offs

- **A real ripwire that daemonizes would now fail instead of hanging.**
  That's the intended direction. 0.7.0 doesn't do this, and its integration
  test returns promptly.
- **A child that never writes again survives the run.** It is documented as
  a non-goal (see above). The tests clean up their own children.
