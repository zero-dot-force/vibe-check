# Recorded ripwire Outputs — `ripwire-adapter` task 1.2

Recorded outputs of the pinned ripwire binary against the two fixture
projects, plus the `--arch` rules file used as a capture input. These
recordings are the golden inputs for the fake-ripwire test harness
(task 1.3), the parser tests (tasks 4.1–4.2), and the hand-computed
`ModuleGraph` goldens (task 1.4). The fixtures themselves are documented
in [`FIXTURES.md`](FIXTURES.md) (task 1.1).

## Layout

```text
recorded/
  arch-rules.txt               # --arch rules file (capture input, committed)
  python/
    version.txt                # `ripwire --version` output
    arch.txt                   # --arch metrics table (XML, root element <arch>)
    deps.xml                   # --deps file-to-file dependency graph (XML)
    cc.json                    # --export=cc.json per-file metrics (CodeCharta)
  typescript/
    version.txt
    arch.txt
    deps.xml
    cc.json
```

Recorded outputs deliberately live **outside** the fixture project
directories: ripwire indexes JSON as a config language, so a recorded
`cc.json` inside a crawl root would change every subsequent capture of
that root and break byte-stability.

## Binary provenance — the validated (version, commit) pair

| Field | Value |
|-------|-------|
| Version | `0.6.9` |
| Commit (`built_from=`) | `bedc0cdcf` (the change pins `bedc0cd`, a prefix of it) |
| Platform | `macos-arm64` |
| Release archive | `ripwire-0.6.9-macos-arm64.tar.gz` |
| Archive sha256 | `38a4342340d2604b500e2063e460972951f0e4f634c77e4e67a0647ca4de5981` (verified against the published release digest) |
| Extracted binary sha256 | `830411ac2132bb3e40622c98de3162e9c12cece06e297577f3836464e1e067bb` (bit-identical to the archive contents; this exact executable produced every recording) |

`version.txt` content, verbatim (91 bytes, trailing newline):

```text
ripwire 0.6.9 (Release, AppleClang 21.0.0.21000101, emit=std::print, built_from=bedc0cdcf)
```

The binary and its bundled 225KB reference README were **outside the
repository** at capture time (a temp directory) and must never be
committed or copied into the repo. For a re-capture, bind `RIPWIRE` to
a binary extracted from the release archive above and verify both
digests first.

## Capture commands (exact, as run)

All commands are run from the repository root. `$RIPWIRE` is the
verified binary from the provenance table above. Each subshell `cd`s
into the crawl root and passes `.` so the recorded `root=` attribute is
`"."` (and `--arch` module paths are `./`-relative), making the
recordings independent of the checkout path.

### Python fixture (src layout — crawl from `src/`, never the project root)

```bash
$RIPWIRE --version > internal/ripwireadapter/testdata/recorded/python/version.txt

(cd internal/ripwireadapter/testdata/python/src && $RIPWIRE . --arch=../../recorded/arch-rules.txt --no-cache) > internal/ripwireadapter/testdata/recorded/python/arch.txt

(cd internal/ripwireadapter/testdata/python/src && $RIPWIRE . --deps --limit=1000 --no-cache) > internal/ripwireadapter/testdata/recorded/python/deps.xml

(cd internal/ripwireadapter/testdata/python/src && $RIPWIRE . --export=cc.json --no-cache) > internal/ripwireadapter/testdata/recorded/python/cc.json
```

### TypeScript fixture (crawl from the directory holding `tsconfig.json`/`package.json`)

```bash
$RIPWIRE --version > internal/ripwireadapter/testdata/recorded/typescript/version.txt

(cd internal/ripwireadapter/testdata/typescript && $RIPWIRE . --arch=../recorded/arch-rules.txt --no-cache) > internal/ripwireadapter/testdata/recorded/typescript/arch.txt

(cd internal/ripwireadapter/testdata/typescript && $RIPWIRE . --deps --limit=1000 --no-cache) > internal/ripwireadapter/testdata/recorded/typescript/deps.xml

(cd internal/ripwireadapter/testdata/typescript && $RIPWIRE . --export=cc.json --no-cache) > internal/ripwireadapter/testdata/recorded/typescript/cc.json
```

`--version` takes no directory argument; its output is per-binary, so
the two `version.txt` files are identical by construction (each was
still captured and verified twice like every other recording).

## Flag choices and why

- **`--limit=1000` (`--deps` only)** — the fixtures contain at most 9
  in-corpus files, so 1000 can never truncate; the recorded roots carry
  `capped="0" has_more="0"` plus the full paging attributes
  (`total= offset= limit=`), proving nothing was cut. Note: `--limit`/
  `--offset` are honored **only** by the paginated verbs — `--arch` and
  `--export` refuse them with exit 1, so `--limit` appears on the
  `--deps` invocation only.
