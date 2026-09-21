---
tag: emit-provenance-metadata
author: jay-flowers
category: gotcha
created_at: 2026-09-21T18:26:20Z
identity: emit-provenance-metadata-20260921T182620-jay-flowers
tier: draft
---

When a feature introduces both an embedded JSON Schema and a hand-rolled Go validator as dual sources of truth (as vibe-check's metrics package does), required-field contracts must be kept in lockstep. In the emit-provenance-metadata change, the JSON Schema declared "required": ["producer","version","generatedAt","input"] on the provenance object, but metrics/validate.go's validateProvenance only required "input", treating producer/version/generatedAt as optional-if-present. This drift slipped through code review's first pass because the existing tests covered only the two endpoints — no provenance at all (omitempty) and full provenance — but never the partial-provenance middle case. A graph with only "input" passed metrics.Validate but failed JSON-Schema validation, violating the design's own R2 risk note. The fix required both the validator change AND a negative test for the partial case. Lesson: whenever a schema and a validator share a required-fields contract, add a negative test for the partial/missing-required-fields case explicitly, and assert the validator's required set matches the schema's required array.
