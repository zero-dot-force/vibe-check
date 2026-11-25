# Recorded ripwire Outputs

Recorded outputs of the pinned ripwire binary against the fixture projects,
plus the `--arch` rules file used as a capture input. These recordings are
the golden inputs for the fake-ripwire test harness, the parser tests, and
the hand-computed `ModuleGraph` goldens (`recorded/golden-notes.md`). The
fixtures themselves are documented in [`FIXTURES.md`](FIXTURES.md).

History: first captured from ripwire 0.6.9 @ `bedc0cd` by
`openspec/changes/archive/2026-11-24-ripwire-adapter` (tasks 1.2–1.4); re-captured for the pin
move by `openspec/changes/archive/2026-11-25-bump-ripwire-pin` (tasks 1.1–1.4, 7.1–7.3); the
Python recordings (`python`, `python-mixed`, derived `python-zero-edge`)
re-captured for the `ring_a`/`ring_b` fixture by
`openspec/changes/archive/2026-11-24-wide-form-cycle-edges` (task 2.2), and again for the
`rel_a`/`rel_b` fixture by `openspec/changes/archive/2026-11-24-relative-import-edges`
(task 3.2) with the same 0.7.0 release binary.

## Layout

```text
recorded/
  arch-rules.txt               # --arch rules file (capture input, committed)
  golden-notes.md              # value-by-value ledger for the goldens
  python/
    version.txt                # `ripwire --version` output
    arch.txt                   # --arch metrics table (XML, root element <arch>)
    deps.xml                   # --deps file-to-file dependency graph (XML)
    cc.json                    # --export=cc.json per-file metrics (CodeCharta)
    golden.json                # hand-computed ModuleGraph golden
  typescript/
    version.txt, arch.txt, deps.xml, cc.json, golden.json
  python-mixed/                # ARCH-HONEST surface: python/ plus one Go file
    version.txt, arch.txt, deps.xml, cc.json
  python-zero-edge/            # synthetic: python/ with an edgeless deps.xml
    version.txt, arch.txt, deps.xml, cc.json
```

Recorded outputs deliberately live **outside** the fixture project
directories: ripwire indexes JSON as a config language, so a recorded
`cc.json` inside a crawl root would change every subsequent capture of
that root and break byte-stability.

## Binary provenance — the validated (version, commit) pair

The recordings were captured from the published ripwire 0.7.0 release
(the first release containing redhat-et/ripwire#533; upstream skipped
0.6.11), the pair `version.go` pins. `version.txt` is the binary's
`--version` output, verbatim.

| Field | Value |
|-------|-------|
| `--version` line | `ripwire 0.7.0 (Release, AppleClang 21.0.0.21000101, emit=std::print, act=on, rows=lines, built_from=5a3df049e)` |
| Release | <https://github.com/redhat-et/ripwire/releases/tag/v0.7.0> (tag commit `5a3df049e032dc537cad2b26eb8c48f3c4b9b670`; contains #533's merge `0b8f0bf8`) |
| Archive | `ripwire-0.7.0-macos-arm64.tar.gz`, sha256 `32e0fcd96ece495b7ab23274e5f14d9d275767ab498299c3c2cd49379e1fcff7` (matches the published `.sha256` asset) |
| Binary sha256 | `36b41c5eb6f213fd6a48f203fd29b0b1f1cb56d24d93bb4923ce12ea5a6de4d7` |
| Platform | `macos-arm64` (AppleClang 21) |

Phase A (`openspec/changes/archive/2026-11-25-bump-ripwire-pin` design D8) captured from a
source build of `main` @ `1edb6fd6a`. Against it, the 0.7.0 re-capture is
byte-identical for `typescript/` and `python-mixed/`; `python/` differs
only because 0.7.0 indexes a `build/` directory whose files git tracks (see
"Reading the recordings").

The binary was extracted and run **outside the repository** and must never be
committed or copied into the repo.

`wide-form-cycle-edges` re-captured `python/` and `python-mixed/` (and
re-derived `python-zero-edge/`) with this same 0.7.0 binary (binary sha256
as above) after adding the `ring_a`/`ring_b` fixture packages; `typescript/`
and every `version.txt` are unchanged. Against the pre-ring recordings,
`python/arch.txt` differs only in the two new `<m>` rows, `fixturepy`
`ca="2"`, the module counts, and `propagation_cost`.

`relative-import-edges` re-captured the same three directories with this
same 0.7.0 binary after adding the `rel_a`/`rel_b` fixture packages, twice
each with byte-identical results; `typescript/` and every `version.txt` are
unchanged. Against the pre-rel recordings, `python/arch.txt` differs only in
the two new `<m>` rows, `fixturepy` `ca="3"`, the module counts, and
`propagation_cost`.

## Capture commands (exact, as run)

All commands run from the repository root. `$RIPWIRE` is the binary from
the provenance table. Each subshell `cd`s into the crawl root and passes `.`
so the recorded `root=` attribute is `"."` and row paths are
root-relative, making the recordings independent of the checkout path.
`version.txt` is the binary's `--version` output.

