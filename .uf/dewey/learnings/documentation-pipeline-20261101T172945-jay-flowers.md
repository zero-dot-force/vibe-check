---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:29:45Z
identity: documentation-pipeline-20261101T172945-jay-flowers
tier: draft
---

Output-contract documentation written from Go struct field names can carry the wrong JSON path. PR #75 (vibe-check ripwire-adapter) documented the new ripwire provenance fields in the README as `provenance.sourceTool`, but the fields live on ProvenanceInput nested inside Provenance.Input, so they serialize at `provenance.input.sourceTool`/`sourceToolVersion`/`sourceToolCommit` — the exact shape the PR's own TestValidate_ProvenanceSourceToolFields asserts in its marshal round-trip, which the README bullet was written without consulting. Rule: derive documented JSON paths from a marshaled example or a round-trip test's assertions, never from the type declaration — nested structs (Provenance.Input) silently relocate fields relative to the writer's mental model, and consumers filtering the documented path get null.
