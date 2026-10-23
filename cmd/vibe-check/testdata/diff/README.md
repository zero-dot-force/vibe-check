# Diff Command Golden Fixtures

This directory holds the schema-valid base/PR `ModuleGraph` JSON fixture
pairs and the committed golden files that pin the `vibe-check diff`
command's `--markdown` and default-table outputs byte-for-byte (tasks 4.1
and 4.2 of the `deterministic-report` OpenSpec change). They back
`TestRunDiff_MarkdownGoldenFiles`, `TestRunDiff_DefaultTableGolden`,
`TestRunDiff_MarkdownSpecScenarios`, `TestRunDiff_MarkdownGateExitCodes`,
`TestDiffCommand_MarkdownGateExitsOne`, and `TestDiffGoldenFixtures_Contract`
in `cmd/vibe-check/diff_test.go`.

The pairs mirror the in-code fixture builders in `diff_test.go`
(`improvementFixtures`, `degradeCycleFixtures`, `commentBandFixtures`), so
the committed goldens and the in-code scenario tests pin the same graphs
through both surfaces. Every fixture passes `metrics.Validate` (schema
version 1.3).

## Fixture pairs

| Pair        | Pins                                                                                                                          |
|-------------|-------------------------------------------------------------------------------------------------------------------------------|
| `improved`  | `ex/a` improves on every gated metric and the base cycle is resolved: APPROVE, direction improving, no reasons, one resolved cycle (the spec's approve-verdict scenario). |
| `declined`  | `ex/a` crosses all three metric gates (ΔInstability +0.30, ΔDistance +0.25, ΔLCOM +2) and a new `ex/b`→`ex/c` cycle appears, plus one added and one removed package: REQUEST_CHANGES, direction degrading, four reasons, one new cycle (the spec's request-changes-verdict scenario). |
| `mixed`     | `ex/a` instability rises by 0.05, below every gate: COMMENT, direction stable, the single materiality reason.                   |

Each pair lives at `<name>-base.json` / `<name>-pr.json`. The `--markdown`
golden for a pair lives at `<name>.golden.md`; the default-table golden
(the spec's default-output-unchanged requirement, pinned on the richest
fixture) lives at `declined.golden.table.txt`.

## Determinism

The diff markdown renderer derives every byte from the two input files and
renders no timestamps and no provenance identity, so the fixtures carry no
provenance discipline: any schema-valid pair renders the same golden
regardless of who produced it or when. Every test run compares against the
committed golden bytes, so each run doubles as a repeated-runs
byte-stability check across machines and time.

## Regenerating the goldens

From the repository root:

```
go test ./cmd/vibe-check/ -update
```

The same single-writer flag regenerates the report goldens under
`../report/`. Review the resulting diff before committing — a golden change
is an output-format change and requires spec justification.
