# Spec: Scaffold Command Deployment

## ADDED Requirements

### Requirement: Scaffold embeds command assets

The scaffold package SHALL embed markdown files from
`assets/commands/*.md` via a new `//go:embed` directive in `embed.go`.

#### Scenario: Command asset embedded at compile time

- **GIVEN** the `assets/commands/` directory contains at least one `.md` file
- **WHEN** the vibe-check binary is built
- **THEN** the `internal/scaffold` package embeds all `.md` files from
  `assets/commands/` in addition to the existing `assets/agents/` files

Note: Go's `//go:embed` with a glob pattern requires at least one
matching file at compile time. The `assets/commands/vibe-check.md`
file serves this role, just as `divisor-entropy.md` does for the
agents directory. The `assets/commands/` directory MUST always contain
at least one `.md` file.

### Requirement: Scaffold deploys commands to .opencode/commands/

The `scaffold.Run()` function SHALL deploy command assets to
`.opencode/commands/` in the target directory, in addition to deploying
agent assets to `.opencode/agents/`.

#### Scenario: Fresh init with commands

- **GIVEN** a project directory with no existing `.opencode/commands/`
- **WHEN** `vibe-check init .` is run
- **THEN** the scaffold creates `.opencode/commands/` with directory
  permission 0o755, writes all embedded command assets with file
  permission 0o644, the deployed content byte-matches the embedded
  source, and `Result.Written` contains the command filenames with
  category prefix (e.g., `commands/vibe-check.md`)

#### Scenario: Existing commands directory

- **GIVEN** a project that already has `.opencode/commands/vibe-check.md`
- **WHEN** `vibe-check init .` is run without `--force`
- **THEN** the scaffold skips the existing file and reports it in the
  Skipped slice of the Result

#### Scenario: Force overwrite

- **GIVEN** a project that already has `.opencode/commands/vibe-check.md`
- **WHEN** `vibe-check init . --force` is run
- **THEN** existing command files are overwritten with file permission
  0o644 and reported in the Forced slice of the Result

#### Scenario: Symlink safety for commands directory

- **GIVEN** `.opencode/commands` is a symlink to an external directory
- **WHEN** `vibe-check init .` is run
- **THEN** the scaffold rejects the symlinked directory and returns an
  error, consistent with the existing symlink-safety behavior for
  `.opencode/agents/`

Note: The scaffold MUST apply the same symlink-safety and containment
checks to `.opencode/commands/` as it does to `.opencode/agents/`.
The `ensureDir()` function already handles this — it is reused for
both deployment targets.

### Requirement: Result struct includes command entries

The `scaffold.Result` struct's Written, Skipped, and Forced slices
SHALL include command file paths alongside agent file paths.

#### Scenario: Mixed result output

- **GIVEN** a project with no existing `.opencode/agents/` or
  `.opencode/commands/` directories
- **WHEN** `vibe-check init .` deploys agents and commands
- **THEN** the Result.Written slice contains entries from both
  categories with category-prefixed paths (e.g.,
  `agents/divisor-entropy.md`, `commands/vibe-check.md`) and the
  CLI output lists all deployed files grouped by category

## MODIFIED Requirements

### Requirement: Scaffold init output

The `vibe-check init` command output SHALL list both agent and command
files in its summary.

#### Scenario: Init with both asset types

- **GIVEN** `vibe-check init .` is run on any project
- **WHEN** the deployment completes successfully
- **THEN** the text output lists all written/skipped/forced files
  with their category prefix, regardless of whether they are agents
  or commands
