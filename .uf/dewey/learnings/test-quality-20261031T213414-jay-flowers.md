---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:34:14Z
identity: test-quality-20261031T213414-jay-flowers
tier: draft
---

Assessing "warning prose can drift silently" feedback requires a per-code prose inventory before accepting the severity rationale — the cited at-risk contents may already be pinned. vibe-check PR #75 (ripwire adapter): the golden comparison projected warnings to codes (assembly_test.go) and golden-notes.md C5 documented codes-not-messages as deliberate, justified by "task 4.5 asserts warning text separately"; verification showed task 4.5's tests pin only the error-budget numbers, while denylist listings (denylist_test.go), parse-error paths (degrade_test.go), and deps-window counts (TestDepsWindowWarning) were independently substring-pinned — the actually-unpinned prose was four STATIC disclosure sentences (directory-as-module, name-resolved-edges, ambiguity-rate, unsupported-metric) with no interpolated values. Rules: (1) a golden ledger's forward-referenced justification ("task N asserts X separately") covers exactly its named scope — read the referenced test before treating coverage as present or absent in either direction; (2) prose-pinning granularity should match message shape: exact-equality for static constant messages (no re-record coupling, no false churn; intentional rewording SHOULD fail a test because the wording is the disclosure), substring pins on load-bearing segments (numbers, names, paths, counts) for interpolated messages, and never full-text goldens for interpolated prose — that couples the golden re-record workflow to wording.
