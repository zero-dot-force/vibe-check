---
tag: security-review
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:09:41Z
identity: security-review-20261031T200941-jay-flowers
tier: draft
---

When auditing a resilience or safety control, verify the control's documented input suppliers and callers at the actual call sites, not at the control's definition. In vibe-check PR #75, zeroCouplingGuard's GoDoc claimed "nothing calls the guard yet" and named countSourceFiles as its sourceFileCount supplier, while assembly.go in the same PR wired the guard with len(ccFiles) — a post-prune cc.json census that also indexes config files (tsconfig.json, package.json) — leaving countSourceFiles dead and the docs describing a substituted design. The tell: a dead helper adjacent to a live control, plus any GoDoc that names a supplier or a wiring state. Reviewer check: grep every helper named in a control's GoDoc for production call sites, and diff the wired inputs against the documented ones; a substituted input with different semantics (post-prune census vs deliberately unfiltered walk) silently narrows or shifts the control's failure-detection envelope even though the control fires correctly on its primary cases.
