## Why

AI coding agents produce structurally fragile code when operating without explicit
architectural constraints. Without convention-pack rules for coupling, cohesion,
complexity, and naming, agents routinely generate high-fan-out packages, circular
dependencies, and oversized files that accumulate technical debt faster than human
developers. The Unbound Force ecosystem already has convention packs for language
idioms (`go.md`, `typescript.md`) and CI policy (`ci.md`), but lacks structural
quality rules that agents can enforce during code generation and review.

Vibe-Check and Gaze now provide the metric engines to measure these properties.
This pack codifies the thresholds so every agent in the ecosystem shares a single
source of truth for structural quality gates.

## What Changes

- Add `.opencode/uf/packs/agent-design.md` containing 10 structural quality rules
  (AD-001 through AD-010)
- Each rule specifies: ID, Name, Rationale, Threshold, Enforcement tool, and
  pass/fail Examples
- Rules cover coupling (Ce, Instability, circular deps), cohesion (LCOM4),
  complexity (cognitive complexity), naming conventions, file size, duplication,
  contract coverage, and test assertion depth
- Enforcement is mapped to four tool categories: vibe-check (AD-002, AD-003,
  AD-004, AD-008, AD-009), gaze (AD-001, AD-006, AD-010), golangci-lint (AD-005),
  and review agents (AD-007)

## Capabilities

### New Capabilities
- `agent-design-rules`: Convention pack defining 10 structural quality rules
  (AD-001 through AD-010) with thresholds, enforcement mappings, and examples

### Modified Capabilities

None

### Removed Capabilities

None

## Constitution Alignment

- **I. Autonomous Collaboration**: PASS — Convention pack is a self-describing
  artifact that agents consume asynchronously without human mediation
- **II. Composability First**: PASS — Pack follows the existing `*-custom.md`
  override pattern; no mandatory dependencies introduced
- **III. Observable Quality**: PASS — Rules map to machine-parseable tool output
  (vibe-check JSON, gaze reports); AD-007 (file size) uses simple line counting
  that any agent can perform
- **IV. Testability**: PASS — No production code changes; pack rules define
  clear pass/fail thresholds that can be verified deterministically
- **V. Security by Default**: N/A — No security implications
- **VI. Metric Fidelity**: PASS — Thresholds use LCOM4 integer values consistent
  with the existing metric model; forward references are explicitly marked
- **VII. Language Agnosticism**: PASS — Rules reference Go-specific tools but
  the pack format is language-agnostic; future language adapters can define
  equivalent packs

## Impact

- **Code**: New file `.opencode/uf/packs/agent-design.md` — no production code changes
- **Agents**: All review council agents and coding agents that read convention packs
  will gain 10 new structural quality gates to enforce
- **Dependencies**: Rules reference existing vibe-check CLI flags
  (`--max-instability`, `--no-circular-deps`, `--max-lcom`) and gaze CLI
  capabilities (CRAP scores, assertion depth, contract coverage). Rules AD-002
  (`--max-ce`) and AD-008 (`--max-duplication`) reference forward-planned
  vibe-check features that are not yet implemented — agents apply these rules
  heuristically until tooling ships
- **Systems**: No CI workflow changes required — convention packs are consumed by
  agents at review time, not by CI pipelines directly
