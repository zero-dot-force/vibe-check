## Why

The current vibe-check reporter produces a dry, clinical table of single-letter metric headers (I, A, D, Ca, Ce) that requires prior knowledge of Martin metrics. Developers unfamiliar with design-quality metrics cannot interpret the report without external research. The gaze-reporter agent demonstrates that emoji-structured, conversational reports with grade scorecards and prioritized recommendations make quality data immediately actionable. Applying this pattern to architectural metrics closes a usability gap and makes vibe-check self-documenting.

## What Changes

- **NEW** `docs/metrics-guide.md` — Plain-language ELI5 guide explaining every metric (Instability, Abstractness, Distance, LCOM4, Ca, Ce, Cycles, Duplication) with emoji severity ranges, concrete examples, and why the developer should care. Linked from every report header.
- **NEW** `.opencode/references/vibe-check-example-report.md` — Definitive formatting reference matching the gaze-reporter example report pattern.
- **MODIFIED** `.opencode/agents/vibe-check-reporter.md` — Complete rewrite of the agent's output format contract:
  - **Emoji vocabulary**: Closed set of 12 emojis (🏗️📊🔗🧩🔄📋📖🏥🟢🟡🔴⚠️) with mandatory usage rules, mirroring gaze's formatting contract.
  - **Full metric names**: Tables use `Instability`, `Abstractness`, `Distance`, `LCOM4` instead of `I`, `A`, `D`.
  - **Legend block**: Every report includes a one-line legend spelling out each column header and linking to the metrics guide.
  - **Health Scorecard**: Grade each architectural dimension (Avg Distance, Max LCOM4, Cycles, Duplication, Instability Spread) with letter grades and severity emojis.
  - **Prioritized Recommendations**: Numbered list prefixed with 🔴/🟡/🟢, starting with action verbs, naming specific packages and concrete metrics.
  - **Zone visualization**: Emoji-prefixed zone labels (🟢 Main Sequence, 🟡 Balanced, 🔴 Zone of Pain, 🔴 Zone of Uselessness).
  - **Duplication section**: Dedicated 📋 section surfacing code duplication blocks from the JSON output.
  - **Conversational tone**: Data-driven, actionable, no pedagogical explanations — matching gaze's tone rules.
- **MODIFIED** `internal/scaffold/assets/agents/vibe-check-reporter.md` — Synchronized copy for `vibe-check init`.

## Capabilities

### New Capabilities

- `vibe-check-metrics-guide`: A plain-language metrics reference document at `docs/metrics-guide.md` that explains each architectural metric (Instability, Abstractness, Distance from Main Sequence, LCOM4, Afferent/Efferent Coupling, Circular Dependencies, Duplication) with ELI5 examples, emoji severity ranges, and actionable takeaways. Linked from every report.
- `revamped-reporter-agent`: A redesigned vibe-check-reporter agent with an emoji-structured formatting contract, full-word column headers, legend block, health scorecard with letter grades, prioritized recommendations, and conversational tone — modeled after the gaze-reporter agent's proven output contract.

### Modified Capabilities

<!-- None — no existing specs to modify -->

## Impact

- **Files changed**: `docs/metrics-guide.md` (new), `.opencode/references/vibe-check-example-report.md` (new), `.opencode/agents/vibe-check-reporter.md` (rewrite), `internal/scaffold/assets/agents/vibe-check-reporter.md` (sync), `internal/scaffold/assets/commands/vibe-check.md` (unchanged, delegates to same agent)
- **No API changes**: The `vibe-check analyze` CLI output format is unchanged. Only the agent's interpretation/reporting layer changes.
- **No dependency changes**: Emoji rendering relies on terminal support (standard in modern terminals). Markdown tables are GitHub-compatible.
- **User-facing**: All `/vibe-check` command users will see the new report format. No breaking changes to the command interface.