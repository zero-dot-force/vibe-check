---
tag: coverage-ratchet
author: jay-flowers
category: pattern
created_at: 2026-10-31T19:30:56Z
identity: coverage-ratchet-20261031T193056-jay-flowers
tier: draft
---

When a feature PR displaces a package coverage floor, the tighten-only ratchet forces authors to recover coverage by testing pre-existing uncovered functions in the same package — producing diff content untraceable to any spec task. In vibe-check PR #75 (ripwire-adapter), cmd/vibe-check gained TestDuplicationPercent and LCOM/duplication flag-wiring tests that no ripwire-adapter task covers, alongside the floor raise 92.7→93.1. The tests are real and behavior-asserting, and the tighten-only ratchet leaves no alternative to adding coverage, so the zero-waste cost is traceability, not dead code. Pattern: when coverage-driven test additions are unavoidable, record the driver in the PR body or a spec task note (e.g., "recovers the cmd/vibe-check floor after the registry refactor") so re-reviews can distinguish gold plating from gate compliance instead of filing scope-discipline findings.
