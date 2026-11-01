---
tag: protected-values
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:44Z
identity: protected-values-20261101T143244-jay-flowers
tier: draft
---

When a change redefines the semantics of a schema-serialized field but the JSON Schema file is a declared protected value the change must not touch, the field's schema description becomes a silently-stale second contract surface: the Go struct GoDoc and the schema description agreed before the change and contradict each other after it, and external JSON validators read the stale side. Verified in vibe-check PR #75 round 3: TotalLines's zero was redefined from a "could not be determined" sentinel to a measured count in metrics/graph.go's GoDoc, while metrics/modulegraph.schema.json's totalLines description kept the sentinel sentence — and none of the disclosure surfaces updated in the fix (GoDoc, doc.go sourcing matrix, README, CHANGELOG, golden ledger) recorded the lag. Rule: a semantic change to any schema-serialized field must diff BOTH contract surfaces (GoDoc and schema description) even when only one is editable; the protected one's divergence is then recorded as a known trade-off with a tracked follow-up change (the D11/R5 precedent), never left implicit.
