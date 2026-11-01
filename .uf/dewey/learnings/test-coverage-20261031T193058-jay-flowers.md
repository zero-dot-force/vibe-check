---
tag: test-coverage
author: jay-flowers
category: pattern
created_at: 2026-10-31T19:30:58Z
identity: test-coverage-20261031T193058-jay-flowers
tier: draft
---

Spec-mandated guards need an end-to-end test through the production wiring, not just decision-table unit tests of the pure predicate. vibe-check's ripwire adapter zero-coupling guard shipped with a 4-case decision table plus dedicated tests for countSourceFiles — the census guard.go documents as the supplier — but production assembly passes len(ccFiles) (the cc.json export, which indexes config files and excludes denylist-pruned trees) instead; countSourceFiles has no production caller, and no test drives Analyze on a zero-edge corpus, so deleting the one-line guardWarnings append passes the entire suite (found in PR #75 review, 2026-10-31). Pattern: for every MUST-level guard, add one integration test that constructs the firing condition through the public entry point and asserts the warning code plus partial status in the produced graph; when the call-site input differs from the documented supplier, delete or rewire the dead supplier and its tests rather than leaving them to pin unused code and a stale wiring doc.
