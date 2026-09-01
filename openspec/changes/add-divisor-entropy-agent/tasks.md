## 1. Scaffold Package Skeleton

- [x] 1.1 Create the `internal/scaffold/` package directory.
- [x] 1.2 Add `internal/scaffold/doc.go` with a package-level GoDoc comment
  describing the scaffold package's role (embedding and deploying agent assets).
- [x] 1.3 Add `internal/scaffold/embed.go` with `import "embed"`,
  `//go:embed assets/agents/*.md`, and `var assetsFS embed.FS`, mirroring the
  `//go:embed` pattern in `metrics/schema.go`.

## 2. Structural Delta & Verdict Engine (pure Go)

- [x] 2.1 Add `metrics/delta.go` defining `Delta` (per-module metric deltas:
  ΔCa, ΔCe, ΔInstability, ΔAbstractness, ΔDistance, ΔLCOM, matched by
  `Module.Path`) and `GraphDelta` (per-module deltas plus `Added []string`,
  `Removed []string` — report-only; packages present in only one graph, MUST
  NOT produce a spurious metric delta or trigger any gate — `NewCycles`,
  `ResolvedCycles`, a typed `EntropyDirection` (defined as
  `type EntropyDirection string` with named constants `improving`/`stable`/
  `degrading`, mirroring `metrics.Status`), and a boolean
  unreliable/partial-measurement flag), with GoDoc documenting ranges/units on
  every exported symbol. Satisfies `diff-command` "Structural Delta
  Computation".
- [x] 2.2 Implement `ComputeDelta(base, pr *metrics.ModuleGraph) GraphDelta`
  as a pure, deterministic function (no I/O, no globals): match modules by
  `Module.Path`, compute per-module deltas (ΔCa, ΔCe, ΔInstability,
  ΔAbstractness, ΔDistance, ΔLCOM), classify packages present only in one
  graph into `Added`/`Removed` (report-only — MUST NOT produce a spurious
  metric delta or contribute to any gate evaluation), classify cycles as new
  (present in PR, absent in base) vs. pre-existing (present in both — MUST NOT
  be flagged) vs. resolved, set `EntropyDirection`, set the unreliable flag
  when either input graph has `Status != "complete"` or carries load-error
  `Warnings`, and handle nil/empty inputs without panicking: if either input
  graph is nil, return a `GraphDelta` with the unreliable/partial-measurement
  flag set (so `DecideVerdict` yields COMMENT, never a false APPROVE), keeping
  the error-less signature; two empty-but-non-nil graphs compare equal. Results
  MUST be emitted in a stable,
  sorted order. Satisfies `diff-command` "Structural Delta Computation", "New
  Circular Dependency Detection", and "Partial-Build Degraded Verdict".
- [x] 2.3 Add `metrics/verdict.go` defining `Verdict` (`APPROVE` / `COMMENT` /
  `REQUEST_CHANGES`), `VerdictThresholds` (with documented default gate values:
  ΔI ≥ 0.15, ΔD ≥ 0.20, ΔLCOM ≥ 2), and
  `DecideVerdict(d GraphDelta, t VerdictThresholds) (Verdict, []string)`
  returning the verdict plus a machine-readable list of triggering reasons.
  When the delta is flagged unreliable (partial build), `DecideVerdict` MUST
  force `COMMENT` (never APPROVE) and emit a machine-readable partial-build
  reason. Delta comparisons round each float to 4 decimal places using
  round-half-away-from-zero (`math.Round(x*1e4)/1e4`); document this rounding
  rule in the `DecideVerdict` GoDoc. The default thresholds are protected
  quality gates — document them as such. Satisfies `diff-command`
  "Deterministic Verdict" and "Partial-Build Degraded Verdict".
