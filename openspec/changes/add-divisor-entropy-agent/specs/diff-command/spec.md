## ADDED Requirements

### Requirement: Structural Delta Computation

The command SHALL compute the exact per-package metric delta between two `ModuleGraph` documents, reading a base graph and a PR graph and reporting the change in every Martin metric per package.

`vibe-check diff <base.json> <pr.json>` reads two `ModuleGraph` JSON documents produced by `vibe-check analyze` and computes, for each package present in either graph, the change in afferent coupling (Ca), efferent coupling (Ce), instability (I), abstractness (A), distance from the main sequence (D), and LCOM. Packages are matched by import path. Packages present only in the PR graph are reported as added; packages present only in the base graph are reported as removed. Added and removed packages are reported for informational purposes ONLY and MUST NOT trigger any delta threshold gate — their absolute quality is enforced by `vibe-check analyze --max-distance` and related flags, not by the entropy divisor's regression tool. Computation is a pure function of the two inputs and MUST be deterministic.

#### Scenario: Delta computed for a changed package

- **WHEN** `vibe-check diff base.json pr.json` runs and a package's instability rose from 0.40 in the base graph to 0.55 in the PR graph
- **THEN** the reported delta for that package records an instability change of +0.15 alongside its Ca, Ce, abstractness, distance, and LCOM changes

#### Scenario: Added and removed packages are classified without triggering gates

- **WHEN** the PR graph contains a package absent from the base graph and the base graph contains a package absent from the PR graph
- **THEN** the first package is reported as added and the second is reported as removed, without a spurious metric delta being fabricated for either, and neither package triggers any REQUEST_CHANGES or COMMENT verdict gate

#### Scenario: Delta computation is deterministic

- **WHEN** `vibe-check diff` is run twice on the same pair of input graphs
- **THEN** both runs produce byte-identical output, including the ordering of per-package rows

### Requirement: New Circular Dependency Detection

The command SHALL classify each cycle as new, pre-existing, or resolved, and MUST NOT report a cycle that exists in both graphs as newly introduced.

A cycle present in the PR graph but absent from the base graph is classified as a new cycle. A cycle present in both graphs is pre-existing and MUST NOT be reported as new. A cycle present in the base graph but absent from the PR graph is classified as resolved. Cycle identity is compared by the set of member packages, independent of the starting node or listed rotation.

#### Scenario: New cycle is detected

- **WHEN** the PR graph contains a cycle among packages `{a, b, c}` that does not appear in the base graph
- **THEN** the command reports `{a, b, c}` in its list of new cycles

#### Scenario: Pre-existing cycle is not reported as new

- **WHEN** an identical cycle among `{a, b}` appears in both the base graph and the PR graph
- **THEN** the command does not report `{a, b}` as a new cycle

#### Scenario: Resolved cycle is recognized

- **WHEN** a cycle among `{a, b}` appears in the base graph but is absent from the PR graph
- **THEN** the command reports `{a, b}` as resolved and does not report it as new

### Requirement: Deterministic Verdict

The command SHALL render exactly one verdict — APPROVE, COMMENT, or REQUEST_CHANGES — by applying documented threshold gates to the computed delta, and the same inputs MUST always yield the same verdict.

The verdict is the output of a pure function over the computed delta and a `VerdictThresholds` value. The command MUST return REQUEST_CHANGES when the PR introduces any new cycle, when any package's instability increase is greater than or equal to the instability threshold (default 0.15), when any package's distance increase is greater than or equal to the distance threshold (default 0.20), or when any package's LCOM increase is greater than or equal to the LCOM threshold (default 2). The command MUST return COMMENT when metric shifts are non-zero but remain below every REQUEST_CHANGES threshold. The command MUST return APPROVE when metrics improve or remain stable. The default threshold values are protected quality gates. The verdict MUST be accompanied by machine-readable reasons naming each gate that fired and the package and metric value responsible.

**Float boundary**: Each float delta (ΔInstability, ΔAbstractness, ΔDistance) MUST be rounded to 4 decimal places (precision 1e-4) using round-half-away-from-zero (`math.Round(x*1e4)/1e4`) before the `≥` comparison is applied. A delta of 0.14999 rounds to 0.1500 and therefore triggers REQUEST_CHANGES at the 0.15 threshold; a delta of 0.14994 rounds to 0.1499 and yields COMMENT. This rule is part of the `DecideVerdict` GoDoc contract and MUST be covered by boundary fixture tests (see tasks.md group 2.4).

