---
pack_id: go-custom
language: Go
version: 1.0.0
---
<!-- scaffolded by uf vdev -->

# Custom Rules: Go

Project-specific Go conventions that extend the canonical
Go convention pack. Rules in this file are loaded alongside
`go.md` by Cobalt-Crush (during implementation) and
all Divisor persona agents (during review).

Use the `CR-NNN` prefix for all custom rules. Use `[MUST]`,
`[SHOULD]`, or `[MAY]` severity indicators per RFC 2119.

## Custom Rules

### CR-001 [SHOULD] Functional Options for Multi-Parameter Constructors

Constructor functions with 3 or more optional parameters MAY use
the functional options pattern (`type Option func(*T)`) instead
of the AP-001 Options struct pattern when the options are
independent and self-documenting. The functional options pattern
is preferred for adapter configuration where sensible defaults
are provided and each option modifies a single field.

**Rationale**: The `metrics.ExternalAdapter` has 5 configurable
timeouts/limits with sensible defaults. Functional options are
more ergonomic for this use case — callers specify only the
overrides they need without constructing a full Options struct.

**Scope**: `metrics` package adapter configuration. New packages
should prefer AP-001 unless the same conditions apply.
