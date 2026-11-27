## MODIFIED Requirements

### Requirement: Subprocess hygiene and data-source constraints

The adapter SHALL invoke ripwire with a sanitized environment
(`metrics.SanitizeEnvironment`), bounded output capture, and context
cancellation that terminates the subprocess. Every ripwire invocation MUST
return in bounded time, including when a process started by ripwire inherits
and keeps its output pipes open. A ripwire run that exits successfully but
leaves its output pipes held open MUST fail closed. The adapter MUST NOT read
ripwire's internal cache, MUST NOT consume ranked agent-facing answers as
metrics input, and MUST NOT vendor or bundle the ripwire binary.

#### Scenario: Credential environment variables not leaked

- **GIVEN** the process environment contains `GITHUB_TOKEN=secret`
- **WHEN** the adapter spawns ripwire
- **THEN** the subprocess environment MUST NOT contain `GITHUB_TOKEN`

#### Scenario: Context cancellation terminates the subprocess

- **GIVEN** a ripwire-backed analysis in flight
- **WHEN** the context is cancelled
- **THEN** the adapter MUST return an error wrapping the context error and the
  ripwire subprocess MUST NOT be left running

#### Scenario: Cancellation returns while a child holds the pipes

- **GIVEN** a ripwire invocation in flight whose child process holds the
  stdout and stderr pipes open
- **WHEN** the context is cancelled
- **THEN** the invocation MUST return an error wrapping the context error
  within the runner's wait delay plus scheduling slack, not block until the
  child exits

#### Scenario: Oversized output returns while a child holds the pipes

- **GIVEN** a ripwire invocation whose child process writes more than the
  capture bound to stdout and holds stderr open
- **WHEN** the stdout capture crosses the bound
- **THEN** the invocation MUST return the oversized-output error naming stdout
  and MUST NOT return truncated output

#### Scenario: Successful exit with abandoned pipes fails closed

- **GIVEN** a ripwire invocation that exits with status 0 while a child
  process it started still holds its output pipes
- **WHEN** the runner's wait delay elapses
- **THEN** the invocation MUST return an error and MUST NOT return the
  captured stdout as a successful result

#### Scenario: Analysis does not depend on ripwire's cache

- **GIVEN** an environment where ripwire's internal cache directory is absent
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** analysis MUST succeed using only the documented structural outputs
  (`--arch`, `--deps`, `--export=cc.json`)
