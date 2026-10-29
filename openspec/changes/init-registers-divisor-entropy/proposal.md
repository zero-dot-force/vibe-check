# Proposal

## Why

`vibe-check init` deploys `divisor-entropy.md` into a target project's
`.opencode/agents/` but does not register the agent in the target's Unbound
Force review policies. `.uf/reviewer-capabilities.yaml` is the canonical
eligibility policy for Divisor personas — the review-dispatch planner fails
closed when a discovered agent has no manifest entry ("discovered agent
divisor-entropy has no reviewer-manifest entry"), so dispatch is blocked until
the manifest is hand-edited. `.uf/review-matrix.yaml` is the canonical model
policy for Divisor review dispatch — advisor list membership is what permits
generated runs, so without an entry the entropy reviewer is never scheduled as
an advisor. Every consuming project must hand-edit both files after running
`init`. This repo hit the same gap and fixed it manually: 25328d5
(capabilities manifest) and 81831e2 (matrix advisor lists).

## What Changes

- Add an idempotent registration step to `vibe-check init`, running after
  asset deployment, that registers `divisor-entropy` in the target's UF
  review policies:
  - `.uf/reviewer-capabilities.yaml`: insert the canonical four-line entry
    (`- agent: divisor-entropy`, `capability: review`, `scopes:`, and
    `- standard`) in sorted position among `capability: review` reviewers.
  - `.uf/review-matrix.yaml`: insert `- divisor-entropy` in sorted position
    in each advisor mode list (`code`, `spec`, `triage`, `feedback`).
- Decision rules (both files):
  - Entry missing, file present → insert in sorted position, preserving all
    existing formatting and comments (the files are hand-curated).
  - Entry already present → skip, regardless of scopes/values. Existing
    entries are never rewritten; `--force` governs asset files only.
  - File absent → skip with a notice in the output. Non-uf-scaffolded
    projects don't consume these files; uf-scaffolded projects always have
    them.
- Only `divisor-entropy` is registered. `vibe-check-reporter` and
  `mx-f-architecture-trend` are not review personas and are out of scope.
- Report registration outcomes using the existing category-prefix convention
  (e.g. `reviewers/divisor-entropy` in Written/Skipped), additive to the
  `--json` payload where a distinct key is needed.
- Line-based insertion in `internal/scaffold/` with the canonical entries as
  embedded constants; no YAML parser and no new dependencies (round-tripping
  would destroy the hand-curated comments).
- Update the README `init` section and add a CHANGELOG entry.

## Capabilities

### New Capabilities
- `init-uf-registration`: `vibe-check init` idempotently registers the
  `divisor-entropy` review agent in the target project's UF review policies
  (`.uf/reviewer-capabilities.yaml` and `.uf/review-matrix.yaml`) with
  skip-when-present, skip-when-file-absent, format-preserving sorted
  insertion, and symlink-safe file handling.

### Modified Capabilities
<!-- None: no main specs exist under openspec/specs/ yet; the repo tracks
     deltas change-locally. Prior precedent (add-diff-init-provenance)
     extends init behavior via a new capability rather than modifying the
     change-local init-command delta. -->

### Removed Capabilities
- None.

## Impact

- Affected code: `internal/scaffold/` (new registration step covering both
  files: canonical entries as constants, line-based sorted insertion, symlink
  guards consistent with the existing `ensureDir`/`verifyContained`/
  `regularFileExists` protections), `cmd/vibe-check/init.go` (report
  registration outcomes in the summary and `--json` payload; update the
  cobra Short/Long help text and exit-code sentence for the new `.uf/`
  writes, `Notices:` section, and exit-2 conditions),
  `internal/scaffold/scaffold_test.go` (registration test cases).
- Affected docs: `README.md` (init section), `CHANGELOG.md`
  (`[Unreleased]` entry), `AGENTS.md` (Project Structure, Architecture
  Layer 3, Build & Test Commands).
- No new dependencies; no breaking changes. The `written`, `skipped`, and
  `forced` payload keys keep their existing semantics for asset files;
  registration outcomes are reported under the new `reviewers/` and
  `advisors/` category prefixes within the same arrays, and the additive
  `notices` key carries absent-file skips.
- Content pipeline: automatic registration completes the `init` deployment
  story by removing the hand-editing step; a mention fits the open
  divisor-entropy announcement story (blog issue #27) rather than a new
  post.

## Constitution Alignment

Assessed against the Vibe-Check constitution (`.specify/memory/constitution.md`).

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PASS — registration is a pure step inside the existing `vibe-check init` artifact pipeline; outcomes flow through the command's self-describing `--json` payload (additive `notices` array alongside `written`/`skipped`/`forced` and the existing provenance object), so any consumer interprets results without consulting another agent. No synchronous interaction is introduced. |
| II. Composability First | PASS — `vibe-check init` remains independently usable; registration auto-detects uf-scaffolded targets (policy files present) and degrades to a notice on non-scaffolded targets, adding value when the Unbound Force scaffold is present without making it a mandatory dependency. |
| III. Observable Quality | PASS — registration outcomes are machine-parseable (`--json` with the additive `notices` array) and human-readable (summary with a `Notices:` section printed only when non-empty); behavior claims are backed by unit tests over canonical fixtures derived from this repo's `.uf/` files. |
| IV. Testability | PASS — coverage strategy defined in design.md (≥80% statement coverage for the new registration code); all tests are isolated unit tests over `t.TempDir()` fixtures exercising observable side effects through the injectable `WriteFile` seam; no external services or network. |
| V. Security by Default | PASS — no new dependencies (the default answer is "do not add"); path safety mirrors and strengthens the existing scaffold guards (Lstat type checks, symlink and non-regular-file refusal with exit 2, no writes through symlinks); existing file modes are preserved. |
| VI. Metric Fidelity | N/A — no metric formulas are added or changed; registration is line-based text splicing in the scaffold layer, not metric computation. |
| VII. Language Agnosticism | N/A — registration touches UF policy YAML in `internal/scaffold/`, not the analysis engine or any language adapter; no language-coupled logic is introduced. |
