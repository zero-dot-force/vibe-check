# Spec Delta

## Purpose

Lets `vibe-check init` complete the deployment story: after writing the
`divisor-entropy` agent into a target project's `.opencode/`, it registers
the agent in the target's Unbound Force review policies so review dispatch
works without hand-editing policy files.

## ADDED Requirements

### Requirement: Register Review Agent After Asset Deployment

The `init` command SHALL, after deploying agent and command assets, attempt
to register `divisor-entropy` in the target's `.uf/reviewer-capabilities.yaml`
and `.uf/review-matrix.yaml`. Registration SHALL run regardless of whether
the asset files themselves were written, skipped, or forced.

#### Scenario: Registration runs on first init into a uf-scaffolded project

- **WHEN** `vibe-check init` runs in a project whose `.uf/` policy files
  exist but contain no `divisor-entropy` entries
- **THEN** both policy files gain the canonical `divisor-entropy` entries
  after the assets are deployed

#### Scenario: Registration runs when every asset was skipped

- **WHEN** `vibe-check init` runs in a project where every asset file
  already existed and was reported as skipped
- **THEN** registration is still attempted and its outcome is reported

### Requirement: Canonical Capabilities Manifest Entry

When `.uf/reviewer-capabilities.yaml` is present and contains no
`divisor-entropy` reviewer entry, the command SHALL insert the canonical
entry in sorted position among the `capability: review` reviewers,
preserving all existing bytes other than the inserted canonical lines.
Presence and sorted-position comparisons use extracted agent names
(whitespace-trimmed, trailing comments stripped, quotes removed), never
raw line equality.

#### Scenario: Entry inserted in sorted position

- **WHEN** the `reviewers:` list contains `divisor-curator` and
  `divisor-guard` entries but no `divisor-entropy` entry
- **THEN** the inserted entry appears between the `divisor-curator` and
  `divisor-guard` entries

#### Scenario: Inserted entry uses the canonical content

- **WHEN** the capabilities manifest gains the `divisor-entropy` entry
- **THEN** the inserted block is the four-line canonical entry —
  `- agent: divisor-entropy`, `capability: review`, `scopes:`, and
  `- standard` — indented to match the surrounding reviewer entries

#### Scenario: Commented or quoted agent name counts as present

- **WHEN** the `reviewers:` section contains an entry whose `- agent:` line
  carries a trailing comment or surrounding quotes around `divisor-entropy`
- **THEN** the entry counts as present and no duplicate entry is inserted

#### Scenario: Reviewers section absent

- **WHEN** `.uf/reviewer-capabilities.yaml` is present but has no
  `reviewers:` list
- **THEN** no entry is inserted, no error is raised, and a notice
  identifying the unexpected file shape is reported

### Requirement: Canonical Advisor List Entries

When `.uf/review-matrix.yaml` is present, the command SHALL insert
`- divisor-entropy` in sorted position in each `advisor` mode list
(`code`, `spec`, `triage`, `feedback`) that exists and lacks the entry.
Presence and sorted-position comparisons are by extracted item name —
whitespace-trimmed, trailing `# …` comments stripped, surrounding quotes
removed — never by raw line equality.

#### Scenario: All four mode lists gain the entry

- **WHEN** the matrix `advisor:` section has `code`, `spec`, `triage`, and
  `feedback` lists, none containing `divisor-entropy`
- **THEN** each list gains `- divisor-entropy` in sorted position between
  `divisor-curator` and `divisor-guard`

#### Scenario: Per-list granularity on partially registered files

- **WHEN** one advisor mode list already contains `- divisor-entropy` and
  the others do not
- **THEN** only the lists missing the entry are modified; the list that
  already has it is left byte-identical

#### Scenario: Commented or quoted item counts as present

- **WHEN** an advisor mode list contains `- divisor-entropy # hand-added` or
  `- "divisor-entropy"`
- **THEN** the item counts as present and no duplicate item is inserted

#### Scenario: Mode list absent

- **WHEN** a mode list (for example `triage:`) is absent from the
  `advisor:` section