- [x] 2.4 [P] Add `metrics/delta_test.go` and `metrics/verdict_test.go`:
  table-driven tests over schema-valid `ModuleGraph` fixtures (each validated
  with `metrics.Validate`) exercising every gate boundary — ΔI `0.14` vs
  `0.15`, ΔD `0.19` vs `0.20`, ΔLCOM `1` vs `2` (guarding `≥` vs `>`
  off-by-one) — plus a new-cycle positive case, a pre-existing-cycle negative
  case, the COMMENT materiality band, and improve/stable → APPROVE. Include a
  float-rounding boundary fixture: a delta of `0.14999` (rounds to `0.1500` →
  REQUEST_CHANGES at the 0.15 threshold) and `0.14994` (rounds to `0.1499` →
  COMMENT), plus a ΔDistance rounding fixture (`0.19999` → `0.2000` →
  REQUEST_CHANGES), verifying the 4-decimal-place round-half-away-from-zero
  rule. Include a cycle rotation-invariance fixture (base cycle `[a,b,c]` vs PR
  `[b,c,a]` → classified as NOT a new cycle). Include a partial-build case (an
  input with `Status:"partial"` and/or a load-error `Warnings` entry → verdict
  COMMENT, unreliable flag `true`, added/removed signal suppressed) and a
  both-complete negative case (full verdict path). Include nil/empty
  contract cases: nil base → `GraphDelta` with the unreliable flag set →
  COMMENT (no panic, never APPROVE); nil pr → same; empty-vs-empty (both
  non-nil) → APPROVE; zero-value `GraphDelta{}` → APPROVE. Assert
  entropy direction (degrading / improving / stable) matches the trigger state.
  Assert determinism (same inputs always yield the same verdict and ordering).
  Follow AGENTS.md Testing Conventions (`TestFunc_Scenario` naming, `got %v,
  want %v` assertions). Target ≥ 80% coverage for the delta/verdict code.

- [x] 2.5 Wire `GOTOOLCHAIN=local` into the `vibe-check analyze` subprocess
  sanitized environment as a forced value (not a host passthrough): add it via
  a testable env-construction seam (e.g. a `packageEnv() []string` helper that
  appends `GOTOOLCHAIN=local` after `SanitizeEnvironment`) in the goadapter
  subprocess path so the analysis subprocess always runs with
  `GOTOOLCHAIN=local` regardless of how `analyze` is invoked. Add acceptance
  tests enumerating BOTH scenarios: (a) the constructed env contains
  `GOTOOLCHAIN=local`; (b) with `t.Setenv("GOTOOLCHAIN","auto")` the
  constructed env still contains `GOTOOLCHAIN=local` and NO `GOTOOLCHAIN=auto`
  (exec last-wins), noting `t.Setenv` precludes `t.Parallel()` on that test.
  Satisfies `analyze-command` "Hermetic Toolchain".

- [x] 2.6 Add an `--output <file>` (short `-o`) flag to `vibe-check analyze`
  (`cmd/vibe-check/analyze.go`) that writes the ModuleGraph JSON to the given
  file path; stdout remains the default sink when the flag is absent. Exit `2`
  with a stderr diagnostic (and no partial graph on stdout) if the output file
  cannot be written. Add table/CLI tests via `bytes.Buffer` and `t.TempDir()`
  asserting file output, the stdout default, and the unwritable-file exit-2
  branch. Satisfies `analyze-command` "Output File Selection".

## 3. `vibe-check diff` CLI Command

- [x] 3.1 Add `cmd/vibe-check/diff.go` with `DiffOptions{ Stdout, Stderr
  io.Writer; BasePath, PRPath string; Thresholds metrics.VerdictThresholds;
  JSON bool }` and `DiffResult{ Delta metrics.GraphDelta; Verdict
  metrics.Verdict; Reasons []string; ExitCode int }`, following the
  AP-002/AP-003 testable pattern used by `analyze.go`.
- [x] 3.2 Implement `RunDiff(ctx context.Context, opts DiffOptions)
  (*DiffResult, error)`: read and `metrics.Validate` both `ModuleGraph` JSON
  inputs, call `metrics.ComputeDelta` then `metrics.DecideVerdict`, and emit a
  human-readable delta table (default) or a JSON payload (`--json`) containing
  the per-module delta table, `newCycles`, `resolvedCycles`, `verdict`,
  `reasons`, entropy direction, and the unreliable-measurement flag. When the
  measurement is unreliable, annotate the affected packages and suppress the
  added/removed-driven signal. Output MUST be deterministic. Satisfies
  `diff-command` "Machine-Readable and Human-Readable Output" and
  "Partial-Build Degraded Verdict".
- [x] 3.3 Map exit codes: `0` when both inputs are valid and a verdict is
  computed (the verdict travels in the payload, not the exit code — a
  REQUEST_CHANGES verdict still exits 0); `2` when either input is
  missing/unreadable OR when either input is readable but fails schema
  validation (both cases: write a diagnostic to stderr and emit NO verdict
  payload). Satisfies `diff-command` "Input Validation and Exit Codes".
