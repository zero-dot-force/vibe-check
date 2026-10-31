# Hand-Computed ModuleGraph Goldens — `ripwire-adapter` task 1.4

`python/golden.json` and `typescript/golden.json` in this directory are the
hand-computed, pinned acceptance targets for analyzing the two fixture
projects (`../python/`, `../typescript/`) through the ripwire adapter. Tasks
4.x/6.x implement against them; task 8.1 asserts against them; the fake
ripwire harness (task 1.3) replays the recorded inputs that produce them.

This file is the computation ledger: for every module and every value in each
golden it names the source (recorded number, fixture fact, or formula), and it
records every convention decision made where `design.md` left the choice open.
Where a convention was open, **the golden's choice is the pinned contract**.

## Shape and scope of the pin

The goldens are the exact bytes of `json.MarshalIndent(graph, "", "  ")` (the
CLI's output format, `cmd/vibe-check/analyze.go` Step 5) plus a trailing
newline, with the `provenance` object **omitted** per `--no-provenance`
semantics (provenance carries the generation timestamp; determinism tests run
with it omitted). Field names and order come from the struct JSON tags in
`metrics/graph.go`, `metrics/module.go`, `metrics/cycle.go`, and
`metrics/warning.go`; `metrics.SchemaVersionCurrent = "1.3"` (`metrics/graph.go`).

Verification performed (task 1.4 validation, no repo code written):

- Both goldens are byte-identical to the marshal of graphs constructed through
  the real `metrics` package with derived values computed by
  `metrics.ComputeInstability` / `ComputeAbstractness` / `ComputeDistance` /
  `ComputeZone` (a scratch program outside the repository, never committed).
- Both goldens pass `metrics.Validate` and validate against the embedded
  `metrics/modulegraph.schema.json`.
- No timestamps or environment-dependent values; re-marshaling is stable.

**Pinned byte-exact:** `schemaVersion`, `language`, `status`, module identity,
order, and all metric values; cycle membership and ordering; the warning code
set (which codes, and how many).

**Deliberately NOT pinned:** warning *message text* — the `Warning` JSON shape
requires a `message` string, so each golden carries the code itself as a
stable placeholder (e.g. `"message": "error-budget"`); the real adapter emits
human-readable messages (task 4.5 asserts warning text separately). Tests
compare **codes, not messages**. Likewise warning *module scoping* is not
pinned (the `module` field is omitted everywhere in the goldens). The goldens
list warnings sorted lexicographically by code as the canonical order;
implementations should emit that order to match byte-for-byte, but the pinned
contract is the code set.

## Sources

| Value family | Source |
|--------------|--------|
| Module set, Ca, Ce | `arch.txt` `<metrics>` table `<m>` rows (authoritative per D6; `<inc>` lists never read for counts — 40-cap) |
| TypeScript type counts | `arch.txt` `types=` / `abstract=` attributes (ripwire interface share) |
| Python type counts | light parse of fixture sources (D3/D8; ripwire's Python abstractness is structurally zero and MUST NOT be used) |
| Instability, Abstractness, Distance, Zone | `metrics.Compute*` (D3: never reimplemented) |
| LCOM4 | light parse of fixture sources (D8) |
| Cycles | directory-grain SCC over `deps.xml` edges (D7), corroborated by TS `<cycles>` |
| Warning codes | design.md D4/D6/D7, Operational Notes, and the change's spec deltas |

Binary provenance for the recordings: ripwire `0.6.9`, `built_from=bedc0cdcf`
(the change pins the `bedc0cd` prefix) — see `README.md` in this directory.

---

## Convention decisions pinned (where design.md left the choice open)

### C1 — Module identity and path normalization (D4)

Python uses dot-separated package style, TypeScript slash-separated style,
`Name` = last path segment. The arch table reports `./`-prefixed paths; the
adapter strips the leading `./` (and, for Python, converts `/` → `.`).

| arch.txt path | golden `path` | golden `name` |
|---------------|---------------|---------------|
| `./fixturepy` | `fixturepy` | `fixturepy` |
| `./fixturepy/cycle_a` | `fixturepy.cycle_a` | `cycle_a` |
| `./fixturepy/cycle_b` | `fixturepy.cycle_b` | `cycle_b` |
| `./fixturepy/operations` | `fixturepy.operations` | `operations` |
| `.` | `.` | `.` |
| `./src` | `src` | `src` |
| `./src/cycle-a` | `src/cycle-a` | `cycle-a` |
| `./src/cycle-b` | `src/cycle-b` | `cycle-b` |

The TypeScript crawl-root module `.` is the one open case: **path stays `.`**
(there is no `./` prefix to strip; `.` is ripwire's own identity for the crawl
root, is non-empty, and is checkout-path independent) and **`Name` = `.`**
(last `/`-segment of `.` under the uniform Name rule). Alternatives rejected:
`""` (empty identity, collides with "unset"), the crawl root's base directory
name (leaks checkout layout into module identity and breaks the
recordings' checkout-independence), and the `package.json` name (a third
identity scheme for one module). This mirrors goadapter, where
`Module.Name = pkg.Name`, the last segment of the import path.

### C2 — LCOM4 aggregation convention (the "average / max / denominator" question)

`internal/goadapter/lcom.go` uses **none** of average-over-classes, max, or a
classes-with-methods denominator. It builds **one union-find over all method
nodes in the package** and reports the number of connected components. Because
its field keys are `TypeName.FieldName`, methods of different types can never
share a field, so the package value is the **sum of per-class component
counts**. The light parse mirrors this (D8: "the same union-find approach as
`goadapter/lcom.go`"):

- **Nodes:** every class-bound method — Python `def` with leading `self` in a
  class body; TS method members in a class body. The *only* exclusion is
  module-level functions (spec: "Module-level functions without a class
  binding are excluded"). Note this is deliberately broader than goadapter's
  node collection, which filters to exported methods on exported receivers;
  the spec's node definition governs the light parse, so `_Internal.hidden`
  **is** a node.
- **Fieldless methods are singleton components.** A method that accesses no
  instance field forms its own connected component (goadapter counts every
  collected method's root; the LCOM4 literature does the same). A class with
  one fieldless method contributes 1.
- **Field keys are class-qualified** (`ClassName.attr`, the analog of
  goadapter's `TypeName.FieldName`), so methods of different classes never
  merge even when attribute names coincide. Neither fixture has cross-class
  attribute-name collisions, so the golden values are invariant to this
  choice — pinned anyway to remove the ambiguity for implementers.
- **Module LCOM = total connected components across the module's classes**
  (sum); a module with no class-bound methods is 0.

### C3 — `totalLines` = 0

`ModuleResult.TotalLines` exists solely to compute duplication percentage for
threshold enforcement, and duplication is an undeclared capability for both
ripwire languages (spec: capability gating — "Duplication detection SHALL
NOT be declared"). No task in the change assigns `totalLines` a source. The
golden pins **0** (the documented "not determined" sentinel in the
`ModuleResult` GoDoc). Alternative rejected: summing `cc.json` per-file `loc`
per directory (Python fixturepy = 1 + 94 = 95, cycle_a = 16, cycle_b = 16,
operations = 1; TS `.` = 5 + 9 = 14, src = 92, cycle-a/b = 16) — ripewire's
`loc` is a *physical* line count that includes config files (`tsconfig.json`,
`package.json` are indexed as config languages), semantics that do not match
goadapter's significant-lines counter and that the design never assigned to
this field.

### C4 — `duplications: []` (non-nil), `extensions` omitted

`ModuleResult.Duplications` GoDoc: "An empty slice (not nil) indicates no
duplications were detected." The golden pins `[]`. Note goadapter *currently*
marshals `null` for duplication-free packages (`detectDuplications` returns a
nil slice — visible in real `analyze` output); that is a pre-existing deviation
from the documented contract that the ripwire adapter must **not** replicate
(task 6.2: non-nil slices). `Extensions` is `omitempty` and the ripwire
adapters define none, so it is absent from the goldens.

### C5 — Warnings shape in the goldens

`{"code": C, "message": C}` with no `module` field, sorted lexicographically
by code. Message text and module scoping are not pinned (see "Shape and scope
of the pin" above).

### C6 — Status, ordering, non-nil slices

`status: "partial"` for both graphs (any-warning graphs are partial — spec,
error-budget requirement). Modules sorted by `path` (lexicographic; note `.`
sorts before `src` because `.` = 0x2E < `s` = 0x73). Cycle members sorted
lexicographically; the single cycle per graph needs no inter-cycle ordering.
`modules`, `cycles`, `warnings` are non-nil arrays in JSON.

---

## Python golden — value-by-value ledger

Module set = the four `<m>` rows of `arch.txt` (`modules="4"`).
`operations/build` is **absent everywhere** (ripwire hard-prunes
denylist-named directories; `deps.xml` `health files="7"` against 9 fixture
`.py` files) — that absence is exactly what the `denylist-pruned` warning
documents.

| path | ca | ce | expTypes | absTypes | I | A | D | lcom | zone |
|------|----|----|----------|----------|---|---|---|------|------|
| `fixturepy` | 0 | 0 | 4 | 1 | 0 | 0.25 | 0.75 | 6 | normal |
| `fixturepy.cycle_a` | 1 | 1 | 0 | 0 | 0.5 | 0 | 0.5 | 0 | normal |
| `fixturepy.cycle_b` | 1 | 1 | 0 | 0 | 0.5 | 0 | 0.5 | 0 | normal |
| `fixturepy.operations` | 0 | 0 | 0 | 0 | 0 | 0 | 1 | 0 | zone-of-pain |

### `fixturepy`

- **ca=0, ce=0** — `arch.txt` `<m path="./fixturepy" ca="0" ce="0">` (table
  authoritative). Note `models.py` imports `abc` (stdlib): that edge is
  out-of-corpus, so the table reports `ce=0` and `cc.json` reports
  `fan_out=0` for `models.py` ("In-corpus files this file depends on").
  D6's "count external targets in Ce only" is a no-op on this fixture — no
  recorded count includes externals, and the golden carries the table value.
- **exportedTypes=4, abstractTypes=1** — light parse of
  `src/fixturepy/models.py` (task 5.1 rules): `Cohesive`, `SplitField`,
  `AbstractThing(abc.ABC)`, `Concrete` are not underscore-prefixed →
  exported (4); `AbstractThing` has an `abc.ABC` base → abstract (1);
  `_Internal` → neither. ripwire's own `types="5" abstract="0"` is
  deliberately **not** used (D3/D8: ripwire's Python abstractness is
  structurally zero; its `types=5` counts all five classes including
  `_Internal`).
- **I** = `ComputeInstability(0,0)` = 0 (zero-denominator convention:
  isolated modules are maximally stable). Agrees with arch `I="0.00"`.
- **A** = `ComputeAbstractness(1,4)` = 0.25.
- **D** = `ComputeDistance(0.25, 0)` = |0.25 + 0 − 1| = 0.75.
- **lcom=6** — class ledger below.
- **zone** = `ComputeZone(0.25, 0, 0.75)` = normal (D ≥ 0.2; A = 0.25 is not
  < 0.2 so not zone-of-pain; I = 0 so not zone-of-uselessness).

**Deliberate divergence from arch.txt:** ripwire reports `A="0.00"
D="1.00" zone="pain"` for this module. A comes from the light parse per
D3/D8, so D and zone differ downstream. This is the one metric family where
the golden intentionally disagrees with the recorded table.

LCOM4 class ledger (`models.py`):

| Class | Methods (nodes) | `self.X` accesses | Components |
|-------|------------------|-------------------|------------|
| `Cohesive` | `set_shared`, `bump_shared`, `read_shared` | all `self.shared` | 1 |
| `SplitField` | `set_x`, `read_x` / `set_y`, `read_y` | `self.x` / `self.y` | 2 |
| `AbstractThing` | `perform` (docstring-only body) | none → singleton | 1 |
| `Concrete` | `describe` (returns a literal) | none → singleton | 1 |
| `_Internal` | `hidden` (returns a literal) | none → singleton | 1 |
| — | `module_function` | module-level → excluded | — |

Module total: 1 + 2 + 1 + 1 + 1 = **6**.

### `fixturepy.cycle_a` / `fixturepy.cycle_b`

- **ca=1, ce=1** — arch table (each package imports the other's `api`).
- **exportedTypes=0, abstractTypes=0** — no class declarations anywhere in
  either package (`__init__.py` is an import + docstring; `api.py` defines a
  module-level function). A = `ComputeAbstractness(0,0)` = 0 (the 0/0
  convention — matches goadapter's behavior for untyped packages).
- **I** = 1/(1+1) = 0.5 (agrees with arch `I="0.50"`). **D** = |0 + 0.5 − 1|
  = 0.5 (agrees with arch `D="0.50"`). **lcom** = 0 (no class-bound
  methods). **zone** = normal (A < 0.2 but I = 0.5 is not < 0.2).
- Divergence: arch `zone="n/a"` (ripwire excludes `types=0` modules from
  its zone corners). `metrics.ComputeZone` has no n/a state — untyped modules
  get A = 0 via the 0/0 convention and classify normally.

### `fixturepy.operations`

- **ca=0, ce=0, types=0** — arch table. Only content is `__init__.py`
  (docstring); `build/` is pruned. **I=0, A=0, D=1, lcom=0**.
- **zone = zone-of-pain** (A = 0 < 0.2 and I = 0 < 0.2, D = 1 ≥ 0.2). Arch
  says `zone="n/a"`; ComputeZone classifies it. Note ripwire's own
  `zone_pain="1"` aggregate counted *fixturepy* (under its A=0.00); under the
  golden's conventions the zone-of-pain module is *operations*. Different A
  sources, not a contradiction.

### Cycle

`deps.xml` file-level edges: `cycle_a/__init__.py` →
`<inc t="fixturepy.cycle_b.api">` and `cycle_b/__init__.py` →
`<inc t="fixturepy.cycle_a.api">` — each one-way at file grain. Mapped to
directories (module = containing directory of the file; the dotted import
name resolves to the `cycle_b` / `cycle_a` package directories), the
directory edge graph is `fixturepy.cycle_a ⇄ fixturepy.cycle_b`, a
strongly-connected pair. Canonical lexicographic membership:

```json
["fixturepy.cycle_a", "fixturepy.cycle_b"]
```

Python `deps.xml` carries **no `<cycles>` element** (README, "Reading the
recordings") — the two-directory cycle is the adapter's own SCC derivation
(D7 / task 4.4), which is precisely why the `cycle-approximation` warning
exists.

### Warning codes carried (and why)

| Code | Present because |
|------|-----------------|
| `cycle-approximation` | cycles present; membership approximate (D7, spec cycle requirement) |
| `denylist-pruned` | `operations/build` hard-pruned (Python-only; spec denylist requirement) |
| `directory-as-module` | every graph carries it (D4) |
| `error-budget` | Python measured budgets: 1.9% lost / 2.2% spurious edges (D6, spec error-budget requirement) |
| `name-resolved-edges` | the Python import graph is name-based (arch note "name-based deps"; `<inc>` targets are dotted module names like `fixturepy.cycle_b.api`, resolved by name to directories — D6) |
| `unsupported-metric` | duplication undeclared for the language; spec capability-gating requires a warning naming duplication as unavailable |

Absent, with reasons: `ambiguity-rate` (no same-named-module or symlink
ambiguity encountered — the R4 triggers are absent from this fixture);
`zero-coupling-guard` (the crawl yields dependency edges — the guard's
precondition "no edges while source files exist" is false);
`ts-abstractness-bias` (Python graph); `parse-error` (every fixture file is
syntactically valid).

---

## TypeScript golden — value-by-value ledger

Module set = the four `<m>` rows of `arch.txt` (`modules="4"`). The root
module `.` contains only `tsconfig.json` and `package.json` — ripwire indexes
both as config languages (they appear in the `cc.json` tree and in
`deps.xml` `health files="5"` alongside 3 dep-capable `.ts` files).

| path | ca | ce | expTypes | absTypes | I | A | D | lcom | zone |
|------|----|----|----------|----------|---|---|---|------|------|
| `.` | 0 | 0 | 0 | 0 | 0 | 0 | 1 | 0 | zone-of-pain |
| `src` | 0 | 0 | 5 | 4 | 0 | 0.8 | 0.19999999999999996 | 2 | main-sequence |
| `src/cycle-a` | 1 | 1 | 0 | 0 | 0.5 | 0 | 0.5 | 0 | normal |
| `src/cycle-b` | 1 | 1 | 0 | 0 | 0.5 | 0 | 0.5 | 0 | normal |

### `.` (crawl root)

- **ca=0, ce=0, types=0, abstract=0** — arch table. Config files only → no
  class-bound methods → **lcom=0**. **I=0, A=0 (0/0), D=1**.
- **zone = zone-of-pain** (A=0, I=0, D=1). Arch says `zone="n/a"` — same
  ComputeZone-has-no-n/a divergence as Python's untyped modules.

### `src`

- **ca=0, ce=0** — arch table (`models.ts` imports nothing).
- **exportedTypes=5, abstractTypes=4** — ripwire interface share (task 6.1):
  arch `types="5" abstract="4"`. The four interfaces (`Point`, `Named`,
  `Repository`, and the nested `Shapes.Point` — nested counts, README
  "Reading the recordings") are abstract; the denominator's fifth type is the
  class `SplitMethods`. Cross-checked against `models.ts`: 3 top-level
  interfaces + 1 nested interface + 1 class = 5 declarations.
- **I** = `ComputeInstability(0,0)` = 0 (agrees with arch `I="0.00"`).
- **A** = `ComputeAbstractness(4,5)` = 0.8 (agrees with arch `A="0.80"`).
- **D** = `ComputeDistance(0.8, 0)` = |0.8 + 0 − 1| = **0.19999999999999996**
  — an IEEE-754 artifact: 1 − 0.8 is one ULP below 0.2 in float64, and the
  result is exactly representable, so Go marshals the shortest round-trip
  decimal `0.19999999999999996`. ripwire prints `D="0.20"` — the values agree
  to display precision; the golden pins the `metrics.ComputeDistance` result
  (D3 mandates the `Compute*` functions; they are never reimplemented).
- **zone = main-sequence** — the subtlest pinned value:
  `ComputeZone(0.8, 0, d)` checks `d < 0.2` first, and
  0.19999999999999996 < 0.2 **is true in float64** (the literal `0.2` is the
  float64 0.2000000000000000111…). ripwire's `zone="ok"` (its own
  main-sequence middle) and `ComputeZone`'s `main-sequence` agree in
  substance. Verified empirically against the real `metrics` package.
- **lcom=2** — class ledger below.

LCOM4 class ledger (`src/models.ts`):

| Class | Methods (nodes) | `this.X` accesses | Components |
|-------|------------------|-------------------|------------|
| `SplitMethods` | `setX`, `readX` / `setY`, `readY` | `this.x` / `this.y` | 2 |

The property initializers `x = 0;` / `y = 0;` are not method members → not
nodes. No other module in the fixture has classes → lcom 0.

### `src/cycle-a` / `src/cycle-b`

- **ca=1, ce=1** — arch table; corroborated by `cc.json`
  (`fan_in=1, fan_out=1` for each `index.ts`).
- **exportedTypes=0, abstractTypes=0** (no type declarations) → A = 0 (0/0).
- **I = 0.5, D = 0.5** (agree with arch). **lcom = 0** (each `index.ts` is a
  const + a module-level function; no class-bound methods). **zone** =
  normal (A < 0.2, I = 0.5 not < 0.2). Arch `zone="n/a"` → same ComputeZone
  divergence as above.

### Cycle

`deps.xml` edges: `src/cycle-a/index.ts` → `<inc t="../cycle-b">` and
`src/cycle-b/index.ts` → `<inc t="../cycle-a">` (relative-path imports).
Resolved relative to the importing file's directory, both map to the
directory pair `src/cycle-a ⇄ src/cycle-b` — a strongly-connected pair.
Canonical lexicographic membership:

```json
["src/cycle-a", "src/cycle-b"]
```

Corroborated by `deps.xml`'s own `<cycles><cycle size="2">` element (file
members `src/cycle-b/index.ts`, `src/cycle-a/index.ts`) — the adapter's
directory-grain SCC matches the listing, exactly the D7 corroboration path.

### Warning codes carried (and why)

| Code | Present because |
|------|-----------------|
| `cycle-approximation` | cycles present (D7, spec cycle requirement) |
| `directory-as-module` | every graph carries it (D4) |
| `error-budget` | TypeScript measured budget: 0.16% unresolved imports (D6, spec error-budget requirement) |
| `ts-abstractness-bias` | `abstract class` undercount (conservative) + type-alias denominator inflation (spec TS-abstractness requirement) |
| `unsupported-metric` | duplication undeclared (spec capability-gating requirement) |

Absent, with reasons: `denylist-pruned` (no denylist-named directories under
the TS crawl root); `name-resolved-edges` (TS fixture imports are relative
paths — `<inc t="../cycle-b">` — not name-resolved; D6 fires this code only
when name resolution is actually used); `ambiguity-rate` (no ambiguity
encountered); `zero-coupling-guard` (edges exist); `parse-error` (all files
parse clean).

---

## Cross-check summary: golden vs `arch.txt`

Ca, Ce, and I agree with `arch.txt` for **all eight modules** (no
divergences). A agrees everywhere except the one deliberate Python
light-parse substitution. D and zone diverge only downstream of A or of
ComputeZone's boundary semantics:

| Module | Field | arch.txt | golden | Reason |
|--------|-------|----------|--------|--------|
| py `fixturepy` | A | 0.00 | 0.25 | light parse per D3/D8 (ripwire Python A structurally zero) |
| py `fixturepy` | D | 1.00 | 0.75 | derives from light-parse A |
| py `fixturepy` | zone | pain | normal | derives from light-parse A |
| py `cycle_a`, `cycle_b` | zone | n/a | normal | ComputeZone has no n/a; untyped → A=0 via 0/0 |
| py `operations` | zone | n/a | zone-of-pain | A=0, I=0, D=1 under ComputeZone |
| ts `.` | zone | n/a | zone-of-pain | same |
| ts `src` | D | 0.20 | 0.19999999999999996 | IEEE-754: 1−0.8 is one ULP below 0.2 |
| ts `src` | zone | ok | main-sequence | d < 0.2 holds in float64 |
| ts `cycle-a`, `cycle-b` | zone | n/a | normal | same as py cycles |

ripwire's `zone_pain`/`zone_useless`/`zone_ok`/`zone_na` aggregates are its
own partition (untyped modules excluded); `metrics.ComputeZone` is the
governed classifier and the goldens follow it exclusively.

## Reuse

When ripwire's validated version moves (re-capture procedure in `README.md`),
re-record the eight inputs, re-run this ledger against the new tables, and
regenerate both goldens — the conventions pinned here (C1–C6) are expected to
remain stable; only the recorded numbers and fixture-derived counts should
change.
