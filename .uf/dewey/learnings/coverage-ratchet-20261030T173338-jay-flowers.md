---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T17:33:38Z
identity: coverage-ratchet-20261030T173338-jay-flowers
tier: draft
---

Go coverage profiles produced under -race use mode: atomic, where any non-zero block Count means the block was executed (counts can exceed 1). Statement-weighted per-package aggregation must treat Count > 0 as covered; assuming mode: set semantics (count of exactly 1) or ignoring counts undercounts coverage. When building tooling over coverage profiles, test the atomic-mode path explicitly with a fixture carrying counts like 7 and 4 (as internal/coverratchet/testdata/atomic.out does).
