---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-10-28T00:58:33Z
identity: ci-cd-20261028T005833-jay-flowers
tier: draft
---

Optional `workflow_call` string inputs arrive in a reusable workflow as EMPTY STRINGS when the caller leaves them unset — never null or absent — so "forward this flag only when the caller set it" must be implemented as a non-empty test (`[ -n "$VAR" ]`), not a presence/definedness test. Canonical shape (from vibe-check's reusable_structural_gate.yml, task 1.4 of openspec/changes/reusable-structural-gate): map the inputs in the step's `env:` block (`MAX_X_DELTA: ${{ inputs.max-x-delta }}`), then conditionally assemble argv in the script — `args=(); if [ -n "$MAX_X_DELTA" ]; then args+=(--max-x "$MAX_X_DELTA"); fi; tool cmd "${args[@]}"`. Two properties this buys: (1) caller-controlled values are never `${{ }}`-interpolated inside `run:` scripts (env routing + shell quoting keeps each value a single argv element — the input-injection defense), and (2) unset inputs forward NO flag, keeping the no-override invocation byte-identical to the reference workflow. `"${args[@]}"` on an empty array is safe under `set -u` on bash ≥ 4.4 (GitHub ubuntu runners ship 5.x). Leave value VALIDATION to the tool: when the CLI already enforces protected semantics (e.g. vibe-check diff's tighten-only `--max-*-delta`, exit 2 on looser), the workflow must not duplicate or risk weakening it — assemble and forward, nothing more.
