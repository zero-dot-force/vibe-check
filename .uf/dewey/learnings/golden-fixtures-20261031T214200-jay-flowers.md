---
tag: golden-fixtures
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:42:00Z
identity: golden-fixtures-20261031T214200-jay-flowers
tier: draft
---

An environment-variable golden-regeneration switch (e.g., WRITE_GOLDEN=1) persists silently across shells, CI environments, and stale terminals; while set, every run of the gated test rewrites the golden and returns early, so the comparison assertions never execute and a real regression passes unnoticed — the verification layer is bypassed with no audit trail. A per-invocation flag (Go's flag.Bool -update pattern, as in cmd/vibe-check/report_test.go:579) cannot persist accidentally. When a package must pick or consolidate a golden-regeneration mechanism, prefer the per-invocation flag over an env var, and keep the single-writer design: the writer test writes under the flag and returns; comparison-only tests t.Skip under the flag so exactly one test owns each golden. Complements learning/golden-fixtures-20261031T193101-jay-flowers (grep for and reuse the existing convention) with the resilience rationale for preferring the flag.
