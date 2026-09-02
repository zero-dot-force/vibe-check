# Tasks: /vibe-check Command and vibe-check-reporter Agent

## 1. Scaffold System Extension

- [x] 1.1 Add `assets/commands/` directory to `internal/scaffold/` with a `.gitkeep` or the command asset file
- [x] 1.2 Add `//go:embed assets/commands/*.md` directive in `internal/scaffold/embed.go` with a new `CommandAssets` variable
- [x] 1.3 Extend `scaffold.Run()` in `internal/scaffold/scaffold.go` to deploy command assets to `.opencode/commands/` in the target directory (same symlink-safe, containment-checked pattern as agent deployment)
- [x] 1.4 Add tests in `internal/scaffold/scaffold_test.go` for command asset deployment: fresh init (verify 0o755 dir, 0o644 files, byte-match content), skip existing, force overwrite, mixed result output with category-prefixed paths. Tests use real embedded assets for contract verification and synthetic `fstest.MapFS` for edge cases. Update existing `TestRun_DeploysEmbeddedAsset` to expect both agent and command files in `Result.Written`
- [x] 1.5 Verify `go build ./...` and `go test -race -count=1 ./internal/scaffold/...` pass

## 2. Agent Asset

- [x] 2.1 Create `internal/scaffold/assets/agents/vibe-check-reporter.md` with YAML frontmatter (`description`, `mode: subagent`, `temperature`, `permission` block with bash allowlist for `vibe-check analyze *` and `git rev-parse *`) and `<!-- scaffolded by vibe-check -->` provenance marker after frontmatter
- [x] 2.2 Write agent body: role definition, source documents to read, mode parsing instructions
- [x] 2.3 Write summary mode instructions: run `vibe-check analyze --output`, parse ModuleGraph JSON, produce traffic-light indicator and aggregate metrics
- [x] 2.4 Write detailed mode instructions: per-package metric table, zone classification, warnings, remediation guidance
- [x] 2.5 Write trending mode instructions: retrieve snapshot from Dewey, compare current vs baseline, show per-package direction (improving/degrading/stable), store new snapshot
- [x] 2.6 Write natural language interpretation guidelines: metric explanations, zone guidance, cycle explanation
- [x] 2.7 Write Dewey snapshot storage instructions: `dewey_store_learning` with tag `vibe-check-snapshot:{module-path}`, commit SHA, module path, timestamp, compact per-package summary (~50 bytes/package). Include deduplication: check for existing snapshot at current commit SHA before storing
- [x] 2.8 Write graceful degradation section: Dewey unavailable handling, analysis error handling (binary not found, timeout, malformed JSON), unrecognized mode handling, input validation (package pattern must match `^[A-Za-z0-9./_-]+$`)
- [x] 2.9 Add contract test for `vibe-check-reporter.md` in `scaffold_test.go`: validate frontmatter (`description`, `mode: subagent`, `temperature`, `permission` block), provenance marker (`<!-- scaffolded by vibe-check -->`), required sections, bash allowlist entries (`vibe-check analyze *`, `git rev-parse *`, catch-all deny)

## 3. Command Asset

- [x] 3.1 Create `internal/scaffold/assets/commands/vibe-check.md` with YAML frontmatter (`description`, `agent: vibe-check-reporter`) and `<!-- scaffolded by vibe-check -->` provenance marker after frontmatter
- [x] 3.2 Write command body: description, usage syntax, mode table (summary/detailed/trending), examples, instructions to pass `$ARGUMENTS` to agent
- [x] 3.3 Add contract test for `vibe-check.md` command asset in `scaffold_test.go`: validate frontmatter (`description`, `agent: vibe-check-reporter`), provenance marker, mode documentation (summary/detailed/trending), `$ARGUMENTS` passthrough instruction

## 4. Integration Verification

- [x] 4.1 Run `go build ./...` to verify embed directives compile
- [x] 4.2 Run `go test -race -count=1 ./...` to verify all tests pass
- [x] 4.3 Run `go vet ./...` to verify no vet issues
- [x] 4.4 Test `go run ./cmd/vibe-check init /tmp/test-project` end-to-end and verify both `.opencode/agents/vibe-check-reporter.md` and `.opencode/commands/vibe-check.md` are deployed
- [x] 4.5 Verify existing `divisor-entropy.md` agent deployment is unchanged

## 5. Documentation Updates

- [x] 5.1 Update `README.md` "Deploying agents: `vibe-check init`" section to document command deployment alongside agents, the new `/vibe-check` command, and the `vibe-check-reporter` agent
- [x] 5.2 Update `AGENTS.md` project structure to include `internal/scaffold/assets/commands/`, `internal/scaffold/assets/agents/vibe-check-reporter.md`, and updated `embed.go`/`scaffold.go` descriptions
- [x] 5.3 Update `internal/scaffold/doc.go` to describe both asset categories (agents and commands), the new embed directive, and the expanded deployment targets
- [x] 5.4 Update `cmd/vibe-check/init.go`: modify `writeInitSummary()` to reference both agents and commands directories, update cobra `Short`/`Long` descriptions, update GoDoc comments on `InitOptions`, `InitResult`, and `RunInit`
- [x] 5.5 Add `CHANGELOG.md` entry for the new command and agent
- [x] 5.6 File a documentation issue against `zero-dot-force/vibe-check` for README/AGENTS.md updates and a website documentation sync issue against `unbound-force/website` (constitution requirement)

<!-- spec-review: passed -->
<!-- code-review: passed -->
