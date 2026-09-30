---
tag: vibe-check-provenance
author: jay-flowers
created_at: 2026-09-30T12:30:19Z
identity: vibe-check-provenance-20260930T123019-jay-flowers
tier: draft
---

When a change has a cross-change dependency (e.g. documenting behavior added by a sibling OpenSpec change), the review council will flag the dependency as a "phantom" if the feature branch was created from a stale local `main`. In the document-provenance-output change, all 10 Divisor reviewers returned REQUEST CHANGES on the premise that `add-diff-init-provenance` did not exist, when in fact it existed as branch commit 4c41b72 and was squash-merged to origin/main as e6e2389 — the reviewers simply could not see it from the stale main-based branch. The fix was `git fetch origin && git checkout main && git merge --ff-only origin/main && git checkout <branch> && git rebase main`. Lesson: always refresh main and rebase before invoking the review council when the change's proposal/design references another change that is unmerged or recently merged.
