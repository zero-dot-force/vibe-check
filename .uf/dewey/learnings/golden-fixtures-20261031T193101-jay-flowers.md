---
tag: golden-fixtures
author: jay-flowers
category: pattern
created_at: 2026-10-31T19:31:01Z
identity: golden-fixtures-20261031T193101-jay-flowers
tier: draft
---

When a Go test package already has a golden-regeneration convention, new golden tests in the same package must reuse it. cmd/vibe-check established `go test ./cmd/vibe-check/ -update` (package-level updateGoldens flag.Bool; comparison-only tests t.Skip under -update to keep a single writer — a lesson bought by an actual first-run race). PR #75's TestRunAnalyze_GoPathByteRegression introduced a parallel WRITE_GOLDEN=1 env-var mechanism, leaving two regeneration conventions in one package. Before adding any golden test, grep the package for an existing update flag or env-var pattern and adopt it; a second mechanism doubles the regeneration knowledge a contributor needs.
