---
tag: structural-entropy
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:37:46Z
identity: structural-entropy-20261031T223746-jay-flowers
tier: draft
---

Assessing "new package unmeasured by CI entropy ceilings" feedback (vibe-check PR #75, 2026-10-31): the repo's CI entropy enforcement is delta-only by design. structural-gate.yml and reusable_structural_gate.yml run whole-module `vibe-check analyze` with explicitly NO --max-* threshold flags (the reusable workflow's header states "gate enforcement belongs solely to diff --gate") plus `diff --gate` under the protected Δ thresholds, and added packages are information-only in the verdict by protected design. A "unmeasured by CI" claim therefore needs verification against those two workflow files before acceptance — the package IS measured (whole-module analyze output in two gates; Dewey drift snapshots cover the repo). The only true residual — no absolute ceiling applies to ANY package — is a repo-wide tighten-only policy question, not a PR fix: wiring --max-instability=0.7 into CI would fail pre-existing internal/goadapter (I=0.9333, Ca=1) and the new internal/ripwireadapter (I=0.9412, Ca=1) today. Also verify the feedback's metric claims with a fresh `vibe-check analyze` in a pinned worktree rather than trusting them: here they verified exactly (Ca=1, Ce=16, I=0.9412, LCOM=2, A=0, D=0.0588, zone main-sequence, profile-matched to goadapter's Ca=1/Ce=14/I=0.9333/LCOM=4).
