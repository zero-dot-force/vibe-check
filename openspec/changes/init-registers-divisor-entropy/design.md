# Design

## Context

See `proposal.md` — Why for the motivation. The relevant current state:
`internal/scaffold/scaffold.go` exposes `Run(Options) (*Result, error)` which
delegates to an unexported `run(agentAssets, commandAssets fs.FS, opts)`
seam, validates the target via `metrics.ValidateProjectPath`, resolves the
root with `filepath.EvalSymlinks`, deploys two asset categories
(`agents/`, `commands/`) with symlink-safe guards (`ensureDir`
Lstat-walk, `verifyContained`, `regularFileExists`), and returns
category-prefixed result names sorted stably. `cmd/vibe-check/init.go`
renders `Result` as a human summary or a `--json` payload
(`written`/`skipped`/`forced`, optional `provenance`) via the AP-001/AP-002
Options/Result/RunX pattern. The two UF policy files are hand-curated YAML
with comments; this repo's own `.uf/reviewer-capabilities.yaml` and
`.uf/review-matrix.yaml` (fixed manually in 25328d5 and 81831e2) are the
canonical format reference. `go.mod` has no YAML dependency.

## Goals / Non-Goals

**Goals:**

- Idempotent, format-preserving, comment-preserving registration of
  `divisor-entropy` in both UF policy files during `vibe-check init`.
- No new dependencies; no YAML parsing or round-tripping.
- Symlink/path safety consistent with the existing scaffold guards.
- Additive output reporting (human and `--json`) with unchanged existing
  key semantics.
- Testable without root or filesystem manipulation, via the existing
  `WriteFile` seam and `t.TempDir()` fixtures.

**Non-Goals:**

- Validating or normalizing the policy files' YAML beyond locating the
  `reviewers:` list and the four `advisor` mode lists.
- Creating `.uf/` or either policy file when absent.
- Registering non-review agents (`vibe-check-reporter`,
  `mx-f-architecture-trend`) or editing any other entry or section
  (`always:`, `profiles`, `runs:`, `risk_augmentation:`, `limits:`).
- Any `--force` interaction with registration (force governs asset files
  only).
- Atomic temp-file-and-rename writes (see D8).

## Decisions

### D1: Registration is a step in `scaffold.run()`, after asset deployment

**Decision**: Add the registration pass in `internal/scaffold/` (new file
`register.go`) and invoke it from `run()` after the categories loop and
before the result slices are sorted. It operates on the same validated,
symlink-resolved `root` as asset deployment.

**Rationale**: The issue places the step "after asset deployment"; living
inside `run()` keeps `cmd/vibe-check/init.go` thin (AP-002: no business
logic in the command layer) and gives tests the same entry point
(`Run`/`run`) they already use.

**Alternative considered**: A separate `scaffold.Register(...)` API called
by `RunInit`. Rejected — two public entry points would let registration and
deployment drift apart (e.g. different validation), and no caller needs
registration alone.

### D2: Canonical entries as Go string constants in a one-entry registration table

**Decision**: Declare the canonical content as package-level constants in
`register.go`: the four-line capabilities entry
(`- agent: divisor-entropy` / `capability: review` / `scopes:` /
`- standard`, plus a scope item `standard`) and the matrix list item
`- divisor-entropy`. A small `registrations` slice (exactly one entry:
`divisor-entropy`) drives the pass.

**Rationale**: The entries are four lines and one line; constants are
diffable, greppable, and compile-time-checked. The one-entry table models
the real concept "review personas this tool deploys" so a future review
agent is a localized addition, without speculative generality.

**Alternative considered**: Embedded template files under
`internal/scaffold/assets/`. Rejected — `//go:embed` assets are deployed
verbatim, and these fragments must be spliced mid-file at derived
indentation, not copied; embedding would imply a deployment relationship
that does not exist.

### D3: Line-based insertion with scope-confined scanning, no YAML parser

**Decision**: Read the file, split into lines (keeping terminators), and:

- **Capabilities**: locate the top-level `reviewers:` key (a line matching
  the top-level-key pattern below whose key is `reviewers`; a trailing
  comment on the key line is tolerated, matching the general top-level
  matcher). Within that section (until the next top-level key), entry
  blocks start at lines whose trimmed form begins with `- agent:`.
  Presence check: any entry in the section whose agent name (extracted per
  the name-extraction rule below) is `divisor-entropy`, regardless of the
  entry's capability or scopes (value-blind skip per the issue). Insertion
  position: before the first entry whose agent name sorts after
  `divisor-entropy` among entries whose `capability:` is `review`; if
  none, after the last `review` entry's block; if there are no `review`
  entries, immediately after the `reviewers:` line.
