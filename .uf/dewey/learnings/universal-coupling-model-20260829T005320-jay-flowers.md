---
tag: universal-coupling-model
author: jay-flowers
category: pattern
created_at: 2026-08-29T00:53:20Z
identity: universal-coupling-model-20260829T005320-jay-flowers
tier: draft
---

When implementing a universal coupling metrics model for Go codebases (vibe-check project), the two-layer architecture pattern (Layer 1 = universal model, Layer 2 = language adapters) proved effective. The key design decisions that worked well: (1) flat metrics package rather than internal/ because the model IS the public API, (2) named metric types like Instability float64 with GoDoc documenting ranges and citations, (3) Module as the universal unit of analysis with language-neutral terminology, (4) JSON-RPC 2.0 over stdin/stdout for external adapters with newline-delimited framing. The ExternalAdapter's functional options pattern (WithAnalyzeTimeout, WithMaxResponseSize, etc.) provided clean configuration without breaking constructor signatures, though it deviates from the AP-001 Options struct convention — this deviation should be documented as a custom rule.
