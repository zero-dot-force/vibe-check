## Context

The vibe-check-reporter agent currently mirrors the raw JSON structure of `vibe-check analyze` output — a table with single-letter column headers (I, A, D, Ca, Ce), plain-text zone labels, and no user education. The gaze-reporter agent (`.opencode/agents/gaze-reporter.md`) has a proven, successful output contract: mandatory emoji section markers, closed emoji vocabulary, grade scorecards, prioritized recommendations, and a conversational tone. This design adapts that contract for architectural metrics.

The agent runs in three modes (summary, detailed, trending) and its output is delivered as markdown in the CLI. No web rendering is involved — terminal emoji support is standard in modern environments.

## Goals / Non-Goals

**Goals:**
- Every report includes a legend block explaining each metric column in plain language
- Every report links to `docs/metrics-guide.md` for ELI5 explanations
- Tables use full metric names, not single letters
- Zones are visualized with severity emojis
- Detailed mode includes a Health Scorecard with letter grades
- Detailed mode includes prioritized, actionable recommendations
- Duplication blocks from JSON output are surfaced in the report
- Emoji usage follows a strict closed-set vocabulary matching gaze-reporter's contract pattern

**Non-Goals:**
- Changing the `vibe-check analyze` CLI output format (JSON schema unchanged)
- Adding new metrics or computation logic
- Changing the trending mode's data pipeline (trend analysis delegation to `mx-f-architecture-trend` unchanged)
- Adding interactive or web-based reporting
- Changing the `vibe-check` command definition in `internal/scaffold/assets/commands/vibe-check.md` (it delegates to the same agent)

## Decisions

### D1: Emoji vocabulary follows gaze-reporter's pattern exactly

The gaze-reporter uses a closed set of 10 emojis with a mandatory formatting contract that prohibits any other emojis. The vibe-check-reporter adopts this same contract with architecture-themed emojis but the same severity markers (🟢🟡🔴) and warning emoji (⚠️).

**Closed set**: 🏗️ 📊 🔗 🧩 🔄 📋 📖 🏥 🟢 🟡 🔴 ⚠️

**Rationale**: Using the same grade-to-emoji mapping (🟢 B+↑, 🟡 B–C, 🔴 C-↓) and same tone rules creates consistency across the Unbound Force tool suite. Developers familiar with `/gaze` immediately understand `/vibe-check` reports.

**Alternatives considered**: A completely independent vocabulary — rejected because it creates cognitive load switching between tools.

### D2: Grade thresholds derived from Martin metrics theory

| Dimension | A | B+ | B | C+ | C | F |
|-----------|---|---|---|---|---|---|
| Avg Distance | D < 0.1 | D < 0.2 | D < 0.3 | D < 0.5 | D < 0.7 | D ≥ 0.7 |
| Max LCOM4 | = 1 | = 2 | = 3 | = 4 | = 5 | ≥ 6 |
| Cycles | 0 | — | — | — | — | ≥ 1 |
| Duplication % | 0% | < 3% | < 5% | < 10% | < 15% | ≥ 15% |
| Instability Spread | None at extremes | — | ≤ 1 extreme | ≤ 2 extremes | — | ≥ 3 extremes |

**Rationale**: Distance thresholds align with Martin's original zone definitions (Main Sequence < 0.3). LCOM4 thresholds follow Hitz & Montazeri (LCOM4=1 cohesive, 2-3 slightly fragmented, 4+ low cohesion). Cycles and duplication are binary/monotonic — any non-zero value indicates a problem. Instability Spread penalizes having too many packages at either I=0.0 or I=1.0. "Extremes" are defined as packages where I ≤ 0.001 or I ≥ 0.999 (floating-point tolerance). Grades B+, B, and C+ are unreachable for Instability Spread because only the extreme-count thresholds apply.

### D3: Metrics guide lives at `docs/metrics-guide.md`

The guide is a standalone markdown file in the repo root's `docs/` directory. This makes it:
- Accessible via GitHub's raw content at `https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md`
- Version-controlled with the codebase
- Discoverable via GitHub's file browser
- Editable without touching agent configuration

**Alternatives considered**: Inline in the agent prompt — rejected because it bloats the agent file and can't be linked externally. In `.opencode/references/` — rejected because the user asked for an HTTP link to the repo.

### D4: Legend block is always present in reports

Every report (summary and detailed) includes a one-line legend after the metadata:

```
📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md
```

**Rationale**: First-time users won't know which report mode to run. The legend makes any mode self-documenting. The full guide link gives deeper education without cluttering the report.

### D5: Recommendations are generated from metric data, not hardcoded

Each recommendation is derived from the JSON output values. The agent maps metric thresholds to recommendation templates:

| Condition | Severity | Template |
|-----------|----------|----------|
| LCOM4 ≥ 4 | 🔴 | Split `<pkg>` — LCOM4 of `<val>` suggests multiple unrelated responsibilities |
| D ≥ 0.5 | 🔴 | Refactor `<pkg>` — distance of `<val>` from main sequence places it in `<zone>` |
| I ≥ 0.95 | 🟡 | Isolate `<pkg>` — instability of `<val>` means it depends on everything but nothing depends on it; appropriate for CLI layers |
| I ≤ 0.05 | 🟡 | Protect `<pkg>` — instability of `<val>` makes it maximally stable; changes ripple widely |
| Ce > 20 | 🟡 | Decouple `<pkg>` — `<val>` efferent couplings; consider interface extraction |
| A = 0, ExportedTypes > 10 | 🟡 | Add interfaces to `<pkg>` — 0 abstractness with `<val>` exported types leaves no room for abstraction |
| Duplication > 0 | 🟢 | Eliminate duplicate blocks in `<file>` — `<count>` blocks with `<pct>`% similarity |
| Cycles > 0 | 🔴 | Break cycle between `<packages>` — circular dependencies prevent independent testing |

**Rationale**: Template-driven recommendations ensure consistency and prevent the LLM from hallucinating irrelevant advice. Thresholds match the CI gate values in `vibe-check analyze` flags.

### D6: Agent scaffolds are synced manually

The `internal/scaffold/assets/agents/vibe-check-reporter.md` file must be kept in sync with `.opencode/agents/vibe-check-reporter.md`. This is a copy-paste operation, not automated generation. The scaffold is the source of truth for `vibe-check init` deployments.

## Risks / Trade-offs

- **Risk**: Long legend line may wrap on narrow terminals → **Mitigation**: Legend uses abbreviations where necessary (e.g., "Abstractness" not "Abstractness from main sequence") and the full guide is one click away. The legend is informational, not load-bearing — reports are readable without it.
- **Risk**: Emoji rendering on older terminals (pre-2018) may show tofu characters → **Mitigation**: Emojis are Unicode standard characters. The report is still readable without them (section names in text provide fallback meaning). Modern macOS, Linux, and Windows Terminal all support emoji rendering.
- **Risk**: Grade thresholds may need tuning based on real-world feedback → **Mitigation**: Thresholds are defined in the agent file as documented constants. Easy to adjust in follow-up changes.
- **Trade-off**: Adding emojis and explanatory text increases report length ~2x. Accepted trade-off for usability — developers can skim by section headers. Summary mode remains concise for quick scans.

## Open Questions

- Should the legend be configurable (on/off)? Decision: No. Keep the output format simple and predictable. Users who don't need the legend can skip it visually.
- Should grades use numeric scores (0-100) instead of letter grades? Decision: Letter grades with emojis. This matches gaze-reporter and is more immediately meaningful than numeric scores without context.