- **`--no-cache`** — forces a cold parse so no per-root TMPDIR cache
  state can influence a recording. A warm-cache run was verified
  byte-identical to the cold run for all three directory-based verbs, so
  this choice removes a variable without changing bytes.
- **No `--exclude`** — the crawl root (the positional argument) already
  is the package root for both fixtures, which is the scoping `--exclude`
  exists to provide in the adapter when the user points at a larger
  tree; nothing inside either crawl root needs pruning. ripwire itself
  hard-prunes source directories named `build`/`dist`/`out`/`target`/
  `vendor`/`third_party`/`captures`.
- **`--arch` rules file** — `arch-rules.txt` defines five layers and no
  `deny` rules, so no violation can fire (`--arch` exits 2 on a
  violation; these captures expect exit 0). Grammar per the bundled
  release's `skills/ripwire-layers/SKILL.md`: `layer NAME = substr`,
  `deny FROM -> TO`, `allow FROM -> TO`, `#` comments; layer substrings
  match root-relative paths (`--help=--arch`). The rules-file path is
  not embedded in the output. Do not add `--baseline`/
  `--baseline-update` — those write a `.ripwire_arch_baseline` sidecar
  into the working directory.

## Byte-stability verification (method and results)

Method: every capture command above was executed **twice**, each run
writing to its own file (run 1 → the recorded path, run 2 → a scratch
file outside the repo), with every invocation asserted to exit 0 with
empty stderr; the two files were then compared with `cmp`:

```bash
$RIPWIRE ... > recorded/python/deps.xml        # run 1
$RIPWIRE ... > /tmp/deps.xml.run2              # run 2 (scratch)
cmp recorded/python/deps.xml /tmp/deps.xml.run2 # must report no difference
```

Results — all eight recordings byte-identical across the two runs:

| Recording | Bytes | Two-run `cmp` |
|-----------|-------|---------------|
| `python/version.txt` | 91 | identical |
| `python/arch.txt` | 1456 | identical |
| `python/deps.xml` | 1629 | identical |
| `python/cc.json` | 1997 | identical |
| `typescript/version.txt` | 91 | identical |
| `typescript/arch.txt` | 1415 | identical |
| `typescript/deps.xml` | 1855 | identical |
| `typescript/cc.json` | 1672 | identical |

The recorded file is the verified run-1 output. The fixture trees were
checksummed before and after all captures and are byte-unchanged; no
`.ripwire*` sidecar files were created.

## Reading the recordings (fixture-specific facts)

- **Python `operations/build` is absent everywhere.** ripwire hard-prunes
  directories named `build` (denylist above): `deps.xml` reports
  `health files="7"` against 9 fixture `.py` files, and `arch.txt`/
  `cc.json` likewise omit `operations/build/**`. This is intended — it
  is the concrete target for the adapter's `denylist-pruned` warning
  (task 3.2).
- **Python `deps.xml` has no `<cycles>` element.** The fixture's import
  cycle resolves file-level as `cycle_a/__init__.py → cycle_b/api.py`
  and `cycle_b/__init__.py → cycle_a/api.py` (each edge one-way at file
  grain). The two-directory cycle exists at directory grain and is the
  adapter's own SCC derivation to make (task 4.4). The TypeScript
  `deps.xml` does carry a `<cycles>` element with the size-2 cycle.
- **Python `arch.txt` reports `abstract="0"`** for every module —
  ripwire's Python abstractness is structurally zero (design D3/D8;
  Python abstractness comes from vibe-check's own light parse).
  TypeScript `./src` reports `types="5" abstract="4"` (interface share
  including the nested `Shapes.Point`).
- **`tsconfig.json` and `package.json` are indexed** (config languages):
  they appear in the TypeScript `cc.json` tree and in `deps.xml`
  `health files="5"` (3 dep-capable `.ts` files).
- **`churn` is `0` in both `cc.json` files** because the fixture files
  were untracked in git at capture time (no history). If the fixtures
  are later committed, a fresh capture's `churn` attributes may become
  non-zero — re-record then, and compare non-churn attributes when
  corroborating against these goldens.

## Re-capture procedure

1. Obtain `ripwire-0.6.9-macos-arm64.tar.gz`, verify its sha256 is
   `38a4342340d2604b500e2063e460972951f0e4f634c77e4e67a0647ca4de5981`,
   extract, and confirm the binary hashes to
   `830411ac2132bb3e40622c98de3162e9c12cece06e297577f3836464e1e067bb`
   and prints the version line above.
2. From the repository root, run the eight commands in
   "Capture commands" twice each into separate files and `cmp` the
   pairs, exactly as described above.
3. Recordings are only valid if both runs are byte-identical, every
   invocation exits 0, and stderr is empty. A nonzero `--arch` exit
   means the committed rules file drifted (a violation fired).
