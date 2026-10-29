---
tag: review-conventions
author: jay-flowers
category: context
created_at: 2026-10-29T15:53:59Z
identity: review-conventions-20261029T155359-jay-flowers
tier: draft
---

In the vibe-check repository, convention-pack rule AD-007's 400-line file-size threshold applies to production source files only; test files follow a separate, larger norm. Existing *_test.go files on main range from ~450 to ~1737 lines (cmd/vibe-check/analyze_test.go 1689, metrics/report/render_test.go 1737, cmd/vibe-check/diff_test.go 1627, metrics/validate_test.go 743) while production files stay under 400. Reviewers should not flag test-file length against AD-007 — flag production files only, or the finding contradicts established repo convention.