### Python fixture (src layout — crawl from `src/`, never the project root)

```bash
(cd internal/ripwireadapter/testdata/python/src && $RIPWIRE . --arch=../../recorded/arch-rules.txt --no-cache) > internal/ripwireadapter/testdata/recorded/python/arch.txt

(cd internal/ripwireadapter/testdata/python/src && $RIPWIRE . --deps --limit=1000 --no-cache) > internal/ripwireadapter/testdata/recorded/python/deps.xml

(cd internal/ripwireadapter/testdata/python/src && $RIPWIRE . --export=cc.json --no-cache) > internal/ripwireadapter/testdata/recorded/python/cc.json
```

### TypeScript fixture (crawl from the directory holding `tsconfig.json`/`package.json`)

```bash
(cd internal/ripwireadapter/testdata/typescript && $RIPWIRE . --arch=../recorded/arch-rules.txt --no-cache) > internal/ripwireadapter/testdata/recorded/typescript/arch.txt

(cd internal/ripwireadapter/testdata/typescript && $RIPWIRE . --deps --limit=1000 --no-cache) > internal/ripwireadapter/testdata/recorded/typescript/deps.xml

(cd internal/ripwireadapter/testdata/typescript && $RIPWIRE . --export=cc.json --no-cache) > internal/ripwireadapter/testdata/recorded/typescript/cc.json
```

### Mixed-language variant (temporary tree, never committed)

```bash
MIXED=$(mktemp -d)
cp -R internal/ripwireadapter/testdata/python "$MIXED/tree"
printf 'package helper\n\nfunc Help() int { return 1 }\n' > "$MIXED/tree/src/fixturepy/cycle_a/helper.go"
RULES=$PWD/internal/ripwireadapter/testdata/recorded/arch-rules.txt
OUT=$PWD/internal/ripwireadapter/testdata/recorded/python-mixed

(cd "$MIXED/tree/src" && $RIPWIRE . --arch="$RULES" --no-cache) > "$OUT/arch.txt"
(cd "$MIXED/tree/src" && $RIPWIRE . --deps --limit=1000 --no-cache) > "$OUT/deps.xml"
(cd "$MIXED/tree/src" && $RIPWIRE . --export=cc.json --no-cache) > "$OUT/cc.json"
```

The temporary tree is not a git checkout, so its `cc.json` `churn` values
are `0` by construction.

### Zero-edge variant (derived, not captured)

