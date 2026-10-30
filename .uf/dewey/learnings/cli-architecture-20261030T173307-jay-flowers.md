---
tag: cli-architecture
author: jay-flowers
category: decision
created_at: 2026-10-30T17:33:07Z
identity: cli-architecture-20261030T173307-jay-flowers
tier: draft
---

vibe-check intentionally carries ~115 lines of near-verbatim CLI scaffolding duplicated across its two binaries (cmd/vibe-check and cmd/coverage-ratchet): exitCodeError + exitCode mapping (main.go in both), provenanceEnvelope + newProvenanceEnvelope (provenance.go in both, differing only in the producerName constant), and versionString + semanticVersion (root.go vs ratchet.go). CS-004 mandates extraction only above two locations, so the two-binary duplication is permitted, and issue #73 tracks extracting internal/climain (exit-code mapping plus a provenance constructor parameterized by producer) before a third binary is added. A new cmd/ binary should mirror the existing cmd/vibe-check scaffolding rather than inventing a competing pattern; the extraction happens once, at the third binary.
