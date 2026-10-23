# Report Command Golden Fixtures

This directory holds the schema-valid `ModuleGraph` JSON fixtures and the
committed golden files that pin the `vibe-check report` command's output
byte-for-byte (tasks 3.3 and 3.4 of the `deterministic-report` OpenSpec
change). They back `TestRunReport_GoldenFiles`, `TestReportCommand_GoldenDetailed`,
`TestReportGoldenFixtures_Contract`, and `TestRunReport_GoldenJSONFiles` in
`cmd/vibe-check/report_test.go`.

The fixtures mirror the renderer fixtures in `metrics/report/render_test.go`
(`cleanMods`, `violationMods`, `detailedFixture`), so the CLI goldens and the
engine exact-output tests pin the same graphs through both surfaces. Every
fixture passes `metrics.Validate` (schema version 1.3).

## Fixtures

| Fixture              | Pins                                                                 |
|----------------------|----------------------------------------------------------------------|
| `clean.json`         | No rule fires: single 🟢 all-clear, "All Clear" status, exit 0.        |
| `violations.json`    | `pkg/monolith` fires 🔴 lcom4-split, `pkg/hub` fires 🟡 ce-decouple: "Needs Attention", exit 1 with the full report still on stdout. |
| `example.json`       | The fictional example-report fixture (the `metrics` `detailedFixture` shape): every detailed section populated — cycle, duplication, Zone of Uselessness, warning. Task 5.4 regenerates `.opencode/references/vibe-check-example-report.md` from this fixture. |
| `no-provenance.json` | The clean graph without a `provenance` object: the header falls back to the rendering binary's producer name and embedded version (the report-command spec's header fallback). |

Each fixture renders in both modes; the golden for a fixture/mode pair lives
at `<fixture>.<mode>.golden.md` (e.g. `example.detailed.golden.md`). The
`--json` output has no detail levels, so it has one golden per fixture at
`<fixture>.golden.json` (e.g. `example.golden.json`): the computed report
model — aggregates, per-dimension grades, overall status, and fired
recommendations with machine-readable `red`/`yellow`/`green` severity —
marshaled from a struct with a stable field order. The JSON model carries no
producer identity, so `no-provenance.golden.json` is byte-identical to
`clean.golden.json`; committing both keeps the fixture table uniform and
exercises the fallback automatically if the model ever gains a provenance
field.

## Determinism

`clean.json`, `violations.json`, and `example.json` pin
`provenance.producer` (`vibe-check`) and `provenance.version` (`1.2.3`), so
their golden headers are stable regardless of the version embedded in the
binary under test. `no-provenance.json` deliberately omits the object, so its
goldens carry the test binary's fallback identity (`vibe-check dev`) — if the
embedded default version ever changes, regenerate.

Reports carry no generation timestamp: the fixtures' `generatedAt`
(`2026-01-15T10:30:00Z`) must never appear in any golden.

## Regenerating the goldens

From the repository root:

```
go test ./cmd/vibe-check/ -update
```

Regeneration runs inside the test binary, so the no-provenance golden records
exactly the fallback identity that binary renders. Review the resulting diff
before committing — a golden change is an output-format change and requires
spec justification.

## Example report regeneration (task 5.4)

The committed example report is regenerated from `example.json` with:

```
go run ./cmd/vibe-check report --mode detailed cmd/vibe-check/testdata/report/example.json
```

The fixture's pinned provenance makes this reproducible regardless of the
built binary's version. The rendered bytes match
`example.detailed.golden.md`; task 5.4 wraps them with the fictional-data
disclaimer when writing `.opencode/references/vibe-check-example-report.md`.
