---
tag: adapter-design
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:44:46Z
identity: adapter-design-20261031T204446-jay-flowers
tier: draft
---

When a hybrid-sourcing adapter defines its metric sourcing matrix (spec design doc, package doc.go, README capability table), the matrix must enumerate every field of the output schema the adapter emits — not only the computed metrics. Schema fields absent from the matrix fall through assembly as silent zero-valued placeholders: no warning, no doc entry, and tests that pin the zero with rationales borrowed from a neighboring unsupported metric. Vibe-check PR #75 (ripwire adapter): the matrix covered Ca/Ce/Instability, Abstractness, LCOM4, Distance/Zone, Cycles, and Duplication-as-unsupported, but ModuleResult.TotalLines was omitted everywhere — assembly hardcoded 0 even though per-file line counts were already parsed (cc.json Loc, used only as a guard census), and the structural-invariants test pinned 0 with the rationale "(duplication unsupported)", conflating the size field with the duplication metric. The zero was schema-legal (the field's GoDoc defines 0 as "could not be determined"), so validation and CI never caught it; only the missing disclosure decision was wrong. Reviewer checklist for any metrics.Adapter implementation: diff the emitted output struct's fields (ModuleResult, ModuleGraph) against the adapter's doc.go sourcing matrix and its warning codes — every field needs a source, an unsupported-metric disclosure, or an explicit documented omission; and a test pinning a zero-valued field must name that field's own reason, never a neighboring metric's.
