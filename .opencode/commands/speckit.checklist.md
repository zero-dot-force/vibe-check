---
description: Generate a custom checklist for the current feature based on user requirements.
---

<protect>
## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Generate checklists that test REQUIREMENTS quality — not implementation behavior. Checklists validate whether the spec artifacts are complete, clear, consistent, and measurable. Items use question format and include quality dimension tags. Each run creates a NEW checklist file (never overwrites existing ones).

## Execution Steps

**STOP HERE. Do NOT proceed to implementation.**

Your job is done. Report the results and prompt the
user. The user will invoke a separate command
(`/speckit.implement`, `/unleash`, or `/cobalt-crush`)
when they are ready to implement.

### 1. Initialize Context

Run `.specify/scripts/bash/check-prerequisites.sh --json` once from repo root and parse JSON for FEATURE_DIR and AVAILABLE_DOCS. Derive absolute paths:

- SPEC = FEATURE_DIR/spec.md
- PLAN = FEATURE_DIR/plan.md (if available)
- TASKS = FEATURE_DIR/tasks.md (if available)
- CHECKLISTS_DIR = FEATURE_DIR/checklists/

Abort with an error message if spec.md is missing. Plan and tasks are optional but enhance checklist quality.

For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

### 2. Ask Clarifying Questions

Before generating, ask up to 3 clarifying questions to scope the checklist:

- What domain or aspect should the checklist focus on? (e.g., data model, UX flows, security, API contracts, error handling)
- Are there specific requirements that feel under-specified?
- What quality dimensions matter most? (Completeness, Clarity, Consistency, Measurability, Coverage)

Wait for the user's responses before proceeding.

### 3. Load Artifacts

Load spec.md and any available plan.md and tasks.md. Focus on the areas identified by the user's answers.

### 4. Generate Checklist

Create a checklist file at `CHECKLISTS_DIR/<domain>.md` where `<domain>` is derived from the user's focus area (e.g., `data-model.md`, `security.md`, `api-contracts.md`).

If a file with that name already exists, append a numeric suffix: `<domain>-2.md`, `<domain>-3.md`, etc. NEVER overwrite existing checklists.

Ensure the `checklists/` directory exists before writing.

**Checklist format:**

```markdown
# Checklist: <Domain>

Feature: <feature name>
Generated: <date>
Focus: <user-specified focus>

## Requirements Quality

- [ ] Are success criteria defined for all user stories? [Completeness]
- [ ] Are error states specified for each workflow step? [Coverage]
- [ ] Is terminology used consistently across spec and plan? [Consistency]
- [ ] Are performance thresholds quantified (not "fast" or "responsive")? [Measurability]
- [ ] Are edge cases enumerated for each functional requirement? [Coverage]
...
```

Each item MUST:
- Use question format: "Are [X] defined for [Y]?"
- Include exactly one quality dimension tag: `[Completeness]`, `[Clarity]`, `[Consistency]`, `[Measurability]`, or `[Coverage]`
- Be specific to the feature, not generic boilerplate

Generate 15-30 items per checklist, proportional to spec complexity.

### 5. Report

Display:
- The checklist file path
- Item count by quality dimension
- Any areas where the spec was too thin to generate meaningful checks

## Operating Principles

- **Checklists test REQUIREMENTS, not code** — "Are X defined?" not "Does X work?"
- **Question format only** — every item is a question
- **Never overwrite** — each run creates a new file
- **Specific over generic** — items reference actual spec content
- **Quality dimensions required** — every item tagged
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
