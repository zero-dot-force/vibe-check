# Tasks

## 1. Fixtures and Pinned Surface

- [x] 1.1 Create fixture projects under `internal/ripwireadapter/testdata/`: (a) a src-layout Python project (package under `src/`, with an `operations/build` source directory, a cohesive class, a split-field class, an `abc.ABC` class, a concrete class, an underscore-prefixed class, and a two-directory import cycle) and (b) a TypeScript project (tsconfig.json, interfaces incl. a nested one, a class with `this.x`/`this.y` method splits, and a two-directory import cycle). Verify: both projects exist on disk with the listed structures and are deterministic (no timestamps in content).
- [x] 1.2 Record ripwire 0.6.9 @ `bedc0cd` outputs for both fixture projects — version output, `--arch` metrics table, `--deps` XML, `--export=cc.json` — into `internal/ripwireadapter/testdata/`. Verify: recorded files exist, are byte-stable across two capture runs, and a README in the testdata dir documents the capture commands and the validated (version, commit) pair.
- [x] 1.3 Build the fake-riprewire test harness: an executable placed on `PATH` during tests that replays the recorded outputs by flag, with switchable variants (validated version, wrong version, garbage version output). Verify: a unit test invoking the fake through the adapter's subprocess runner gets the recorded bytes.
- [x] 1.4 Hand-compute expected `ModuleGraph` goldens for both fixture projects from the recorded outputs and fixture structure (Ca/Ce per directory, cycle membership, TS interface share, Python abstractness, LCOM4 values). Verify: golden JSON files exist and are committed alongside the recorded inputs.

## 2. Adapter Core — Discovery, Version Pin, Subprocess Hygiene

- [x] 2.1 Scaffold `internal/ripwireadapter/` (`doc.go`, `adapter.go`): `NewPython()` and `NewTypeScript()` constructors returning distinct `metrics.Adapter` implementations; `Language()` returns `"python"`/`"typescript"`; `Capabilities()` returns exactly ca, ce, instability, abstractness, distance, lcom, circular (no duplication); compile-time interface compliance check. Verify: unit tests assert language identifiers and the exact capability set per language.
- [x] 2.2 Implement the fail-closed version pin: `exec.LookPath("ripwire")` at `Analyze` time; read the binary's version and commit; compare against pinned constants (`0.6.9`, `bedc0cd`); absent binary → error with install guidance naming https://github.com/redhat-et/ripwire and the validated version; mismatch → error naming both versions; unparseable output → error. Verify: tests via the fake harness cover absent (PATH without fake), matching, mismatching, and garbage version outputs; no ModuleGraph is produced on any failure path.
- [x] 2.3 Implement the subprocess runner: `exec.CommandContext` with `metrics.SanitizeEnvironment(nil)`, working directory at the crawl root, bounded stdout/stderr capture (oversized output is an error, never truncation), context cancellation terminating the subprocess. Verify: tests assert `GITHUB_TOKEN` is not passed, a cancelled context returns an error wrapping the context error with no orphaned process, and oversized output errors.
- [x] 2.4 Add optional provenance-input fields to `metrics.ProvenanceInput` (`sourceTool`, `sourceToolVersion`, `sourceToolCommit`, all `omitempty` with GoDoc), leaving `modulegraph.schema.json` and `metrics/validate.go` untouched. Verify: tests marshal/unmarshal the fields, `metrics.Validate` passes with them present and absent, and existing metrics tests still pass.

## 3. Crawl Scoping and Pre-Checks