- **Matrix**: locate the top-level `advisor:` key, then each mode key
  (`code:`, `spec:`, `triage:`, `feedback:`) at one indent level below it.
  Each mode list is the run of `- <name>` item lines below the mode key.
  Presence check per list: an item whose extracted name is
  `divisor-entropy`. Insertion: before the first item whose name sorts
  after `divisor-entropy`, else after the last item.
- **Name extraction**: agent and item names are compared after extraction,
  never by raw line equality — trim surrounding whitespace, strip a
  trailing `# …` comment, and remove surrounding single or double quotes
  before comparing. A `- divisor-entropy # hand-added` item and a
  `- "divisor-entropy"` scalar therefore count as present, preventing
  duplicate insertion into hand-curated files.

The inserted capabilities block is rendered at the `- ` marker indentation
observed on neighboring reviewer entries (fallback: 2 spaces, the UF
scaffold default); its `capability:`/`scopes:` lines sit at marker+2 and
the scope item at marker+4. The matrix item is rendered at the indentation
of neighboring items in that list (fallback: 4 spaces). Comments and blank
lines are skipped over structurally, never rewritten. Top-level keys are
lines matching `^[A-Za-z_][A-Za-z0-9_]*:` at zero indentation, matched
with `strings` functions.

The implementation adds no imports to `internal/scaffold`: it uses only
packages the package already imports (`os.ReadFile` + `strings.Split`
preserve the line terminators; top-level-key and name matching use
`strings` functions, not `regexp`; line scanning does not use `bufio`).
This keeps the package's efferent coupling unchanged so the entropy
dogfood gate (task 4.2) observes a zero delta.

**Rationale**: No YAML parser exists in `go.mod` and round-tripping would
destroy the hand-curated comments the issue explicitly protects. Scoping
detection to the `reviewers:` section and the `advisor` mode lists is load
bearing: `- divisor-entropy` could legitimately appear elsewhere (for
example a hand-added `risk_augmentation.agents` entry) and a whole-file
search would falsely report the entry present.

**Alternative considered**: A YAML parser with comment-preserving
round-trip (for example `yaml.Node`). Rejected — new dependency for a
four-line splice, and comment fidelity is not guaranteed by any mainstream
Go YAML library.

### D4: Per-list granularity for the matrix, per-file for the capabilities manifest

**Decision**: The capabilities manifest is one registration unit (one
entry). The matrix is four units internally — each mode list is checked and
spliced independently — reported as one outcome name: `advisors/divisor-entropy`
is *written* when at least one list gained the item, *skipped* when the file
is present with at least one mode list and every existing mode list already
contains it. A file that is absent or lacks the `advisor:` section or all
mode lists is notice-only (D6) — never reported as written or skipped.

**Rationale**: A partially registered matrix (entry in `code` but not
`triage`) is exactly the fail-closed gap this change fixes; per-file
skip-on-any-occurrence would leave it broken. Reporting stays at file
granularity because consumers act on files, and four per-mode names would
be noise.

### D5: Reporting via two category-prefixed names plus a `Notices` field

**Decision**: Report `reviewers/divisor-entropy` (capabilities manifest)
and `advisors/divisor-entropy` (review matrix) through the existing
`Result.Written`/`Result.Skipped` slices — the names follow the existing
`category/name` convention (`agents/…`, `commands/…`). Add a
`Notices []string` field to `scaffold.Result` and `InitResult`, an
additive `notices` array to the `--json` payload (after `forced`,
normalized to `[]` never `null`), and a `Notices:` section in the human
summary printed only when non-empty. A file-absent or
unexpected-shape skip emits a notice naming the file; it does NOT appear in
`written` or `skipped`.

**Rationale**: `skipped` currently promises "file existed and was left
untouched" — scripting consumers would be misled by absent files landing
there. The issue's example name (`reviewers/divisor-entropy`) maps the
category prefix to the section being edited; `advisors/` is the matrix's
section. Printing `Notices:` only when non-empty keeps every run on a
uf-scaffolded target free of a permanently empty section.

**Alternative considered**: Reporting absent files in `skipped` with a
notice. Rejected — overloads `skipped` semantics for both human and JSON
consumers.

### D6: Absent files and unexpected shapes are notice-skips, never creations

**Decision**: Registration never creates `.uf/`, either policy file, a
`reviewers:` list, or an `advisor` mode list. `.uf/` absent → one notice
per policy file (or per-file when only one is missing); file present but
`reviewers:`/`advisor:`/a mode list absent → notice naming the file and
the unexpected shape; both are successes (exit 0).

