---
tag: opsxadd-diff-init-provenance
author: jay-flowers
created_at: 2026-09-29T22:00:32Z
identity: opsxadd-diff-init-provenance-20260929T220032-jay-flowers
tier: draft
---

When extending a proven provenance-metadata pattern from one CLI command to sibling commands in the vibe-check codebase, the winning move is to first extract a shared helper (a producerName constant plus a newProvenanceEnvelope() constructor) and then refactor the existing command to consume it, rather than copy the inline producer/version/timestamp assembly. This makes the three JSON-emitting commands (analyze, diff, init) source their envelope from one place, eliminating the hardcoded "vibe-check" literal and semanticVersion()/time.Now() duplication. For byte-reproducible output, pair the additive omitempty pointer field with a --no-provenance flag that nils the pointer before marshaling, and guard the determinism guarantee with a byte-identical-output test across two runs — critically, the --no-provenance path must not call time.Now() at all, or the determinism test will flake. A useful cross-flag equivalence test asserts that verdict/reasons/entropyDirection/modules are identical with and without provenance, proving provenance is pure metadata and never a metric.
