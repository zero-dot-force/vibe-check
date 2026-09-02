---
description: "Analyze architectural health metrics for a Go codebase"
agent: vibe-check-reporter
---
<!-- scaffolded by vibe-check -->

Analyze the architectural health of this Go codebase using coupling,
cohesion, and design quality metrics.

## Usage

```
/vibe-check [mode] [package-pattern]
```

## Modes

| Mode | Description |
|------|-------------|
| `summary` | Traffic-light health indicator with aggregate metrics (default) |
| `detailed` | Per-package metric table with zone classification |
| `trending` | Historical comparison against previous snapshots (requires Dewey) |

## Examples

```
/vibe-check                          # Summary mode, all packages
/vibe-check summary                  # Same as above
/vibe-check detailed                 # Per-package breakdown
/vibe-check detailed ./internal/...  # Per-package for internal only
/vibe-check trending                 # Compare against last snapshot
```

## Arguments

Pass `$ARGUMENTS` to the agent for mode selection and package pattern.
