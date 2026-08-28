---
description: Convert existing tasks into actionable, dependency-ordered GitHub issues for the feature based on available design artifacts.
tools: ['github/github-mcp-server/issue_write']
---

<protect>
## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Read tasks.md and create GitHub issues for each task. Issues are created in dependency order with proper labels, cross-references, and traceability back to spec requirements. Requires a GitHub remote URL and uses the GitHub MCP server for issue creation.

## Execution Steps

### 1. Initialize Context

Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` once from repo root and parse JSON for FEATURE_DIR and AVAILABLE_DOCS. Derive absolute paths:

- TASKS = FEATURE_DIR/tasks.md
- SPEC = FEATURE_DIR/spec.md (for requirement cross-references)

Abort with an error message if tasks.md is missing. Instruct the user to run `/speckit.tasks` first.

For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

### 2. Validate GitHub Remote

Run `git remote get-url origin` to get the repository URL. Parse the owner and repo name from the URL.

**CRITICAL**: Validate that the remote URL points to a real GitHub repository. If the remote is not a GitHub URL, **STOP** with an error:
> "No GitHub remote found. This command requires a GitHub repository."

Display the resolved owner/repo to the user and ask for confirmation before creating any issues.

### 3. Load and Parse Tasks

Read tasks.md. For each task, extract:

- Task title
- Description/acceptance criteria
- Phase/group membership
- Dependencies (other tasks that must complete first)
- Requirement references ([FR-xxx], [US-xxx] labels)
- Estimated complexity (if present)

### 4. Create Issues in Dependency Order

For each task, create a GitHub issue using the GitHub MCP server:

- **Title**: Task title from tasks.md
- **Body**: Include description, acceptance criteria, requirement cross-references, and phase information
- **Labels**: Add appropriate labels (e.g., `feature`, `phase-1`, `spec-generated`)
- **Dependencies**: Reference dependency issues by number in the body

Create issues in dependency order — tasks with no dependencies first, then tasks that depend on already-created issues.

**NEVER create issues in a repository that does not match the validated remote URL.**

### 5. Update tasks.md

After all issues are created, update tasks.md to include GitHub issue references:

- Add issue numbers next to each task: `- [x] Task title (#123)`
- Add a summary section at the bottom listing all created issues

### 6. Report

Display:
- Total issues created
- Issue numbers and titles
- Any tasks that were skipped (with reasons)
- Link to the repository's issues page

## Operating Principles

- **NEVER create issues in unrelated repositories** — validate the remote URL first
- **Ask for confirmation** — show the user what will be created before creating
- **Dependency order** — create parent/dependency issues first
- **Traceability** — every issue links back to spec requirements
- **Idempotent awareness** — if issues already exist for some tasks, skip them
</protect>

## Guardrails

- This command creates **GitHub issues via** the MCP API.
  It does NOT write local files.
- Issues MUST only be created in the repository matching
  the current Git remote. NEVER create issues in
  unrelated repositories.
- Do NOT modify source code, spec artifacts, or any
  local files.

## Context

$ARGUMENTS
