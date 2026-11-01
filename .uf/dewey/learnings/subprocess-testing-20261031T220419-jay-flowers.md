---
tag: subprocess-testing
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:04:19Z
identity: subprocess-testing-20261031T220419-jay-flowers
tier: draft
---

When a subprocess runner implements a symmetric two-stream bounding contract — each pipe read through the same bounded reader, a per-stream oversized flag, and a stream label in the error — a single oversized-output test driven through a stdout-only fixture leaves the stderr half unpinned: a regression that drops the stderr bound or mislabels the stream passes the whole suite, silently reintroducing the full-pipe deadlock or unbounded capture the contract prevents. Parametrize the oversized-output test over both streams (a fake variant or probe that floods stderr while stdout stays silent) and assert the error names the offending stream. Verified in vibe-check PR 75 (internal/ripwireadapter): TestRunner_OversizedOutput covered only the stdout path because the fake-ripwire double's --version replays to stdout and writes nothing to stderr on success, leaving readBounded(stderrPipe), the stderrOversized flag, and the 'stderr' error-label selection untested despite the task 2.3 mandate and runner.go's documented two-pipe deadlock-avoidance rationale.
