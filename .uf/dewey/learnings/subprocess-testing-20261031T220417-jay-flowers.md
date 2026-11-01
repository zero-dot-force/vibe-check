---
tag: subprocess-testing
author: jay-flowers
category: pattern
created_at: 2026-10-31T22:04:17Z
identity: subprocess-testing-20261031T220417-jay-flowers
tier: draft
---

When a subprocess runner bounds multiple output streams symmetrically (one bounded reader per pipe, per-stream oversized flags, and a stream label in the surfaced error), the oversized-path tests must drive EACH stream across the bound — code that looks covered by symmetry is not: the sibling branch, its flag, and its label selection ship unverified. In vibe-check's internal/ripwireadapter (PR #75), readBounded ran on both stdout and stderr, but every fake-riprewire --version variant wrote success output to stdout only, so TestRunner_OversizedOutput could cross only the stdout bound; the stderr readBounded call, the stderrOversized flag, and the "stderr" label selection had zero coverage. Rules: (1) for every per-stream branch of a symmetric contract, give the test double a variant that floods that stream (and only that stream) beyond the bound; (2) assert the stream LABEL in the error, not just errors.Is(sentinel) and the bound value — the label-selection expression (e.g. map[bool]string{true: "stdout", false: "stderr"}[stdoutOversized]) regresses silently otherwise; (3) installing the flooding variant through an env-sanitizing runner requires baking the double's control variables into a PATH wrapper script, since the sanitized environment cannot carry them.
