# Design

## Context

See `proposal.md` — Why. The current state that shapes this design:

- `internal/scaffold` deploys `.opencode/` assets via a `categories` table
  (`assets/agents` → `.opencode/agents`, `assets/commands` → `.opencode/commands`),
  each entry globbed as `*.md`, written through symlink-safe helpers (`ensureDir`
  with refusal + containment checks, `regularFileExists`, `writeFile` seam) with
  skip/force semantics, then performs UF policy registration. `Options` carries
  `TargetDir`, `Force`, and the `WriteFile` test seam; `Result` reports
  `Written`/`Skipped`/`Forced` with category-prefixed paths plus `Notices`.
- `cmd/vibe-check/init.go` follows AP-001/AP-002: `InitOptions`/`InitResult`
  around a testable `RunInit`; `writeInitSummary` renders the human summary,
  `writeInitJSON` the `written`/`skipped`/`forced`/`notices` payload; every
  failure path exits 2 (`exitCodeError` plumbing in `main.go`).
- `cmd/vibe-check/root.go` already provides `semanticVersion()`: ldflags-injected
  `version` → `debug.ReadBuildInfo` main-module-version fallback, so
  `go install ...@v0.3.0` yields `v0.3.0`. Dev builds yield `dev` or `(devel)`.
- The callee — `reusable_structural_gate.yml` (shipped v0.3.0) — declares 7 string
  inputs (`version` default `latest`, `path` default `.`, `timeout` default `5m`,
  `go-version-file` default `go.mod`, three tighten-only `max-*-delta` overrides),
  `permissions: contents: read` only, and requires a `pull_request`-triggered caller
  (`github.event.pull_request.base.sha`). Existing release tags: v0.1.0, v0.2.0,
  v0.3.0 — a caller pinned below v0.3.0 fails at dispatch with "workflow not found".
- `go.mod`: `golang.org/x/mod v0.39.0 // indirect`;
  `go.yaml.in/yaml/v3 v3.0.4` is already in the module graph (go.sum).
- Precedent: unbound-force/gaze#306 (`gaze init --ci`) — the pattern being ported.

## Goals / Non-Goals

**Goals:**

- One command — `vibe-check init --ci .` — yields a working, correctly pinned,
  advisory structural gate on pull requests for any repository.
- Correct-by-construction pinning: the generator resolves its own release tag and
  fails closed rather than emitting a caller pinned to a tag that lacks the
  reusable workflow.
- The generated file is user property: never overwritten, even under `--force`.
- Plain `vibe-check init` stays byte-identical (no regression risk to the existing
  surface).
- The caller is forward-compatible with the planned Python/TypeScript adapters.
- The repo's own self-gate and the generated caller share one file name.

**Non-Goals:**

- `--ci-only` variant (skip `.opencode/` scaffolding) — deliberately rejected,
  matching gaze #306.
- A `/vibe-check thresholds` guided-threshold-authoring command (gaze phase-2 idea).
- Monorepo git-root detection — root-only assumption, covered by the D8 notice and
  noted as a known limitation in the README.
- Changing the reusable workflow's input contract, thresholds, or any `metrics/`
  or `cmd/vibe-check` gate semantics.
- YAML surgery on existing consumer workflows, test jobs, artifacts, or secrets in
  the generated caller.

## Decisions

