<!--
  [P] marks tasks eligible for parallel execution.
-->

## 1. Failing tests first

- [x] 1.1 Add a test-level deadline helper to `runner_test.go` that runs `runner.run` in a goroutine and fails the test (with a message naming the hang) if it does not return within a bound. Verify: the helper is used by every test in 1.2–1.4.
- [x] 1.2 Make `TestRunner_OversizedOutput` deterministic: the flood script backgrounds its writer (`cat blob >&N & wait`) so a child holds both pipes, and the blob is 1 MiB so the child can never drain into the pipe buffer. Verify: on the unchanged runner the stdout case fails by deadline every run.
- [x] 1.3 Add `TestRunner_CancelWithPipeHoldingChild`: a script that backgrounds `sleep` holding both pipes and waits; cancel the context mid-flight. Assert `errors.Is(err, context.Canceled)`, nil stdout, and return within the deadline. The script records the child PID and the test kills it in cleanup. Verify: fails by deadline on the unchanged runner.
- [x] 1.4 Add `TestRunner_SuccessfulExitWithAbandonedPipes`: a script that backgrounds `sleep` holding both pipes, prints output, and exits 0. Assert an error, nil stdout, and return within the deadline; kill the child in cleanup. Verify: fails by deadline on the unchanged runner.

## 2. Runner fix (design D1–D3)

- [x] 2.1 Replace the `StdoutPipe`/`StderrPipe` readers in `runner.go` with a per-stream `boundedWriter` assigned to `cmd.Stdout`/`cmd.Stderr`; on crossing the bound it marks oversized, kills the process once, and returns `errOutputOversized` without buffering the excess. Verify: 1.2 passes.
- [x] 2.2 Add a `waitDelay` field (default 2s, D2) set as `cmd.WaitDelay`; keep the existing precedence (cancellation, oversize, exit status) and add `exec.ErrWaitDelay` on a successful exit as an error naming abandoned pipes (D3). Update GoDoc on `ripwireRunner` and `run`. Verify: 1.3 and 1.4 pass, and every existing runner test passes unchanged.
- [x] 2.3 Stress the fix: `go test -race -count=300 -timeout 300s -run 'TestRunner_(OversizedOutput|CancelledContext|NonzeroExit|EnvironmentAndWorkingDirectory)$'`, plus `-count=20` of the two wait-delay tests (`CancelWithPipeHoldingChild`, `SuccessfulExitWithAbandonedPipes`, about 2s each by design). Verify: passes with no hang and leaves no `sleep`/`cat` children running.

## 3. Gates and docs

- [x] 3.1 Run the CI commands from `.github/workflows/ci.yml` (build, vet, `go test -race -count=1` with coverage, coverage ratchet, golangci-lint at the CI-pinned version). Verify: all pass.
- [x] 3.2 Add a CHANGELOG Unreleased → Fixed entry. Assess AGENTS.md (runner.go line) and README for impact. Verify: entry present; other docs updated only if they describe the changed behavior.
- [x] 3.3 Run the review council and resolve any REQUEST CHANGES. Verify: all reviewers APPROVE.
