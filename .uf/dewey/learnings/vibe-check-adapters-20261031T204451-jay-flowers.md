---
tag: vibe-check-adapters
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:44:51Z
identity: vibe-check-adapters-20261031T204451-jay-flowers
tier: draft
---

A language adapter must never silently zero a ModuleGraph schema field it does not measure: either derive it from parsed data or apply the package's unsupported-metric disclosure convention (zero-value + machine-coded warning + doc.go sourcing-matrix entry). In vibe-check PR #75's ripwireadapter, Duplication followed the convention but TotalLines was hardcoded to 0 even though parseCCOutput had already parsed per-file Loc from cc.json; the gap propagated into assembly_test.go (an assertion rationale citing "duplication unsupported" on a TotalLines check) and both recorded goldens (reflect.DeepEqual on Modules pins the zeros, so populating the field later breaks the goldens too). Review checklist for any new metrics.Adapter: diff every ModuleResult field against (a) the adapter's capability map, (b) its warning emissions, and (c) its doc.go sourcing matrix — each field must appear in at least one; verify test rationales name the field they assert; and when populating from a tool census, filter non-source File nodes first (the ripwire cc census indexes package.json/tsconfig.json as File nodes) and disclose the tool's line-count semantics where they differ from the reference adapter's definition.