**Threshold direction**: `vibe-check diff` uses inclusive `≥` for its threshold comparisons whereas `vibe-check analyze` uses strict `>` for its `--max-*` flags. This asymmetry is intentional: a regression of exactly the threshold amount is already too much for entropy enforcement, so it must trigger REQUEST_CHANGES. This distinction MUST be documented in the `DecideVerdict` GoDoc.

**Nil/empty contract**: `ComputeDelta` and `DecideVerdict` MUST handle nil or empty inputs without panicking. If either input graph is nil, `ComputeDelta` returns a `GraphDelta` with the unreliable/partial-measurement flag set — so `DecideVerdict` yields COMMENT, never a false APPROVE — while keeping its error-less signature. Two empty-but-non-nil graphs compare equal, and `DecideVerdict` on a zero-value `GraphDelta` returns APPROVE. These contracts MUST be covered by table tests (nil base → unreliable → COMMENT, nil pr → unreliable → COMMENT, empty-vs-empty → APPROVE, zero `GraphDelta` → APPROVE).

**Tighten-only overrides**: Threshold-override flags (`--max-instability-delta`, `--max-distance-delta`, `--max-lcom-delta`) are TIGHTEN-ONLY. Any override value looser than the protected default (e.g. a larger max-instability-delta than 0.15) MUST be rejected with exit 2 and a diagnostic to standard error. This prevents callers from weakening protected quality gates.

**Entropy direction**: The `GraphDelta` payload MUST carry an `EntropyDirection` field computed deterministically as: `degrading` when any REQUEST_CHANGES trigger fired; `improving` when at least one metric delta is negative AND no metric delta is positive above the 1e-4 rounding noise AND no trigger fired; `stable` otherwise (no material change, or offsetting improvements and regressions that fire no gate). Because both `improving` and `stable` map to APPROVE, this field is reporting-only and never affects the verdict; `design.md` documents the identical rule.

**Added/removed packages in thresholds**: Packages present in only one graph (added or removed) MUST NOT produce a spurious metric delta and MUST NOT contribute to any threshold gate evaluation.

#### Scenario: Improvement or stability yields APPROVE

- **WHEN** every package's instability, distance, and LCOM are unchanged or reduced and no new cycle is introduced
- **THEN** the verdict is APPROVE

#### Scenario: New cycle forces REQUEST_CHANGES

- **WHEN** the PR introduces a new cycle even though all numeric metrics are otherwise stable
- **THEN** the verdict is REQUEST_CHANGES and the reasons name the new cycle

#### Scenario: Instability increase at the threshold triggers REQUEST_CHANGES

- **WHEN** a package's instability increases by exactly 0.15 (the default threshold, which rounds to 0.1500 at 4 decimal places)
- **THEN** the verdict is REQUEST_CHANGES and the reasons name that package and its instability delta

#### Scenario: Instability increase just below the threshold yields COMMENT

- **WHEN** the largest instability increase after rounding to 4 decimal places is 0.1499, no distance increase reaches 0.20, no LCOM increase reaches 2, and no new cycle is introduced
- **THEN** the verdict is COMMENT

#### Scenario: Distance increase boundary is inclusive

- **WHEN** a package's distance from the main sequence increases by exactly 0.20 while a separate run increases it by only 0.19
- **THEN** the 0.20 case yields REQUEST_CHANGES and the 0.19 case yields COMMENT

#### Scenario: LCOM increase boundary is inclusive

- **WHEN** a package's LCOM increases by exactly 2 while a separate run increases it by only 1
- **THEN** the 2 case yields REQUEST_CHANGES and the 1 case yields COMMENT

#### Scenario: Nil or empty input produces a safe result

- **WHEN** `ComputeDelta` is called with a nil base or nil PR argument, or `DecideVerdict` is called with a zero-value `GraphDelta`
- **THEN** the function returns a documented empty result or APPROVE respectively and does not panic

#### Scenario: Tighten-only override is enforced

- **WHEN** a caller passes a threshold-override flag value that is looser than the protected default (e.g. `--max-instability-delta 0.30`)
- **THEN** the command exits with code 2 and writes a diagnostic to standard error without emitting any verdict payload

#### Scenario: Entropy direction is degrading when a trigger fires

- **WHEN** the computed delta contains at least one REQUEST_CHANGES-triggering condition
- **THEN** the `EntropyDirection` field in the output payload is `degrading`

