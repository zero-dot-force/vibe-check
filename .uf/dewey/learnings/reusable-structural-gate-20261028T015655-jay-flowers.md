---
tag: reusable-structural-gate
author: jay-flowers
category: pattern
created_at: 2026-10-28T01:56:55Z
identity: reusable-structural-gate-20261028T015655-jay-flowers
tier: draft
---

vibe-check reusable-workflow authorship pattern that shipped clean through both council reviews (opsx/reusable-structural-gate, 2026-10-28), reusable as a checklist for the next workflow_call surface: (1) declare permissions at BOTH workflow and job level, minimum contents:read — reusable workflows cannot escalate caller permissions; (2) hardcode the release base repo URL in curl commands because the github context inside a workflow_call refers to the CALLER's repo, not the workflow's home; (3) verify checksums fail-closed by grepping the exact asset line out of checksums.txt and piping to sha256sum --check - under set -euo pipefail with shell: bash; (4) install tools and write ephemeral snapshots outside the workspace ($RUNNER_TEMP + GITHUB_PATH exposure); (5) checkout caller head and base (base via github.event.pull_request.base.sha into a path: base subdirectory, persist-credentials: false on both); (6) satisfy the caller's toolchain via setup-go go-version-file input pointing at the caller's go.mod; (7) pass every input through env: indirection and quote "$VAR" in run: scripts (workflow_call inputs arrive as empty strings, and ${{ }} interpolation in run: is an injection vector); (8) let the gate step own its exit code — no if/continue-on-error on it — while the markdown report step runs if: always() with continue-on-error: true. Each element traces to a spec-review advisory or code-review finding on this change.
