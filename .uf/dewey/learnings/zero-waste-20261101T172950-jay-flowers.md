---
tag: zero-waste
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:50Z
identity: zero-waste-20261101T172950-jay-flowers
tier: draft
---

Final-assembly zero-waste sweeps must grep every derived field's readers, not just every helper's callers. In vibe-check's ripwireadapter (PR #75), the sweep that fixed first-generation orphans (countSourceFiles wiring, cycle-cap constants, checkRipwirePin removal) missed a second-generation instance: mappedModule.InternalOut/InternalIn/ExternalOut — derived counters incremented in production but read only by tests, whose GoDoc still claimed they "feed cycle derivation and the Ce-only external accounting" while deriveCycles actually consumed mapping.Edges and Ca/Ce came verbatim from the arch table. Pattern: superseded design decisions leave derived state behind. When a field's value is computed from data that a later decision made authoritative elsewhere, grep the field name across non-test files before marking the assembly task complete — and treat any GoDoc naming a consumer as a claim to verify, not documentation to trust.
