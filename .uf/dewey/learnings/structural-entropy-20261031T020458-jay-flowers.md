---
tag: structural-entropy
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:04:58Z
identity: structural-entropy-20261031T020458-jay-flowers
tier: draft
---

Entropy-divisor reviews of uncommitted-local diffs must baseline the base-side graph at HEAD (the parent commit of the uncommitted work), not at the diverged upstream base. Empirical check on opsx/ripwire-adapter: git merge-base HEAD upstream/main resolved to HEAD itself — the branch tip is an ancestor of upstream/main (the fork is simply behind upstream), so upstream/main contains newer out-of-scope commits and a graph diff against it would attribute those reverted upstream changes to the PR. Correct measurement: analyze the working tree as the PR side and a detached worktree at the pinned HEAD SHA as the base side, using the same vibe-check binary for both.
