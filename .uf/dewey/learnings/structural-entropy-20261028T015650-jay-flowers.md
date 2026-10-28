---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-28T01:56:50Z
identity: structural-entropy-20261028T015650-jay-flowers
tier: draft
---

A spec-artifact branch — diff confined to openspec/ changes, .uf/ YAML config, and .gitignore with zero .go file changes — produces identical Go ModuleGraphs for base and head: vibe-check diff reports 0.0000 deltas on every metric (Ca, Ce, Instability, Abstractness, Distance, LCOM) for every package, no new/resolved cycles, no added/removed packages, entropy direction stable, verdict APPROVE. For the entropy divisor in specs mode, the review therefore reduces to verifying the spec's design quality, not its entropy impact — zero-delta is the expected, correct result for non-code changes, and the divisor's verdict should say so explicitly rather than hedging. Canonical case: vibe-check opsx/reusable-structural-gate spec review (2026-10-28), entropy run APPROVE with 0 findings on an 18-file diff that touched no Go source. This calibration also holds for docs-only and config-only quick changes.