**Rationale**: Non-uf-scaffolded projects don't consume these files, and
fabricating structure the file's curator didn't write would violate the
never-rewrite principle from the opposite direction. Fail-soft here is
safe: the dispatcher's fail-closed behavior surfaces the gap loudly at
review time, and the notice tells the user exactly which file to look at.

### D7: Path safety mirrors the existing guards; refusals are errors

**Decision**: Before touching a policy file, Lstat the `.uf` directory
(requiring an existing real directory — no creation, unlike `ensureDir`;
a symlink or non-directory component is refused) and Lstat the file
(requiring a regular non-symlink file: no `ModeSymlink` and
`Mode().IsRegular()` — stricter than `regularFileExists`, which rejects
only symlinks and would pass a FIFO or device node that `os.WriteFile`
would block on indefinitely). Any symlink, non-directory `.uf` component,
or non-regular policy file → wrapped error → exit code 2, target
untouched. Because both checks pass on real directories/files inside the
already-resolved root, containment holds absent a concurrent swap of the
checked path between the Lstat and the write — the same TOCTOU window the
existing asset path has (`os.WriteFile` has no `O_NOFOLLOW`).

**Rationale**: Rewriting through a symlinked `.uf` component or policy file
would modify a file outside the validated root — the same escape the
existing guards exist to prevent. Consistency demands the same refusal
semantics (error, not notice).

### D8: Rewrite via the existing `WriteFile` seam, preserving file mode

**Decision**: Registration rewrites a policy file with the same injectable
`Options.WriteFile` seam used for assets (`os.WriteFile` by default),
passing the file's existing mode (captured from the Lstat in D7) so the
mode is preserved. The new content is computed fully in memory and written
with a single `WriteFile` call. Splice edge case: if the insertion point is
after the file's final line and that line lacks a trailing newline, the
missing terminator is appended (otherwise two entries would share a line);
this is the only byte change permitted beyond the inserted lines, and it
cannot occur in canonical files (which end with a trailing comment). A
write failure can leave the target truncated or partial; the returned
error (and an accompanying notice when output is available) states that
the policy file may be partially written and can be restored from version
control.

**Rationale**: Reusing the seam keeps the I/O-failure exit-2 branch
testable unconditionally and the write path uniform. Mode preservation is
least-surprise for a hand-curated file. Temp-file-and-rename would buy
crash atomicity but bypass the seam (losing deterministic failure
injection) and diverge from the accepted asset-write behavior; the
single-write window is small and consumers' policy files are normally
committed to git.

**Alternative considered**: Write to `file.tmp` + `os.Rename`. Rejected
for the reasons above; revisit only if a corruption incident occurs.

### D9: Registration errors propagate as exit 2 (partial deployment is precedented)

**Decision**: I/O failures and symlink refusals during registration return
from `run()` as wrapped errors, so `RunInit` reports exit 2 with nothing on
stdout — even though asset deployment earlier in the same run already
succeeded.

**Rationale**: A security refusal or I/O failure must be loud, not demoted
to a notice. Partial-state-plus-error is already the behavior when a later
asset category fails after an earlier one succeeded, so no new semantics
are introduced. Re-running `init` is idempotent and completes the remainder.

### D10: Presence detection is value-blind and scope-confined

**Decision**: An existing `divisor-entropy` entry suppresses insertion
regardless of its capability, scopes, or values — including a hand-added
`capability: content` entry — and detection only looks inside the
`reviewers:` section (capabilities) or the specific `advisor` mode list
(matrix), never the whole file. When a present capabilities entry lacks
`capability: review` or has empty scopes, the skip is accompanied by an
advisory notice (for example `divisor-entropy present with capability:
content; review dispatch may fail closed`) so the dispatch gap is visible
at init time rather than at review-dispatch time; the entry itself is
never rewritten.

**Rationale**: The issue's rule "Entry already present → skip, regardless
of scopes/values. Existing entries are never rewritten" makes the user's
curation sovereign; silently "fixing" a hand-edited entry would be a
rewrite by another name. Scope confinement prevents false skips from
same-named items in unrelated lists (see D3).

## Risks / Trade-offs

- [Line-based scanner mis-reads an unusual hand-curated shape] → Detection
  is confined to the two structures that matter; anything unlocatable
  (missing section, missing mode list) is a notice-skip, never a guess; the
  file is never rewritten when its shape is not understood. Fixtures in
  tests are the canonical UF-scaffold shapes.
