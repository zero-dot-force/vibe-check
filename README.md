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
current directory (`.`). It writes the analysis as JSON to stdout.

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
