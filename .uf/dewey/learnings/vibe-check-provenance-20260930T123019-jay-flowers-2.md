---
tag: vibe-check-provenance
author: jay-flowers
created_at: 2026-09-30T12:30:19Z
identity: vibe-check-provenance-20260930T123019-jay-flowers-2
tier: draft
---

The vibe-check provenance envelope has a field-shape asymmetry that is easy to mis-document: `analyze --json` emits four fields (`producer`, `version`, `generatedAt`, and an `input` sub-object with `path`/`modulePath`) via `metrics.Provenance` in metrics/graph.go, while `diff --json` and `init --json` emit only the three shared fields (`producer`, `version`, `generatedAt`) via the smaller `provenanceEnvelope` struct in cmd/vibe-check/provenance.go (producerName const = "vibe-check", generatedAt = RFC 3339 UTC). When writing specs or README prose that must state this shape, pin the distinction with both a positive enumeration of the three shared fields AND an explicit negative assertion ("neither diff nor init describes an `input` field, which is analyze-only"). Avoid the words "identical" or "the same object" which invite conflation and caused a HIGH review finding.
