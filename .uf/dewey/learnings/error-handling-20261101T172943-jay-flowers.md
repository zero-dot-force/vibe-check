---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:29:43Z
identity: error-handling-20261101T172943-jay-flowers
tier: draft
---

When a review remediation establishes a deterministic-error convention for wire-format parsing, sweep every parser in the package for the pattern it replaces. In vibe-check PR #75 (ripwireadapter), the deterministic-errors fix landed as the ordered intAttr/requireInts helper in parse_deps.go — whose doc comment explicitly says map iteration picks an arbitrary missing-attribute label — while parse_arch.go's aggregate-attribute check kept a map[string]bool range, preserving the exact non-determinism the fix targeted, one file away. Consolidation checklist: after introducing an ordered-validation helper, grep the package for `range` over diagnostic maps and for sibling fail-closed attribute checks; a convention that lives in one helper but not its sibling call sites is half-landed and will resurface in the next review round.