#### Scenario: Entropy direction is improving when metrics net-improve

- **WHEN** all deltas are negative or zero and no trigger fires
- **THEN** the `EntropyDirection` field in the output payload is `improving`

### Requirement: Partial-Build Degraded Verdict

The command SHALL inspect each input `ModuleGraph`'s `Warnings` slice and `Status` field before computing a verdict. MUST handle unreliable inputs safely.

When either input graph has `Status != "complete"` or carries load-error `Warnings`, the measurement is considered unreliable. In this case `RunDiff` MUST return COMMENT (never APPROVE), MUST annotate the affected packages in the output, and MUST suppress added/removed-driven signal. The `GraphDelta` payload MUST carry a boolean partial/unreliable-measurement flag indicating this condition.

#### Scenario: Partial-build input yields a degraded verdict

- **WHEN** either the base or PR `ModuleGraph` has `Status` equal to `"partial"` or contains load-error `Warnings`
- **THEN** the command returns COMMENT (not APPROVE), annotates the affected packages, and sets the unreliable-measurement flag in the output payload

#### Scenario: Both inputs complete yields a confident verdict

- **WHEN** both the base and PR `ModuleGraph` have `Status == "complete"` and no load-error warnings
- **THEN** the command computes a full verdict normally without suppressing signal

### Requirement: Machine-Readable and Human-Readable Output

The command SHALL emit a deterministic machine-readable JSON payload under `--json` and a human-readable table by default, both reporting the per-package delta, the new and resolved cycles, the entropy direction, the verdict, and the verdict reasons.

Under `--json`, the command writes a single JSON object to standard output containing the per-package delta rows, the new-cycle and resolved-cycle lists, the overall entropy direction (improving, stable, or degrading), the verdict, the reasons, and the unreliable-measurement flag. Without `--json`, the command writes an equivalent human-readable table. All collections MUST be emitted in a stable, sorted order so that output is byte-stable across runs and platforms.

Note: Like the underlying `analyze` command, the `diff --json` payload does not include provenance metadata (producer, version, timestamp, compared SHAs) in this change. This is a known gap tracked as a follow-up; the compared SHAs are reported in the human-readable output and the agent's narrative, not in the machine payload.

#### Scenario: JSON payload carries verdict, reasons, and deltas

- **WHEN** `vibe-check diff --json base.json pr.json` runs
- **THEN** the emitted JSON object contains the verdict, the list of reasons, the per-package delta rows, the new-cycle list, the entropy direction, and the unreliable-measurement flag, with arrays in sorted order

#### Scenario: Human-readable table is the default

- **WHEN** `vibe-check diff base.json pr.json` runs without `--json`
- **THEN** a readable per-package delta table containing per-package rows and the verdict are written to standard output

### Requirement: Input Validation and Exit Codes

The command SHALL validate both input documents before computing anything and MUST exit with a non-zero code when an input cannot be read or fails schema validation, reporting success through the payload verdict rather than through the exit code. No verdict payload MUST be emitted when exiting with code 2.

Each input path is read and validated against the `ModuleGraph` schema before any delta is computed. When either input is missing or unreadable, the command writes a diagnostic to standard error and exits with code 2. When either input is readable but fails schema validation, the command writes a diagnostic to standard error and exits with code 2. Both missing/unreadable and readable-but-invalid cases produce exit code 2 and no payload. When both inputs are valid, the command exits 0 regardless of the verdict; the verdict (including REQUEST_CHANGES) is conveyed in the output payload, not encoded in the exit code, so that a caller can distinguish a tool failure from a legitimate REQUEST_CHANGES verdict.

#### Scenario: Missing or unreadable input exits non-zero with no payload

- **WHEN** either input file path does not exist or cannot be read
- **THEN** the command writes a diagnostic to standard error, exits with code 2, and emits no verdict payload

#### Scenario: Readable but schema-invalid input exits non-zero with no payload

- **WHEN** either input file is readable but contains JSON that fails `ModuleGraph` schema validation
- **THEN** the command writes a diagnostic to standard error, exits with code 2, and emits no verdict payload

#### Scenario: Valid inputs exit zero even for REQUEST_CHANGES

- **WHEN** both inputs are valid and the computed verdict is REQUEST_CHANGES
- **THEN** the command exits 0 and conveys REQUEST_CHANGES in the output payload