- **THEN** that list is not created and a notice identifies the missing
  list, while the remaining existing mode lists are still spliced

### Requirement: Skip When Entry Already Present

Registration operates per unit: the capabilities manifest is one registration
unit, and each `advisor` mode list is its own registration unit. When a
registration unit already contains the `divisor-entropy` entry, the command
SHALL skip registration for that unit and leave it byte-identical —
regardless of the entry's scopes or values. `--force` SHALL NOT cause an
existing entry to be rewritten.

#### Scenario: Canonical entry already present

- **WHEN** `.uf/reviewer-capabilities.yaml` already contains the canonical
  `divisor-entropy` entry and `vibe-check init` runs
- **THEN** the file is left byte-identical and the registration is reported
  as skipped

#### Scenario: Partially registered matrix is completed, not skipped

- **WHEN** one advisor mode list already contains `- divisor-entropy` and
  the others do not, and `vibe-check init` runs
- **THEN** the lists missing the entry are modified, the pre-registered list
  is left byte-identical, and the review matrix is reported as written

#### Scenario: Fully registered matrix reported as skipped

- **WHEN** the review matrix is present with at least one advisor mode list
  and every existing mode list already contains `- divisor-entropy`
- **THEN** the file is left byte-identical and the review matrix is reported
  as skipped

#### Scenario: Absent or shape-failed units are notice-only

- **WHEN** a policy file is absent, the capabilities manifest has no
  `reviewers:` list, or the review matrix lacks the `advisor:` section or
  all of its mode lists
- **THEN** its outcome is reported through a notice only — never as
  written or skipped

#### Scenario: Non-canonical scopes are never rewritten

- **WHEN** an existing `divisor-entropy` entry declares different scopes or
  values than the canonical entry
- **THEN** the entry is left unchanged and reported as skipped

#### Scenario: Force does not affect registration

- **WHEN** `vibe-check init --force` runs in a project whose policy files
  already contain the `divisor-entropy` entries
- **THEN** the asset files are overwritten but the policy entries are left
  byte-identical and reported as skipped

### Requirement: Advisory Notice for Non-Review Presence

When a present capabilities entry lacks `capability: review` or has empty
scopes, the skip SHALL be accompanied by an advisory notice identifying the
gap; the entry itself is never rewritten.

#### Scenario: Present entry with non-review capability emits an advisory notice

- **WHEN** the capabilities manifest contains a `divisor-entropy` entry whose
  capability is not `review` (for example `capability: content`) or whose
  scopes are empty
- **THEN** the entry is left unchanged and reported as skipped, and an
  advisory notice states that the present entry may not satisfy review
  dispatch

### Requirement: Skip When Policy File Absent

When a policy file does not exist, the command SHALL skip registration for
it, report a notice in the output identifying the absent file, create
nothing, and treat the absence as success (exit code 0). An absent file
SHALL NOT be reported as written or skipped.

#### Scenario: Non-uf-scaffolded target

- **WHEN** `vibe-check init` runs in a project with no `.uf/` directory
- **THEN** assets are deployed, a notice reports each absent policy file,
  no `.uf/` paths are created, and the command exits 0

#### Scenario: One policy file present, one absent

- **WHEN** `.uf/review-matrix.yaml` exists but
  `.uf/reviewer-capabilities.yaml` does not
- **THEN** the matrix is registered normally and a notice reports the
  absent capabilities manifest

### Requirement: Preserve Hand-Curated Formatting

Insertion SHALL be line-based: every pre-existing byte — including comments,
blank lines, indentation, key order, and trailing-newline state — SHALL be
preserved except for the inserted canonical lines and, when the insertion
point follows a final line lacking a terminator, that line's appended
terminator (without which the inserted block would share a line with the
final existing entry). The rewritten file SHALL retain its existing file
mode.

#### Scenario: Comments and formatting preserved

- **WHEN** a policy file containing header comments, inline comments, and
  blank lines gains the `divisor-entropy` entry