- [x] 3.4 Add `diffCmd() *cobra.Command` (args `<base.json> <pr.json>`, flags
  `--json` and threshold-override flags `--max-instability-delta`,
  `--max-distance-delta`, `--max-lcom-delta`) wiring flags into `RunDiff`.
  Threshold overrides are TIGHTEN-ONLY: any value looser than the protected
  default (e.g. `--max-instability-delta 0.30` > 0.15) MUST be rejected with
  exit code 2 and a diagnostic to stderr before any verdict is computed.
  Register via `cmd.AddCommand(diffCmd())` in `rootCmd()`.
- [x] 3.5 [P] Add `cmd/vibe-check/diff_test.go` driving `RunDiff` with
  `bytes.Buffer`: assert the JSON payload's `verdict`/`reasons`/`newCycles`/
  `resolvedCycles`, the per-module delta rows array (with specific Δ values in
  stable sorted order), the unreliable-measurement flag, and `EntropyDirection`
  for an improvement fixture (APPROVE, direction `improving`), a degradation
  fixture (REQUEST_CHANGES, direction `degrading`), a COMMENT-band fixture
  (COMMENT, direction `stable`), and a partial-build fixture (COMMENT,
  unreliable flag `true`, added/removed suppressed). Assert exit code
  `0` for all valid-input cases (verdict carried in payload, not exit
  code). Assert exit code `2` separately for: (a) missing/unreadable base
  file, (b) missing/unreadable PR file, (c) readable-but-schema-invalid base,
  (d) readable-but-schema-invalid PR — and in each case assert NO verdict
  payload is emitted. Assert that `RunDiff` without `--json` writes the default
  human-readable per-package delta table with a verdict line. Assert tighten-
  only override enforcement: a looser-than-default threshold flag exits 2 with
  no payload. Assert CLI determinism: invoking `RunDiff` twice on the same
  inputs yields byte-identical stdout. Follow AGENTS.md Testing Conventions
  (`TestFunc_Scenario` naming, `got %v, want %v` assertions). Target ≥ 80%
  coverage for the new CLI code.

## 4. Embedded divisor-entropy Agent Asset

- [x] 4.1 Author `internal/scaffold/assets/agents/divisor-entropy.md`
  frontmatter: a one-line `description`, `mode: subagent`, `temperature: 0.1`
  (matching `divisor-architect`), `edit: deny`, `webfetch: deny`, and a
  granular `bash` allowlist (`"*": "deny"` plus `allow` for `git merge-base *`,
  `git rev-parse *`, `git worktree add *`, `git worktree remove *`,
  `git worktree prune`, `git fetch origin *`, `git check-ref-format *`,
  `vibe-check analyze *`, and `vibe-check diff *` — nine entries). Mirror the
  existing granular-allowlist precedent in
  `.opencode/agents/divisor-curator.md` (the first divisor to enable bash).
  Satisfies spec `divisor-entropy-agent` Requirement "Minimal Tool
  Permissions".
- [x] 4.2 Insert the AP-006 provenance marker
  `<!-- scaffolded by vibe-check -->` (a stable, version-less marker)
  immediately after the frontmatter, then write the agent
  body: `# Role`, an optional `## Step 0: Prior Learnings` (Dewey-optional),
  and `## Source Documents` (AGENTS.md, constitution,
  `.opencode/uf/packs/severity.md`, other packs) matching the existing divisor
  structure.
- [x] 4.3 Write `## Code Review Mode` describing the delta workflow, ordering
  ref validation so the regex gate runs FIRST: reject any base ref that does
  not match `^[A-Za-z0-9._/-]+$` (the primary metacharacter defense, applied
  before the ref reaches any command), THEN run `git check-ref-format` for
  ref-semantic validation (rejecting `..`, `@{`, leading `-`, `.lock`), THEN
  resolve it with `git rev-parse --verify --end-of-options "<ref>^{commit}"`
  (so a ref can never be parsed as a flag or chained command). Default the base
  to `git merge-base HEAD origin/main`; if the base ref must be fetched, run
  `git fetch origin <ref>` only after that same regex has validated the ref
  (note `git fetch` takes no `--end-of-options`, so the regex is the
  load-bearing pre-fetch guard). Create an isolated `git worktree` at the
  resolved SHA in a unique temp directory outside the repo (running
  `git worktree prune` first to clear any stale entry), run
  `vibe-check analyze --output <file>` (NOT gaze/goda) with a bounded
  `--timeout` to materialize a JSON graph for both the base worktree and the PR
  checkout using the same binary (using the `--output` flag rather than shell
  redirection, which the allowlist forbids), run `vibe-check diff` on the two
  resulting JSON graph files, then remove the worktree with
  `git worktree remove --force`. Satisfies "Structural Delta Computation"
  and "Base Branch Isolation".
