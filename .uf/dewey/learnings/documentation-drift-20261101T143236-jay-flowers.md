---
tag: documentation-drift
author: jay-flowers
category: pattern
created_at: 2026-11-01T14:32:36Z
identity: documentation-drift-20261101T143236-jay-flowers
tier: draft
---

When a change redefines a schema field's zero-value semantics, sweep every parallel prose description of that field, not only the code-adjacent docs. A JSON Schema's description annotation is a published contract surface that duplicates field semantics verbatim and drifts silently: goldens pin values, schema validation pins constraints, but nothing pins description prose. Enumerate the full surface set before editing: struct-field GoDoc, JSON Schema property description, adapter doc.go sourcing matrix, README capability table, CHANGELOG entry. Vibe-check PR #75 round 3: metrics/graph.go redefined ModuleResult.TotalLines zero as a measured count ("not a 'could not be determined' sentinel"), but metrics/modulegraph.schema.json's totalLines description still reads "Zero indicates the line count could not be determined" — four of five surfaces updated, the fifth left contradicting the others.
