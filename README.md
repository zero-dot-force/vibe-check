# Vibe-Check

Vibe-Check is a design-quality and architectural-metrics tool for Go. It computes the
Martin package-coupling metrics suite — afferent coupling (Ca), efferent coupling (Ce),
instability, abstractness, and distance from the main sequence — plus LCOM4 cohesion and
circular-dependency detection, and emits the results as JSON.

## Install

```bash
go install github.com/zero-dot-force/vibe-check/cmd/vibe-check@v0.2.0
```

This resolves against the `v0.2.0` release tag, which must be published in the repository
for module-version installs to work. When a binary is built this way (without release
`ldflags`), `--version` falls back to the toolchain's build info so it still reports a
meaningful module version (commit and build date may show as `none`/`unknown` for
module-cache installs).

### Pre-built binaries

Download the latest binary for your platform from
[GitHub Releases](https://github.com/zero-dot-force/vibe-check/releases).

Binaries are available for linux, darwin, and windows on amd64 and arm64.
Each release includes a `checksums.txt` file for artifact verification.

**macOS users**: downloaded binaries are not notarized. If Gatekeeper blocks
the binary, right-click → Open in Finder, or run:

```bash
xattr -d com.apple.quarantine ./vibe-check
```

To build from a checkout instead:

```bash
go build ./cmd/vibe-check
```

## Usage

```bash
vibe-check analyze [path]
```

`analyze` takes a single optional path to a Go module directory and defaults to the
current directory (`.`). It writes the analysis as JSON to stdout, or to a file with
`--output`/`-o`. Analysis forces `GOTOOLCHAIN=local` for its internal `go/packages`
subprocess, so it never downloads a Go toolchain named by the target module's `go.mod`
(a trusted module that requires a newer-than-local toolchain must be built manually).

```bash
# Analyze the current module
vibe-check analyze

# Analyze a module elsewhere, failing CI if any package is too unstable
vibe-check analyze ./myproject --max-instability 0.8 --no-circular-deps
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--max-instability` | float | unset (no check) | Fail if any module's instability exceeds this value. Must be in `[0.0, 1.0]`. |
| `--max-distance` | float | unset (no check) | Fail if any module's distance from the main sequence exceeds this value. Must be in `[0.0, 1.0]`. |
| `--max-lcom` | int | unset (no check) | Fail if any module's LCOM4 exceeds this value. Must be `>= 1`. |
| `--no-circular-deps` | bool | `false` | Treat any detected circular dependency as a violation. |
| `--no-provenance` | bool | `false` | Omit the `provenance` object from output for byte-reproducible JSON. |
| `--timeout` | duration | none | Bound total analysis time (e.g., `30s`, `2m`). No timeout by default. |
| `--output`, `-o` | string | stdout | Write the ModuleGraph JSON to a file instead of stdout. Exits `2` (with no partial stdout) if the file cannot be written. |
| `--version` | — | — | Print version, commit, and build date, then exit. Use on the root command: `vibe-check --version`. |

Threshold comparisons use strict greater-than: a metric exactly equal to the threshold
**passes**. Invalid flag values (for example, out of range) exit with code `2` before any
analysis runs.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success — analysis completed with no threshold violations. |
| `1` | Policy failure — analysis succeeded, but one or more thresholds were exceeded. |
| `2` | Tool failure — invalid arguments/flags, adapter error, timeout, or interrupt signal. |

Violation messages are written to stderr; the JSON report is still written to stdout even
when violations cause a non-zero exit, so CI can capture the full results while failing the
gate.

## Output

The analysis is printed to stdout as pretty-printed JSON with `schemaVersion` `"1.2"`. Each
entry in `modules[]` carries `ca`, `ce`, `instability`, `abstractness`, `distance`, `lcom`,
`exportedTypes`, `abstractTypes`, `zone`, and an optional `extensions` object. For Go, the
extensions may include `go.interfaceWidth` (method count per exported interface) and
`go.interfaceProximity` (`"producer"` or `"consumer"` per interface). Detected cycles appear
in `cycles` as lexicographically-sorted sets of package paths.

A top-level `provenance` object records how the graph was produced: `producer` (`"vibe-check"`),
`version` (the tool version), `generatedAt` (an RFC 3339 UTC timestamp), and `input`
(`path` and `modulePath`). Pass `--no-provenance` to omit it for byte-reproducible output.

```json
{
  "schemaVersion": "1.2",
  "language": "go",
  "modules": [
    {
      "path": "github.com/you/proj/pkg",
      "name": "pkg",
      "ca": 2,
      "ce": 5,
      "instability": 0.71,
      "abstractness": 0.25,
      "distance": 0.04,
      "lcom": 1,
      "exportedTypes": 4,
      "abstractTypes": 1,
      "zone": "main-sequence",
      "extensions": {
        "go.interfaceWidth": { "Reader": 1 },
        "go.interfaceProximity": { "Reader": "consumer" }
      }
    }
  ],
  "cycles": [],
  "warnings": [],
  "status": "complete",
  "provenance": {
    "producer": "vibe-check",
    "version": "0.1.0",
    "generatedAt": "2026-09-21T12:00:00Z",
    "input": {
      "path": ".",
      "modulePath": "github.com/you/proj"
    }
  }
}
```

The full JSON Schema is at [`metrics/modulegraph.schema.json`](metrics/modulegraph.schema.json).

### A note on terminology

Vibe-Check analyzes Go **packages**, but the universal output schema names each unit of
analysis a `module`. Throughout vibe-check's output — every `modules[]` entry in the JSON
and each `VIOLATION: module ...` line on stderr — one `module` corresponds to one Go
package.

## Rendering a health report: `vibe-check report`

`report` renders a deterministic architectural health report from a ModuleGraph JSON
snapshot previously emitted by `analyze`. It never analyzes code — reading, validating,
and rendering are pure — so identical input files always render byte-identical reports
(no timestamps, modules in path order).

```bash
# Snapshot first — report renders an existing ModuleGraph, it never analyzes
vibe-check analyze --output graph.json .

# Summary report (default): header, metric legend, status line, recommendations
vibe-check report graph.json

# Detailed report: adds the package metrics table, coupling and cohesion
# summaries, cycle chains, duplication table, health scorecard, and warnings
vibe-check report --mode detailed graph.json

# Machine-readable report model (mutually exclusive with --mode)
vibe-check report --json graph.json
```

`--mode` selects the detail level. `summary` (the default) renders the report header,
the metric legend, a status summary line ("All Clear" or "Needs Attention"), and
prioritized recommendations. `detailed` additionally renders the package metrics table,
coupling and cohesion summaries, circular-dependency chains, the duplication table, a
health scorecard that grades five dimensions (average distance, max LCOM4, cycle count,
duplication, instability spread) on an A–F scale, and warnings.

`--json` emits the computed report model — aggregates, per-dimension grades, overall
status, and fired recommendations — as a single JSON object, for consumers that should
not parse markdown. It is mutually exclusive with `--mode`: passing both exits `2` with
no output.

Exit codes: `0` when no 🔴 or 🟡 recommendation fires, `1` when any fires (the full
report is still written to stdout), and `2` on tool failure — an unknown `--mode` value,
`--json` together with `--mode`, or an unreadable or schema-invalid input (error on
stderr, nothing on stdout).

Recommendations come from a fixed rule table — LCOM4 ≥ 4, distance ≥ 0.5, any circular
dependency, or duplication ≥ 5% produce 🔴 entries; instability ≥ 0.95 or ≤ 0.05, more
than 20 efferent couplings, or fully concrete packages with more than ten exported types
produce 🟡 entries — ordered by severity, then metric extremity, then package path, and
capped at five. Health grades are advisory: vibe-check's enforcement gates are the
`diff` verdict thresholds, not the report.

## Comparing snapshots: `vibe-check diff`

`diff` compares two `analyze` JSON snapshots — a base and a PR — and reports the change in
design-quality metrics plus a verdict:

```bash
vibe-check analyze --output base.json .        # on the base revision
vibe-check analyze --output pr.json .          # on the PR revision
vibe-check diff base.json pr.json              # add --json for machine-readable output
vibe-check diff --gate base.json pr.json       # exit 1 on REQUEST_CHANGES (CI regression gate)
vibe-check diff --markdown base.json pr.json   # byte-stable markdown report (display-only)
```

`diff` computes per-module deltas (Ca, Ce, instability, abstractness, distance, LCOM), new
and resolved cycles, an entropy direction (`improving`, `stable`, or `degrading`), and a
verdict — `APPROVE`, `COMMENT`, or `REQUEST_CHANGES`. The protected gates are
Δinstability ≥ 0.15, Δdistance ≥ 0.20, ΔLCOM ≥ 2, or a new circular dependency (any of
which yields `REQUEST_CHANGES`); smaller non-zero shifts yield `COMMENT`; an improving or
flat delta yields `APPROVE`. A partial/unreliable input is always downgraded to `COMMENT`.

When run with `--json`, `diff` also emits a top-level `provenance` object with `producer` (the constant
`"vibe-check"`), `version` (the tool version), and `generatedAt` (an RFC 3339 UTC timestamp); pass
`--no-provenance` to omit it for byte-reproducible output.

Pass `--markdown` to render the comparison as a byte-stable markdown report — the
verdict, its triggering reasons, the entropy direction, the per-module deltas, and the
new and resolved cycles — instead of the default table. It is mutually exclusive with
`--json` (passing both exits `2`) and is display-only: it never affects the verdict, the
protected thresholds, or the exit code, so it is never a gate input. Under `--gate`, a
`REQUEST_CHANGES` verdict still exits `1` with the markdown report written to stdout.

By default it exits `0` whenever both inputs are valid — the verdict is data in the payload, so a
`REQUEST_CHANGES` verdict still exits `0` — and `2` when an input is missing, unreadable, or
schema-invalid, or when a `--max-instability-delta`, `--max-distance-delta`, or
`--max-lcom-delta` override is looser than the protected default (overrides may only tighten).
With `--gate`, a `REQUEST_CHANGES` verdict instead exits `1` (an `APPROVE` or `COMMENT` verdict
still exits `0`), so CI can block a PR on structural regression. The verdict and reasons are
still written to the payload; only the exit code changes. See
[`.github/workflows/structural-gate.yml`](.github/workflows/structural-gate.yml) for the
pull-request regression gate that runs `analyze` on both the PR head and base and then
`diff --gate`; its Report step appends the `diff --markdown` report to the PR step
summary, so the report appears in the Checks UI even when the gate fails.

This gate is advisory until a repository admin enables branch protection and marks
`structural-gate` required — it fails the check but does not by itself block a merge.

## CI integration

Any repository can run the same structural gate on its pull requests with a single SHA-pinned
job — no build-from-source, no secrets. The called workflow
([`reusable_structural_gate.yml`](.github/workflows/reusable_structural_gate.yml)) installs the
released vibe-check binary (checksum-verified against the release's `checksums.txt`), checks out
your PR head and base revisions itself, analyzes both, and fails the job on a `REQUEST_CHANGES`
verdict from `diff --gate` under the protected thresholds. The markdown diff report is appended
to the job step summary even when the gate fails, and the gate job runs on `ubuntu-latest`, so
the analyzed module must build on linux/amd64. As with the self-gate, this gate is advisory
until a repository admin enables branch protection and marks the gate's check (`Structural
Gate`) required — it fails the check but does not by itself block a merge.

```yaml
jobs:
  structural-gate:
    # Required when the calling workflow triggers on events other than
    # pull_request; see "Trigger context" below.
    if: github.event_name == 'pull_request'
    # Grant exactly the scope the callee declares; never broader.
    permissions:
      contents: read
    # Pin the reference to a full commit SHA of this repository.
    uses: zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@<commit-sha>
```

A job that calls a reusable workflow takes no `steps`, `runs-on`, or `timeout-minutes` — the
called workflow owns the runner, its timeout, and its steps. Inputs are passed with `with:`;
omitted inputs take their declared defaults, so the minimal call above needs no `with:` block:

```yaml
jobs:
  structural-gate:
    if: github.event_name == 'pull_request'
    permissions:
      contents: read
    uses: zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@<commit-sha>
    with:
      version: v0.2.0                      # pin the analyzer release
      path: services/api                    # module directory in a monorepo
      go-version-file: services/api/go.mod  # must select that same module
      max-instability-delta: "0.10"         # tighten-only; quotes keep it a string
```

### Inputs

| Input | Type | Default | Effect |
|-------|------|---------|--------|
| `version` | string | `latest` | vibe-check release tag to install. `latest` resolves the newest published release; an explicit tag (for example `v0.2.0`) gives reproducible gates. A nonexistent tag fails the job before any analysis runs. |
| `path` | string | `.` | Go module directory to analyze, relative to the repository root — `.` for a root module, or a subdirectory such as `services/api` in a monorepo. The same subpath under the base checkout is analyzed for the base snapshot. |
| `timeout` | string | `5m` | Per-analysis timeout (Go duration string) forwarded to `vibe-check analyze` for both the head and base snapshots. |
| `go-version-file` | string | `go.mod` | Path to the caller's `go.mod` (or `go.work`), relative to the head-checkout root, whose Go version `actions/setup-go` installs. Must select the same module `path` points at — see the coupling note below. |
| `max-instability-delta` | string | unset | Optional tighten-only instability-delta threshold forwarded to `diff --gate` as `--max-instability-delta`. Unset keeps the protected default (`0.15`); a looser value is rejected by the CLI with exit `2`. |
| `max-distance-delta` | string | unset | Optional tighten-only distance-delta threshold forwarded to `diff --gate` as `--max-distance-delta`. Unset keeps the protected default (`0.20`); a looser value is rejected by the CLI with exit `2`. |
| `max-lcom-delta` | string | unset | Optional tighten-only LCOM-delta threshold forwarded to `diff --gate` as `--max-lcom-delta`. Unset keeps the protected default (`2`); a looser value is rejected by the CLI with exit `2`. |

All seven inputs are strings: quote numeric-looking values (for example
`max-instability-delta: "0.10"`) or GitHub rejects them at dispatch time. The three
`max-*-delta` inputs are tighten-only — a value looser than the protected default fails the job
with exit `2` instead of weakening the gate.

### Caller permissions

The called workflow declares only `contents: read` — the scope its two checkout steps need; the
release download is anonymous HTTPS and needs no token. A called workflow cannot elevate beyond
what the caller grants, so a caller using a top-level default-deny (`permissions: {}`) must
grant that scope at the job level or the called job fails at startup before any step runs:

```yaml
permissions: {}

jobs:
  structural-gate:
    if: github.event_name == 'pull_request'
    permissions:
      contents: read  # mirrors the callee's declaration; never grant broader
    uses: zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@<commit-sha>
```

Never grant more than the callee declares at the SHA you have pinned.

### Trigger context

The gate resolves its base revision from the caller's `pull_request` event context
(`github.event.pull_request.base.sha`). If the calling workflow triggers on events other than
`pull_request` — `push`, `schedule`, `workflow_dispatch` — guard the job with
`if: github.event_name == 'pull_request'` as in the snippets above. Without the guard, the PR
context is empty on those events: the base checkout falls back to the event SHA, base equals
head, the delta is zero, and the gate passes vacuously — a green check that proves nothing. A
workflow that triggers only on `pull_request` does not need the guard.

### `path` / `go-version-file` coupling

`go-version-file` resolves from the head-checkout root, not relative to `path`. It must select
the same module the `path` input points at — that module's `go.mod` (or `go.work`): when `path`
is `services/api`, pass `go-version-file: services/api/go.mod`. `analyze` forces
`GOTOOLCHAIN=local`, so the installed toolchain must satisfy the analyzed module or the
analysis fails as a visible tool failure.

### Version pinning

Two independent pins govern a gate run:

- The `uses:` reference pins the workflow logic. Pin a full commit SHA — immutable and
  reviewable, like SHA-pinning an action. (vibe-check's own CI calls the workflow with the
  same-repo local form `./.github/workflows/reusable_structural_gate.yml`, which resolves
  against the caller's ref; external callers use the cross-repo form above.)
- The `version` input pins the analyzer binary. The default `latest` silently moves with each
  release; an explicit tag such as `v0.2.0` gives reproducible gates. The downloaded archive
  is verified against the release's `checksums.txt` before any analysis runs, and a missing or
  nonexistent tag fails the job — no silent fallback.

### Trust model

The gate itself runs only the released, checksum-verified vibe-check binary — it never checks
out or builds vibe-check's source. `vibe-check analyze` compiles your module, which is the same
trust posture as running Go tests on a fork PR in your own CI; there is deliberately no fork
guard, and callers who want one can add their own job-level `if:`. No secrets are required or
read, so fork PRs work with zero configuration. Both checkout steps set
`persist-credentials: false`: the job token is used only for the clone and is never persisted
into `.git/config`, where compiled build code could otherwise read it.

vibe-check dogfoods this contract on every pull request — the `structural-gate-dogfood` job in
[`ci.yml`](.github/workflows/ci.yml) calls the reusable workflow end-to-end.

## Deploying agents and commands: `vibe-check init`

`init` deploys the embedded agent and command assets into a project's `.opencode/` directory and
registers the `divisor-entropy` reviewer into the project's Unbound Force policy files:

```bash
vibe-check init [path]     # path defaults to "."; --force to overwrite, --json for machine output
```

It writes assets to two directories:

- `.opencode/agents/` — the `divisor-entropy` agent (structural-entropy reviewer that runs
  `analyze` + `diff` across a base↔PR pair) and the `vibe-check-reporter` agent (interactive
  metrics interpreter with summary, detailed, and trending modes).
- `.opencode/commands/` — the `/vibe-check` slash command, which delegates to the
  `vibe-check-reporter` agent for conversational architectural analysis.

When the project carries Unbound Force policy files, `init` also registers the reviewer
surgically, preserving every other line:

- `.uf/reviewer-capabilities.yaml` — a `divisor-entropy` entry under `reviewers:` with the
  `review` capability.
- `.uf/review-matrix.yaml` — a `divisor-entropy` item in each of the `advisor` mode lists
  (`code`, `spec`, `triage`, `feedback`).

Registration is idempotent: present entries are reported as `reviewers/divisor-entropy` and
`advisors/divisor-entropy` in the `Skipped` list and left byte-identical. Absent policy files or
unexpected shapes (a missing `reviewers:`/`advisor:` section or mode list) are reported in a
`Notices` list — also the `notices` array under `--json` — and never created or guessed. `--force`
affects asset overwrite only, never registration.

Existing files are skipped unless `--force` is given. Exit code `0` on success (including
when every asset is skipped and notice-only policy gaps), `2` on an invalid target path, a policy
file that cannot be safely rewritten (symlinked or non-regular), or I/O failure.

When run with `--json`, `init` also emits a top-level `provenance` object with `producer` (the constant
`"vibe-check"`), `version` (the tool version), and `generatedAt` (an RFC 3339 UTC timestamp); pass
`--no-provenance` to omit it for byte-reproducible output.

## Known limitations

- `analyze` accepts a single path argument (defaulting to `.`); it does not take multiple
  package patterns.
- There is no default timeout — set `--timeout` to bound long analyses (for example, on large
  monorepos).
- Large-repository performance is a P2 item: `go/packages` loads full ASTs and type
  information into memory, so very large monorepos may be slow.

## License

Apache 2.0.
