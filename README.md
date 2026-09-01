# Vibe-Check

Vibe-Check is a design-quality and architectural-metrics tool for Go. It computes the
Martin package-coupling metrics suite — afferent coupling (Ca), efferent coupling (Ce),
instability, abstractness, and distance from the main sequence — plus LCOM4 cohesion and
circular-dependency detection, and emits the results as JSON.

## Install

```bash
go install github.com/zero-dot-force/vibe-check/cmd/vibe-check@v0.1.0
```

This resolves against the `v0.1.0` release tag, which must be published in the repository
for module-version installs to work. When a binary is built this way (without release
`ldflags`), `--version` falls back to the toolchain's build info so it still reports a
meaningful module version (commit and build date may show as `none`/`unknown` for
module-cache installs).

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

The analysis is printed to stdout as pretty-printed JSON with `schemaVersion` `"1.1"`. Each
entry in `modules[]` carries `ca`, `ce`, `instability`, `abstractness`, `distance`, `lcom`,
`exportedTypes`, `abstractTypes`, `zone`, and an optional `extensions` object. For Go, the
extensions may include `go.interfaceWidth` (method count per exported interface) and
`go.interfaceProximity` (`"producer"` or `"consumer"` per interface). Detected cycles appear
in `cycles` as lexicographically-sorted sets of package paths.

```json
{
  "schemaVersion": "1.1",
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
  "status": "complete"
}
```

The full JSON Schema is at [`metrics/modulegraph.schema.json`](metrics/modulegraph.schema.json).

### A note on terminology

Vibe-Check analyzes Go **packages**, but the universal output schema names each unit of
analysis a `module`. Throughout vibe-check's output — every `modules[]` entry in the JSON
and each `VIOLATION: module ...` line on stderr — one `module` corresponds to one Go
package.

## Comparing snapshots: `vibe-check diff`

`diff` compares two `analyze` JSON snapshots — a base and a PR — and reports the change in
design-quality metrics plus a verdict:

```bash
vibe-check analyze --output base.json ./...   # on the base revision
vibe-check analyze --output pr.json  ./...     # on the PR revision
vibe-check diff base.json pr.json              # add --json for machine-readable output
```

`diff` computes per-module deltas (Ca, Ce, instability, abstractness, distance, LCOM), new
and resolved cycles, an entropy direction (`improving`, `stable`, or `degrading`), and a
verdict — `APPROVE`, `COMMENT`, or `REQUEST_CHANGES`. The protected gates are
Δinstability ≥ 0.15, Δdistance ≥ 0.20, ΔLCOM ≥ 2, or a new circular dependency (any of
which yields `REQUEST_CHANGES`); smaller non-zero shifts yield `COMMENT`; an improving or
flat delta yields `APPROVE`. A partial/unreliable input is always downgraded to `COMMENT`.

It exits `0` whenever both inputs are valid — the verdict is data in the payload, so a
`REQUEST_CHANGES` verdict still exits `0` — and `2` when an input is missing, unreadable, or
schema-invalid, or when a `--max-instability-delta`, `--max-distance-delta`, or
`--max-lcom-delta` override is looser than the protected default (overrides may only tighten).

## Deploying agents: `vibe-check init`

`init` deploys the embedded Review Council agent assets into a project's `.opencode/agents/`
directory:

```bash
vibe-check init [path]     # path defaults to "."; --force to overwrite, --json for machine output
```

It writes the bundled `divisor-entropy` agent — a structural-entropy reviewer that runs
`analyze` + `diff` across a base↔PR pair in an isolated worktree — and skips files that
already exist unless `--force` is given. It exits `0` on success (including when every asset
is skipped) and `2` on an invalid target path or I/O failure.

## Known limitations

- `analyze` accepts a single path argument (defaulting to `.`); it does not take multiple
  package patterns.
- There is no default timeout — set `--timeout` to bound long analyses (for example, on large
  monorepos).
- Provenance metadata (producer, version, timestamp, input) is not yet emitted in the output;
  it is deferred to a follow-up.
- Large-repository performance is a P2 item: `go/packages` loads full ASTs and type
  information into memory, so very large monorepos may be slow.

## License

Apache 2.0.