`python-zero-edge/` copies `python/`'s `version.txt`, `arch.txt`, and
`cc.json` verbatim. Its `deps.xml` is `python/deps.xml` with the window
emptied: the root's `files=`, `shown=`, `total=`, and `next_offset=` set to
`0`; `<godfiles>` emptied to `total="0" shown="0"`; every `<f>` file row and
any `<cycles>` element removed; a trailing newline added. `<health>` and
`<pruned_packages>` (absent in the pinned release's `python/` capture) are
kept as captured. It drives the zero-coupling guard
test (an edgeless crawl over a non-empty corpus).

## Flag choices and why

- **`--limit=1000` (`--deps` only)** — the fixtures contain at most 22
  in-corpus files, so 1000 can never truncate; the recorded roots carry
  `capped="0" has_more="0"` plus the full paging attributes
  (`total= offset= limit=`), proving nothing was cut. In the pinned release
  the same limit also sizes the `<cycles>` listing (max(20, `--limit`)).
  `--limit`/`--offset` are honored **only** by the paginated verbs —
  `--arch` and `--export` refuse them with exit 1.
- **`--no-cache`** — forces a cold parse so no per-root cache state can
  influence a recording.
- **No `--exclude`** — the crawl root (the positional argument) already is
  the package root for every fixture. ripwire itself hard-prunes source
  directories named `build`/`dist`/`out`/`target`/`vendor`/`third_party`/
  `captures`.
- **`--arch` rules file** — `arch-rules.txt` defines five layers and no
  `deny` rules, so no violation can fire (`--arch` exits 2 on a violation;
  these captures expect exit 0). The rules-file path is not embedded in
  the output. Do not add `--baseline`/`--baseline-update` — those write a
  `.ripwire_arch_baseline` sidecar into the working directory.

## Byte-stability verification (method and results)

Method: every capture command was executed **twice**, each run writing to
its own scratch directory outside the repo, every invocation asserted to
exit 0, and the two runs compared with `diff -r`. The fixture trees were
checksummed before and after all captures and are byte-unchanged; no
`.ripwire*` sidecar files were created.

stderr is empty for every invocation **except** `python-mixed` `--arch`,
which prints one expected disclosure line (exit 0):
`ripwire arch: 1 file(s) in go: this build resolves almost none of their
imports to files, so an edge through one was mostly never judged:
violations= is a floor`. The adapter's runner ignores stderr on success.

The Python recordings were captured **after** the fixture files were
committed (`cc.json` records git `churn`, which is `0` for untracked files).

Results — every recording byte-identical across the two runs:

| Recording | Bytes |
|-----------|-------|
| `python/arch.txt` | 3177 |
| `python/deps.xml` | 2870 |
| `python/cc.json` | 4421 |
| `typescript/arch.txt` | 2172 |
| `typescript/deps.xml` | 2123 |
| `typescript/cc.json` | 1672 |
| `python-mixed/arch.txt` | 4544 |
| `python-mixed/deps.xml` | 4203 |
| `python-mixed/cc.json` | 4522 |
| `*/version.txt` | 111 |

## Reading the recordings (fixture-specific facts)

- **Row paths carry no `./` prefix** in the pinned release
  (`fixturepy/cycle_a`; 0.6.9 printed `./fixturepy/cycle_a`). The crawl
  root row is still `.`. The mapping accepts both forms.
- **Python `operations/build` depends on git.** ripwire 0.7.0 indexes a
  `build/` directory whose files git tracks and prunes it otherwise. The
  `python/` capture runs inside this repository's checkout, so
  `fixturepy/operations/build` is an isolated module (`ca="0"
  ce="0"`), `cc.json` lists its two files, and there is no
  `<pruned_packages>`; the adapter's pre-check drops a denylist-named
  directory ripwire indexed, so the graph carries no `denylist-pruned`
  warning. The `python-mixed/` capture runs on a temporary copy outside
  git: there `<pruned_packages count="1">` names the directory and the
  warning lists it once (union of that disclosure and the pre-check).
- **The three import-fix packages show at directory grain:** `wide`
  `ce="2"` (gap 1: `from fixturepy import cycle_b` reaches the `cycle_b`
  subpackage), `alias` `ce="1"` (gap 2), `collide` `ce="0"` (gap 4).
- **`--deps` records imports as written, not as resolved.** `wide/
  consumer.py` lists a single `<inc t="fixturepy"/>`; the resolved member
  edge to `cycle_b/__init__.py` shows only in that file's `afferent=` and
  in `cc.json` `fan_out=2`. Name-resolving `<inc t=>` text alone cannot
  see it; the adapter recovers it with its own member-import light parse,
  anchored to that `<inc t="fixturepy"/>` (`openspec/changes/
  wide-form-cycle-edges`). `ring_a/__init__.py` (`<inc t="fixturepy"/>`)
  is the same shape: its member edge to `ring_b` is the only edge closing
  the `ring_a`/`ring_b` cycle.
- **Relative imports are recorded as written too.** `rel_a/__init__.py`
  lists `<inc t="..rel_b.api"/>` and `rel_b/__init__.py` lists
  `<inc t=".."/>`; ripwire's `--arch` resolves both (and the `rel_a` member
  of `from .. import rel_a`), so `fixturepy/rel_b` reads `ce="2"`. The
  adapter resolves the targets against the importing package and recovers
  the member with its light parse (`openspec/changes/archive/2026-11-24-relative-import-edges`);
  together they are the only edges closing the `rel_a`/`rel_b` cycle.
- **Python `deps.xml` has no `<cycles>` element.** The fixture's import
  cycle resolves file-level as `cycle_a/__init__.py → cycle_b/api.py` and
  `cycle_b/__init__.py → cycle_a/api.py` (each edge one-way at file grain);
  the ring pair likewise (`ring_a/__init__.py → ring_b/__init__.py →
  ring_a/api.py`), and the rel pair (`rel_b/__init__.py →
  rel_a/__init__.py → rel_b/api.py`).
  The two-directory cycle exists at directory grain and is the adapter's
  own SCC derivation. The TypeScript `deps.xml` carries a `<cycles>`
  element with the size-2 cycle and `<health shape="tangled"
  cyclic_files="2">` (0.6.9 printed `vertical`; the adapter does not
  consume shape).
- **Python `arch.txt` reports `abstract="0"`** for every module —
  ripwire's Python abstractness is structurally zero (Python abstractness
  comes from vibe-check's own light parse). TypeScript `src` reports
  `types="5" abstract="4"` (interface share including the nested
  `Shapes.Point`).
- **`python-mixed` `fixturepy/cycle_a`** reads `ca="2" ce="1" types="0"
  abstract="0" A="0.00" zone="unknown" no_file_edges="go"` — no `I`, `D`,
  or `reachable` — and the root carries `no_file_edges="go"
  unmeasured_files="1" graph_partial="1"`.
- **`tsconfig.json` and `package.json` are indexed** (config languages):
  they appear in the TypeScript `cc.json` tree and in `deps.xml`
  `health files="5"` (3 dep-capable `.ts` files).

## Re-capture procedure

1. Obtain the binary named in the provenance table (download the
   release archive, verify its published sha256, extract outside the
   repository) and confirm it prints the expected `--version` line.
2. From the repository root, run every command in "Capture commands"
   twice into separate scratch directories, `diff -r` the pairs, and
   rebuild `python-zero-edge/deps.xml` as described.
3. Recordings are only valid if both runs are byte-identical, every
   invocation exits 0, and stderr is empty apart from the documented
   `python-mixed` line. A nonzero `--arch` exit means the committed rules
   file drifted (a violation fired).
4. Commit any fixture changes **before** capturing, so `cc.json` `churn`
   matches the committed tree.
