## Context

`internal/ripwireadapter` (change `ripwire-adapter`, PR #75) sources Python and
TypeScript coupling data from a PATH-discovered ripwire binary and fails closed
on any (version, commit) other than `0.6.9` @ `bedc0cd`. Upstream has since
landed all four Python import-resolution fixes from redhat-et/ripwire#478 —
gaps 2/3/4 in v0.6.10, gap 1 (redhat-et/ripwire#533) on `main` only — so the
pin has to move to reach users. See `proposal.md` for the fix table.

### Evidence base (probed 2026-11-19)

A Release build of ripwire `main` @ `1edb6fd6a` (contains #533; reports
`ripwire 0.6.10-dev.65+g1edb6fd6a (Release, …, built_from=1edb6fd6a)`) was run
with the committed capture commands (`testdata/recorded/README.md`) against
the two committed fixtures plus two probe variants, and the outputs were fed
through the adapter's real parsers (`parseArchOutput`, `parseDepsOutput`,
`mapToModules`) in an isolated worktree:

| Probe | `--arch` / `--deps` delta vs the 0.6.9 recordings | Real-parser result |
|-------|--------------------------------------------------|--------------------|
| Python fixture | row paths lose `./` (`./fixturepy` → `fixturepy`); `<pruned_packages>` element added; root gains `root="."`, loses `new_violations=`; legend comment rewritten; **all `ca`/`ce`/`types`/`abstract` identical**; `cc.json` byte-identical | parses; mapping identical to the golden |
| TypeScript fixture | same path and root changes; `<health>` gains `cyclic_files="2"`, `shape` `vertical` → `tangled`; coupling identical; `cc.json` byte-identical | parses; mapping identical |
| Python + one `.go` file (`cycle_a/helper.go`) | root `no_file_edges="go" unmeasured_files="1" graph_partial="1"`; `<metrics zone_unknown="1" masked_files="1">`; row `cycle_a`: `ca="1" ce="1" types="0" abstract="0" A="0.00" zone="unknown" no_file_edges="go"` — **no `I`, `D`, `reachable`** | **fails**: `module row "fixturepy/cycle_a" missing I attribute` |
| Python + package-less `dist/` | `--arch` lists only `operations/build` under `<pruned_packages>`; `dist` appears only in `--skipped` as `why="pruned-dir"` | parses |

The fixtures do not exercise any of the four fixes, which is why their
coupling is unchanged; the fixes' effect is measured separately (D6).

## Goals / Non-Goals

### Goals

- Users running `vibe-check analyze` on Python/TypeScript get the four upstream
  import fixes, through a pin to a published release.
- A Python/TypeScript tree containing files in a language ripwire marks
  `no_file_edges` analyzes, with the floor approximation surfaced as a warning,
  instead of failing.
- Error-budget warnings state numbers measured against the pinned binary.
- All development and review happen before the upstream release; the release
  only unblocks the final re-pin and re-capture.

### Non-Goals

- Consuming `--dump-graph` (#68 / redhat-et/ripwire#477; not landed upstream).
- Any change to Layer 1 (`metrics/`), `modulegraph.schema.json`, verdict
  thresholds, the report engine, language detection, or the Go adapter.
- Calling `--skipped` (an extra subprocess per analysis) — not needed; see D5.
- Supporting more than one ripwire version at a time.
- Using `--quality-delta` or any other new upstream verb.

## Decisions

### D1: Pin target — the first release containing redhat-et/ripwire#533

The pin moves once, directly to the first published release whose tag contains
#533's merge commit `0b8f0bf8`, expected `0.6.11` (`CMakeLists.txt` on `main`
still declares `0.6.10`; the release number is upstream's choice and the pin
follows whatever it is). `0.6.10` is skipped as an intermediate pin: it lacks
gap 1 — the largest measured effect (django Ce 60% → 100% exact) — and every
pin move re-records every golden.

The pin semantics are unchanged: exact version string match plus `built_from=`
commit-prefix match (7 hex characters, as today). Development builds stay
refused by construction: their version token is `X.Y.Z-dev.N+g<sha>`, never
equal to a release version, and the existing parser already takes the second
whitespace-separated token verbatim.

**Single source of truth.** `pinnedRipwireVersion` / `pinnedRipwireCommit`
remain the only code sites. GoDoc that currently names `0.6.9` (`doc.go`,
`adapter.go`, `parse_arch.go`, `parse_deps.go`, `denylist.go`, `assembly.go`,
`cycles.go`) is reworded to "the pinned ripwire release" so the next bump does
not touch it; tests that hardcode `0.6.9`/`bedc0cd` reference the constants or
the recorded `version.txt` instead. Prose that must name the version (README,
CHANGELOG, AGENTS.md, testdata README) is updated by hand and checked by grep in
the task list.

*Alternative rejected:* accepting a version range (`>= 0.6.11`). Upstream is
0.x with no semver policy and changed the `--arch` surface in a minor release;
a range would have admitted exactly the break this change fixes.

### D2: Narrow the `--arch` parse contract to consumed attributes

Today `parseArchOutput` requires every row attribute (`ca ce types abstract I A
D zone`) and seven `<metrics>` aggregates (`modules typed_modules zone_pain
zone_useless zone_ok zone_na propagation_cost`). Downstream, only `Path`,
`Ca`, `Ce`, `Types`, and `Abstract` are read (`mapping.go:94-95`; `I`, `D`, and
zone are recomputed from `metrics.Compute*` in `assembly.go:202-205`). The
unconsumed fields are exactly what ARCH-HONEST omits.

New contract:

- **Required on every row, fail-closed** (unchanged): `path`, `types`,
  `abstract`; duplicate paths remain an error.
- **`ca` / `ce`**: required on every row **except** a row with
  `zone="unknown"`, where an absent value is read as a floor of `0` and the row
  is flagged `CouplingFloor=true`. A present value on an unknown row is also a
  floor and flags the row. A missing `ca`/`ce` on any other row stays an error.
- **`zone`**: read only to recognize `"unknown"`; its other values are not
  interpreted. Missing `zone` is tolerated (treated as not-unknown, which keeps
  `ca`/`ce` required — fail-closed in the direction that matters).
- **`no_file_edges`**: read per row (comma-separated language list) and carried
  on the flagged row for the warning text.
- **Dropped from the wire structs**: `I`, `A`, `D`, `reachable`, `isolated`, and
  all `<metrics>` aggregates. The `<metrics>` element itself stays required.
- **New: `<pruned_packages>`** under `<arch>`: optional; each `<p path=…>` is
  collected (D5). `count=` vs `shown=` mismatch (ripwire caps the listing at 8)
  is recorded so the warning can say the list is partial.

*Alternative rejected:* modeling every ARCH-HONEST attribute (`zone_unknown`,
`masked_files`, `graph_partial`, …). It adds parse surface the adapter never
reads and re-creates the brittleness that caused the break.

Row paths: ripwire now prints `fixturepy` where it printed `./fixturepy`.
`moduleIdentity` already strips an optional `./`, so both forms map to the same
identity; a table test pins that both forms (and the root row `.`) are
accepted. The parser stays verbatim; normalization stays in mapping.

### D3: `coupling-floor` warning

When at least one mapped module is flagged, the graph carries one warning:

- Code: `coupling-floor` (new, stable).
- Message: `ripwire resolves almost no imports for <langs>; coupling for <n>
  module(s) is a floor, not an exact count: <id1>, <id2>, …` — languages are the
  sorted union of the flagged rows' `no_file_edges=` values, identities sorted
  lexicographically, so the text is deterministic.
- Values: the floors flow into `metrics.ComputeInstability` /
  `ComputeDistance` / `ComputeZone` unchanged. Instability over a floor pair is
  not itself a bound in either direction; the warning says the inputs are
  floors and the status is `partial` (every warning forces `partial`,
  `budget.go:graphStatusForWarnings`), so no consumer reads it as exact.

*Alternatives rejected:* (a) dropping flagged modules — silently shrinks the
graph, which is worse than a labelled floor; (b) failing the analysis — the
current behavior, and the reason for this decision; (c) a per-module marker —
needs a schema change to `ModuleResult`, which is a protected value.

### D4: Determinism unaffected

Nothing in D2/D3/D5 introduces map-iteration order or time into output: the
warning lists are sorted, and the existing two-run byte-equality tests cover
the new recording (D7).

### D5: Denylist warning = filesystem pre-check ∪ ripwire `<pruned_packages>`

ripwire 0.6.10+ rows pruned **packages** (directories with `__init__.py`) on
`--arch`, capped at 8 with an exact `count=`; plain pruned directories
(`why="pruned-dir"`) appear only in `--skipped`. The existing filesystem
pre-check (`detectDenylistPruned`) already finds both classes from a hardcoded
name list, so it stays. The warning lists the sorted, de-duplicated union of
both sources, with ripwire's paths taken as crawl-root-relative (ripwire prints
them relative to `root="."`, which is the crawl root). If ripwire's
`count=` exceeds its shown rows, the message appends `(ripwire reports N pruned
packages; see ripwire --skipped)`.

