## ADDED Requirements

### Requirement: Hermetic Toolchain

The `vibe-check analyze` command SHALL force `GOTOOLCHAIN=local` in the analysis subprocess environment so that an untrusted target module's `go.mod` `toolchain` directive cannot trigger a Go toolchain download and execution during metric collection.

The forced value MUST be set by the binary itself (added to the sanitized subprocess environment with a fixed value, not passed through from the host environment). This ensures the guarantee holds regardless of how `analyze` is invoked and does not depend on the caller — including the `divisor-entropy` agent, whose permission allowlist only permits the bare `vibe-check analyze *` form and cannot prefix an inline environment variable.

This requirement hardens `analyze` because the `divisor-entropy` agent runs it against a PR ref whose `go.mod` may be attacker-influenced. It complements — and does not replace — the trusted-refs-only operating constraint documented for the agent. Privileged CI environments SHOULD additionally export `GOTOOLCHAIN=local` ambiently as defense-in-depth.

Forcing `GOTOOLCHAIN=local` is a deliberate compatibility trade-off: a standalone `vibe-check analyze` run against a (trusted) module whose `go.mod` `toolchain` directive requires a newer-than-local Go toolchain will now FAIL rather than silently download and run that toolchain. Within the `divisor-entropy` agent this failure degrades safely to a COMMENT verdict, but it bites operators who run `analyze` directly on such a module — they must install a compatible local toolchain. This trade-off is accepted and the gate MUST NOT be weakened to restore the download behavior.

#### Scenario: Untrusted go.mod toolchain directive cannot download a toolchain

- **WHEN** `vibe-check analyze` is run against a target module whose `go.mod` declares a `toolchain` directive naming a Go version other than the running toolchain
- **THEN** the analysis subprocess environment contains `GOTOOLCHAIN=local` and the analysis proceeds with the local toolchain without downloading or executing the directive-named toolchain

#### Scenario: Forced value is not overridable by the host environment

- **WHEN** the process that invokes `vibe-check analyze` has `GOTOOLCHAIN=auto` (or any other value) set in its own environment
- **THEN** the analysis subprocess still runs with `GOTOOLCHAIN=local`, because the binary sets the forced value in the sanitized subprocess environment rather than inheriting it from the host

### Requirement: Output File Selection

The `vibe-check analyze` command SHALL support an `--output <file>` flag (short form `-o`) that writes the ModuleGraph JSON to the named file. When the flag is absent, the command SHALL write the ModuleGraph JSON to stdout as before. This lets the `divisor-entropy` agent materialize base and PR graphs to files — which `vibe-check diff` consumes as positional arguments — without relying on shell redirection, which the agent's permission allowlist forbids.

#### Scenario: Output flag writes the graph to a file

- **WHEN** `vibe-check analyze --output graph.json <target>` is run against a valid target
- **THEN** the ModuleGraph JSON is written to `graph.json` and NOT to stdout, and the command exits 0

#### Scenario: Absent output flag preserves stdout behavior

- **WHEN** `vibe-check analyze <target>` is run without the `--output` flag
- **THEN** the ModuleGraph JSON is written to stdout exactly as before

#### Scenario: Unwritable output path fails cleanly

- **WHEN** `vibe-check analyze --output <path> <target>` is given an output path that cannot be created or written
- **THEN** the command exits 2 with a diagnostic on stderr and does NOT emit a partial ModuleGraph on stdout
