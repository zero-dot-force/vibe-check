---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:58Z
identity: test-quality-20261101T172958-jay-flowers
tier: draft
---

A prose pin on a multi-clause disclosure warning is complete only when every load-bearing clause has its own assertion. In vibe-check PR #75 (head 6a5e15e), the cycle-approximation warning's prose remediation pinned the corroboration-verdict clause ("does not corroborate") but left the numeric caps clause ("20 SCCs × 12 members") — the rationale the spec scenario requires — unguarded, even though the constants (ripwireCycleMaxSCCs/ripwireCycleMaxMembers) are interpolated into the message via fmt.Sprintf and the analogous error-budget numbers (1.9%/2.2%/0.16%) are pinned in budget_test.go. Constants interpolated into warning text are unpinned disclosure unless a test asserts the rendered substring or the constant values directly (the TestDenylistNamesMatchRipwire pattern). Reviewer checklist: per warning code, enumerate the numeric and identifying clauses of the emitted message and grep for a pin of each.
