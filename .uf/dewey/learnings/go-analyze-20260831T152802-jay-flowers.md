---
tag: go-analyze
author: jay-flowers
category: pattern
created_at: 2026-08-31T15:28:02Z
identity: go-analyze-20260831T152802-jay-flowers
tier: draft
---

Curator doc-gate review of vibe-check `opsx/go-analyze` (ships user-facing `vibe-check analyze` CLI, first working adapter). Key triage findings for future reviewers: (1) The content pipeline IS tracked — issues #18 (docs: README+CHANGELOG), #19 (blog), #22 (tutorial) all exist — so do NOT file duplicates; reference/refresh instead. (2) All three content issues have `labels: []`; the repo uses title prefixes (`docs:`/`blog:`/`tutorial:`) but never created the GitHub labels, so `gh issue list --label docs` returns EMPTY. Any agent following the AGENTS.md documentation gate (which searches by label) will falsely conclude no docs issue exists and file a duplicate. Fix: create+apply the labels. (3) Issue #18's README scope is STALE — it predates the CLI (from the universal-coupling-model change) and describes a LIBRARY install (`go get`), omitting `go install .../cmd/vibe-check@v0.1.0`, `vibe-check analyze`, threshold flags. CHANGELOG half of #18 is done; README half is not. (4) The competitive superlative "the Martin metrics suite that no single OSS tool computes for Go" (AGENTS.md:9, proposal Why) is unsubstantiated (rivals: goda, go-arch-lint) and will propagate into README/blog — Envoy previously flagged it as "HIGH the moment it's published". (5) go-analyze tasks.md Section 9 has NO README task — README deferred entirely to #18. Pattern: when a repo tracks content work by title-prefix instead of labels, the label-based doc-gate search silently breaks; verify labels exist AND that stale tracking issues cover the CURRENT user-facing surface, not a prior change's.
