<!--
  [P] marks tasks eligible for parallel execution.
  Phase A (groups 1–6) runs now against a Release build of ripwire `main`
  (>= 1edb6fd, contains redhat-et/ripwire#533). Phase B (group 7) is blocked
  until upstream publishes the first release containing #533 (design D8).
-->

## 1. Capture baseline from a ripwire `main` build (Phase A)

- [x] 1.1 Build ripwire `main` (>= `1edb6fd`) as a Release binary outside the repository (`cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build --target ripwire`) and record its `--version` line. Verify: the version line contains `-dev.` and `built_from=` a commit that has `0b8f0bf8` (#533) as an ancestor.
- [x] 1.2 Extend the Python fixture with `fixturepy/wide/` (`from fixturepy import cycle_b`), `fixturepy/alias/` (`import fixturepy.cycle_a.api as ca`), and `fixturepy/collide/` (`import logging` beside a sibling `logging.py`), each with an `__init__.py`; document them in `testdata/FIXTURES.md` (design D7). Verify: a capture shows `wide ce=2`, `alias ce=1`, `collide ce=0` on the `--arch` rows.
- [x] 1.3 Create the mixed-language recordings `testdata/recorded/python-mixed/`, captured from a temporary copy of the Python fixture plus `src/fixturepy/cycle_a/helper.go` and replayed over the standard `python/` tree (the `python-zero-edge` pattern — no duplicated fixture tree is committed); document it in `FIXTURES.md`. Verify: its `--arch` capture carries `zone="unknown" no_file_edges="go"` on `fixturepy/cycle_a` with no `I`/`D` attribute.
- [x] 1.4 Re-capture every recording (`python`, `python-zero-edge`, `typescript`, `python-mixed`) with the committed capture commands, twice each, `cmp`-ing the pairs; write a provisional release-shaped `version.txt` and mark it provisional in `testdata/README.md`. Verify: all pairs byte-identical, every invocation exit 0 with empty stderr, fixture trees checksum-unchanged.

## 2. Parse contract (design D2)

- [x] 2.1 Rewrite `parse_arch.go` to the consumed-attribute contract: require `path`/`types`/`abstract` always and `ca`/`ce` unless `zone="unknown"`; read `zone` only to recognize `unknown`; carry `no_file_edges` and a `CouplingFloor` flag on the row; drop `I`/`A`/`D`/`reachable`/`isolated` and all `<metrics>` aggregates from the wire and table structs; parse optional `<pruned_packages count= shown=>` with its `<p path=>` rows. Update GoDoc. Verify: `parse_arch_test.go` table cases — unknown row without `ca`/`ce` (floor 0, flagged), unknown row with floors (flagged), non-unknown row missing `ca` (error naming row and attribute), missing `zone` (requires `ca`/`ce`), absent `propagation_cost` (ok), `<pruned_packages>` with `count > shown`, duplicate path (error) — all pass.
- [x] 2.2 [P] Add a `mapping_test.go` table case asserting `./fixturepy/cycle_a`, `fixturepy/cycle_a`, and `.` map to the expected identities for both languages. Verify: test passes against the unchanged `moduleIdentity`.
- [x] 2.3 [P] Confirm `parse_deps.go` tolerates `<pruned_packages>` and the new `<health>`/root attributes on the Phase A recordings; update its GoDoc to name "the pinned ripwire release". Verify: `parse_deps_test.go` passes on every re-captured `deps.xml`.

## 3. Coupling-floor warning (design D3)

- [x] 3.1 Carry the row's `CouplingFloor` flag and `no_file_edges` languages through `mapToModules` onto `mappedModule`. Verify: a mapping test on the `python-mixed` recording marks only `fixturepy.cycle_a`.
- [x] 3.2 Emit one `coupling-floor` warning in `Analyze` assembly when any module is flagged: sorted identities, sorted language union, deterministic message per D3; no warning otherwise. Verify: assembly tests for flagged/unflagged tables; the message is byte-stable across two runs.

## 4. Denylist union (design D5)

- [x] 4.1 Make the `denylist-pruned` warning list the sorted, de-duplicated union of `detectDenylistPruned` results and the `--arch` `<pruned_packages>` paths, appending ripwire's total when `count > shown`. Verify: `denylist_test.go` cases — both sources agree (listed once), pre-check only (plain `dist/`), ripwire only, `count > shown` disclosure, neither (no warning).

## 5. Error-budget re-measurement (design D6)

- [x] 5.1 Write the measurement harness under `openspec/changes/bump-ripwire-pin/measurement/`: a Python `ast` ground-truth resolver for in-tree imports and a comparator against `--deps --limit=N` file edges (lost %, spurious %); a TypeScript unresolved-target counter over `--deps`; a `README.md` with exact commands. Run Python with `python3 -I`. Verify: the ground-truth side reproduces the spike's pip @ `1000636` corpus (668 `.py` files, 77 package dirs) and rollup @ `ad39ddc` corpus (274 src `.ts` files), and the harness runs end-to-end on the `main` build.
- [x] 5.2 Measure pip @ `1000636`, django (pinned commit recorded in `results.md`), and rollup @ `ad39ddc` against the `main` build; write `measurement/results.md`. Verify: `results.md` lists target commits, binary version line, and figures. **Gate**: if any figure is worse than its 0.6.9 baseline (Python 1.9% lost / 2.2% spurious; TS 0.16% unresolved), stop and report — do not proceed.
- [x] 5.3 Update `budget.go` messages to the measured figures and name the pinned version via the `pinnedRipwireVersion` constant. Verify: `budget_test.go` asserts the new text; no literal version string in `budget.go`.

## 6. Goldens, docs, single-source pin (Phase A)

- [x] 6.1 Re-derive and commit the hand-computed goldens for `python`, `typescript`, and `python-mixed` from the Phase A recordings; record each value's derivation in `golden-notes.md`. Verify: integration tests through the fake harness match the goldens, including `wide`/`alias`/`collide` values and the `coupling-floor` warning on `python-mixed`.
- [x] 6.2 [P] Reword GoDoc naming `0.6.9`/`bedc0cd` in `doc.go`, `adapter.go`, `parse_arch.go`, `parse_deps.go`, `denylist.go`, `assembly.go`, and `cycles.go` to version-neutral text; replace hardcoded `0.6.9`/`bedc0cd` in tests (`version_test.go`, `parse_*_test.go`, `integration_test.go`, `assembly_test.go`, `fake_ripwire_test.go`, `cmd/vibe-check/cli_ripwire_test.go`) with the pin constants or recorded `version.txt`. Leave `metrics/` untouched (its `0.6.9` strings are illustrative examples, and Layer 1 is out of scope). Verify: `grep -rn '0\.6\.9\|bedc0cd' --include=*.go internal/ripwireadapter cmd` returns only the deliberate previous-pin refusal test.
- [x] 6.3 [P] Add the version-test cases from the spec: previous pin `0.6.9`/`bedc0cdcf` refused; dev build `0.6.10-dev.65+g1edb6fd6a` refused. Verify: `version_test.go` passes.
- [x] 6.4 Run the two-run byte-equality determinism tests across all recordings, then CI parity from `.github/workflows/ci.yml` (build, vet, `go test -race -count=1`, coverage ratchet, golangci-lint) in a ripwire-free shell. Verify: every command exits 0; no coverage floor lowered. (Done 2026-11-19: all pass; golangci-lint is clean on changed packages — local v2.15.0 reports two `errcheck` findings in untouched `metrics/external.go` that predate this branch; CI pins v2.12.2.)
- [x] 6.5 Draft README (ripwire install prerequisite, metric-sources/caveats incl. `coupling-floor`), CHANGELOG `[Unreleased]` (Changed: pin; Added: `coupling-floor`; note that coupling values shift across the upgrade boundary and base snapshots should be re-taken), and AGENTS.md (`version.go` line) with a `<RELEASE>` placeholder for the version. Verify: `grep -n '<RELEASE>'` lists every placeholder for Phase B.
- [x] 6.6 Open the PR as **draft** after the review council approves Phase A, file the documentation issue (AGENTS.md documentation gate), and note in the PR body that it is blocked on the upstream release. Verify: PR is draft; docs issue linked.

## 7. Finalize on the upstream release (Phase B — blocked on upstream)

- [x] 7.1 When upstream publishes the first release containing `0b8f0bf8` (#533), download the macOS arm64 archive, verify its published sha256, extract outside the repository, and confirm `--version` reports the release version and a `built_from=` commit containing #533. Verify: digests and version line recorded in `testdata/README.md`'s provenance table.
- [x] 7.2 Set `pinnedRipwireVersion` / `pinnedRipwireCommit` to the release pair; replace the provisional `version.txt` files with real captures. Verify: version tests pass; the dev build from 1.1 is now refused.
- [x] 7.3 Re-capture every recording from the release binary (twice, `cmp`). Verify: coupling content is identical to Phase A — only `version.txt` and legend comments may differ; **any other diff reopens review** (design D8). (Done 2026-11-24 with ripwire 0.7.0: `typescript/` and `python-mixed/` identical; `python/` differs because 0.7.0 indexes the git-tracked `operations/build` — review reopened; the adapter, goldens, docs, spec, and design D5/D8 amended, see design "Phase B outcome".)
- [x] 7.4 Re-run the measurement harness against the release binary; update `results.md` and `budget.go` if any figure moved. Verify: the 5.2 gate still holds.
- [x] 7.5 Replace every `<RELEASE>` placeholder; run CI parity (as 6.4) and the real-binary integration test with the release binary on PATH. Verify: all pass; `grep '<RELEASE>'` is empty. (Done 2026-11-24: build, vet, race tests, ratchet 8/8 ripwire-free; real-binary tests pass with 0.7.0 on PATH; lint clean apart from the two pre-existing `metrics/external.go` errcheck findings under local v2.15.0.)
- [x] 7.6 Re-run the review council on the Phase B diff, then mark the PR ready. Verify: council APPROVE; no code change between approval and ready. (Done 2026-11-24: council approved 7421b3e; PR #96 marked ready with no further change and merged as 58a6f3a.)

## Workflow follow-up

- On PR merge: comment the adopted pin on redhat-et/ripwire#478 and #68.
- #80/#81 content should cite `measurement/results.md`.
- Candidate follow-up (out of scope, D-risks): `vibe-check diff` warns when base and PR `sourceToolVersion` differ.
