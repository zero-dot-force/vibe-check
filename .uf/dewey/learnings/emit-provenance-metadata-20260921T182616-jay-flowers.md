---
tag: emit-provenance-metadata
author: jay-flowers
category: pattern
created_at: 2026-09-21T18:26:16Z
identity: emit-provenance-metadata-20260921T182616-jay-flowers
tier: draft
---

To add provenance metadata to a language-agnostic analysis pipeline without breaking the adapter contract, split ownership across layers: the language adapter (Layer 2) populates only the Input identity fields (the analyzed path and the resolved module path, since only the adapter knows how to resolve module identity), while the CLI (Layer 3) populates the build-identity fields (producer constant, semantic version, RFC3339 timestamp). This keeps the metrics.Adapter interface unchanged — a core project invariant that adding a new language must not require core engine changes — and lets a single adapter serve multiple CLI frontends. The layer that owns a field is the layer with the information necessary to fill it. For byte-reproducible output, expose a --no-provenance flag that nils the pointer before marshaling, and verify determinism with a byte-identical-output test across two runs.
