---
tag: subprocess-testing
author: jay-flowers
category: pattern
created_at: 2026-10-31T22:04:25Z
identity: subprocess-testing-20261031T220425-jay-flowers
tier: draft
---

When a subprocess runner bounds output on multiple pipes through a shared read helper, one oversized-output test on a single stream verifies the helper, not the wiring — and leaves the per-stream contract unobserved: the flag propagation into the oversize condition, the stream label in the error, and that stream's bound enforcement. In vibe-check PR 75 (internal/ripwireadapter), TestRunner_OversizedOutput shrank the capture bound and drove the fake's stdout-only --version replay, so the stderr-oversized branch (readBounded(stderrPipe), the stderrOversized flag, the "stderr" label) had zero coverage even though the runner's GoDoc documents bounded capture on both streams as a memory-exhaustion defense. Rule: for every pipe a runner bounds, drive the oversized path per stream — the fake must flood that specific stream beyond the bound and beyond the OS pipe buffer (so the concurrent-drain/kill hygiene is pinned, not just the flag), stay at or below the bound on the other stream (so the correct label is selected), and the test must assert errors.Is on the oversize sentinel, the stream label, the bound value, and nil output. This mirrors the project's recorded copy-and-drift pattern: unobserved contracts are the ones that drift.
