---
tag: go-testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T00:19:14Z
identity: go-testing-20261101T001914-jay-flowers
tier: draft
---

When a fail-closed validator reports the first missing attribute, iterate an ordered slice of {label, value} pairs instead of a map: Go randomizes map iteration order, so the reported "missing X attribute" varies run to run and the error contract cannot be pinned by tests. vibe-check PR #75 replaced requireInts(map[string]*int) with variadic intAttr pairs in declaration order (internal/ripwireadapter/parse_deps.go) and pinned the first-error choice in the malformed-input table's wantErr column ("<deps> missing files attribute") — turning the determinism fix into a regression-tested contract. Pattern: fix the ordering, then pin the chosen first error in the test table so a future refactor back to map iteration (or a reordering) fails a test rather than silently reintroducing nondeterministic diagnostics.
