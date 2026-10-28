---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-10-28T12:14:48Z
identity: openspec-20261028T121448-jay-flowers
tier: draft
---

A change's proposal.md Impact section is its scope contract — the file list the review council reviews against. Before submitting a spec'd feature branch, reconcile the actual diff (`git diff --name-only <base>...HEAD`) against the Impact section's New/Modified file lists: every file present in the diff but absent from Impact is an unauthorized rider, regardless of walkthrough disclosure (declaration is disclosure, not authorization). Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — three review-governance riders (.uf/review-matrix.yaml, .uf/reviewer-capabilities.yaml, .gitignore) rode the branch undeclared through two approving council gates because neither gate performs diff-vs-Impact reconciliation; the gap surfaced only in the retrospective and became a HIGH Guard finding at the final code review, blocking an otherwise-approved PR on a documentation-only fix. Distinction that keeps the check low-noise: retrospective learnings (.uf/dewey/learnings/) are process byproducts with an established home, not riders — the discriminator is whether the file changes review or governance behavior. The reconciliation is one command at submission time and converts a blocking governance finding into a pre-submission chore.
