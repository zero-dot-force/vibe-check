---
tag: subprocess-testing
author: jay-flowers
category: pattern
created_at: 2026-10-31T22:04:08Z
identity: subprocess-testing-20261031T220408-jay-flowers
tier: draft
---

When a subprocess runner bounds BOTH stdout and stderr with per-stream oversized flags and a stream-label selection in the error message (vibe-check internal/ripwireadapter/runner.go, PR #75), a single oversized-output test that floods only stdout leaves three constructs unpinned: the readBounded call on the stderr pipe, the stderrOversized flag, and the bool→label expression naming the stream. The gap compounds when the test double's default verb writes nothing to stderr (the fake's --version replays stdout-only recordings), making the stderr branch unreachable without a new double. Rules: (1) parametrize the oversized test over both streams; (2) a stderr-flooding double need not extend the committed fake — an inline probe script in the installProbeScript pattern (t.TempDir + chmod 0o755 + PATH prepend) works, or reuse an existing fail-closed fake variant whose short stderr message still crosses a shrunk bound, which additionally pins that oversize outranks exit status; (3) assert errors.Is(errOutputOversized), the stream label, and the bound as a phrase ("more than N bytes"), because a bare strings.Contains(err, digit) can pass vacuously when the error embeds a TempDir path containing that digit; (4) assert the label on the stdout case too — a label-swap regression otherwise passes the suite.
