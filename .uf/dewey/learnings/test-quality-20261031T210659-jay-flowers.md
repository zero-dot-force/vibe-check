---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:06:59Z
identity: test-quality-20261031T210659-jay-flowers
tier: draft
---

When a review thread asks to consolidate duplicated helpers, inventory what the existing tests actually pin before judging the fix's risk: if the only coverage is short-input pass-through (strings.Contains) and sentinel classification (errors.Is), the consolidation is break-safe but semantics-blind. In vibe-check PR #75 (ripwire-adapter), previewVersionOutput (cap 120, TrimSpace) and previewBytes (cap 512, no trim) duplicated the bounded-preview pattern with zero direct tests of either cap — merging to either constant, dropping trim, or dropping the "<empty>" fallback would have passed the entire suite green. Rule: a consolidation of security-relevant behavior is only as safe as the contract test that pins the merged semantics; land the extraction with a table-driven test of the shared helper (bound, truncation marker, empty-fallback, quoting, trim policy, multi-byte boundary), parameterized per call site when the duplicates served consumers with different noise profiles. Contrast within the same PR is instructive: the outer subprocess capture bounds got dedicated oversize tests (including a shrinkable maxOutputBytes field existing expressly for tests) while the inner preview caps got none — make every bounding layer testable and tested, not just the outermost.
