---
description: Identify underspecified areas in the current feature spec by asking up to 5 highly targeted clarification questions and encoding answers back into the spec.
---

<protect>
## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Scan spec.md for ambiguities across 10 taxonomy categories, ask up to 5 targeted clarification questions (one at a time), provide recommended answers with reasoning, and integrate answers directly into the relevant spec.md sections. Record all Q&A in a `## Clarifications` section.

## Execution Steps

**STOP HERE. Do NOT proceed to implementation.**

Your job is done. Report the results and prompt the
user. The user will invoke a separate command
(`/speckit.implement`, `/unleash`, or `/cobalt-crush`)
when they are ready to implement.

### 1. Initialize Context

Run `.specify/scripts/bash/check-prerequisites.sh --json` once from repo root and parse JSON for FEATURE_DIR and AVAILABLE_DOCS. Derive absolute paths:

- SPEC = FEATURE_DIR/spec.md

Abort with an error message if spec.md is missing. Instruct the user to run `/speckit.specify` first.

For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

### 2. Load and Analyze Spec

Read spec.md fully. Scan for ambiguities across 10 taxonomy categories:

1. **Functional Scope** — unclear feature boundaries, missing "out of scope"
2. **Data Model** — undefined fields, missing types, unclear relationships
3. **UX Flow** — missing states, undefined transitions, unclear navigation
4. **Non-Functional** — vague performance targets, missing SLAs
5. **Integration** — undefined API contracts, missing auth flows
6. **Edge Cases** — unhandled error states, boundary conditions
7. **Constraints** — missing technical/business constraints
8. **Terminology** — undefined or inconsistently used terms
9. **Completion Signals** — unclear definition of done, missing acceptance criteria
10. **Placeholders** — TODO markers, TBD notes, "to be determined" text

Rank discovered ambiguities by impact (how much implementation risk they create).

### 3. Ask Clarification Questions

Ask up to 5 questions, **one at a time**. For each question:

1. State the ambiguity category and location in spec.md
2. Ask the specific question
3. Provide a **recommended answer** with reasoning (the user can accept, modify, or reject)
4. Wait for the user's response before asking the next question

If fewer than 5 ambiguities are found, ask fewer questions. If the spec is exceptionally clear, report that with a summary and skip to step 5.

### 4. Integrate Answers

For each answered question:

1. Update the relevant section of spec.md with the clarified information
2. Add the Q&A to a `## Clarifications` section at the end of spec.md (create the section if it doesn't exist)

**Clarifications section format:**

```markdown
## Clarifications

### Q1: [Category] — [Short question summary]
**Question**: [Full question text]
**Answer**: [User's answer or accepted recommendation]
**Integrated into**: [Section reference, e.g., "## Data Model > User entity"]
```

### 5. Report

Display:
- Number of ambiguities found by category
- Questions asked and answers received
- Sections of spec.md that were updated
- Remaining ambiguities (if any) that weren't addressed in this session

## Operating Principles

- **One question at a time** — don't overwhelm the user
- **Always provide a recommended answer** — the user should be able to just say "yes"
- **Integrate immediately** — don't defer updates to spec.md
- **Preserve existing content** — add clarifications, don't rewrite sections
- **5 question maximum** — respect the user's time; more sessions can follow
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
