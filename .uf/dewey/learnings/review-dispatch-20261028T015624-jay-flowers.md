---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:24Z
identity: review-dispatch-20261028T015624-jay-flowers
tier: draft
---

Review-tooling config riders (dispatch-roster membership, reviewer-capabilities registration, review-matrix advisor-list entries, .gitignore additions for tool config) committed on a spec'd feature branch need an authorization record beyond PR-walkthrough declaration. Declaration is disclosure, not authorization: the first code review of a council-roster change cannot both adjudicate and authorize it, because that lets the council modify its own composition in flight. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — the branch carries three .uf riders (review-matrix +divisor-entropy in all four advisor mode lists, reviewer-capabilities registration, .gitignore review-matrix.override entry) declared in the walkthrough but absent from proposal.md's Impact section; the prior add-divisor-entropy-agent spec is the missing authorization vehicle. Rule: tooling riders belong either on a dedicated chore/spec branch or in the proposal's Impact section naming each file and citing the governing spec — record them as an erratum before archive when they ride a feature branch.
