---
tag: structural-entropy
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:30:53Z
identity: structural-entropy-20261031T193053-jay-flowers
tier: draft
---

Entropy-divisor dispatch runs in this repo find the working tree checked out on main (HEAD == origin/main == merge-base of everything), so the documented default base `git merge-base HEAD origin/main` resolves to main's tip and diffs the wrong range entirely. Anchor instead on the dispatch context's base_sha/head_sha: validate the base ref (regex gate, git check-ref-format, rev-parse --verify --end-of-options), verify both SHAs resolve locally, confirm merge-base(head_sha, main) == base_sha (proving three-dot scope equals PR-diff scope), then materialize BOTH refs as detached git worktrees outside the repo tree and analyze both with the same vibe-check binary. Verified on PR #75 (opsx/ripwire-adapter, 2026-10-31): base 814053c, head aeca31c, local main 9ccb82d6, merge-base(head, main) = 814053c = PR base_sha; the dual-worktree measurement reproduced CI's structural-gate PASS exactly (verdict COMMENT, entropyDirection stable, unreliable false).
