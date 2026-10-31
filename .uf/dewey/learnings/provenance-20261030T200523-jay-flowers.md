---
tag: provenance
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:05:23Z
identity: provenance-20261030T200523-jay-flowers
tier: draft
---

metrics.Validate is a hand-rolled required-fields validator, not JSON-Schema enforcement: validateProvenance requires provenance.input.path and modulePath to be non-empty strings whenever provenance is present, but tolerates unknown extra fields inside input — while modulegraph.schema.json sets additionalProperties:false there. Two consequences: (1) new language adapters must define non-Go modulePath semantics or validation fails on every provenance-enabled graph; (2) optional provenance-input fields can ride the Go struct without schema evolution (the CommitSHA/Branch precedent from architectural-drift-tracking), at the cost of external schema validators rejecting them.
