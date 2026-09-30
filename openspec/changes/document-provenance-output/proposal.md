## Why

The provenance work introduced two user-facing surfaces — a top-level `provenance`
object and a `--no-provenance` flag — but their documentation is incomplete. The
`emit-provenance-metadata` change (issue #24) documented the `analyze` path in
README, while the follow-up `add-diff-init-provenance` change (issue #25) updated
`CHANGELOG.md` and `AGENTS.md` but left README's `diff` and `init` sections
silent. Issue #34 tracks the `analyze` documentation gate, and issue #36 tracks the
`diff`/`init` gap this change closes. The `provenance` output field and
`--no-provenance` flag must be documented consistently for every JSON-emitting
command so users and downstream consumers can discover and rely on them.

## What Changes

- Verify README's output section documents the `provenance` object (`producer`,
  `version`, `generatedAt`, `input{path,modulePath}`) — already present for
  `analyze`; confirm wording is accurate.
- Verify README's flag table lists `--no-provenance` with its byte-reproducibility
  purpose — already present for `analyze`.
- Confirm README's JSON example reflects `schemaVersion` `"1.2"` and includes an
  example `provenance` object.
- Add documentation for the `provenance` object and `--no-provenance` flag to
  README's `vibe-check diff` and `vibe-check init` sections (the actual gap left
  by `add-diff-init-provenance`).
- Confirm `CHANGELOG.md` has entries under `[Unreleased]` for both provenance
  changes (analyze and diff/init).

## Capabilities

### New Capabilities

- `provenance-output-documentation`: README and CHANGELOG accurately document the
  `provenance` output field (`producer`, `version`, `generatedAt`; `analyze` also
  records `input{path,modulePath}`) and the `--no-provenance` flag for `analyze`,
  `diff`, and `init`, including the `schemaVersion` `1.2` JSON example and the
  `[Unreleased]` changelog entries.

### Modified Capabilities

<!-- None: this is a documentation-only change. No live specs under
     openspec/specs/ and no runtime behavior changes. -->

### Removed Capabilities

<!-- None. -->

## Impact

- **Docs**: `README.md` (output section, flag table, JSON example, and the
  `diff`/`init` sections) and confirmation of `CHANGELOG.md` `[Unreleased]`
  entries. No code, API, schema, or dependency changes.
- **Dependency note**: README `diff`/`init` provenance documentation describes
  behavior added by `openspec/changes/add-diff-init-provenance/` (merged into
  `main` as commit `e6e2389`). The documentation is accurate against the current
  tree.
- **Risk**: Documentation is prose, not executable — a prose/behavior drift could
  go undetected. Mitigated by a spec that pins each documentation requirement to a
  concrete, verifiable location (see `specs/provenance-output-documentation/`).

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | N/A — documentation-only; no agent artifacts or inter-agent exchange formats. |
| II. Composability First | N/A — no new dependencies or coupling introduced. |
| III. Observable Quality | PASS — README becomes the single source of truth for the `provenance` field and `--no-provenance` flag across `analyze`, `diff`, and `init`. |
| IV. Testability | PASS — each requirement is pinned to a grep-able README/CHANGELOG location and verified by tasks against actual `--json` output. |
| V. Security by Default | N/A — no secrets, permissions, or dependency surface. |
| VI. Metric Fidelity | PASS — documents that provenance is metadata, not a metric, and that `--no-provenance` preserves byte-reproducible metric payloads. |
| VII. Language Agnosticism | N/A — no adapter or core-engine changes. |
