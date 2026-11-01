---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-10-31T23:46:17Z
identity: documentation-20261031T234617-jay-flowers
tier: draft
---

A review-feedback fix that corrects a stale contract string in one user-facing surface must sweep every sibling documentation surface for the same literal before the feedback item is closed. Canonical case: vibe-check PR #75's fix layer corrected the ModuleGraph schema version 1.2→1.3 in the CLI help text (cmd/vibe-check/analyze.go, main.go) but left README.md's Output section (prose and example JSON both saying "1.2") and three AGENTS.md references (schema file comment, validate.go acceptance comment, Architecture Layer 1) still at 1.2 — despite metrics.SchemaVersionCurrent being "1.3" and the committed goldens emitting 1.3. The same sweep discipline applies to capability descriptions (root command updated to "Go, Python, and TypeScript" while README's opening line still said "tool for Go") and flag enumerations (--max-duplication added to the help enumeration but absent from README's Flags table). Closing a doc-accuracy feedback item should include grepping the repo for the stale literal across README, AGENTS.md, and any reference tables, not just the surface the feedback named.