- [x] 3.1 Implement package-root detection and crawl scoping: Python src-layout (`<root>/src` with `__init__.py` packages) vs flat layout; TypeScript root (tsconfig.json/package.json directory); pass `--exclude` scoping so the crawl starts at the package root, never the repo root. Verify: the src-layout fixture yields non-zero coupling (the pip failure mode), flat and TS fixtures resolve their roots, and root-detection unit tests pass.
- [x] 3.2 Implement the crawl denylist pre-check: detect source-bearing directories under the crawl scope named `build`/`dist`/`out`/`target`/`vendor`/`third_party`/`captures` and emit a `Warning` (code `denylist-pruned`) listing them. Verify: the fixture's `operations/build` directory triggers the warning; a clean fixture emits none.
- [x] 3.3 Implement the zero-coupling guard: when a crawl yields no dependency edges while source files exist in scope, refuse to emit a status-`complete` all-zero graph (error or `partial` + `zero-coupling-guard` warning). Verify: a synthetic empty-edge crawl test asserts the guard fires.

## 4. Parsing and Mapping

- [x] 4.1 Implement the `--arch` metrics-table parser: read per-unit Ca/Ce/I from the `<metrics>` table and aggregate attributes; never read per-file `<inc>` lists for counts (40-cap). Verify: parser tests against the recorded fixtures assert values and that a fixture with a 40-plus `<inc>` list still yields uncapped counts.
- [x] 4.2 Implement the `--deps` XML parser (stdlib `encoding/xml`) and the `--export=cc.json` per-file fan-in/out reader (stdlib `encoding/json`), both adapter-internal. Verify: parser tests against the recorded fixtures assert the parsed edge and fan-in/out sets.
- [x] 4.3 Implement directory-module mapping and aggregation: D4 identity rules (Python dot-style, TS slash-style, `Name` = last segment), per-file values summed into directories, edges deduplicated per directory pair, externals counted in Ce only; emit `name-resolved-edges`, `ambiguity-rate`, and `directory-as-module` warnings. Verify: mapping the same recorded output twice yields identical module coupling; a unit test asserts each warning code.
- [x] 4.4 Implement cycle derivation: Tarjan SCC over the directory-module edge graph, canonical lexicographic ordering (members and cycle slice), corroboration against ripwire's capped cycle listing (20 SCCs × 12 members, no total), and a `cycle-approximation` warning. Verify: the two-directory cycle in each fixture is detected with canonical ordering and the warning is present; a constructed multi-SCC adjacency test checks ordering.
- [x] 4.5 Implement error-budget warnings: emit `error-budget` warnings carrying the spike-measured budgets (Python 1.9% lost / 2.2% spurious; TS 0.16% unresolved) per language. Verify: tests assert the warning text and that any-warning graphs have status `partial`.

## 5. Light Language Parse

