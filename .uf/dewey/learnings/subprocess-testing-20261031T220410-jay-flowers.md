---
tag: subprocess-testing
author: jay-flowers
category: pattern
created_at: 2026-10-31T22:04:10Z
identity: subprocess-testing-20261031T220410-jay-flowers
tier: draft
---

Symmetric two-stream subprocess contracts need per-stream oversize tests, and the bound's GoDoc must name every stream it bounds. In vibe-check PR 75's ripwireadapter runner, TestRunner_OversizedOutput shrank the unexported maxOutputBytes to 8 bytes but drove only the fake's --version — a stdout-only replay — so the stderr readBounded call, the stderrOversized flag, and the stdout/stderr label in the errOutputOversized message went unverified, while the bound's field/const/run-contract GoDoc described a stdout-only bound even though readBounded applies it to both pipes (the type doc in the same file correctly promised two-stream bounding — the docs drifted within one file). Structural amplifier: when the bound is an unexported field with no setter, ONLY in-package tests can shrink it, so CLI/external tests can never close the gap. Pattern: table-drive the oversize test across both streams with a tiny per-stream probe script (cheaper than adding a fake-harness variant, and it leaves the fake's fail-closed variant contract untouched), assert the stream label ('bytes of stdout'/'bytes of stderr') rather than just the sentinel and the bound digit — the label is the only observable that distinguishes the branches — and keep every doc site's stream scope identical to the implementation's.
