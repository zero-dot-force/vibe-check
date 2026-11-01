---
tag: golden-fixtures
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:34:04Z
identity: golden-fixtures-20261031T213404-jay-flowers
tier: draft
---

A two-run byte-equality determinism test on a module graph proves warning prose is stable, not correct: a silently gutted or hollowed-out disclosure passes determinism checks unchanged. In vibe-check's ripwireadapter (PR #75, head aeca31c), TestIntegration_Determinism pinned run-to-run byte equality of the full JSON while the golden comparison reduced Warnings to code-only lists and both fixture goldens stored message=code placeholders — leaving four disclosure warnings (directory-as-module, name-resolved-edges, ambiguity-rate, unsupported-metric) with no content net anywhere in the package. Pattern for warnings-as-disclosure adapters: the golden layer pins the machine contract (codes, sorted, non-nil, status derivation); each warning's load-bearing prose gets a substring assertion at its emitting unit — numbers/paths/listings for dynamic messages, one semantic substring per static disclosure code. Flake safety is verifiable up front: message construction there was deterministic (sorted listings, constant budgets, canonical cycle ordering), so substring pins are byte-stable by construction. Reviewer shortcut when assessing "is the prose pinned anywhere": inventory assertions per warning CODE, not per test file — grep for the code constant and for .Message substring checks together; a code-only golden comparison plus message=code placeholder goldens means no layer pins prose even though the golden files visually contain a message field.