Why union rather than either alone: ripwire's list is authoritative for
packages but incomplete for plain directories; the pre-check is complete for
the names it knows but goes stale silently if upstream changes its denylist.
The union costs no subprocess (it reuses the `--arch` call already made).

**Amendment (Phase B, ripwire 0.7.0).** A denylist name is now a candidate,
not a verdict: 0.7.0 indexes a `build/` whose files git tracks. The
pre-check therefore drops any candidate that has an `--arch` module row at or
beneath it (it was crawled), and keeps descending into it so a nested
candidate ripwire did not index is still reported. Using ripwire's own rows
rather than re-implementing its git query keeps the adapter free of a git
dependency and correct if upstream extends the rule to other names.
Untracked files inside a tracked `build/` are left out by ripwire and not
reported by the pre-check (ripwire's `--skipped` lists them as
`why=untracked-build`).

### D6: Re-measured error budgets with a committed harness

`errorBudgetWarning` hardcodes the spike's figures (Python 1.9% lost / 2.2%
spurious; TypeScript 0.16% unresolved), measured on 0.6.9. Four fixes later
they are stale. The spike harness was throwaway, so this change rebuilds it
reproducibly:

- Location: `openspec/changes/bump-ripwire-pin/measurement/` — spec-phase
  evidence that archives with the change; not production code, not built by
  CI.
