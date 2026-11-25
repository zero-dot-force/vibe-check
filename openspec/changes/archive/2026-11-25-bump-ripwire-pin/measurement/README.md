# Error-budget measurement harness

`openspec/changes/bump-ripwire-pin` design D6, tasks 5.1–5.2 (Phase A) and 7.4
(Phase B). Spec-phase evidence that archives with the change: never built,
imported, or run by CI. Results are in [`results.md`](results.md).

| File | Purpose |
|------|---------|
| `measure_python.py` | `ast` ground truth vs ripwire `--arch` (directory Ca/Ce) and `--export=cc.json` (per-file fan-out). Method and conventions in its docstring. |
| `measure_typescript.py` | ripwire's own `imports_unresolved=` over the TS/JS import rows of `--deps`. |
| `arch-rules.txt` | `--arch` rules input (a copy of `internal/ripwireadapter/testdata/recorded/arch-rules.txt`; no rule can fire). |

Both scripts only **parse** the measured trees — nothing in them is imported or
executed. Run them with `python3 -I` (isolated mode) from a directory outside
the measured tree.

## Targets

| Target | Commit | Crawl root (as the adapter chooses it) |
|--------|--------|------------------------------------------|
| pypa/pip | `1000636317b856a7a8d0ae6662c3eb30c2a9bfed` (the Phase 0 spike's) | `src/` (src layout) |
| django/django | `03b95b8fbab990ebbe399a4676e4ad581fe9acfd` (`main`, 2026-11-19) | repository root (flat layout) |
| rollup/rollup | `ad39ddc2856a229bf28762a7a9399757527cbffc` (the spike's) | repository root (`tsconfig.json`) |

## Commands

```bash
fetch() { mkdir -p "$1" && git -C "$1" init -q && git -C "$1" remote add origin "https://github.com/$2.git" \
  && git -C "$1" fetch -q --depth 1 origin "$3" && git -C "$1" checkout -q --detach FETCH_HEAD; }
fetch pip    pypa/pip      1000636317b856a7a8d0ae6662c3eb30c2a9bfed
fetch django django/django 03b95b8fbab990ebbe399a4676e4ad581fe9acfd
fetch rollup rollup/rollup ad39ddc2856a229bf28762a7a9399757527cbffc

H=path/to/openspec/changes/bump-ripwire-pin/measurement
python3 -I "$H/measure_python.py"     "$RIPWIRE" pip/src --json --tracked-build
python3 -I "$H/measure_python.py"     "$RIPWIRE" django  --json --tracked-build
python3 -I "$H/measure_typescript.py" "$RIPWIRE" rollup  --json
```

`--tracked-build` mirrors ripwire 0.7.0's crawl rule (a `build/` directory
whose files git tracks is indexed); pass it for 0.7.0 and later, omit it for
0.6.9, 0.6.10, and the Phase A `main` build, which prune every `build/`.

Run every target once per binary: the pinned release, and ripwire 0.6.9 as the
like-for-like baseline (release archive `ripwire-0.6.9-macos-arm64.tar.gz`,
sha256 `38a4342340d2604b500e2063e460972951f0e4f634c77e4e67a0647ca4de5981`,
binary sha256 `830411ac2132bb3e40622c98de3162e9c12cece06e297577f3836464e1e067bb`).

## Sanity checks

- **Fixture**: on `internal/ripwireadapter/testdata/python/src` the pinned
  build scores 100% exact Ca/Ce with 0% lost/spurious, and 0.6.9 shows exactly
  its known gaps (the aliased edge lost, the `logging` collision spurious).
- **Corpus**: the ground truth sees pip's `src/` as 418 `.py` files in 54
  directories; the spike's 668 files / 77 packages are pip's whole repository,
  while its coupling comparison (~53 directories) used the `src/` crawl.
  rollup's `src/` holds 274 `.ts` files, as in the spike.
