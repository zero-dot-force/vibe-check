---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:30:00Z
identity: test-quality-20261101T173000-jay-flowers
tier: draft
---

Precedence chains implemented as first-match-wins ordered slices need adjacent-pair tests, not one representative pair. vibe-check's detectLanguage (cmd/vibe-check/detect.go) implements the spec'd precedence go.mod → Python markers → tsconfig.json as an ordered markers slice; TestDetectLanguage_MarkerPriority pins only go-beats-tsconfig, so reordering the slice (hoisting tsconfig.json above the Python markers) passes every test while flipping mixed Python/TypeScript monorepos to the TypeScript adapter. For any ordered selection list, exercise each adjacent pair — or assert the complete order with a multi-marker tree — so slice reordering is visible to the suite.
