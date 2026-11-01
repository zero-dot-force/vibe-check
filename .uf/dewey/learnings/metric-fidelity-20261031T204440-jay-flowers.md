---
tag: metric-fidelity
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:44:40Z
identity: metric-fidelity-20261031T204440-jay-flowers
tier: draft
---

When an adapter hardcodes a schema field to zero, audit the field's downstream consumers and the adapter's own disclosure convention before treating the zero as inert. vibe-check PR #75: ripwireadapter hardcoded ModuleResult.TotalLines=0 (implicitly reasoned as "duplication unsupported"), but metrics/report uses TotalLines as the module-size sort comparator (render.go:582) and sums it into the graph total-lines aggregate (render.go:801), so every ripwire-backed graph reports 0 total lines and a degenerate size ranking in `vibe-check report` — silently, because the unsupported-metric warning text names only duplication. The Go adapter (parity reference) populates the same field from real data (goadapter/adapter.go countTotalLines), and ripwire's cc.json already carried the per-file Loc that would have populated it. Rule: a zero-valued measured field is an undocumented approximation (Constitution VI). Either populate it from already-parsed data (documenting the sourcing semantics in the doc.go matrix) or run the full disclosure triple — warning text naming the field, doc.go sourcing-matrix entry, README caveat row — and never let a test rationale attribute the zero to the wrong feature (assembly_test.go pinned TotalLines==0 with "(duplication unsupported)").
