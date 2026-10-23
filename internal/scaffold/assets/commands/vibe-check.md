---
description: "Analyze architectural health metrics for a Go codebase"
agent: vibe-check-reporter
---
<!-- scaffolded by vibe-check -->

Analyze the architectural health of this Go codebase using coupling,
cohesion, and design quality metrics. Summary and detailed reports are
rendered by `vibe-check report` from a fresh analysis snapshot — the
same deterministic markdown the CLI emits — with the agent adding
violation explanations and remediation advice.

## Usage

```
/vibe-check [mode] [directory]
```

`[directory]` is the project directory to analyze and defaults to `.`
(the repository root). Pass a directory path, never a package pattern:
`vibe-check analyze` validates its argument as a directory, so patterns
such as `./internal/...` are rejected with exit code 2.

## Modes

| Mode | Description |
|------|-------------|
| `summary` | Report header, metric legend, status summary line, and prioritized recommendations (default) |
| `detailed` | Additionally renders the package metrics table, coupling summary, cohesion summary, circular dependency chains, duplication table, health scorecard, and warnings |
| `trending` | Historical comparison against previous snapshots, computed by the agent from Dewey-stored history |

## Examples

```
/vibe-check                          # Summary mode, current directory (default)
/vibe-check summary                  # Same as above
/vibe-check detailed .               # Detailed mode, current directory
/vibe-check detailed ./internal      # Detailed mode, scoped to the ./internal directory
/vibe-check trending                 # Compare against previous snapshots
```

## Arguments

Pass `$ARGUMENTS` to the agent for mode selection and the project
directory.
