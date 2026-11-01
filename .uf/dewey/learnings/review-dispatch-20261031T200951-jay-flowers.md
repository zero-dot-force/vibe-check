---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:09:51Z
identity: review-dispatch-20261031T200951-jay-flowers
tier: draft
---

Entropy-divisor operational trap in address-feedback dispatches (hit on PR #75): the review session's checkout was on main (9ccb82d6), ahead of BOTH the PR base (814053c) and the frozen PR head (aeca31cf) — analyzing "the PR checkout" per the naive code-review-mode recipe would have measured the wrong tree (main carrying post-PR commits). Recipe: before measuring, resolve HEAD, the PR branch tip, and the frozen head SHA (git rev-parse --verify --end-of-options "<ref>^{commit}"), and confirm merge-base(head, main) equals the caller-provided base_sha; when HEAD ≠ frozen head, materialize BOTH base and head as detached git worktrees outside the repo tree, analyze both with the same vibe-check binary via --output, diff the two graphs, then remove both worktrees unconditionally.
