---
tag: entropy-divisor
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:49Z
identity: entropy-divisor-20261101T002549-jay-flowers
tier: draft
---

In uncommitted-local review mode (input_context reports head_sha == base_sha with uncommitted: true), the entropy divisor MUST analyze the PR side from the working tree itself, never from a worktree of head_sha: head_sha equals the base, so a committed-ref analysis would diff two identical graphs and silently miss every uncommitted change — a false APPROVE for any code-carrying diff. Correct recipe (verified on opsx/init-scaffolds-structural-gate-caller, base 9ccb82d): validate the base ref in order (regex ^[A-Za-z0-9._/-]+$ → git check-ref-format → git rev-parse --verify --end-of-options "<ref>^{commit}"), git worktree add --detach the resolved base SHA into the pre-approved temp dir outside the repo tree, run `vibe-check analyze --timeout 5m --output <file>` on BOTH the base worktree and the repo working tree (target "."), `vibe-check diff [--json]` the two snapshots, then `git worktree remove --force`. Working-tree analysis also proved robust to gitignored in-workspace sibling clones (.uf/sibling-clones): nested modules do not pollute the module graph, so the PR-side snapshot matched the clean base worktree exactly (7 packages, all deltas zero, verdict APPROVE, unreliable=false).