- **THEN** all comments, blank lines, and unrelated entries are preserved
  verbatim and the only difference is the inserted canonical lines (plus an
  appended terminator only when insertion follows an unterminated final
  line)

#### Scenario: Unterminated final line gains a terminator

- **WHEN** the insertion point follows the file's final line and that line
  lacks a trailing newline
- **THEN** the missing terminator is appended before the inserted canonical
  lines so the inserted block does not share a line with the final existing
  entry

#### Scenario: Existing file mode preserved

- **WHEN** a policy file with mode `0o600` gains the `divisor-entropy`
  entry
- **THEN** the file still has mode `0o600` after the rewrite

### Requirement: Registration Outcome Reporting

The command SHALL report each policy file's registration outcome using the
category-prefixed names `reviewers/divisor-entropy` (capabilities manifest)
and `advisors/divisor-entropy` (review matrix) in the same written/skipped
lists used for assets, in both the human summary and the `--json` payload.
Existing payload keys SHALL keep their semantics; absent-file notices SHALL
be reported additively without altering existing keys.

#### Scenario: First registration reported as written

- **WHEN** `vibe-check init --json` registers `divisor-entropy` in both
  policy files of a uf-scaffolded project
- **THEN** the `written` array contains `reviewers/divisor-entropy` and
  `advisors/divisor-entropy`, and the `skipped` and `forced` arrays do not

#### Scenario: Existing payload keys unchanged

- **WHEN** `vibe-check init --json` runs with registration active
- **THEN** the payload still contains the `written`, `skipped`, `forced`,
  and `provenance` keys with their existing semantics for asset files;
  registration outcomes are reported under the `reviewers/` and
  `advisors/` category prefixes within the same arrays, and the additive
  `notices` key carries absent-file skips

#### Scenario: Absent-file notice appears in output

- **WHEN** a policy file is absent and registration is skipped for it
- **THEN** the command output includes a notice naming the absent file

### Requirement: Symlink and Non-Regular File Refusal

The command MUST NOT write through symlinks or non-regular files when
registering. If the `.uf` directory is a symlink or not a directory, or a
policy file is a symlink or not a regular file, the command SHALL refuse
the registration with an error and exit code 2, leaving the target
untouched.

#### Scenario: Symlinked policy file refused

- **WHEN** `.uf/reviewer-capabilities.yaml` is a symlink and
  `vibe-check init` runs
- **THEN** the command reports the refusal, exits 2, and the symlink and
  its target are unmodified

#### Scenario: Symlinked .uf directory refused

- **WHEN** the target's `.uf` directory is itself a symlink
- **THEN** the command reports the refusal and exits 2

### Requirement: Registration I/O Failure Handling

A failure to read or write a present policy file SHALL be reported as an
error with exit code 2, naming the failed file and operation, consistent
with the command's existing I/O failure semantics. Because the rewrite is a
single truncating write, the error (or an accompanying notice) SHALL state
that the policy file may be partially written and that it can be restored
from version control.

#### Scenario: Write failure during registration

- **WHEN** rewriting a policy file fails due to an I/O error
- **THEN** the command reports the failed path and operation and exits 2

### Requirement: Only the Entropy Review Persona Is Registered

Registration SHALL touch only the `divisor-entropy` entries described in
this spec. No other agent entry — and no other section of either policy
file (including model overrides, profiles, runs, and risk augmentation) —
SHALL be added, removed, or rewritten.

#### Scenario: Non-review agents are not registered

- **WHEN** `vibe-check init` deploys `vibe-check-reporter` and
  `mx-f-architecture-trend` alongside `divisor-entropy` into a
  uf-scaffolded project
- **THEN** the policy files gain only the `divisor-entropy` entries

#### Scenario: Unrelated sections untouched

- **WHEN** `.uf/review-matrix.yaml` gains `- divisor-entropy` in its
  advisor lists
- **THEN** the `always:`, `profiles:`, `defaults:`, `runs:`,
  `risk_augmentation:`, and `limits:` sections are byte-identical
