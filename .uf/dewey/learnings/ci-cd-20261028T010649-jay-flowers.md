---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:06:49Z
identity: ci-cd-20261028T010649-jay-flowers
tier: draft
---

A reusable workflow that reads PR-scoped context (github.event.pull_request.base.sha) produces a SILENT VACUOUS GREEN GATE when its caller triggers on non-PR events: on push/schedule/workflow_dispatch the PR context is empty, actions/checkout with an empty ref falls back to the event SHA, base == head, the delta is zero, and the gate passes — a green check that proves nothing, with no error anywhere. Fix at the CALLER: any job consuming such a workflow in a workflow that triggers on more than pull_request must carry `if: github.event_name == 'pull_request'` (bare form or ${{ }} form; both valid — prefer the bare form to match house style). Canonical case: vibe-check's ci.yml dogfood job for reusable_structural_gate.yml (task 2.1 of openspec/changes/reusable-structural-gate, spec-review advisory). Companion mechanics for same-repo dogfooding: job-level `uses:` cannot contain expressions, so the reference must be the static local form ./.github/workflows/<file>.yml (GitHub rejects ./path@ref), which resolves against the caller's ref — the PR's own reviewed code, so no SHA pin is possible or needed (CI-001's SHA requirement applies to external references; the local reference's trust posture is stronger, not weaker). Unlike GitHub (which validates uses:/with: only at dispatch), actionlint DOES resolve local ./ references at lint time — it checks the callee file exists, has a workflow_call trigger, and that with: inputs match declared inputs — so actionlint exit 0 on a caller with a local reference is meaningful pre-merge evidence, while cross-repo references remain dispatch-time-only (design D10's dogfood rationale).
