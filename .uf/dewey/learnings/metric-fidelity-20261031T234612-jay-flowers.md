---
tag: metric-fidelity
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:46:12Z
identity: metric-fidelity-20261031T234612-jay-flowers
tier: draft
---

When an adapter change redefines the zero-value semantics of a shared metric field (from "not determined" sentinel to measured value), the universal type's GoDoc contract must be updated in the same change. In vibe-check PR #75's feedback round, TotalLines was newly populated from cc.json loc sums and the golden ledger explicitly declared that a config-only module's 0 is "the honest count of its source lines, not the 'not determined' sentinel" — yet metrics/graph.go's ModuleResult.TotalLines GoDoc still reads "Zero indicates the line count could not be determined." Per-adapter disclosures (the adapter's doc.go sourcing matrix) do not reach consumers who read the metrics type contract, so two authoritative documents now disagree about what zero means. Rule: a semantic change to a field's zero value is a contract change; update the type-level GoDoc (and schema description if one exists) in the same commit, and let the golden notes reference the updated contract rather than contradicting it.
