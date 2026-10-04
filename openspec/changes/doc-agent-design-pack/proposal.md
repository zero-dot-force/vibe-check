## Why

The agent-design convention pack (AD-001 through AD-010) defines 10 structural quality rules that Unbound Force agents enforce during code generation and review. These rules are already deployed at `.opencode/uf/packs/agent-design.md`, but there is no standalone documentation covering the rule reference, enforcement tool mapping, override mechanism, forward references, or LCOM4 semantics. Without this documentation, new contributors and agent maintainers cannot understand the full enforcement model.

## What Changes

- Add documentation covering all 10 rule IDs (AD-001–AD-010), their thresholds, severity levels, and enforcement tools
- Document the `agent-design-custom.md` override mechanism (CR-NNN prefix, scoped overrides with justification)
- Clarify enforcement tool mapping: vibe-check (AD-002, AD-003, AD-004, AD-008, AD-009), gaze (AD-001, AD-006, AD-010), golangci-lint (AD-005), review agents (AD-007)
- Note forward references for AD-002 (`--max-ce`) and AD-008 (`--max-duplication`) — planned but not yet implemented
- Document LCOM4 integer semantics (connected components, not a 0.0–1.0 cohesion float)

## Capabilities

### New Capabilities

- `agent-design-reference`: Full rule reference for the agent-design convention pack (AD-001 through AD-010) with rule IDs, thresholds, severity levels, enforcement tools, and rationale

### Modified Capabilities

<!-- No existing capabilities are modified — this is a new documentation addition -->

## Impact

- **Documentation**: New doc file at `docs/agent-design-pack.md` documenting the agent-design convention pack
- **No code changes**: The convention pack itself (`agent-design.md`, `agent-design-custom.md`) is already shipped and requires no modification
- **No API or dependency changes**