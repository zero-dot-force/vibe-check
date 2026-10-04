## Context

The agent-design convention pack (`.opencode/uf/packs/agent-design.md`) is already deployed and contains 10 structural quality rules (AD-001 through AD-010). It is consumed by Cobalt-Crush (implementation) and all Divisor persona agents (review). The companion file (`.opencode/uf/packs/agent-design-custom.md`) provides a project-specific override mechanism.

This change adds standalone documentation — not in the pack files themselves — covering the full enforcement model. The existing pack files serve as runtime configuration for agents; the documentation serves as a human-readable reference for contributors and agent maintainers.

## Goals / Non-Goals

**Goals:**
- Create a standalone documentation file covering all 10 rules with IDs, thresholds, severity levels, enforcement tools, and rationale
- Document the `agent-design-custom.md` override mechanism (CR-NNN prefix, scoped overrides)
- Clarify enforcement tool mapping across vibe-check, gaze, golangci-lint, and review agents
- Document forward references for AD-002 (`--max-ce`) and AD-008 (`--max-duplication`)
- Document LCOM4 integer semantics for AD-009

**Non-Goals:**
- Modifying the existing pack files (`agent-design.md`, `agent-design-custom.md`)
- Implementing forward-reference features (`--max-ce`, `--max-duplication`)
- Adding new rules beyond AD-001–AD-010
- Changing enforcement thresholds

## Decisions

### Documentation placement: `docs/agent-design-pack.md`

The file will live at `docs/agent-design-pack.md` in the repository root. This keeps documentation near the code it describes and follows the convention of other Go projects. The existing `.opencode/uf/packs/` directory is runtime configuration, not documentation.

**Alternatives considered:**
- Placing docs inside `.opencode/uf/packs/`: Rejected — that directory is for agent runtime config, not human docs
- Using a wiki or Confluence page: Rejected — documentation should be versioned alongside the code

### Documentation structure

The document is organized into the following sections:
1. **Introduction**: Purpose and target audience (contributors and agent maintainers)
2. **Rule Reference Table**: A compact table of all 10 rules with ID, name, threshold, severity, enforcement tool, and rationale
3. **Rule-by-Rule Detail**: Per-rule sections covering rationale, compliance examples (PASS/FAIL), and tool-specific flags
4. **Enforcement Tool Matrix**: A cross-reference showing which tool enforces which rules, including notes on forward references
5. **Override Mechanism**: How to use `agent-design-custom.md` for project-specific threshold overrides
6. **LCOM4 Semantics**: Clarification that LCOM4 is an integer connected-component count, not a float ratio
7. **Authority Statement**: The pack file is authoritative; the documentation is a derivative reference

## Risks / Trade-offs

- **Forward references may become outdated**: AD-002 (`--max-ce`) and AD-008 (`--max-duplication`) reference planned features. The documentation notes these as forward references. If those features ship, the doc should be updated. → Mitigation: Link to the tracking issue for `--max-duplication` (#13). AD-002 (`--max-ce`) has no tracking issue at this time.
- **Duplication with pack files**: The documentation mirrors information already in `agent-design.md`. If a rule changes, both the pack file and the doc need updating. → Mitigation: The doc explicitly states the pack file is authoritative; the doc is a derivative reference.