- [x] 4.4 Document that the deterministic verdict is computed by
  `vibe-check diff` (the tested Go `DecideVerdict` gates: new cycle → REQUEST
  CHANGES; ΔI ≥ 0.15 → REQUEST CHANGES; ΔD ≥ 0.20 → REQUEST CHANGES;
  ΔLCOM ≥ 2 → REQUEST CHANGES; smaller shifts → COMMENT; improve/stable →
  APPROVE) — the agent reports and explains the tool's verdict rather than
  computing thresholds in-prompt. Satisfies "New Circular Dependency
  Detection" and "Deterministic Verdict Thresholds".
- [x] 4.5 Write `## Out of Scope` (delegating other personas' domains) and the
  `## Output Format` requiring base+PR SHAs, the per-package delta table
  (Ca, Ce, Instability, Abstractness, Distance, LCOM: base → PR, Δ) sourced
  from the `vibe-check diff` JSON, the new-cycle list, entropy direction,
  standard `### [SEVERITY]` finding blocks (per `severity.md`), a domain Score
  (1–10), and a final verdict line. Satisfies "Auditable Delta Report".
- [x] 4.6 Write `## Decision Criteria` and a graceful-degradation clause:
  return COMMENT (never a false APPROVE, never a crash) when the binary is
  missing, the base ref cannot be resolved/analyzed, the PR ref fails to
  analyze, worktree creation fails, or `vibe-check diff` cannot produce a
  verdict. Satisfies "Graceful Degradation".
- [x] 4.7 Write a `## Security / Operating Constraints` section in the asset
  documenting that `vibe-check analyze` executes the target's build tooling
  (compilation and cgo) — a residual code-execution surface that
  `GOTOOLCHAIN=local` does NOT close (it only closes the toolchain-directive
  download vector) — and that the divisor therefore MUST only run on refs the
  CI context already trusts and MUST NOT be wired into untrusted-fork-PR CI.
  Satisfies "Trusted-Refs-Only Operating Constraint".

## 5. Scaffold Writer Logic

- [x] 5.1 In `internal/scaffold/scaffold.go`, define `Options{ TargetDir
  string; Force bool }` and `Result{ Written []string; Skipped []string;
  Forced []string }` with GoDoc on every exported symbol.
- [x] 5.2 Implement `Run(opts Options) (*Result, error)` (named `Run`, not
  `Scaffold`, to avoid the package/function stutter and follow the AP-002
  canonical entry-point name): validate `TargetDir` using
  `metrics.ValidateProjectPath` (which requires an existing directory and
  rejects traversal); resolve symlinks using a deepest-existing-ancestor
  strategy — starting from the already-validated root, walk toward the
  destination finding the deepest directory that actually exists, then create
  each new component incrementally and verify containment after creation (or
  use `O_NOFOLLOW` semantics) so a symlink placed at `.opencode/` or
  `.opencode/agents/` cannot redirect writes outside the validated root; add an
  injectable write/filesystem seam to `Options` (e.g. `WriteFile func(path
  string, data []byte, perm fs.FileMode) error`) so the I/O-failure exit-2
  branch is testable unconditionally without requiring root access; walk
  `assetsFS`; create the `.opencode/agents/` tree with mode `0o755`; write
  each asset with mode `0o644`. Satisfies init-command "Deploy Embedded Agent
  Assets", "Target Path Validation", and "Safe File Permissions".
- [x] 5.3 Implement skip-existing default (record in `Skipped`) and `Force`
  overwrite (record in `Forced`); on overwrite, normalize the destination file
  mode to `0o644` (e.g. `O_TRUNC` write followed by `Chmod`) so a pre-existing
  loose mode cannot persist; wrap I/O errors with
  `fmt.Errorf("context: %w", err)` naming the operation and path. `Result`
  slices MUST be returned in stable, sorted order. Satisfies "Idempotent
  Skip-Existing Behavior" and "Force Overwrite".

## 6. Scaffold Package Tests

- [x] 6.1 [P] Add `internal/scaffold/scaffold_test.go` using `t.TempDir()`:
  assert assets land in `.opencode/agents/`, directory mode is `0o755`, file
  mode is `0o644`, nested dirs are created, default run skips an existing file
  (reported in `Skipped`), and `Force` overwrites (reported in `Forced`).
- [x] 6.2 [P] Add a `Force`-overwrite permission test: pre-create the target
  file with a loose mode (e.g. `0o666`), run with `Force`, and assert the
  resulting file mode is normalized to `0o644`.
- [x] 6.3 [P] Add a path-traversal / nonexistent-root rejection test (a
  `..`-escaping or nonexistent `TargetDir` returns an error and writes
  nothing) and a symlink-escape test (a `.opencode/agents` resolving outside
  the root is rejected), guarding the symlink case with a `runtime.GOOS` check
  / `t.Skip` on platforms without reliable symlink support.
- [x] 6.4 [P] Add an embedded-asset contract test: parse the embedded
  `divisor-entropy.md` frontmatter and assert `description` is non-empty,
  `mode: subagent`, `temperature: 0.1`, `edit: deny`, `webfetch: deny`, and a
  granular `bash` block whose catch-all is `deny` and whose allowlist is
  **exactly** (set-equality — failing on any missing *or* extra entry) the nine
  entries `git merge-base *`, `git rev-parse *`, `git worktree add *`,
  `git worktree remove *`, `git worktree prune`, `git fetch origin *`,
  `git check-ref-format *`, `vibe-check analyze *`, and `vibe-check diff *`;
  assert the AP-006 provenance marker prefix `<!-- scaffolded by vibe-check` is
  present, the embedded filename matches the `divisor-*.md` discovery glob, and
  the Source Documents, Code Review Mode, Output Format, Decision Criteria, and
  Security / Operating Constraints sections exist. The bash allowlist
  set-equality assertion MUST use stdlib string/byte scanning only to extract
  the `bash:` frontmatter block region — asserting presence of each exact
  quoted allow key, absence of any other allow entry, and `"*": "deny"` — with
  NO third-party YAML library and no new module dependency. Target ≥ 80%
  coverage for the package.
- [x] 6.5 [P] Add a `Result` sort-order test over a synthetic multi-entry
  `fs.FS` (at least three assets with non-alphabetical source ordering): assert
  that `Run` returns `Written`, `Skipped`, and `Forced` slices in stable sorted
  (ascending lexicographic) order regardless of the `fs.FS` walk order.

## 7. `vibe-check init` CLI Command

- [x] 7.1 Add `cmd/vibe-check/init.go` with `InitOptions{ Stdout, Stderr
  io.Writer; Path string; Force bool; JSON bool }` and
  `InitResult{ Written, Skipped, Forced []string; ExitCode int }`, following
  the AP-002/AP-003 testable pattern used by `analyze.go`.
- [x] 7.2 Implement `RunInit(ctx context.Context, opts InitOptions)
  (*InitResult, error)`: default `Path` to `.`, honor `ctx.Err()` between
  writes, call `scaffold.Run`, and print a human summary or (with `--json`) a
  JSON object containing `written`, `skipped`, and `forced`. Satisfies
  "Machine-Readable Summary".
- [x] 7.3 Map exit codes: `0` on success (including all-skipped) and `2` on
  invalid path or I/O failure. Satisfies "Target Path Validation" and
  "Success Exit Code".
- [x] 7.4 Add `initCmd() *cobra.Command` (arg `[path]`, flags `--force`,
  `--json`) wiring flags into `RunInit`; register it via
  `cmd.AddCommand(initCmd())` in `rootCmd()` (`cmd/vibe-check/root.go`).

## 8. `vibe-check init` CLI Tests

- [x] 8.1 [P] Add `cmd/vibe-check/init_test.go` driving `RunInit` with
  `bytes.Buffer` stdout/stderr: assert the human summary lists written files, a
  second run reports skips, and `--force` reports forced overwrites.
- [x] 8.2 [P] Assert the `--json` output unmarshals and carries the specific
  expected values per lifecycle stage: `written == ["divisor-entropy.md"]`
  with empty `skipped`/`forced` on first run; the file in `skipped` on a
  second run; the file in `forced` under `--force`.
- [x] 8.3 [P] Assert exit codes: `0` on success and `2` on an invalid/
  traversal path (asserting nothing is written). Exercise the I/O-failure
  exit-2 branch deterministically via the injected write/filesystem seam added
  to `scaffold.Options` in task 5.2 — no `os.Geteuid()==0` guard required.
  Target ≥ 80% coverage for the new CLI code.

## 9. Dogfood & Agent Behavior Validation

- [x] 9.1 Generate vibe-check's own copy by running
  `go run ./cmd/vibe-check init .` so `.opencode/agents/divisor-entropy.md` is
  produced from the embedded source of truth (do not hand-author a second
  copy). Satisfies "deployed via vibe-check init" and Review Council
  auto-discovery.
- [x] 9.2 Add improvement, degradation, COMMENT-band, and partial-build
  fixtures under `metrics/testdata/` (and/or `internal/scaffold/testdata/`) as
  schema-valid `ModuleGraph` JSON pairs that each pass `metrics.Validate`, with
  each fixture pinned to the specific gate it exercises (new cycle, ΔI, ΔD,
  ΔLCOM, stable, or a partial build with `Status:"partial"`/load-error
  `Warnings`).
- [x] 9.3 Add a documented validation checklist asserting the entropy divisor,
  driven by `vibe-check diff`, yields APPROVE for the improvement fixture and
  REQUEST CHANGES for the degradation fixture (acceptance criterion #7 from
  issue #3). For the agent's runtime degradation/isolation behaviors that are
  not reachable by the Go tests, document (labelled "manual-only" with a brief
  rationale) that the agent returns COMMENT (never a false APPROVE) for
  binary-missing, PR-won't-build, and partial-build inputs, and removes the
  worktree even on failure. The numeric verdict logic itself is covered by the
  automated table-driven tests in task 2.4.

## 10. Validation & CI Parity Gate

- [x] 10.1 Run `go build ./...` and `go build ./cmd/vibe-check` — both succeed.
- [x] 10.2 Run `go test -race -count=1 ./...` — all tests pass with the race
  detector enabled.
- [x] 10.3 Run `go vet ./...` and `golangci-lint run ./...` — no findings.
- [x] 10.4 Verify GoDoc comments exist on every exported symbol in
  `metrics/delta.go`, `metrics/verdict.go`, `internal/scaffold/`, and the new
  `cmd/vibe-check/` files.

## 11. Documentation

- [x] 11.1 Update `AGENTS.md` Project Structure to add `metrics/delta.go`,
  `metrics/verdict.go`, `internal/scaffold/` (embed.go, scaffold.go,
  assets/agents/divisor-entropy.md, doc.go, testdata/), and
  `cmd/vibe-check/init.go` + `cmd/vibe-check/diff.go`; update the Architecture
  (Layer 1 delta/verdict, Layer 3 CLI) and Build & Test Commands sections to
  mention `vibe-check init`, `vibe-check diff`, the new `vibe-check analyze
  --output`/`-o` flag, and the forced `GOTOOLCHAIN=local` analysis behavior.
- [x] 11.2 Add a `CHANGELOG.md` entry under the unreleased section for the
  `divisor-entropy` agent, the `vibe-check init` command, the
  `vibe-check diff` command, the new `vibe-check analyze --output`/`-o` flag,
  and the forced `GOTOOLCHAIN=local` hardening of `vibe-check analyze` (noting
  the compatibility trade-off for a trusted module whose `go.mod` `toolchain`
  directive requires a newer-than-local Go).
- [x] 11.3 Update `README.md` if the project description or usage changes to
  cover `vibe-check init` and `vibe-check diff`.
- [x] 11.4 File a documentation issue in `unbound-force/website` for the
  user-facing `vibe-check init` / `vibe-check diff` commands, the new
  `vibe-check analyze --output` flag and forced `GOTOOLCHAIN=local` behavior,
  and the entropy divisor (per the Documentation/Website Sync gate) before PR
  merge.
- [x] 11.5 File a provenance follow-up issue covering the new `vibe-check diff
  --json` and `vibe-check init --json` payloads (adding producer/version/
  timestamp metadata to satisfy Constitution Principles I and III), referenced
  by the Constitution Alignment note in `proposal.md`.

<!-- spec-review: passed -->
<!-- code-review: passed -->
