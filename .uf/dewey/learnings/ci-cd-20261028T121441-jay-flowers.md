---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-10-28T12:14:41Z
identity: ci-cd-20261028T121441-jay-flowers
tier: draft
---

Caller-controlled `workflow_call` inputs never cross a trust boundary beyond the calling workflow file itself. On `pull_request` events (including forks), the running workflow file — and therefore its `with:` inputs — come from the PR's merge ref, so a fork-PR author who could abuse an input already controls the entire file and could replace the job outright; on `pull_request_target`, the file and its inputs both come from the trusted base. Consequence for adversarial severity assessment: an input like a release `version` used to construct a download URL (e.g. `releases/download/$VERSION`) cannot escalate privilege beyond controls-the-workflow-file, and URL-redirecting values still fail against the checksum gate unless the attacker also controls the release transport (the documented out-of-scope threat). The env-routing discipline (`env:` block + quoted \"$VAR\", never `${{ }}` interpolation in `run:`) remains mandatory, but its security value is containment and single-argv semantics against accidental shell interpretation — not a trust-boundary defense. Verified in vibe-check's reusable_structural_gate.yml adversarial review (opsx/reusable-structural-gate, 2026-10-28): enumerating all seven inputs (version, path, timeout, go-version-file, three max-*-delta) found zero boundary-crossing vectors despite URL-constructing and path-joining usage.
