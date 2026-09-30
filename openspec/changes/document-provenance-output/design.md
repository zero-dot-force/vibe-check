## Context

The provenance feature spans three JSON-emitting CLI commands:

- `vibe-check analyze` — documents producer/version/generatedAt/input, landed via
  `emit-provenance-metadata` (schema `1.1` → `1.2`).
- `vibe-check diff --json` — emits a three-field envelope (producer/version/
  generatedAt), added by `add-diff-init-provenance` (merged into `main`).
- `vibe-check init --json` — emits the same three-field envelope, added by the
  same change.

`add-diff-init-provenance` updated `CHANGELOG.md` and `AGENTS.md` but did not touch
`README.md`. As a result, README's `diff` and `init` sections describe only the
metric/verdict and deployment behavior, with no mention of the `provenance` object
or the `--no-provenance` flag. Issue #34 tracks the `analyze` documentation gate;
issue #36 tracks the `diff`/`init` gap this change closes.

Current README state (on `main`):

- Output section (lines ~84–86) already documents the `provenance` object for
  `analyze`.
- Flag table (line ~54) already lists `--no-provenance` for `analyze`.
- JSON example already shows `schemaVersion: "1.2"` and a `provenance` object.
- `diff` section (lines ~135–156) does NOT mention provenance or `--no-provenance`.
- `init` section (lines ~158–175) does NOT mention provenance or `--no-provenance`.

## Goals / Non-Goals

**Goals:**

- Make README the single source of truth for the `provenance` field and
  `--no-provenance` flag across `analyze`, `diff`, and `init`.
- Ensure the JSON example and flag table remain accurate against schema `1.2`.
- Confirm CHANGELOG `[Unreleased]` carries an entry for each provenance change.
- Pin each documentation requirement to a verifiable location via a spec, so a
  future drift is detectable.

**Non-Goals:**

- No code, schema, or dependency changes.
- No new documentation beyond README/CHANGELOG (e.g., no separate docs site).
- No wording/grammar overhaul of unrelated README sections.

## Decisions

**Decision 1 — Treat the documentation as one capability across all three commands.**

Rationale: the `--no-provenance` flag and its purpose are identical across
`analyze`, `diff`, and `init`, and all three commands share the same three core
provenance fields (`producer`, `version`, `generatedAt`). `analyze` additionally
carries an `input` sub-object (`path`, `modulePath`); `diff` and `init` do not.
Documenting them together avoids three near-duplicate prose blocks while keeping
the field-shape asymmetry explicit. Alternatives considered: three separate
capabilities (more surface, no benefit for prose); documenting only `analyze`
(leaves the real gap unaddressed).

**Decision 2 — Add prose to the existing `diff` and `init` sections rather than
expanding the single flag table.**

Rationale: the flag table documents `analyze` flags only; `diff` and `init`
describe their flags inline in their own sections. Adding a short `--no-provenance`
mention to each section matches the existing structure. Alternative: a combined
"provenance flags" table (rejects — diverges from the current README layout).

**Decision 3 — Spec pins each requirement to a concrete, grep-able location.**

Rationale: prose cannot be unit-tested. The spec expresses requirements such that
"the README flag table SHALL list `--no-provenance`" and "the `diff` section SHALL
mention `provenance`" can be verified by inspection of the pinned README/CHANGELOG
sections, plus the `openspec validate` gate. An automated doc-contract test is out
of scope — this change is documentation-only and adds no code. Alternative: no spec
(rejects — documentation is a project invariant per AGENTS.md's documentation gate).

**Decision 4 — Base this change on the already-merged `add-diff-init-provenance`.**

Rationale: README `diff`/`init` provenance prose describes behavior added by
`add-diff-init-provenance`, which is merged into `main` (commit `e6e2389`). The
documentation branch is authored against the current tree, so the prose and code
cannot diverge. Alternative (superseded): gate this change behind the code change
in the same PR — no longer needed now that the code has landed.

## Risks / Trade-offs

- [Prose/behavior drift] → Mitigation: spec pins each requirement to a concrete
  README/CHANGELOG location; tasks include a verification pass against the emitted
  JSON from each command.
- [Field-shape drift] → Mitigation: tasks verify `diff`/`init` `--json` output
  against the current `main` tree, confirming the three-field envelope omits
  `input`; the spec pins the shape explicitly.
- [Redundant/near-duplicate prose across three sections] → Mitigation: keep the
  `diff`/`init` additions one short sentence each, naming the three fields and
  their semantics (`producer` as the constant `"vibe-check"`, `version`, RFC 3339
  UTC `generatedAt`) and noting `input` is `analyze`-only, rather than deferring
  to the richer `analyze` output section.