- [Non-atomic in-place rewrite could truncate a policy file on a crash
  mid-write] → Accepted (D8): single `WriteFile`, tiny window, policy files
  are normally git-tracked; assets and this repo's copies remain
  recoverable sources for regeneration.
- [Registration failure exits 2 after assets were already deployed] →
  Accepted (D9): re-running `init` is idempotent and finishes the job;
  error text names the failed path and operation.
- [Advisors reported as one name while four lists are spliced] → Per-list
  outcomes are internally exact (D4); the single reported name reflects
  file-level action, which is what consumers script against. A partial
  splice is still a correct file (each list is independently valid).
- [Future review agents need code changes to register] → Deliberate
  (D2): one-entry table; adding an agent is a localized, reviewable change
  rather than data-driven magic.

## Migration Plan

Pure additive behavior on an existing command; no flags. The `--json`
payload gains an additive `notices` array — an output-schema change that
triggers the constitution's Website Documentation Sync rule (tasks 5.4 and
5.5 file the website and repo documentation issues). Consumers who already
hand-registered (as this repo did in
25328d5 and 81831e2) see `reviewers/divisor-entropy` and
`advisors/divisor-entropy` reported as skipped on their next `init` run,
with byte-identical policy files. Rollback is reverting to the previous
binary; no data migration exists.

## Test Strategy

**Coverage target**: ≥80% statement coverage for the new registration code
in `internal/scaffold/`, verified via `go test -race -count=1 -cover
./internal/scaffold/` in the CI parity tasks.

**Classification**: all new tests are unit tests isolated in `t.TempDir()`
with hand-written fixture policy files derived from this repo's canonical
`.uf/` files (comments, blank lines, and trailing
`# scaffolded by uf vdev` included). No integration-only behaviors.

**Cases** (one table or focused test per bullet, per the issue's list):

- Insert-when-missing for each file: capabilities entry appears between
  `divisor-curator` and `divisor-guard`; `- divisor-entropy` appears in
  sorted position in each of the four advisor mode lists.
- Skip-when-present for each file: byte-identical file after re-run,
  including a variant whose existing entry has non-canonical scopes, and a
  run with `Force: true`. Name extraction is exercised: a
  `- divisor-entropy # hand-added` item and a quoted
  `- "divisor-entropy"` item each count as present (no duplicate
  insertion). A present capabilities entry with `capability: content`
  emits the advisory notice while remaining byte-identical.
- Skip-when-file-absent: no `.uf/` at all (two notices, nothing created,
  exit 0), and one-file-present/one-absent (matrix registered + one
  notice).
- Sorted insertion per mode list, including the per-list-granularity case
  (one list pre-registered, only the other three spliced).
- Comment/format preservation: assert the rewritten file equals the
  fixture with exactly the canonical lines inserted (byte-exact expected
  output), preserving header/inline comments, blank lines, and the
  trailing scaffold comment.
- Unterminated final line: a fixture whose insertion point follows a final
  line lacking a newline gains the terminator, and the inserted block
  starts on its own line (D8).
- File-mode preservation: fixture written `0o600` stays `0o600`.
- Symlink safety: symlinked policy file and symlinked `.uf` directory each
  produce an error (exit-2 path via `RunInit` or `Run`) with the target
  untouched — consistent with the existing `ensureDir`/`verifyContained`
  guard tests. A FIFO (or other non-regular, non-symlink file) at a policy
  path is refused promptly by the Lstat `IsRegular()` check, never
  reaching a blocking write.
- Unexpected shape: a capabilities file without `reviewers:` and a matrix
  with no mode lists at all each produce notices and no edits; a matrix
  missing one mode list still splices the remaining lists (per-list
  granularity, D4) with a notice identifying the missing list.
- Reporting: result slices carry `reviewers/divisor-entropy` and
  `advisors/divisor-entropy` in written/skipped as appropriate;
  `--json` payload contains the additive `notices` array, unchanged
  `written`/`skipped`/`forced` keys; human summary prints `Notices:` only
  when non-empty.
- I/O failure: a `WriteFile` seam that fails selectively on the policy
  path (after assets are pre-deployed so asset writes don't trip it first)
  yields a wrapped error and the exit-2 branch.
- Idempotency: full second `Run` reports both registration names as
  skipped with byte-identical files.

**Dogfood**: run `go run ./cmd/vibe-check init .` on this repository —
both entries already exist (25328d5, 81831e2), so the expected outcome is
both names reported skipped and byte-identical `.uf/` files; this also
exercises the notice-free path on a uf-scaffolded target.
