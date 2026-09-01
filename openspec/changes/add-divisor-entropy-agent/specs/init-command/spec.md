## ADDED Requirements

### Requirement: Deploy Embedded Agent Assets

The `vibe-check init [path]` command SHALL write the agent assets embedded in the
binary into `<path>/.opencode/agents/`, creating the `.opencode/` and
`.opencode/agents/` directories when they do not exist. The target `path` root
MUST already exist and be a directory; the command validates the root and does
not create it. The `path` argument SHALL default to the current directory (`.`)
when omitted.

#### Scenario: Init into a project without agents

- **WHEN** a user runs `vibe-check init` in a project that has no
  `.opencode/agents/` directory
- **THEN** the command creates `.opencode/agents/` and writes `divisor-entropy.md`
  into it, reporting the file as written

#### Scenario: Default path is the current directory

- **WHEN** a user runs `vibe-check init` with no path argument
- **THEN** the command targets the current working directory

### Requirement: Idempotent Skip-Existing Behavior

By default the command SHALL skip any asset whose destination file already exists,
leaving the existing file unmodified, and SHALL report each skipped file.
Skipping an existing file MUST NOT be treated as an error.

#### Scenario: Existing agent file is skipped

- **WHEN** `.opencode/agents/divisor-entropy.md` already exists and the user runs
  `vibe-check init` without `--force`
- **THEN** the command leaves the file unchanged, reports it as skipped, and
  exits 0

### Requirement: Force Overwrite

When invoked with `--force`, the command SHALL overwrite existing destination
files with the embedded asset content, SHALL normalize each overwritten file's
mode to `0o644`, and SHALL report each overwritten file as forced.

#### Scenario: Force overwrites an existing file

- **WHEN** `.opencode/agents/divisor-entropy.md` exists and the user runs
  `vibe-check init --force`
- **THEN** the command overwrites the file with the embedded content, resets its
  mode to `0o644`, and reports it as forced

### Requirement: Machine-Readable Summary

When invoked with `--json`, the command SHALL emit a JSON object summarizing the
files written, skipped, and forced. Without `--json`, it SHALL emit a
human-readable summary.

#### Scenario: JSON summary output

- **WHEN** a user runs `vibe-check init --json`
- **THEN** the command prints a JSON object with `written`, `skipped`, and
  `forced` arrays describing the outcome

#### Scenario: Summary reports the agent as written on first run

- **WHEN** `vibe-check init --json` runs in a project without
  `.opencode/agents/divisor-entropy.md`
- **THEN** the `written` array contains `divisor-entropy.md` and the `skipped`
  and `forced` arrays are empty

### Requirement: Safe File Permissions

The command SHALL create directories with mode `0o755` and write asset files with
mode `0o644`.

#### Scenario: Created directory and files use safe modes

- **WHEN** the command creates `.opencode/agents/` and writes an asset
- **THEN** the directory has mode `0o755` and the asset file has mode `0o644`

### Requirement: Target Path Validation

The command SHALL validate the target path and reject unsafe or traversal paths.
On an invalid path or an I/O failure the command SHALL exit with code 2.

Symlink resolution MUST use a deepest-existing-ancestor strategy: because
`.opencode/agents/` may not exist on first run, `EvalSymlinks` on the full leaf
path would fail. The implementation MUST walk from the validated root to find the
deepest existing ancestor, create each new directory component incrementally, and
verify containment after creation. Alternatively, `O_NOFOLLOW` semantics MAY be
used. This prevents a symlink placed at `.opencode/` or `.opencode/agents/` from
redirecting writes outside the validated root. An injectable write/filesystem seam
MUST be available in `scaffold.Options` so that the I/O-failure exit-2 branch is
testable unconditionally without requiring root access.

#### Scenario: Reject path traversal

- **WHEN** a user supplies a target path that escapes the intended location (for
  example one containing `..` segments that fail validation)
- **THEN** the command reports the invalid path and exits with code 2

#### Scenario: I/O failure exits non-zero

- **WHEN** the command cannot create the directory or write a file
- **THEN** it reports the failed operation and path and exits with code 2

#### Scenario: Nonexistent target root is rejected

- **WHEN** a user supplies a `path` that does not exist or is not a directory
- **THEN** the command reports the invalid path and exits with code 2

#### Scenario: Destination symlink escaping the root is rejected

- **WHEN** the resolved `.opencode/agents` destination would fall outside the
  validated target root (for example via a symlink placed at `.opencode/` before
  the command runs)
- **THEN** the command detects the escape using deepest-existing-ancestor
  resolution, reports the unsafe path, and exits with code 2

#### Scenario: I/O failure is testable via injected seam

- **WHEN** an injected filesystem writer seam in `scaffold.Options` is configured
  to return an error
- **THEN** the command exits with code 2 without requiring elevated privileges or
  filesystem manipulation

### Requirement: Success Exit Code

The command SHALL exit with code 0 on success, including runs where every asset
was skipped because it already existed.

#### Scenario: Successful run exits zero

- **WHEN** the command completes without an invalid path or I/O error
- **THEN** it exits with code 0 regardless of whether files were written,
  skipped, or forced