- Python: `ast`-based ground truth of every import statement in pip @
  `1000636` resolved to in-tree files (the spike's method), compared with the
  pinned binary's `--deps --limit=N` file graph → lost % and spurious %.
  A django run is added because the wide-form fix's effect was measured there.
- TypeScript: rollup @ `ad39ddc`, unresolved-target rate from `--deps`.
- Output: a `results.md` with commands, target commits, binary version line,
  and figures; the warning text cites the figures and the pinned version.

Gate: if a re-measured budget is **worse** than the 0.6.9 figure for the same
target, implementation stops and reports instead of shipping the bump (AGENTS.md
gatekeeping — a fidelity regression is not ours to accept silently).

### D7: Fixtures — exercise the fixes, add an ARCH-HONEST recording

- Extend the Python fixture with three packages, one per fix, so each shows
  independently at directory grain (probed on the `main` build):

  | Package | Import | Fix | Probed `--arch` row |
  |---------|--------|-----|---------------------|
  | `fixturepy/wide/` | `from fixturepy import cycle_b` | gap 1 | `ce="2"` (→ `fixturepy`, → `fixturepy/cycle_b`) |
  | `fixturepy/alias/` | `import fixturepy.cycle_a.api as ca` | gap 2 | `ce="1"` (→ `fixturepy/cycle_a`) |
  | `fixturepy/collide/` | `import logging` beside a sibling `logging.py` | gap 4 | `ce="0"` (no false sibling edge) |

  Gap 1 must import a **subpackage**: at directory grain `pkg/__init__.py` and
  `pkg/mod.py` are one module, so `from pkg import mod` for a plain module file
  is invisible to the fix. The goldens then pin the improved resolution; a pin
  regression shows up as a golden diff, not only as a budget drift.
- Add `testdata/recorded/python-mixed/` — the Python fixture plus one `.go` file
  — captured from the pinned binary; the fake-ripwire harness replays it like the
  others. Integration test asserts analysis succeeds, the `coupling-floor`
  warning names the module and `go`, and status is `partial`.
- Parser unit tests cover: unknown row without `ca`/`ce`; unknown row with
  floors; non-unknown row missing `ca` (still an error); missing `zone`; both
  path forms; `<pruned_packages>` with `count > shown`.

### D8: Two-phase capture — develop on `main`, finalize on the release

Phase A (now): recordings are captured from a Release build of ripwire `main`
(≥ `1edb6fd`) with the existing capture commands; all code, tests, goldens,
and the measurement are developed and reviewed against them. The version pin
cannot match a dev build, so integration tests run through the fake harness,
whose `version.txt` is temporarily a hand-written release-shaped line marked
in the testdata README as provisional.

Phase B (after upstream publishes the release): download the release archive,
verify its published digest, re-capture every recording twice (byte-stability
method unchanged), confirm the coupling content is identical to Phase A's
(only `version.txt` and legend comments may differ — any other diff reopens
review), set the pin constants, update the provenance table in the testdata
README, re-run the measurement harness against the release binary, and run CI
parity.

The PR stays in draft through Phase A and is marked ready only after Phase B.

**Phase B outcome (2026-11-24).** Upstream skipped 0.6.11 and released
**0.7.0** @ `5a3df049e`, 681 commits past the Phase A build; the pin is
`0.7.0` / `5a3df04`. The re-capture was byte-identical for `typescript/` and
`python-mixed/` (apart from `version.txt`). `python/` differed, which reopened
review per this decision: 0.7.0 indexes a `build/` directory whose files git
tracks (upstream "a `build/` directory whose files git tracks is indexed";
`dist/`, `out/`, `target/` stay pruned), and the `python/` capture runs inside
this repository's checkout, so `fixturepy/operations/build` became an
isolated eighth module with no `<pruned_packages>`. Every other row is
identical. The adapter follows in D5's amendment; the measurement harness
gained `--tracked-build` to mirror the rule, and every figure held (pip
`src/` now has 54 directories, 100% exact).

## Risks / Trade-offs

- **Upstream changes the surface again before the release.** Phase B's
  re-capture diff catches it; anything beyond version/legend text reopens
  review. Accepted: the cost is a second capture, not a wrong ship.
- **Release timing is upstream's.** Phase A carries no user value until Phase
  B. Accepted: the work is review-ready, and the alternative (pinning a dev
  commit) would ship an unreproducible install instruction.
- **Coupling deltas across the upgrade boundary.** A `vibe-check diff` whose
  base snapshot came from 0.6.9 and PR snapshot from the new pin can show
  coupling changes caused by better resolution, not code. Mitigation:
  CHANGELOG states it and recommends re-snapshotting the base; provenance
  already records `sourceToolVersion`, so the mismatch is detectable. Making
  `diff` warn on mismatched `sourceToolVersion` is a sensible follow-up but a
  Layer-3 change outside this scope.
- **Floor semantics could be misread** as exact by a consumer ignoring
  warnings. Mitigated by `partial` status; a per-module schema marker is
  deliberately out of scope (protected schema).
- **Hardcoded denylist drift** remains possible for plain directories (ripwire
  rows only packages on `--arch`). Accepted; calling `--skipped` per analysis
  costs a subprocess for a rare case.
- **Wide-form member edges are invisible to cycle derivation** (found in
  Phase A implementation, task 1.2). Gap 1 reaches Ca/Ce — the `--arch`
  table counts ripwire's resolved edges — but not the adapter's own edge
  set: `--deps` records an import as written (`from fixturepy import
  cycle_b` → `<inc t="fixturepy"/>`, the imported name dropped), and no
  verb exposes the resolved target files (`cc.json` carries only per-file
  `fan_in`/`fan_out` counts). The adapter's directory edges — and so its
  Tarjan SCC — therefore miss a cycle closed only through a wide-form
  subpackage import. This is not a regression: on 0.6.9 Ca/Ce and the edge
  set both missed the edge; after the bump only the edge set does, so a
  module's Ce can exceed its internal out-edges. Pinned in
  `mapping_test.go` and disclosed in the README `cycle-approximation`
  row. Remedies, both out of scope here: resolve `from X import n` members
  in vibe-check's own Python light parse (consistent with the hybrid
  split), or consume resolved edges from ripwire's `--dump-graph`
  (redhat-et/ripwire#477, #68) once it lands.
- **ripwire's `<cycles>` listing cap changed** (found in Phase A,
  redhat-et/ripwire#479 item 2): the release shows max(20, `--limit`)
  SCCs, so under the adapter's `--limit=1000` the 20-SCC cap the
  `cycle-approximation` message quoted no longer applies.
  `ripwireCycleMaxSCCs` now derives from the adapter's `depsWindowLimit`;
  the 12-member cap is unchanged and now disclosed upstream by
  `members_capped=1`.
- **Gap 4's disclosed floor** (upstream): a script inside a package can lose a
  true sibling edge (`src/__init__.py` + `python src/train.py`). Inherited
  from upstream; folded into the re-measured Python budget rather than a
  separate warning.

## Constitution alignment

See `proposal.md`. Observable Quality is the decision driver for D3 and D6
(approximations surface as machine-readable warnings with measured figures);
Composability First constrains D1/D5 (ripwire stays optional, no extra
subprocess); Testability drives D7/D8 (hermetic replay of real recordings).
