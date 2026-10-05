<!-- spec-review: passed -->

## 1. CLI: `--store` Flag on `vibe-check analyze`

- [x] 1.1 Add `--store` boolean flag to the `analyze` cobra command in `cmd/vibe-check/analyze.go`
- [x] 1.2 Implement git metadata resolution: `commitSHA` (via `git rev-parse HEAD`) and `branch` (via `git rev-parse --abbrev-ref HEAD`), using empty strings when not in a git repo
- [x] 1.3 Enrich the `ProvenanceInput` / `Provenance` with snapshot fields (`commitSHA`, `branch`, `modulePath`) when `--store` is set
- [x] 1.4 Ensure `--no-provenance` takes precedence over `--store` (no provenance when both are set)
- [x] 1.5 Add table-driven tests for `--store` output covering: (a) git repo on a named branch — verify `commitSHA` is 40-char hex and `branch` matches; (b) outside a git repo — verify `commitSHA` and `branch` are empty strings; (c) `--no-provenance` + `--store` — verify no provenance object is emitted; (d) detached HEAD — verify `branch` is empty or `HEAD`; (e) `modulePath` resolves to the Go module path from `go.mod`; (f) missing `go.mod` — verify `modulePath` is empty with a warning

## 2. Agent: `mx-f-architecture-trend.md`

- [x] 2.1 Create `internal/scaffold/assets/agents/mx-f-architecture-trend.md` agent prompt with role, tools (Dewey MCP + bash), and modes (snapshot-store, trend-report, retention-enforce)
- [x] 2.2 Implement snapshot storage workflow: invoke `vibe-check analyze --store`, capture output, call `dewey_store_learning` with tag `vibe-check-snapshot` and category `metric-snapshot`
- [x] 2.3 Implement snapshot retrieval: call `dewey_semantic_search` / `dewey_get_page` filtered by module path, parse timestamps, order by `generatedAt`
- [x] 2.4 Implement trend detection: compute linear regression (slope + R²) per package per metric over 7/30/90-day windows
- [x] 2.5 Implement trend classification: improving (negative slope, |slope| ≥ threshold, R² ≥ 0.5), degrading (positive slope, same thresholds), stable (R² < 0.5 or |slope| < threshold), insufficient data (< 3 snapshots)
- [x] 2.6 Implement drift alert logic: sustained degradation (N consecutive snapshot decreases, N=5 default) and predictive alert (projected threshold crossing within M days, M=30 default)
- [x] 2.7 Implement retention policy: identify daily snapshots > 90 days (prune unless Sunday), weekly-Sunday snapshots > 1 year (prune), invoke removal
- [x] 2.8 Implement sparkline rendering: Unicode block character (▁▂▃▄▅▆▇█) normalization to metric range, max 30 chars
- [x] 2.9 Implement health report output: overall classification, per-package table with sparklines and trends, drift alerts, projected crossings
- [x] 2.10 Implement Dewey-unavailable graceful degradation per existing trending-mode spec

## 3. Scaffold: Embed Agent Asset

- [x] 3.1 Register `mx-f-architecture-trend.md` in `internal/scaffold/assets/` Go embedding (`assets/agents/` via `//go:embed`)
- [x] 3.2 Update scaffold contract tests in `internal/scaffold/scaffold_test.go` to verify the new agent asset is embedded and writable
- [x] 3.3 Verify `vibe-check init` deploys the new agent to `.opencode/agents/mx-f-architecture-trend.md`

## 4. Trending Mode Extension

- [x] 4.1 Update the `vibe-check-reporter.md` agent's trending mode section to delegate time-series work to `mx-f-architecture-trend.md` when ≥2 snapshots exist
- [x] 4.2 Add sparkline output format to the reporter agent's trending output template
- [x] 4.3 Add multi-window (7/30/90-day) classification display to trending output
- [x] 4.4 Integrate drift alert results into trending mode output when available

## 5. Validation & Documentation

- [x] 5.1 Run `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...` and resolve any failures
- [x] 5.2 Add `CHANGELOG.md` entry for architectural drift tracking feature
- [x] 5.3 Assess `README.md` impact for `--store` flag documentation and file a documentation issue for user-facing changes (trending mode extension, `--store` flag)
<!-- code-review: passed -->