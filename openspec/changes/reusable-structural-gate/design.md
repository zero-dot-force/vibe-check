# Design

## Context

`.github/workflows/structural-gate.yml` (see `openspec/changes/ci-regression-gate`
and `openspec/changes/deterministic-report`) is vibe-check's build-from-source
self-gate: `pull_request`-triggered, fork-guarded, it builds `bin/vibe-check`
from the PR's own source, analyzes head and the base SHA (checked out into
`./base`), runs `diff --gate`, and appends `diff --markdown` to
`$GITHUB_STEP_SUMMARY`. It cannot be consumed via `uses:` — no `workflow_call`
trigger.

The distribution channel exists: `.goreleaser.yaml` + `release.yml` (#23)
publish per-platform archives (`vibe-check_<os>_<arch>.tar.gz`, zip on Windows)
and a SHA-256 `checksums.txt` to GitHub Releases (v0.2.0 current). The gate
semantics are stable and protected in `metrics.DecideVerdict` (ΔInstability
≥ 0.15, ΔDistance ≥ 0.20, ΔLCOM ≥ 2, new cycle → `REQUEST_CHANGES`), and
`cmd/vibe-check/diff.go` already provides `--gate` exit semantics (0/1/2),
tighten-only `--max-*-delta` overrides, and the byte-stable `--markdown`
report. `analyze` forces `GOTOOLCHAIN=local` for its `go/packages` subprocess,
so the caller module's `go.mod` must be satisfied by an installed toolchain.

Precedent: `unbound-force/gaze#300` added `reusable_gaze_ai_report.yml`
(`workflow_call`; caller uploads a coverage artifact; pinned toolchains;
threshold enforcement; step summary; fork-PR safe), dogfooded it in `test.yml`,
documented the consumer contract in `docs/guides/ci-integration.md`;
`unbound-force/unbound-force#752` consumes it pinned by immutable SHA.

## Goals / Non-Goals

**Goals:**

- A `workflow_call` reusable workflow any repo can consume via a SHA-pinned
  `uses:` reference, running the **released** binary — never building
  vibe-check from source.
- Preserve the gate contract exactly: protected thresholds by default,
  `--gate` exit semantics (REQUEST_CHANGES → job failure), tighten-only
  overrides, step-summary report.
- No secrets; minimal (`contents: read`) permissions; documented trust model.
- Continuous proof the consumer path works: a dogfood job in vibe-check's own
  CI.
- Consumer contract documented in README; website docs issue filed.

**Non-Goals:**

- Changing `structural-gate.yml` (the self-gate), its fork guard, or its
  build-from-source dogfooding — it remains vibe-check's authoritative gate.
- Any production Go source change; any change to gate thresholds or `diff`/
  `analyze` command surfaces.
- Composite action, Homebrew/Scoop install paths, or binary signing changes.
- Authoring website documentation content (only filing the issue).
- Callers triggering on events other than `pull_request` (the gate's semantics
  are PR head-vs-base; other event shapes are the caller's problem to arrange).

## Decisions

### D1: Reusable workflow (`workflow_call`), not a composite action

**Choice:** `.github/workflows/reusable_structural_gate.yml` with an
`on: workflow_call` trigger.

**Rationale:** Matches the gaze precedent. A reusable workflow is a first-class
job: it carries its own `runs-on`, `timeout-minutes`, and permissions
declaration, and callers pin it by immutable SHA at the job level — the
supply-chain property the org already relies on (`unbound-force#752` pins gaze
by SHA). A composite action would execute inside the caller's job, inheriting
its step context and permissions with no independent job identity or bounded
timeout, and would be harder for consumers to reason about and for the callee
to keep least-privileged.

**Alternatives considered:**

- **Composite action (`.github/actions/structural-gate/action.yml`)** — viable
  and simpler for callers already inside a job, but no job-level SHA pinning,
  no independent permissions/timeout, and it deviates from the established
  gaze precedent. Rejected.
- **Asking consumers to copy the self-gate YAML (status quo)** — ~60 lines of
  duplicated workflow per repo, build-from-source on every run. Rejected (the
  problem being solved).

### D2: Self-checkout snapshot acquisition, not caller-uploaded artifacts

**Choice:** The reusable workflow checks out the caller's PR head and base
(base SHA into `./base`, mirroring `structural-gate.yml`) and runs
`vibe-check analyze --output` on both itself.

**Rationale:** gaze's artifact handoff works because coverage is a byproduct of
the caller's test job — the caller uploads something it already produces. A
vibe-check snapshot can only be produced by running `vibe-check analyze`, so an
artifact-handoff design would force the caller to install vibe-check and run
`analyze` itself — reintroducing the copy-paste/install burden this change
removes. Self-checkout keeps the caller's contract to a single `uses:` line.
The PR event context (`github.event.pull_request.base.sha`) propagates into the
called workflow, so the callee can resolve both revisions without extra inputs.

**Alternatives considered:**

- **Caller-uploaded snapshot artifacts (gaze pattern)** — circular: producing
  the artifact requires the tool the workflow exists to provide. Rejected.
- **`workflow_run`-based base resolution** — more event plumbing, no benefit.
  Rejected.

### D3: Released binary via GitHub Releases download URLs, checksum-verified

**Choice:** Download `vibe-check_linux_amd64.tar.gz` and `checksums.txt` with
`curl -fsSL` from `https://github.com/zero-dot-force/vibe-check/releases/...`,
then `sha256sum --check --ignore-missing checksums.txt` before extraction.

**Rationale:** The release URLs for a public repo need no authentication and no
cross-repo token, sidestepping any `GITHUB_TOKEN` scope questions in the
caller's context. Version selection: a pinned `version` input resolves to
`releases/download/<tag>/…`; the default resolves via the permanent
`releases/latest/download/…` redirect endpoint, which needs no API call.
Checksum verification guards against corrupted or mutated artifacts and fails
the job before any analysis runs. The job pins `runs-on: ubuntu-latest`, so the
asset name is constant and the verification tooling (`sha256sum`, `tar`) is
guaranteed present.

**Alternatives considered:**

- **`gh release download`** — requires a token with read access and adds a CLI
  dependency for what a plain HTTPS GET does anonymously. Rejected.
- **Verifying cosign signatures** — the releases are cosign-signed via
  org-infra, but `checksums.txt` is the documented verification channel
  (README) and signature verification would add cosign installation to every
  consumer run. Possible future hardening; out of scope.

### D4: `version` input semantics (release tag; default latest)

**Choice:** `version` is a release tag (e.g. `v0.2.1`); the literal default
`latest` selects the newest published release. README guidance recommends
pinning an explicit tag for reproducible gates, mirroring the SHA-pinning
philosophy applied to the workflow reference itself.

**Rationale:** Defaulting to `latest` gives zero-config adoption; explicit tags
give determinism. A missing/nonexistent tag makes `curl` fail (`-f`) and the
job fails before analysis — no silent fallback.

### D5: Go toolchain via `setup-go` with a `go-version-file` input

**Choice:** `actions/setup-go` (SHA-pinned, reusing the SHAs already verified
in `structural-gate.yml`) with `go-version-file: ${{ inputs.go-version-file }}`
(default `go.mod`, relative to the head checkout root).

**Rationale:** `analyze` forces `GOTOOLCHAIN=local` — it never downloads a
toolchain named by the target module's `go.mod` — so the workflow must install
a satisfying toolchain first. `go-version-file` reads the caller's declared
version, exactly like the self-gate. Monorepo callers pass their module's
`go.mod` path. The base checkout is analyzed with the same installed toolchain;
if the base module would need a newer Go than the head's `go.mod` names (e.g.
the PR lowers the `go` directive), the base analysis fails as a visible tool
failure rather than silently succeeding — acceptable and rare.

**Alternatives considered:**

- **`go-version` input with a raw version** — duplicates what `go.mod` already
  declares; drifts. `go-version-file` is strictly better. A raw-version input
  remains possible for callers via a `go.mod` they control; not needed.

### D6: No fork guard; trust model documented in the header

**Choice:** The reusable workflow imposes no fork guard and documents its trust
model in the workflow header comment: `vibe-check analyze` compiles the
caller's target module, which is the same trust posture as running Go tests on
a fork PR in the consumer's CI; the gate itself runs only the released,
checksum-verified binary (trusted code), and the job holds a read-only,
no-secrets token.

**Rationale:** The self-gate's fork guard exists because that workflow builds
vibe-check from the PR's own source inside vibe-check's repo. Here the binary
is the immutable release artifact, so consuming the released binary removes the
build-untrusted-source concern entirely; compiling the caller's PR code is a
posture every CI that runs `go test` already accepts. No secrets means fork
PRs work with zero configuration. Callers who want to skip forks can add their
own job-level `if:`.

### D7: Permissions — callee declares `contents: read`; consumer contract covers default-deny callers

**Choice:** The called job declares `permissions: contents: read` (needed for
the two `actions/checkout` steps; the release download needs no token). The
README consumer contract documents that callers using a top-level
`permissions: {}` (default-deny) must grant the job-level `contents: read` the
callee declares — otherwise the called job fails at startup — and that callers
must never grant more than the callee declares at their pinned SHA.

**Rationale:** A called workflow cannot elevate beyond the caller's grant, but
it runs with whatever the caller's job grants; least-privilege is maintained by
declaring exactly what is needed and teaching the mirror-grant pattern already
used in `release.yml` and recorded as a prior learning (empty top-level
permissions cause `startup_failure` before any step runs).

### D8: File name `reusable_structural_gate.yml` (CI-010) over the issue's suggested hyphen

**Choice:** `.github/workflows/reusable_structural_gate.yml`.

**Rationale:** Convention CI-010 [MUST] requires the `reusable_` underscore
prefix for reusable workflows; the issue's suggested
`reusable-structural-gate.yml` is superseded by that rule. The gaze precedent
(`reusable_gaze_ai_report.yml`) uses the same convention. The change's
OpenSpec directory name (`reusable-structural-gate`) is unaffected — OpenSpec
uses kebab-case paths regardless.

### D9: Threshold overrides forwarded, tighten-only enforcement left to `diff`

**Choice:** Optional inputs `max-instability-delta`, `max-distance-delta`, and
`max-lcom-delta` are forwarded to `diff --gate` only when the caller sets them
(the run step assembles flag arguments conditionally). The workflow performs no
threshold validation of its own.

**Rationale:** `cmd/vibe-check/diff.go` already enforces tighten-only semantics
(a looser value exits 2 with no stdout), and that enforcement is a protected
gate. Reimplementing it in workflow bash would duplicate (and risk weakening)
a protected check — DRY and gatekeeping both forbid it. An unset input must not
forward its default value, because forwarding the default explicitly is
harmless but forwarding nothing keeps the diff invocation identical to the
self-gate's.

### D10: Dogfood job in `ci.yml` via same-repo `uses:`; self-gate untouched

**Choice:** Add a job to `.github/workflows/ci.yml` that calls the reusable
workflow with a same-repository local reference
(`zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@<ref>`
resolved as `./.github/workflows/reusable_structural_gate.yml` for same-repo
callers), running on every PR alongside `lint`/`test`.

**Rationale:** A reusable workflow's `uses:` filename and `with:` inputs are
validated by GitHub only at dispatch time — never at commit or PR-CI time — so
inspection alone never proves the consumer path. A dogfood job dispatches the
workflow end-to-end on every vibe-check PR, continuously proving resolution,
input wiring, download, and gate execution (the gaze precedent dogfoods in
`test.yml`). It is additive: `structural-gate.yml` keeps building vibe-check
from the PR's own source, so analyzer changes made in a PR are still gated by
that PR's own build — the released binary used by the dogfood job lags by
design, and that is acceptable because the dogfood job's purpose is consumer
smoke-testing, not gating.

**Alternatives considered:**

- **Replace the self-gate with the reusable workflow** — loses PR-source
  dogfooding and would gate vibe-check with a stale released binary. Rejected.
- **No dogfood; wait for an external consumer** — leaves the consumer contract
  unproven until someone else adopts it. Rejected.

### D11: Bounded execution and step shape

**Choice:** Job-level `timeout-minutes` (self-gate parity, 25) bounds the whole
job; the `timeout` input (duration string, default `5m`) is forwarded to both
`analyze` invocations, mirroring the self-gate. Step order: install (download
+ verify + extract) → checkout head → checkout base → `setup-go` → analyze
head → analyze base → gate (`diff --gate`, forwarding any override inputs) →
report (`diff --markdown >> $GITHUB_STEP_SUMMARY`, `if: always()`,
`continue-on-error: true`, plain `--markdown` without `--gate` so a
REQUEST_CHANGES verdict still exits 0 in the render step).

**Rationale:** The report step reuses the exact semantics proven in
`structural-gate.yml` and the deterministic-report learning: plain
`diff --markdown` exits 0 even on REQUEST_CHANGES (the verdict travels in the
payload without `--gate`), exit 2 (tool failure) writes no stdout so nothing is
appended, and `continue-on-error` covers the exit-2 render-failure case while
the gate step remains the authoritative conclusion.

## Risks / Trade-offs

- **[Default-deny callers hit `startup_failure`]** → README consumer contract
  documents the required job-level `contents: read` grant (D7); the workflow
  header repeats it. The dogfood job proves the happy path every PR.
- **[Dispatch-time-only validation of `uses:`/`with:`]** → the dogfood job
  dispatches the workflow end-to-end on every PR (D10), so regressions in the
  contract surface immediately in vibe-check's own CI rather than in a
  consumer's.
- **[Checksum file shares the release transport]** → `checksums.txt` verifies
  artifact integrity, not release-account compromise; version pinning (D4) and
  the org's cosign signing of releases bound the blast radius. Documented as
  guidance to pin versions.
- **[Caller module must analyze on linux/amd64]** → the gate job pins
  `runs-on: ubuntu-latest` (constant asset name, guaranteed `sha256sum`/`tar`);
  modules that only build on another OS fail visibly. Same constraint as the
  self-gate; acceptable, documented.
- **[Base analysis under head's toolchain]** → if a PR lowers the `go`
  directive below what base needs, base analysis fails as a tool failure
  (D5). Visible, rare, and honest — preferable to silent toolchain downloads.
- **[Latest-by-default drift]** → `version: latest` silently moves with
  releases; README recommends pinning a tag, mirroring SHA-pin guidance for the
  `uses:` reference itself.

## Migration Plan

Purely additive: one new workflow file, one additive `ci.yml` job, README and
CHANGELOG updates, one external docs issue. No production source, schema, or
gate-threshold changes; no consumer coordination required (no one consumes the
new workflow until they choose to). Rollback is deleting the dogfood job and
the workflow file; the self-gate is never touched.
