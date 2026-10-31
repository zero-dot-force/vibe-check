// Package ripwireadapter implements ripwire-backed language adapters for the
// vibe-check metrics toolkit, covering Python and TypeScript codebases. It
// implements the [metrics.Adapter] interface as two language instances:
// [NewPython] creates an adapter whose Language method reports "python", and
// [NewTypeScript] creates one reporting "typescript". Both instances share a
// single analysis pipeline and differ only in language identifier,
// crawl-root rules, abstractness source, and light-parse dialect.
//
// The backing tool is ripwire, an optional external binary discovered on PATH
// and pinned to the validated surface: ripwire 0.6.9 at commit bedc0cd. The
// binary is never vendored or bundled, and the package's full test suite
// passes without it installed — subprocess behaviors are exercised by
// replaying committed ripwire 0.6.9 output recordings through a fake ripwire
// binary that the tests install on PATH.
//
// # Hybrid Metric Sourcing
//
// Metrics are sourced per the validated Phase 0 spike matrix, hybridizing
// ripwire's structural coupling data with vibe-check's own light language
// parse:
//
//   - Ca, Ce, Instability: ripwire --arch metrics-table aggregates. Only
//     uncapped sources are trusted; per-file <inc> lists are capped at 40
//     entries and are never read for counts.
//   - Abstractness: for Python, the adapter's own light parse of class
//     declarations (bases referencing abc.ABC, ABCMeta, or typing.Protocol
//     are abstract; ripwire's Python abstractness is structurally zero and
//     is never used). For TypeScript, ripwire's interface share, including
//     nested interfaces.
//   - LCOM4: the adapter's own light parse for both languages (Python
//     self.<attr> and TypeScript this.<prop> field usage; connected
//     components per directory module).
//   - Distance and Zone: derived via metrics.Compute* functions — never
//     reimplemented.
//   - Circular dependencies: the adapter's own Tarjan strongly-connected-
//     component analysis over ripwire --deps edges, corroborated against
//     ripwire's capped cycle listing.
//   - Duplication: unsupported. It is never declared as a capability, is
//     zero-valued in results, and is reported as an "unsupported-metric"
//     warning.
//
// Every known mapping approximation (the directory-as-module unit, measured
// import-graph error budgets, cycle-membership approximation, TypeScript
// abstractness bias) is surfaced as a machine-coded [metrics.Warning], and
// graphs carrying warnings have status "partial".
//
// # Version Pin
//
// ripwire is 0.x with no upstream semver policy, so before analyzing the
// adapter validates the exact (version, commit) pair "0.6.9" at "bedc0cd"
// and fails closed on an absent binary, a version mismatch, or unparseable
// version output. The ModuleGraph provenance input records the source tool
// identity ("ripwire", version, commit) on every ripwire-backed graph.
//
// # Directory as Module
//
// The unit of analysis is a directory (a Python package directory or a
// TypeScript source directory), not a file. Python module paths use
// dot-separated package style (pkg.sub); TypeScript paths use
// slash-separated style (src/pkg/sub). Module.Name is the last path segment,
// and files within a directory are merged into one module.
//
// # Security
//
// ripwire is invoked via exec.CommandContext with the sanitized environment
// from [metrics.SanitizeEnvironment] so no host credentials reach the
// subprocess, with its working directory set to the crawl root, and with
// bounded stdout/stderr capture — oversized output is an error, never a
// truncation. Context cancellation terminates the subprocess. The adapter
// never reads ripwire's internal cache, never consumes its ranked
// agent-facing answers as metrics input, and never vendors or bundles the
// binary.
package ripwireadapter
