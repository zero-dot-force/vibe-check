---
description: Perform a non-destructive cross-artifact consistency and quality analysis across spec.md, plan.md, and tasks.md after task generation.
---

<protect>
## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Perform a read-only cross-artifact consistency and quality analysis across spec.md, plan.md, and tasks.md. This command runs after `/speckit.tasks` produces a complete `tasks.md`. It detects duplication, ambiguity, underspecification, constitution alignment issues, coverage gaps, and inconsistencies across all three artifacts. It produces a Markdown analysis report with severity-ranked findings and optional remediation suggestions.

## Execution Steps

**STOP HERE. Do NOT proceed to implementation.**

Your job is done. Report the results and prompt the
user. The user will invoke a separate command
(`/speckit.implement`, `/unleash`, or `/cobalt-crush`)
when they are ready to implement.

### 1. Initialize Analysis Context

Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` once from repo root and parse JSON for FEATURE_DIR and AVAILABLE_DOCS. Derive absolute paths:

- SPEC = FEATURE_DIR/spec.md
- PLAN = FEATURE_DIR/plan.md
- TASKS = FEATURE_DIR/tasks.md
- CONSTITUTION = .specify/memory/constitution.md

Abort with an error message if any required file is missing. Instruct the user to run `/speckit.tasks` first if tasks.md is absent.

For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

### 2. Load Artifacts

Load the full content of spec.md, plan.md, tasks.md, and constitution.md.

### 3. Run Detection Passes

Perform 6 independent detection passes across the loaded artifacts:

**Pass 1 — Duplication Detection**
- Identify requirements, user stories, or tasks that appear duplicated across artifacts
- Flag tasks that restate the same requirement with different wording
- Check for redundant acceptance criteria

**Pass 2 — Ambiguity Detection**
- Find vague language: "should", "might", "could", "appropriate", "as needed", "etc."
- Identify requirements without measurable success criteria
- Flag undefined terms or acronyms

**Pass 3 — Underspecification Detection**
- Find requirements referenced in plan.md or tasks.md but not defined in spec.md
- Identify tasks without clear completion criteria
- Flag missing error handling or edge case coverage

**Pass 4 — Constitution Alignment**
- Verify each constitution principle is addressed in the spec and plan
- Check that the testing strategy aligns with Constitution Principle IV (Testability)
- Flag deviations from constitution constraints

**Pass 5 — Coverage Gap Detection**
- Map requirements (FR-xxx, US-xxx) to tasks — flag unmapped requirements
- Map tasks to requirements — flag orphan tasks
- Identify phases or areas with no test coverage

**Pass 6 — Inconsistency Detection**
- Check for contradictory requirements across artifacts
- Verify terminology consistency (same concept, same name)
- Flag mismatched priorities or ordering between plan and tasks

### 4. Format Analysis Report

Produce a Markdown report (no file writes) with:

#### Cross-Artifact Analysis Report

| ID | Pass | Severity | Location | Summary | Recommendation |
|----|------|----------|----------|---------|----------------|

Severity levels:
- **CRITICAL**: Blocks implementation — must be resolved
- **HIGH**: Likely to cause rework — should be resolved
- **MEDIUM**: Quality concern — recommended to address
- **LOW**: Minor improvement opportunity

**Summary Statistics:**

| Pass | CRITICAL | HIGH | MEDIUM | LOW |
|------|----------|------|--------|-----|
| Duplication | | | | |
| Ambiguity | | | | |
| Underspecification | | | | |
| Constitution | | | | |
| Coverage Gaps | | | | |
| Inconsistency | | | | |

**Coverage Matrix:**
- Requirements mapped to tasks: N/M
- Tasks mapped to requirements: N/M
- Constitution principles addressed: N/M

### 5. Provide Next Actions

At end of report, output a concise Next Actions block:

- If CRITICAL issues exist: Recommend resolving before `/speckit.implement`
- If only LOW/MEDIUM: User may proceed, with improvement suggestions
- Provide explicit command suggestions: e.g., "Run `/speckit.clarify` to resolve ambiguities", "Update spec.md to add missing requirements"

### 6. Offer Remediation

Ask the user: "Would you like me to suggest concrete remediation edits for the top N issues?" (Do NOT apply them automatically.)

## Operating Principles

- **NEVER modify files** — this is read-only analysis
- **NEVER hallucinate missing sections** — report accurately what is absent
- **Constitution violations are always CRITICAL**
- **Missing coverage mapping is HIGH** — traceability matters
- **Report zero issues gracefully** — emit success report with statistics
- **Deterministic results** — rerunning without changes should produce consistent findings
- **Token-efficient output** — limit findings table to 40 rows; summarize overflow
</protect>

## Guardrails

- **NEVER modify source code** — this command updates
  spec artifacts ONLY. Implementation changes belong in
  `/speckit.implement`, `/unleash`, or `/cobalt-crush`.
- **NEVER modify test files, Go source, Markdown agents,
  convention packs, or config files** outside the
  `specs/NNN-*/` feature directory.
- The ONLY files this command may write are:
  - `FEATURE_SPEC` (the spec.md file)
  - Files within `FEATURE_DIR` (spec artifacts:
    plan.md, tasks.md, research.md, data-model.md,
    quickstart.md, contracts/, checklists/)
- The user needs to review the plan before
  implementation begins. Implementing without review
  defeats the purpose of the spec-first workflow.

## Context

$ARGUMENTS
