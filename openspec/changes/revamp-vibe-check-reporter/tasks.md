## 1. Metrics Guide

- [ ] 1.1 Create `docs/metrics-guide.md` with ELI5 explanations for every metric (Instability, Abstractness, Distance from Main Sequence, LCOM4, Afferent Coupling, Efferent Coupling, Circular Dependencies, Code Duplication)
- [ ] 1.2 Each metric section includes: "What it is" (plain-language definition), "Scale" (range + extremes meaning), "Why care" (practical impact), and emoji severity ranges
- [ ] 1.3 Include at least one link to external Martin metrics reference
- [ ] 1.4 Verify the guide renders correctly on GitHub: `https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md`

## 2. Example Report Reference

- [ ] 2.1 Create `.opencode/references/vibe-check-example-report.md` with a concrete example report matching the new output format (following the gaze example-report.md pattern)
- [ ] 2.2 Example report SHALL include emoji title, metadata, legend block, per-package table with full metric names, health scorecard, prioritized recommendations, and duplication section
- [ ] 2.3 Use fictional data with clearly unrealistic values so no one copies the numbers

## 3. Agent Rewrite

- [ ] 3.1 Add emoji formatting contract section (closed vocabulary, mandatory usage rules, grade-to-emoji mapping) — modeled on gaze-reporter's `FORMATTING CONTRACT — MANDATORY, NON-NEGOTIABLE` block
- [ ] 3.2 Add legend block template to be included in every report output format spec
- [ ] 3.3 Redesign Summary Mode output format with 🏗️ title, 📖 legend, aggregate metrics, and traffic-light indicator
- [ ] 3.4 Redesign Detailed Mode output format with full metric names, emoji zone indicators, 📊 metric table, 🏥 Health Scorecard, prioritized recommendations, and 📋 duplication section
- [ ] 3.5 Add recommendation generation rules mapping JSON metric values to severity-prefixed recommendation templates
- [ ] 3.6 Add grade computation rules for each scorecard dimension (Distance, LCOM4, Cycles, Duplication, Instability Balance) per design D2
- [ ] 3.7 Add tone rules (conversational, no pedagogy, no filler, no puns, no slang) — matching gaze-reporter's tone section
- [ ] 3.8 Keep Trending Mode output format spec updated to use emojis, full names, and legend block (pipeline delegation to mx-f-architecture-trend unchanged)
- [ ] 3.9 Ensure security constraints (bash allowlist, input validation) are preserved verbatim

## 4. Scaffold Synchronization

- [ ] 4.1 Copy the finalized `.opencode/agents/vibe-check-reporter.md` to `internal/scaffold/assets/agents/vibe-check-reporter.md` (byte-for-byte identical)
- [ ] 4.2 Verify `vibe-check init` produces the new agent format in a temporary project directory

## 5. Verification

- [ ] 5.1 Run `/vibe-check` (summary mode) and confirm output includes 🏗️ title, 📖 legend, and guide link
- [ ] 5.2 Run `/vibe-check detailed` and confirm output includes full metric names, emoji zone indicators, 🏥 Health Scorecard, and prioritized recommendations
- [ ] 5.3 Run `/vibe-check detailed` and confirm emoji section markers appear in the agent's natural language interpretation
- [ ] 5.4 Verify no single-letter column headers (I, A, D) appear in report output
- [ ] 5.5 Run CI: `go test -race -count=1 ./...` and `go vet ./...` — all SHALL pass
- [ ] 5.6 Run scaffold tests: `go test -race -count=1 ./internal/scaffold/...` — all SHALL pass

## 6. Documentation Impact

- [ ] 6.1 Add a "Changed" entry to CHANGELOG.md describing the revamped reporter output format (emoji contract, legend, scorecard, full metric names)
- [ ] 6.2 File a documentation issue in `unbound-force/website` for the user-facing `/vibe-check` report format change (per constitution Website Documentation Sync)
- [ ] 6.3 Assess whether README.md needs updating (reference to new metrics guide, updated agent description)
- [ ] 6.4 Validate AGENTS.md project structure reflects `docs/metrics-guide.md` and `.opencode/references/vibe-check-example-report.md`

<!-- spec-review: passed -->