- [x] 5.1 Implement the Python abstractness light parse: `class Name(Bases)` detection; abstract when bases reference `abc.ABC`, `ABCMeta`, or `typing.Protocol`; exported when the name is not underscore-prefixed; ripwire's structurally-zero Python abstractness never used. Verify: fixture tests assert the `abc.ABC` class counts in both `AbstractTypes` and `ExportedTypes`, the concrete class only in `ExportedTypes`, and `_Internal` in neither.
- [x] 5.2 Implement the LCOM4 light parse for both languages: class-bound methods as nodes (Python `def` with leading `self` in a class body; TS method members by brace depth), edges on shared instance-field access (`self.X` / `this.X`), connected components via union-find per directory module; module-level functions excluded. Verify: cohesive class → LCOM 1, split-field class → LCOM 2, module with no bound methods → LCOM 0, in both language fixtures.
- [x] 5.3 Implement unparseable-file degradation: a file the light parse cannot process is skipped with a `parse-error` warning, its affected metrics zeroed, and the graph status set to `partial` (mirroring `goadapter`'s partial-build pattern). Verify: a fixture with a deliberately malformed file produces the warning and a `partial` graph.

## 6. TypeScript Abstractness and Graph Assembly

- [x] 6.1 Implement TypeScript abstractness from ripwire interface share: `AbstractTypes` counts interfaces including nested ones; `ExportedTypes` reflects the type-count denominator; emit a `ts-abstractness-bias` warning documenting the `abstract class` undercount and type-alias denominator inflation. Verify: the TS fixture's interface counts (incl. nested) match the hand-computed golden and the bias warning is present.
- [x] 6.2 Implement `Analyze` assembly for both languages: hybrid merge of ripwire coupling data with light-parse results, derived metrics via `metrics.Compute*` (never reimplemented), `unsupported-metric` warning for duplication, warnings assembled, status `partial` when any warning exists, modules sorted by path, non-nil module/cycle/warning slices, `SchemaVersion` = current, `Language` set, provenance input populated with `sourceTool`/`sourceToolVersion`/`sourceToolCommit` and analyzed path/module identity. Verify: assembled graphs for both fixtures pass `metrics.Validate`, match the hand-computed goldens, and provenance fields are present.

## 7. CLI — Detection, Registry, Go Unchanged

- [x] 7.1 Implement `detectLanguage(path)` as a pure function in `cmd/vibe-check`: `go.mod` → `go`; else `pyproject.toml`/`setup.py`/`setup.cfg` → `python`; else `tsconfig.json` → `typescript`; else deterministic `.py` vs `.ts`/`.tsx` census with ties defaulting to `go`. Verify: table-driven tests cover every marker, the census majority, the tie default, and the markerless fallback to the existing Go error path.
- [x] 7.2 Wire the registry into `RunAnalyze`: construct a `metrics.Registry` with the Go adapter plus both ripwire adapter instances and select via `detectLanguage` (go-analyze D10's deferred refactor). Verify: unit tests assert Go targets select the native adapter without invoking ripwire and Python/TS targets select the ripwire instances.
- [x] 7.3 Implement the absent-binary CLI path: Python/TS target without ripwire on PATH exits with code 2 and an error naming ripwire, the install location, and the validated version. Verify: a CLI test with a ripwire-free `PATH` asserts exit code 2, the guidance text, and no ModuleGraph on stdout.
- [x] 7.4 Add the Go-path regression test: analyze a Go testdata module before-change vs after-change (recorded golden) and assert byte-identical output under `--no-provenance`. Verify: the regression test passes with ripwire absent from `PATH`.

## 8. Integration, Determinism, CI Parity, Docs

- [x] 8.1 Write full-pipeline integration tests via the fake ripwire on `PATH` for both fixture projects: assert specific Ca/Ce/cycle/LCOM/abstractness values against the goldens, all expected warning codes, status `partial`, and `metrics.Validate` passing. Verify: `go test ./internal/ripwireadapter/ -run Integration` passes without ripwire installed.
- [x] 8.2 Write the two-run byte-equality determinism tests: run `Analyze` twice per fixture project with provenance omitted and assert byte-identical ModuleGraph JSON. Verify: determinism tests pass under `-race -count=1`.
- [x] 8.3 Add the optional real-binary integration test guarded by `exec.LookPath("ripwire")` + `t.Skip` when absent (corroborates fixtures against a real 0.6.9 install). Verify: the test skips cleanly on a ripwire-free machine and passes locally where ripwire 0.6.9 is installed.
- [x] 8.4 Run CI parity from `.github/workflows/ci.yml` locally without ripwire installed: `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...` all pass; every exported symbol in new code has GoDoc. Verify: the three commands exit 0 in a ripwire-free shell.
- [x] 8.5 Update documentation: `AGENTS.md` project-structure section for `internal/ripwireadapter/`, `CHANGELOG.md` entry for the optional multi-language adapter, README usage note for Python/TypeScript analysis + ripwire install guidance, and resolve documentation issue #70 before PR merge. Verify: documented commands run as written and #70 is closed or explicitly tracked in the PR body.

## Workflow follow-up

- Run `/opsx-apply` (or `/uf.unleash`) to implement this change; artifacts are ready.
- The PR for this change carries `Closes #69`; documentation issue #70 must be resolved before merge.
- Upstream ask #68 (ripwire version/semver stability) is tracked separately and is not a blocker.
