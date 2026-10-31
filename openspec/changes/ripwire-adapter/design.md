# Design

## Context

Vibe-Check's Layer 1 (`metrics/`) provides the `Adapter` interface
(`Analyze`/`Language`/`Capabilities`), a DI `Registry`, the `ModuleGraph` model
(schema version `1.3`), `Warning{Code, Message, Module}`, `Provenance` with
`ProvenanceInput` (which already carries optional `commitSHA`/`branch` fields
beyond the schema file's strict listing, per architectural-drift-tracking), and
subprocess security primitives (`ValidateProjectPath`, `SanitizeEnvironment`).
Layer 2 has exactly one adapter: `internal/goadapter`, which natively owns
package resolution, type classification, LCOM4, cycles, and Go extensions.
Layer 3's `RunAnalyze` (`cmd/vibe-check/analyze.go`) instantiates the Go adapter
directly — go-analyze D10 explicitly deferred registry wiring until
multi-language support landed. `metrics.ExternalAdapter` exists for JSON-RPC
subprocess analyzers; ripwire is not one (its outputs are a metrics table, XML
deps, and JSON exports), so that path does not apply.

The Phase 0 spike (#67, GO 2026-10-30) validated ripwire 0.6.9 @ `bedc0cd`
against pip (Python) and a TypeScript corpus with these findings, which this
design treats as hard constraints:

- **Capability matrix**: Python Ca/Ce/I/cycles GOOD-ENOUGH (81%/85% exact, MAE
  ~0.2); ripwire's Python Abstractness is structurally zero. TS Ca/Ce/I/cycles
  GOOD-ENOUGH (26/26 exact); TS Abstractness via interface share is exact
  (incl. nested), with `abstract class` undercounting (conservative) and type
  aliases inflating the denominator.
- **Crawl root**: src-layout repo-root crawls silently yield `ca=ce=0` (pip);
  directories named `build`/`dist`/`out`/`target`/`vendor`/`third_party`/
  `captures` are hard-pruned from the crawl (pip's `operations/build` vanished).
- **Output surface**: Ca/Ce/I live in the `--arch` `<metrics>` table and
  aggregate attributes; per-file `<inc>` lists are capped at 40 entries per
  file; `--export=cc.json` provides per-file fan-in/out; `--deps` is XML-only;
  cycle listings cap at 20 SCCs × 12 members with no `total=`; TS Ce counts
  type-only and side-effect imports, excludes dynamic `import()` and externals,
  and edges relative `package.json` imports to the JSON's directory.
- **Error budgets**: Python import graph 1.9% lost / 2.2% spurious; TS 0.16%
  unresolved. Re-runs were byte-identical.
- **Versioning**: ripwire is 0.x with no semver policy upstream.

## Goals / Non-Goals

**Goals:**

- Two ripwire-backed `metrics.Adapter` instances (`python`, `typescript`) that
  satisfy every requirement in `specs/ripwire-adapter/spec.md`.
- Hybrid sourcing exactly per the spike matrix: ripwire for coupling/cycles
  (both) and TS abstractness; vibe-check's own light parse for LCOM4 field
  usage (both) and Python abstractness.
- Deterministic, schema-valid `ModuleGraph` output with mandatory approximation
  warnings and source-tool provenance.
- A test suite that fully passes without ripwire installed (fixture replay),
  keeping CI and releases zero-dependency.
- Go analysis byte-identical to pre-change behavior.

**Non-Goals:**

- JavaScript analysis (TypeScript only; JS is future work behind the same
  adapter pattern).
- Full-fidelity Python/TypeScript parsing (tree-sitter or language servers) —
  the light parse is deliberately approximate and warning-annotated.
- Duplication detection for Python/TypeScript (capability not declared;
  zero-valued with a warning per the adapter-interface contract).
- Any change to verdict thresholds, `modulegraph.schema.json`, the report
  engine, or `metrics.Compute*` functions (protected values).
- A `--language` override flag (deterministic detection only; recorded as a
  possible follow-up if real repos need it).
- Changes to `vibe-check diff`, `report`, or `init` — they consume `ModuleGraph`
  generically and work unchanged.
- Resolving upstream ask #68 (version/semver stability) — ship on the current
  surface with `Warning`s.

## Decisions

### D1: One adapter package, two language instances

**Decision**: A single `internal/ripwireadapter/` package holds the shared core
(subprocess runner, output parsing, mapping, warnings) and exposes two
constructors — `NewPython()` and `NewTypeScript()` — each returning a distinct
`metrics.Adapter` whose `Language()` returns `"python"` / `"typescript"`. The
two instances share all pipeline code and differ only in language identifier,
crawl-root rules, abstractness source, and light-parse dialect.

**Rationale**: The languages share ~90% of the pipeline (discovery, version
pin, invocation, parsing, mapping, cycle detection, warnings). Two packages
would duplicate all of it. Keeping the package `internal/` mirrors `goadapter`
and prevents external coupling to an explicitly optional component.

**Alternatives considered**:
- Two packages (`internal/pyadapter`, `internal/tsadapter`) — rejected:
  duplication of the ripwire plumbing.
- `metrics.ExternalAdapter` (JSON-RPC) — rejected: ripwire does not implement
  the JSON-RPC analyzer protocol; its outputs are XML/tables/JSON files that
  need adapter-internal parsing anyway.

### D2: PATH discovery with a fail-closed version pin

**Decision**: At `Analyze` time the adapter resolves `ripwire` via
`exec.LookPath("ripwire")`. It then reads the binary's version and commit and
compares them against pinned constants — `validatedRipwireVersion = "0.6.9"`,
`validatedRipwireCommit = "bedc0cd"` — failing closed with an actionable error
(absent binary: install guidance pointing at
https://github.com/redhat-et/ripwire plus the validated version; mismatch: name
both versions; unparseable output: refuse to analyze with an unknown tool).

**Rationale**: Upstream is 0.x with no semver policy, so vibe-check cannot
reason about compatibility of any surface it has not validated; an exact
(version, commit) pin is the only honest contract. PATH discovery keeps
releases zero-dependency and activation automatic. Discovery at `Analyze` time
(not construction) keeps `Capabilities()` and registry construction side-effect
free and cheap.

**Alternatives considered**:
- `RIPWIRE_PATH` environment override — rejected: PATH injection in tests
  covers the need; a public env var is config surface the issue does not ask
  for.
- Best-effort version warning instead of failure — rejected: the issue
  mandates fail-closed; a silently-degraded graph would violate Metric
  Fidelity.

### D3: Hybrid capability matrix (source of truth per metric)

**Decision**:

| Metric | Python source | TypeScript source |
|--------|---------------|-------------------|
| Ca, Ce, Instability | ripwire `--arch` metrics table | ripwire `--arch` metrics table |
| Abstractness | own light parse (classes/ABCs) | ripwire interface share |
| Distance, Zone | `metrics.Compute*` (derived) | `metrics.Compute*` (derived) |
| LCOM4 | own light parse (`self.X`) | own light parse (`this.X`) |
| Cycles | ripwire `--deps` edges + own SCC | ripwire `--deps` edges + own SCC |
| Duplication | unsupported (zero + warning) | unsupported (zero + warning) |

The adapter MUST use `metrics.ComputeInstability` / `ComputeAbstractness` /
`ComputeDistance` / `ComputeZone` for derived values — never reimplement them.

**Rationale**: This is the Phase 0 GO decision verbatim: ripwire supplies what
the spike validated as GOOD-ENOUGH; vibe-check owns what ripwire cannot do
(Python abstractness is structurally zero upstream) and what it should not do
(LCOM4 field usage needs language semantics the spike did not validate in
ripwire). Mirroring `goadapter` — which owns LCOM4/type classification via its
own AST — keeps metric ownership in the adapter layer.

### D4: Directory-as-module unit with deterministic identity

**Decision**: The unit of analysis is a directory (Python package directory /
TypeScript source directory), not a file. `Module.Path` is the directory's
path relative to the crawl root, normalized: Python uses dot-separated package
style (`pip._internal.utils`), TypeScript uses slash-separated style
(`src/pkg/sub`). `Module.Name` is the last segment. Every graph carries one
`Warning` (code `directory-as-module`) documenting that files within a
directory are merged into one module.

**Rationale**: Martin coupling metrics are meaningless at file granularity for
these languages, and ripwire's per-file data aggregates cleanly to
directories. The differing path styles match each ecosystem's native idiom,
keeping output legible to the module's users.

**Alternatives considered**:
- File-as-module — rejected: explodes module counts and makes Ca/Ce noise.
- ripwire's internal unit naming (if any) — rejected: not stable contract
  surface; the adapter owns the mapping.

### D5: Subprocess invocation hygiene

**Decision**: ripwire is invoked via `exec.CommandContext` with
`metrics.SanitizeEnvironment(nil)` (PATH/HOME/LANG only), working directory
set to the crawl root, stdout/stderr captured into bounded buffers (constant
caps, documented in GoDoc; oversized output is an error, not a truncation), and
cancellation propagated through the context so the subprocess dies with the
analysis. The version check runs first as a cheap standalone invocation.
`--deps` XML and `--export=cc.json` are parsed with stdlib `encoding/xml` /
`encoding/json` inside the adapter package — no new Go module dependencies.

**Rationale**: Matches the security posture `goadapter` applies to
`go/packages` subprocesses and the bounds `metrics.ExternalAdapter`
established (timeouts, size caps). Bounded-capture-as-error (rather than
silent truncation) preserves Metric Fidelity: a truncated metrics table must
not produce wrong numbers.

### D6: Coupling mapping rules (uncapped sources only)

**Decision**: Per-unit Ca/Ce/I are read from the `--arch` `<metrics>` table
and its aggregate attributes; per-file fan-in/out comes from
`--export=cc.json`; per-file `<inc>` lists are never used for counts (40-cap).
Files aggregate to their directory module by deterministic rules: sum per-file
fan-in/fan-out, deduplicate edges between the same pair of directory modules,
and count external (out-of-scope) targets in Ce only. Name-resolved edges
(imports resolved by name rather than path) and any ambiguity the mapping
encounters emit warnings (`name-resolved-edges`, `ambiguity-rate`) plus the
spike's measured error budgets (`error-budget`: Python 1.9% lost / 2.2%
spurious; TS 0.16% unresolved).

**Rationale**: The 40-cap on `<inc>` silently undercounts Ca if trusted; the
table and cc.json are uncapped. The exact field names of the 0.6.9 output are
pinned by committed golden fixtures (see D12) rather than restated here, so
the parser is validated against recorded reality.

**Alternatives considered**:
- Deriving all coupling from `--deps` edges alone — rejected: the spike
  validated the table's aggregates as the accurate per-unit source; edges are
  the cycle/corroboration source.

### D7: Cycle detection — own SCC over ripwire edges, capped listing as corroboration

**Decision**: The adapter runs its own strongly-connected-component analysis
(Tarjan, mirroring `goadapter/cycles.go`'s approach but implemented within
`ripwireadapter` over its directory-module edge graph) on the `--deps` XML
edge set. Ripwire's own cycle listing (capped at 20 SCCs × 12 members, no
`total=`) is treated as approximate: the adapter corroborates its derived
cycles against the listing where present and always emits a
`cycle-approximation` warning stating that membership is approximate due to
listing caps and edge error budgets. `--deps --limit` is set to a documented
high bound and `--exclude` scopes the crawl to the package root.

**Rationale**: Deriving SCCs from the full edge set is deterministic,
uncapped, and reuses a proven algorithm shape; the capped listing then adds a
cross-check rather than being a bottleneck. Keeping the implementation local
to `ripwireadapter` (instead of extracting a shared SCC helper out of
`goadapter`) keeps the native Go adapter's file set untouched, honoring "along
side, never replacing" with minimal blast radius — the algorithm is ~50 lines.

**Alternatives considered**:
- Use ripwire's cycle listing directly — rejected: caps mean unlisted cycles
  and truncated membership with no total count.
- Extract `internal/scc` shared with `goadapter` — rejected for this change:
  refactoring the protected Go path is out of scope; revisit if a third
  adapter lands.

### D8: Light language parse (Python abstractness, LCOM4 both languages)

**Decision**: A stdlib-only, line/token-oriented light parse owned by the
adapter package:

- **Python abstractness**: `class Name(Bases)` declarations; abstract when a
  base references `abc.ABC`, `ABCMeta`, or `typing.Protocol`; exported when
  the name is not underscore-prefixed.
- **LCOM4 (both)**: nodes are methods bound to a class (Python `def` with a
  leading `self` parameter inside a class body; TypeScript method members
  inside a class body, tracked by brace depth); edges connect methods sharing
  at least one instance-field access (Python `self.<attr>`, TypeScript
  `this.<prop>`); LCOM4 = connected components per directory module via the
  same union-find approach as `goadapter/lcom.go`.
- Unparseable files are skipped with a warning (mirroring `goadapter`'s
  partial-build pattern: the module stays in the graph, the affected metric is
  zeroed, status is `partial`).

**Rationale**: The issue assigns LCOM4 field usage and Python abstractness to
vibe-check's own parse; a full parser (tree-sitter, language server) would
violate the zero-new-dependency constraint. The parse rules above are the
deliberately light subset the spike scoped; their approximations are the
documented metric definition for these languages, not silent degradation.

**Alternatives considered**:
- tree-sitter / Pyright / tsc APIs — rejected: new heavyweight dependencies.
- Sourcing LCOM4 from ripwire — rejected: not part of the validated surface.

### D9: Crawl-root detection and the zero-coupling guard

**Decision**: The adapter resolves the language-appropriate package root
before invoking ripwire. Python: if `<root>/src` contains top-level packages
(`__init__.py` directories), the crawl root is `<root>/src` (src-layout);
otherwise the root itself (flat layout). TypeScript: the crawl root is the
directory containing `tsconfig.json`/`package.json`. `--exclude` scopes the
crawl to that root. As a backstop, if the crawl yields no dependency edges
while source files exist in scope, the adapter refuses to present an all-zero
`complete` graph (error, or `partial` + warning).

**Rationale**: The pip failure mode — src-layout repo-root crawls silently
  returning `ca=ce=0` — is exactly the "fabricated metrics presented as real"
  violation `goadapter` guards against with `errTotalLoadFailure`. The guard
  makes root misdetection loud instead of silent.

### D10: CLI wiring — registry adoption and language detection

**Decision**: `RunAnalyze` replaces direct `goadapter.New()` instantiation with
a `metrics.Registry` populated at construction with the Go adapter plus both
ripwire adapter instances, and selects via a pure, unit-testable
`detectLanguage(path)`: `go.mod` → `go`; else `pyproject.toml` / `setup.py` /
`setup.cfg` → `python`; else `tsconfig.json` → `typescript`; else majority
census of `.py` vs `.ts`/`.tsx` files (deterministic walk, ties → Go default,
preserving the pre-change error path). This is the multi-adapter refactor
go-analyze D10 explicitly predicted and deferred.

**Rationale**: The registry exists for exactly this scenario; detection is a
pure function over the filesystem so it is trivially table-testable. Go
targets never construct or invoke ripwire, guaranteeing the unchanged-Go
requirement by construction.

**Alternatives considered**:
- Keep direct instantiation with an if/else over languages — rejected: the
  registry is the designed seam and keeps `RunAnalyze` open to future adapters
  without edits.
- `--language` flag — rejected (Non-Goal): deterministic detection covers the
  acceptance criteria; a flag is config surface to justify later.

### D11: Provenance source-tool fields (schema file untouched)

**Decision**: `metrics.ProvenanceInput` gains three optional `omitempty`
string fields — `sourceTool`, `sourceToolVersion`, `sourceToolCommit` —
populated by the ripwire adapter (`"ripwire"`, `"0.6.9"`, `"bedc0cd"`),
omitted by the Go adapter. `modulegraph.schema.json`, `metrics/validate.go`'s
required-fields contract, the verdict thresholds, and the report engine are
untouched; ripwire graphs pass `metrics.Validate` unchanged (the hand-rolled
validator checks required fields only). This follows the established
`commitSHA`/`branch` precedent from architectural-drift-tracking: optional
provenance-input fields riding the Go struct with the schema file left strict.

**Rationale**: The issue requires provenance to carry the ripwire version, and
the schema file is a protected value. The commit is included (not just the
version) because upstream has no semver policy — two builds can share a
version string — and archived snapshots feed drift tracking, which must be
able to distinguish measurement instruments. Since the pin fails closed, only
the validated (version, commit) pair can ever appear in output.

**Alternatives considered**:
- Carry the version in a `Warning` — rejected: warnings are accuracy caveats;
  conflation would poison warning semantics and the warning is not part of the
  provenance contract consumers read.
- Compose it into `Provenance.Version` — rejected: that field is CLI-owned
  vibe-check identity; overloading it breaks the diff/report provenance story.
- Schema bump to add the fields to the schema file — rejected: protected
  value; revisit only via a dedicated future change if external schema
  validation of these fields becomes a requirement (recorded as a known
  trade-off in Risks).

### D12: Test strategy — golden fixtures and a fake ripwire

**Decision**: `internal/ripwireadapter/testdata/` holds (a) two fixture
projects — a src-layout Python project and a TypeScript project, each with
known structure; (b) recorded ripwire 0.6.9 outputs for each project
(`--arch`, `--deps`, `--export=cc.json`, version output); and (c) expected
`ModuleGraph` goldens. Tests exercise the full subprocess path by prepending a
test directory to `PATH` containing an executable fake `ripwire` that replays
the recorded outputs (correct version by default; wrong version / garbage
variants for fail-closed tests). The absent-binary test runs with a `PATH`
that lacks the fake. A two-run byte-equality determinism test runs under
`--no-provenance` semantics. An optional real-binary integration test is
guarded by `exec.LookPath` and `t.Skip`s when ripwire is absent, so CI never
requires it.

**Rationale**: CI parity (`go build ./...`, `go test -race -count=1 ./...`,
`go vet ./...`) must pass on runners without ripwire; fixture replay makes the
subprocess, parsing, mapping, warnings, and provenance paths fully hermetic
and deterministic while the golden files pin the validated 0.6.9 output
shapes. The fake binary is test scaffolding, never shipped or vendored.

## Coverage Strategy

**Unit tests**: `internal/ripwireadapter/` — version pin (match/mismatch/
unparseable/absent), crawl-root detection (src-layout/flat/TS), denylist
pre-check, XML/JSON parsing against recorded fixtures, directory aggregation
determinism, SCC cycle derivation + corroboration warning, light parse
(Python ABC/concrete/underscore classes; LCOM cohesive/split/no-methods; TS
`this.X`), warning code presence, provenance fields, sorted output.
`cmd/vibe-check/` — `detectLanguage` table tests (all markers, census, tie
default), registry selection, Go-path regression (byte-equal against a
recorded pre-change golden), absent-binary exit code 2 + guidance text.

**Integration**: full-pipeline analysis of both fixture projects via the fake
riprewire on `PATH`; assert specific metric values, warning codes, status
`partial`, and `metrics.Validate` passing.

**Determinism**: `Analyze()` twice per fixture project; byte-identical
ModuleGraph JSON with provenance omitted.

**CI parity gate**: derive from `.github/workflows/ci.yml` at implementation
time and run locally: `go build ./...`, `go test -race -count=1 ./...`,
`go vet ./...` — all without ripwire installed.

## Operational Notes

**Warning codes** (machine-readable, stable): `unsupported-metric`,
`denylist-pruned`, `zero-coupling-guard`, `name-resolved-edges`,
`ambiguity-rate`, `error-budget`, `cycle-approximation`,
`ts-abstractness-bias`, `directory-as-module`, `parse-error`.

**Error message content**: every terminal error names the failing operation,
the ripwire version validated by vibe-check, and a remediation (install URL
for absent binary; "install ripwire 0.6.9" guidance for mismatch; relative
paths for filesystem errors), mirroring the `goadapter` error-message
convention.

**TS Ce semantics** (spike-validated, adapter must preserve): type-only and
side-effect imports count toward Ce; dynamic `import()` and externals are
excluded; relative `package.json` imports edge to the JSON's directory.

**Never** (hard constraints): read ripwire's internal cache; consume ranked
agent-facing answers as metrics input; vendor or bundle the ripwire binary.

**Performance**: one ripwire invocation per analysis phase (version, arch,
deps, export) with bounded capture; acceptable for CI-sized projects; the
`--timeout` flag bounds the whole analysis as it does for Go today.

## Risks / Trade-offs

**[R1] ripwire 0.x drift breaks the pin on every upstream release** →
Accepted by design: fail-closed is the honest contract without upstream
semver (issue #68 tracks the upstream ask). Mitigation: the mismatch error is
actionable, and re-validation is a fixture-refresh procedure (record new
goldens, bump two constants) documented in the adapter's GoDoc.

**[R2] Output-shape drift within a "validated" version** → Mitigation: golden
fixtures pin the exact parsed shapes; any parse mismatch fails tests loudly
rather than producing wrong metrics.

**[R3] Light-parse inaccuracy for Python/TS idioms** (decorators, nested
classes, computed member access, TS spread/destructuring) → Mitigation: the
parse is the documented metric definition with rules kept deliberately small;
unparseable files degrade to warnings + zeroed metrics + `partial` status, and
the measured error budgets are surfaced as warnings rather than hidden.

**[R4] Directory aggregation ambiguity** (same-named modules in different
trees, symlinked dirs) → Mitigation: deterministic path normalization,
ambiguity-rate warning, and the census/detection rules are pure and
table-tested.

**[R5] Provenance fields exist on the Go struct but not in the schema file's
strict listing** → Known trade-off inherited from the `commitSHA`/`branch`
precedent: `metrics.Validate` (the enforced contract) passes; external
consumers validating raw JSON against `modulegraph.schema.json` with
`additionalProperties: false` would reject these optional fields. Mitigation:
fields are `omitempty` (absent on Go graphs, droppable via `--no-provenance`);
a future dedicated schema change can formalize them if external schema
validation becomes a real consumer.

**[R6] Mixed-language monorepos get one adapter per invocation** → Detection
precedence is deterministic (go > python > typescript) and documented;
analyzing the other language requires pointing `analyze` at the subdirectory.
Accepted for this change; a `--language` flag is a recorded follow-up.

## Migration Plan

Purely additive: no existing flags, outputs, or exit codes change for Go
targets (regression-tested byte-equal). Python/TypeScript targets previously
failed with "no Go packages found"; they now activate the ripwire path, which
itself requires ripwire on PATH — the new failure mode is a strictly more
actionable error. Rollback is a plain revert; no data or schema migrations
are involved.

## Open Questions

- The exact `--arch`/`--deps`/cc.json field names of the 0.6.9 surface: to be
  pinned by recording the golden fixtures at implementation start (does not
  change the approach — the parser is written against the fixtures).
- Whether ripwire's version output reports the commit directly or it must be
  correlated another way: resolved while implementing D2's version check
  against the real binary; the fail-closed behavior is fixed either way.
