---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-10-28T14:17:18Z
identity: ci-cd-20261028T141718-jay-flowers
tier: draft
---

A same-repo dogfood consumer of a local-reference reusable workflow (`uses: ./.github/workflows/foo.yml`) resolves the callee from the PR's merge ref — so on a fork pull request the fork author controls BOTH the calling job and the called workflow file, executing arbitrary callee YAML in your CI. That is safe without a fork guard under `pull_request` (not `pull_request_target`) semantics: fork-PR workflow runs are maintainer-approval-gated, the GITHUB_TOKEN is forced read-only, and no secrets are exposed — the attacker-modified callee gains nothing beyond editing the caller's own CI in the same PR, which is the same privilege. The pattern inverts under `pull_request_target` (base-ref workflow + write token + secrets): checkout-of-PR-head inside such a workflow is the classic exfiltration vector, and fork guards ARE required there. Canonical case: vibe-check ci.yml's structural-gate-dogfood job (opsx/reusable-structural-gate, 2026-10-28) deliberately carries no fork guard, matching the callee's D6 trust model (analyze compiling fork-PR code = same trust as running go test on it, with read-only token and persist-credentials: false); the build-from-source self-gate (structural-gate.yml) retains its fork guard for a different reason — it builds the PR's own source into the binary used as the gate.
