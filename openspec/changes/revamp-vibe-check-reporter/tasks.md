## 1. Metrics Guide

- [x] 1.1 Create `docs/metrics-guide.md` with ELI5 explanations for every metric (Instability, Abstractness, Distance from Main Sequence, LCOM4, Afferent Coupling, Efferent Coupling, Circular Dependencies, Code Duplication)
- [x] 1.2 Each metric section includes: "What it is" (plain-language definition), "Scale" (range + extremes meaning), "Why care" (practical impact), and emoji severity ranges
- [x] 1.3 Include at least one link to external Martin metrics reference
- [ ] 1.4 Verify the guide renders correctly on GitHub: `https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md`

## 2. Example Report Reference

- [x] 2.1 Create `.opencode/references/vibe-check-example-report.md` with a concrete example report matching the new output format (following the gaze example-report.md pattern)
- [x] 2.2 Example report SHALL include emoji title, metadata, legend block, per-package table with full metric names, health scorecard, prioritized recommendations, and duplication section
- [x] 2.3 Use fictional data with clearly unrealistic values so no one copies the numbers

## 3. Agent Rewrite

- [x] 3.1 Add emoji formatting contract section (closed vocabulary, mandatory usage rules, grade-to-emoji mapping) — modeled on gaze-reporter's `FORMATTING CONTRACT — MANDATORY, NON-NEGOTIABLE` block
- [x] 3.2 Add legend block template to be included in every report output format spec
- [x] 3.3 Redesign Summary Mode output format with 🏗️ title, 📖 legend, aggregate metrics, and traffic-light indicator
- [x] 3.4 Redesign Detailed Mode output format with full metric names, emoji zone indicators, 📊 metric table, 🏥 Health Scorecard, prioritized recommendations, and 📋 duplication section
- [x] 3.5 Add recommendation generation rules mapping JSON metric values to severity-prefixed recommendation templates
- [x] 3.6 Add grade computation rules for each scorecard dimension (Distance, LCOM4, Cycles, Duplication, Instability Spread) per design D2
- [x] 3.7 Add tone rules (conversational, no pedagogy, no filler, no puns, no slang) — matching gaze-reporter's tone section
- [x] 3.8 Keep Trending Mode output format spec updated to use emojis, full names, and legend block (pipeline delegation to mx-f-architecture-trend unchanged)
- [x] 3.9 Ensure security constraints (bash allowlist, input validation) are preserved verbatim

## 4. Scaffold Synchronization

- [x] 4.1 Copy the finalized `.opencode/agents/vibe-check-reporter.md` to `internal/scaffold/assets/agents/vibe-check-reporter.md` (byte-for-byte identical)
- [x] 4.2 Verify `vibe-check init` produces the new agent format in a temporary project directory

## 5. Verification

- [ ] 5.1 Run `/vibe-check` (summary mode) and confirm output includes 🏗️ title, 📖 legend, and guide link
- [ ] 5.2 Run `/vibe-check detailed` and confirm output includes full metric names, emoji zone indicators, 🏥 Health Scorecard, and prioritized recommendations
- [ ] 5.3 Run `/vibe-check detailed` and confirm emoji section markers appear in the agent's natural language interpretation
- [ ] 5.4 Verify no single-letter column headers (I, A, D) appear in report output
- [x] 5.5 Run CI: `go test -race -count=1 ./...` and `go vet ./...` — all SHALL pass
- [x] 5.6 Run scaffold tests: `go test -race -count=1 ./internal/scaffold/...` — all SHALL pass

## 6. Documentation Impact

- [x] 6.1 Add a "Changed" entry to CHANGELOG.md describing the revamped reporter output format (emoji contract, legend, scorecard, full metric names)
- [ ] 6.2 File a documentation issue in `unbound-force/website` for the user-facing `/vibe-check` report format change (per constitution Website Documentation Sync)
- [x] 6.3 Assess whether README.md needs updating (reference to new metrics guide, updated agent description) — no change needed; README describes the binary, not the reporter agent
- [x] 6.4 Validate AGENTS.md project structure reflects `docs/metrics-guide.md` and `.opencode/references/vibe-check-example-report.md` — new files are in `docs/` and `.opencode/references/`, both documented in project structure

<!-- spec-review: passed -->