The maintainer-settled decisions D1–D9 (issue #77, 2026-10-31) are the contract;
they are recorded here with their rationale. D3 and D6 carry maintainer-approved
amendments from the `/uf.unleash` spec-review decision gate (recorded in place,
below — this is the change's decision log). D10–D14 are implementation decisions
this design adds.

### D1: Additive `--ci` flag

**Choice:** `--ci` additionally emits the caller workflow; plain `vibe-check init`
behavior stays byte-identical. No `--ci-only` variant.

**Rationale:** Zero regression risk to the existing surface; the flag composes with
`--force` and `--json` exactly as today. `--ci-only` would double the flag matrix for
a marginal convenience — gaze #306 deliberately skipped it too.

### D2: Self-contained starter

**Choice:** The generated workflow is a single job calling the reusable workflow —
no YAML surgery on existing consumer workflows, no test job, no artifacts, no
secrets.

**Rationale:** vibe-check's gate is far simpler than gaze's (the callee installs the
released binary and checks out both refs itself), so a minimal known-good shape is
the whole value: correct permissions, correct trigger context, correct pinning.
**Alternatives:** a multi-job caller with a caller-side test job (rejected — the
callee owns everything; extra jobs are consumer choices); emitting a patch to edit an
existing workflow (rejected — fragile YAML surgery).

### D3: SHA for the ref, tag for the version input

> **Maintainer amendment (2026-10-31, via the `/uf.unleash` spec-review
> decision gate — recorded here as the decision log).** The original D3 pinned
> both the `uses:` ref and the `with: version:` input to the same release tag.
> The spec review's Adversary finding (supply-chain: a mutable tag is
> name-addressed, not content-addressed, and can be repointed) and Guard
> finding (constitution-alignment: tag pinning violates CI-001 [MUST] and
> Constitution Principle V, and the rejection rationale rested on a false
> premise — `.goreleaser.yaml` already injects `-X main.commit={{.Commit}}`, so
> release binaries know their release commit) together showed the compliant
> SHA+tag form is achievable. The maintainer approved amending D3 to SHA+tag;
> this is maintainer-approved input on a D1–D9 decision.

**Choice:** The `uses:` ref is pinned to the full 40-character commit SHA of
the release; the `with: version:` input stays pinned to the release tag.

**Rationale:** A commit SHA is content-addressed — immutable and reviewable —
which is exactly what CI-001 [MUST], Constitution Principle V ("CI pipelines
MUST pin actions and reusable workflows by commit SHA, not mutable tags"), and
the README's published consumer guidance ("### Version pinning": "Pin a full
commit SHA — immutable and reviewable") require; the emitted artifact no longer
contradicts published guidance. Workflow logic and analyzer binary still move
together: the SHA is the release tag's own commit (resolved by peeling the tag
via `refs/tags/<tag>^{}`), so both pins name the same release.

**SHA resolution per distribution channel:**

- **Release binary (the default path, fully offline):** the ldflags-injected
  `main.commit` (`.goreleaser.yaml` `-X main.commit={{.Commit}}`, declared in
  `cmd/vibe-check/main.go`) — a release binary knows its own release commit
  with no network access.
- **`go install` and `--vibe-version` paths:** the tag is known but the commit
  is not (module-proxy builds set no ldflags, so `commit` stays `"none"`;
  an override names a release other than the binary's own, so the embedded
  commit does not apply) — the commit is resolved at init time via
  `git ls-remote refs/tags/<tag>^{}`. When offline this fails closed with an
  error shaped like the D6 contract.

**Alternative (the original D3, rejected):** pinning both refs to the same
tag. It solves tag *existence* (the installer knows its own release) but not
ref *integrity*: a tag can be repointed (repo compromise, maintainer-account
takeover), silently redirecting consumer CI to different workflow logic under
`contents: read`. The original rejection of SHA pinning — "the installer
cannot know an arbitrary commit SHA's relationship to that release" — was true
only for the `go install` channel; the release-binary channel carries the
commit via ldflags, and the remaining channels can resolve it at init time.

### D4: Advisory defaults; thresholds are a human decision

**Choice:** No threshold inputs are emitted. A commented-out `TODO` block shows the
three tighten-only overrides. Protected defaults (0.15 / 0.20 / 2) apply while
unset; looser values are ignored by the gate.

**Rationale:** No tool picks governance gates (protected-threshold policy); the
comment teaches the tightening path without deciding it. **Alternative:** emitting
the three inputs commented-in at protected values (rejected — redundant with the
callee's defaults and implies the generator endorses specific numbers).

### D5: File name `vibe-check-structural-gate.yaml`

**Choice:** Both the generated caller and this repo's own self-gate (currently
`.github/workflows/structural-gate.yml`) use this name. The self-gate rename is a
file rename only — its `name:` field (`Structural Gate — Design Quality
Regression`) and job names stay unchanged.

**Rationale:** One name to teach. Branch protection keys on check names (from
`name:`/job names, not the file name) and Actions run history is unaffected, so the
rename is behavior-free. `git mv` preserves file history.

### D6: Version resolution from the binary itself, fail closed

> **Maintainer amendment (2026-10-31, spec-review decision gate):** D6
> originally resolved only the tag. It is amended (with D3) to resolve the
> release commit SHA as well — offline via the ldflags-injected commit on the
> release-binary path, via `git ls-remote refs/tags/<tag>^{}` at init time
> otherwise — so both pins are resolved and validated before any writes.

**Choice:** Resolve the tag via `semanticVersion()` (ldflags →
`debug.ReadBuildInfo` fallback; `go install ...@v0.3.0` yields `v0.3.0`).
Resolve the release commit SHA per D3's channel table — the ldflags-injected
`main.commit` when it is a full 40-hex SHA and no `--vibe-version` override is
present (release-binary path, fully offline), else `git ls-remote
refs/tags/<tag>^{}` at init time (`go install` and override paths). Validate
both with `golang.org/x/mod/semver` plus a hex check: the tag must be a valid
`v`-prefixed semver `>= v0.3.0`, the SHA a full 40-character hex string. Dev
builds, unparseable values, below-floor versions, or an unresolvable SHA
(offline `git ls-remote`) exit **2** with messages shaped like:

```
cannot resolve a vibe-check release tag >= v0.3.0 (got %q) — install vibe-check from a release tag or pass --vibe-version
cannot resolve the release commit SHA for tag %q (%v) — install vibe-check from a release binary or rerun with network access
```

Validation happens **before any filesystem writes** (atomic — no partial state).
`--vibe-version` override carries the same validation (its SHA is always
resolved via `git ls-remote`, since the embedded commit belongs to the binary's
own release, not the override); passing it without `--ci` is a usage error.

**Rationale:** The biggest copy-paste footgun is pinning a tag that predates the
reusable workflow; an installer that knows its own version eliminates the failure
class. The floor `v0.3.0` is a protected gate value (first release containing the
callee) — it may only be raised deliberately, never to make a build pass. SHA
resolution closes the one link in the generated pipeline that had no content
verification: a mutable tag can be repointed, a commit SHA cannot.
**Alternatives:** `latest` pin (rejected — not reproducible); asking the user for a
tag interactively (rejected — non-deterministic installer); accepting any semver
without a floor (rejected — reintroduces the run-time failure the feature exists to
remove); a tag-pinned `uses:` ref with no SHA resolution (rejected per the D3
amendment — violates CI-001/Principle V).

### D7: Skip-if-present, even under `--force`

**Choice:** The generated file is never clobbered. `--force` continues to apply
only to `.opencode/` asset overwrite (existing documented semantics). A rerun with
the file present reports it as skipped.

**Rationale:** Users edit thresholds and pins in this file; overwriting would
destroy human decisions. **Alternative:** `--force` overriding the caller too
(rejected — `--force` is documented as `.opencode/`-scoped and has no
`--force-ci` companion; changing that meaning silently is a footgun).

### D8: Forward-compatible with future language adapters

**Choice:** The generated caller is language-neutral (passes no `go-version-file`;
the callee's `go.mod` default applies today). Manifest detection is a small seam —
`go.mod` → go today; future entries (`pyproject.toml` → python,
`package.json` → ts) extend a table, not a redesign. No manifest found at the
target root → notice + still emit, e.g.:

```
no Go module manifest (go.mod) at target root — the reusable gate currently analyzes Go modules; Python/TypeScript adapters will reuse this same caller workflow.
```

**Rationale:** RFC P2/P3 plan Python and TS adapters; the caller must not hardcode
Go beyond what the callee's current defaults already assume. **Alternative:**
emitting `go-version-file: go.mod` explicitly (rejected — it would need editing
when adapters land, and the D8 notice covers the non-Go case).

### D9: Pull-request-only trigger

**Choice:** `on: pull_request: branches: [main]`, a caller-owned concurrency group
(CI-012 pattern), and `permissions: contents: read` (exactly the callee's declared
scope, CI-021). No event-shape `if:` guard and no fork guard in the caller.

**Rationale:** The gate requires `github.event.pull_request.base.sha`, which a
`pull_request`-only trigger always provides. The callee is fork-safe by design
(read-only token, `persist-credentials: false`, no secrets), so a fork guard would
only reduce coverage. **Alternative:** also triggering on `push` with an event
guard (rejected — vacuous-pass hazard the README's "Trigger context" section
exists to prevent).

**Known limitation (documented per the spec-review decision, not a D9 change):**
the `branches: [main]` filter is hardcoded — GitHub's branch filter cannot
express "the repository's default branch", and a static, offline-capable
template cannot discover it. Repositories whose default branch is `master`,
`trunk`, or `develop` must edit `branches:` or the gate never triggers (the
same silent-no-op class the vacuous-pass guard above exists to prevent, which
is why the limitation is surfaced rather than left implicit). It is documented
in the README's known-limitations entry and in the init next-steps output
("edit `branches:` when your default branch is not `main`") rather than made
adaptive: template adaptivity would require repo introspection (network or
`gh`) at init time, which the offline release-binary path does not have.

### D10: Template + placeholder substitution rendering

**Choice:** The template lives at
`internal/scaffold/assets/ci/vibe-check-structural-gate.yaml` with two distinct
placeholders (amended per D3) — `__VIBE_CHECK_VERSION_SHA__`, substituted
exactly once into the `uses:` ref, and `__VIBE_CHECK_VERSION_TAG__`,
substituted exactly once into the `with: version:` pin — rendered by simple
string substitution. Draft (the issue's YAML body, amended to carry the
two-pin shape):

```yaml
# vibe-check structural gate caller — generated by `vibe-check init --ci`.
# This file is yours: vibe-check never overwrites it (reruns report "skipped").
# Edit freely to uncomment tighten-only thresholds, bump the pinned release,
# or adjust `branches:` when your default branch is not `main`.
# Docs: https://github.com/zero-dot-force/vibe-check#ci-integration
name: vibe-check structural gate
on:
  pull_request:
    branches: [main]
concurrency:
  group: vibe-check-structural-gate-${{ github.ref }}
  cancel-in-progress: true
permissions:
  contents: read          # exactly the callee's declared scope (CI-021)
jobs:
  structural-gate:
    # Two pins, one release: `uses:` carries the release commit SHA (immutable);
    # `with: version:` carries the release tag — see README "### Version pinning".
    uses: zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@__VIBE_CHECK_VERSION_SHA__
    with:
      version: "__VIBE_CHECK_VERSION_TAG__"
      # TODO: tighten-only overrides — uncomment to tighten beyond the protected
      # defaults (instability 0.15 / distance 0.20 / LCOM 2). Values must be
      # STRICTLY tighter; looser values are ignored by the gate.
      # max-instability-delta: "0.10"
      # max-distance-delta: "0.15"
      # max-lcom-delta: "1"
```

**Rationale:** Comments are load-bearing in this artifact (ownership, pinning
shape, tightening guidance); a template engine or struct→YAML marshaling would
drop them. The placeholder set is closed and tiny — one per pin, each
substituted once, so the two pins cannot drift apart in rendering.
**Alternatives:** `text/template` (rejected — unnecessary power for two
substitutions); `yaml.Marshal` from a struct (rejected — loses comments and key
order).

### D11: Separate embed filesystem, excluded from mirror semantics

**Choice:** `internal/scaffold/embed.go` gains `//go:embed assets/ci/*.yaml` →
`ciAssetsFS`. The `ci/` subtree is **excluded** from the `categories` table,
`assetPaths()`, `assetContent()`, and `TestEmbeddedAssetsMatchSource` — CI
templates have no `.opencode/` mirror. A new `deployCIWorkflow` path reuses the
existing helpers (`ensureDir` with symlink refusal + containment,
`regularFileExists`, the `WriteFile` seam, 0o644 files) and lists the file with a
`workflows/` prefix in `Written`/`Skipped`, so the JSON payload and summary lists
carry it with no payload-schema change. The CI deployment runs **before** the
`.opencode/` categories loop (spec-review decision, write order): the most
likely *new* failure mode in an existing repository is a `.github` path conflict
(a regular file at `.github`, a symlinked `workflows/`), and deploying the
caller first fails it before any `.opencode/` mutation — the cheapest
mitigation for the partial-state window that D6's atomicity guarantee
deliberately does not cover.

**Rationale:** The categories loop globs `*.md` and mirrors into `.opencode/`;
CI templates deploy to `.github/workflows/` and are user-editable property (D7) —
different target, different overwrite semantics, one shared writer vocabulary.
**Alternatives:** adding a `category` entry for CI (rejected — would entangle
`.opencode/` mirroring, `--force` semantics, and the `*.md` glob); a separate
package (rejected — the helpers are the point of reuse).

### D12: Validation placement and exit-code mapping

**Choice:**

- `RunInit` (the testable entry point) rejects `--vibe-version` without `--ci` as
  a usage error before any work: an `exitCodeError` with code 2 whose message
  names the flag combination — keeping the command's documented 0/2 exit contract
  (`SilenceUsage` is already the house pattern; the message is the usage
  guidance). Cobra's conventional usage-error exit 1 was considered and rejected
  because the init command documents only 0 and 2.
- `init.go` resolves **both pins** when `--ci` is set (amended per D3/D6): the
  tag — the override if non-empty, else `semanticVersion()` — and the release
  commit SHA — the ldflags-injected `main.commit` when it is a full 40-hex SHA
  and no override is present (release-binary path, offline), else
  `git ls-remote refs/tags/<tag>^{}` resolved at init time. Both are wired
  into `scaffold.Options.VibeVersion` and `scaffold.Options.VibeCommit`
  alongside `Options.CI`. `internal/scaffold` cannot import package `main`, so
  both resolutions live in the CLI layer; the tag→SHA resolution is an
  injectable seam on the init command's options so `RunInit` stays
  unit-testable offline.
- `internal/scaffold` owns validation (`golang.org/x/mod/semver.IsValid` +
  floor comparison for the tag; a full 40-character hex check for the SHA) in
  its run path, executed **before any writes** when `Options.CI` is set. This
  guarantees the no-partial-state property regardless of caller, and keeps the
  validation unit-testable in-package. On failure the existing error mapping
  yields exit 2 with the D6 message and nothing on stdout.

**Rationale:** Each layer owns what it can enforce: the CLI owns flag-combination
rules and pin *resolution*; the scaffold owns the write-atomicity invariant.
**Alternative:** validating only in `init.go` (rejected — `scaffold.Run` would be
able to emit a caller with an unvalidated pin).

### D13: Dependency promotions (no new module requirements)

**Choice:** Promote `golang.org/x/mod` (semver validation, runtime) from
indirect to direct, and add `go.yaml.in/yaml/v3` (test-only YAML parse of the
rendered output) as a direct require. No version changes; no new module
*requirements*.

**Corrected dependency facts** (per the spec-review factual-accuracy and
dependency-declaration findings; the original D13 claim of "no new module
downloads" was wrong for yaml/v3): `golang.org/x/mod` is listed indirect in
`go.mod` and its module zip hash (`h1:`) is already in `go.sum` — promoting it
downloads nothing. `go.yaml.in/yaml/v3` is in the module *graph* (its
`/go.mod` hash is recorded in `go.sum`, reached via `golang.org/x/tools`) but
is absent from `go.mod`, and its module zip has never been fetched — `go.sum`
carries no `h1:` line for it. The test-only import therefore performs one
module-zip download at first build and adds a full `h1:` hash line to
`go.sum`; task 2.1's verification expects that go.sum delta rather than
assuming "no new downloads".

**Rationale:** Honest dependency declaration for code the repo now imports
directly; `go.yaml.in/yaml/v3` validates template syntax in tests without adding
an actionlint dependency to CI. **Alternative:** a separate test-only module for
yaml (rejected — overkill for one import).

### D14: Self-gate rename mechanics

**Choice:** `git mv .github/workflows/structural-gate.yml
.github/workflows/vibe-check-structural-gate.yaml` — contents byte-identical
(`name:` field and job names unchanged). Update path references at:
`AGENTS.md` project-structure listing (line ~291), `README.md` link (line ~234),
`ci.yml` comments (lines ~4 and ~62), `reusable_structural_gate.yml` comment
(line ~173). Historical artifacts deliberately untouched: archived/unarchived
`openspec/changes/*`, old CHANGELOG entries, `.uf/dewey/learnings/`. CHANGELOG
gains a `Changed` entry for the rename alongside the `Added` entry for the
feature.

**Rationale:** `git mv` preserves history; check names come from `name:`/job
names, so branch protection and Actions history are unaffected. Rewriting history
references would falsify the record.

## Implementation surface

Verified against the tree at design time; adapt if the code has drifted, without
changing D1–D9:

- `internal/scaffold/scaffold.go` — add `Options.CI`, `Options.VibeVersion`,
  and `Options.VibeCommit`; pin validation (D6/D12) before any writes;
  `deployCIWorkflow` reusing the existing helpers (running before the
  `.opencode/` categories loop per D11); manifest-detection seam + notice (D8).
  Target: `.github/workflows/vibe-check-structural-gate.yaml` under the target
  root, listed with a `workflows/` prefix in `Written`/`Skipped`.
- `internal/scaffold/embed.go` — `//go:embed assets/ci/*.yaml` → `ciAssetsFS`
  (excluded from the categories table, `assetPaths()`, `assetContent()`, and
  `TestEmbeddedAssetsMatchSource` per D11).
- `internal/scaffold/assets/ci/vibe-check-structural-gate.yaml` — the template
  (D10).
- `cmd/vibe-check/init.go` — `--ci` and `--vibe-version` flags; usage-error
  check; tag + commit-SHA wiring (D12, with the tag→SHA resolution as an
  injectable seam); new CI next-steps section in the summary output (commit the
  workflow; gate is advisory until branch protection marks the check required;
  uncomment thresholds to tighten; edit `branches:` when the default branch is
  not `main`; bump pins by editing the file since reruns skip); `--json` payload
  includes the workflow path in `written`/`skipped` via the existing arrays.
- `go.mod` / `go.sum` — the two indirect → direct promotions (D13).

## Risks / Trade-offs

- [Template drifts from the README consumer example or the callee's contract] →
  golden-file test pins the rendered output; YAML parse test validates syntax;
  `actionlint` over the golden and the renamed self-gate is a task verification
  step (precedent: reusable-structural-gate task 4.2).
- [`go install` from a branch or commit yields `dev`/`(devel)` → scaffolding
  fails closed] → intended (D6); the error message states both remediations
  (install from a release tag, or pass `--vibe-version`). Trade-off accepted: a
  dev build refusing to scaffold is better than a caller pinned to a nonexistent
  relationship between ref and release.
- [`go install` build or `--vibe-version` override run offline → tag→SHA
  resolution fails] → intended (D3/D6 fail-closed); the error names the tag and
  the remediation (install from a release binary, or rerun with network access).
  Trade-off accepted: a refused scaffold is better than a mutable-tag `uses:`
  ref with no content verification.
- [Version floor goes stale relative to new releases] → the floor is a protected
  gate value; raising it is a deliberate, reviewed change; it must never be
  lowered to unblock a build.
- [Rename breaks external deep links to the old path] → GitHub serves web
  redirects for renamed files; in-repo references are updated (D14); historical
  artifacts intentionally keep old paths.
- [Partial writes on non-version I/O failures] → only the version-validation
  error is mandated atomic (D6); other write failures keep the existing per-file
  semantics — unchanged risk, documented.
- [`go.yaml.in/yaml/v3` direct but test-only] → honest declaration, no runtime
  cost; alternative rejected as overkill (D13).
- [Consumer repo has an existing differently-named caller workflow] → out of
  scope; skip-if-present (D7) guards only the vibe-check-emitted name.

## Migration Plan

- Consumers: nothing to migrate — the flag is additive; existing callers of the
  reusable workflow are unaffected.
- Self-gate rename: single `git mv` plus reference updates in the same change;
  branch protection and check history unaffected (D5/D14).
- Rollback: revert the change; consumer-side generated files are user-owned and
  unaffected by a vibe-check rollback.

## Test strategy

Per the issue's test list (repo conventions: standard library `testing`,
table-driven, `t.TempDir()`, testdata goldens):

- **Pin-resolution matrix**: `dev` / empty / no-`v`-prefix / garbage /
  `v0.2.9` tag → error; `v0.3.0` / `v0.4.1` with a 40-hex SHA → ok; missing /
  short / non-hex SHA → error; `--vibe-version` override honored (SHA resolved
  for the override tag); offline tag→SHA resolution failure → exit 2 with zero
  writes; override without `--ci` → usage error.
- **Golden-file test** of the rendered workflow (testdata golden, repo
  convention).
- **YAML parse** of rendered output (`go.yaml.in/yaml/v3`, test-only) — validates
  template syntax without a new module.
- **Skip-if-present rerun** — with and without `--force`, content unchanged.
- **Plain `vibe-check init` writes nothing under `.github/`**.
- **Version-error atomicity** — invalid version + `--ci` errors with zero files
  written.
- **No-`go.mod` notice** — notice emitted, workflow still written.
- **`assetPaths()` excludes `ci/`** (guard for the mirror test);
  `TestEmbeddedAssetsMatchSource` continues to cover only agents/commands.
- **JSON payload includes the workflow in written/skipped.**
- **Manual verification step in tasks**: `actionlint` over the generated golden
  and the renamed self-gate.
- **CI parity** (derived from `.github/workflows/ci.yml`, run locally):
  `go build ./...`, `go vet ./...`,
  `go test -race -count=1 -coverprofile=coverage.out ./...`,
  `go run ./cmd/coverage-ratchet --profile coverage.out --floors coverage-floors.json`,
  golangci-lint (v2.12.2 per the CI action pin). All touched packages already
  have committed coverage floors; floors must not drop.
