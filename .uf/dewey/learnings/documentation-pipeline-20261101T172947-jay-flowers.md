---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:47Z
identity: documentation-pipeline-20261101T172947-jay-flowers
tier: draft
---

Protected-file deferrals need the same issue trail as review-caught gaps. When a design deliberately leaves a published contract surface unaligned with new output fields (vibe-check PR #75 design D11/R5: provenance sourceTool fields added to the Go struct while the protected modulegraph.schema.json's provenance.input stays additionalProperties:false and therefore rejects them), the deferred formalization needs an owning follow-up issue at PR time. The repo's own precedent set the bar: #85 was filed for the lesser totalLines description drift in the same schema file, while the normative provenance-field rejection got no issue. User-facing docs that advertise the fields (the README provenance bullet) satisfy the design's "if external schema validation becomes a real consumer" condition by construction. Symmetric rule: every declared in-PR trade-off against a protected file gets an owning issue, or the post-PR alignment is unowned.
