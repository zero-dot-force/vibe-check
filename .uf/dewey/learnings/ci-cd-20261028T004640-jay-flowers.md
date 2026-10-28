---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-28T00:46:40Z
identity: ci-cd-20261028T004640-jay-flowers
tier: draft
---

`actions/checkout` persists the job's GITHUB_TOKEN into the checkout's `.git/config` for the remainder of the job BY DEFAULT (`persist-credentials` defaults to true). Any step that later executes code over that working tree can read the token: the Go toolchain compiling a fork-PR module (go/packages type-aware loading runs build-time code — cgo, generated files, custom build steps), make targets, codegen scripts. In workflows whose core activity is compiling the caller's possibly-untrusted code — reusable gates that deliberately accept fork PRs — every checkout step MUST set `persist-credentials: false` whenever nothing after the clone touches a git remote (no push, no submodule update, no further fetch): the token is needed only for the clone itself (Constitution V: secrets scoped to the narrowest context needed). Spec-review advisory applied in vibe-check's reusable_structural_gate.yml (task 1.3 of openspec/changes/reusable-structural-gate); the fork-guarded self-gate (structural-gate.yml) skips fork PRs entirely, so the untrusted-code-over-checkout exposure is specific to reusable workflows without a fork guard.
