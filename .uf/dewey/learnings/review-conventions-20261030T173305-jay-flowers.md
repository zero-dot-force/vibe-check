---
tag: review-conventions
author: jay-flowers
category: gotcha
created_at: 2026-10-30T17:33:05Z
identity: review-conventions-20261030T173305-jay-flowers
tier: draft
---

In vibe-check, convention-pack rule AD-007 (no source file exceeds 400 lines) is applied to production source files only. Committed test files routinely exceed it: metrics/report/render_test.go (1737 lines), cmd/vibe-check/analyze_test.go (1689), cmd/vibe-check/diff_test.go (1627), metrics/report/report_test.go (1243), metrics/external_test.go (793), cmd/coverage-ratchet/ratchet_test.go (714), internal/goadapter/adapter_test.go (482). Architect reviewers must not flag test-file length against AD-007 — that is a false MEDIUM against established repo practice.
