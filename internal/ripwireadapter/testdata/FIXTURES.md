# Ripwire Adapter Test Fixtures

Static-analysis fixtures for the ripwire adapter — task 1.1 of the
`ripwire-adapter` OpenSpec change (`openspec/changes/archive/2026-11-24-ripwire-adapter/`;
rationale in design.md D8/D12).

Properties shared by every fixture file:

- **Deterministic** — hand-authored content only; no timestamps, dates,
  or environment-dependent values anywhere in the tree.
- **Static, not runtime** — runtime importability is not required. The
  fixtures exercise static import-graph and AST analysis. All Python is
  syntactically valid Python 3; all TypeScript is syntactically valid
  TypeScript.
- **Source only** — no build artifacts (`__pycache__`, `node_modules`,
  emitted output) belong in this tree.

## Python fixture: `python/` (src layout)

| File | Scenario exercised |
|------|--------------------|
| `python/pyproject.toml` | Minimal project identity (`fixturepy`) for the src-layout project. |
| `python/src/fixturepy/__init__.py` | Package root marker for the src layout. |
| `python/src/fixturepy/models.py` | Classification and LCOM scenarios — see table below. |
| `python/src/fixturepy/operations/__init__.py` | Regular subpackage marker. |
| `python/src/fixturepy/operations/build/__init__.py` | Real Python package inside a directory named `build` (denylist name). git tracks it, so the pinned release indexes it in the `python` recordings (an isolated module); the `python-mixed` recordings, captured outside git, show it pruned. |
| `python/src/fixturepy/operations/build/steps.py` | Source file inside the denylist-named directory; the `denylist-pruned` warning's concrete target when it is pruned (non-git capture), and the pre-check's indexed-candidate case when it is not. |
| `python/src/fixturepy/cycle_a/__init__.py` | Cycle side A: `from fixturepy.cycle_b.api import helper_b` (import edge A → B). |
| `python/src/fixturepy/cycle_a/api.py` | Defines `helper_a`, imported by cycle side B. |
| `python/src/fixturepy/cycle_b/__init__.py` | Cycle side B: `from fixturepy.cycle_a.api import helper_a` (import edge B → A). |
| `python/src/fixturepy/cycle_b/api.py` | Defines `helper_b`, imported by cycle side A. |
| `python/src/fixturepy/wide/__init__.py` | Package marker for the wide-form import scenario. |
| `python/src/fixturepy/wide/consumer.py` | `from fixturepy import cycle_b` — wide-form from-import of a **subpackage** (ripwire#478 gap 1): resolves to an edge into `fixturepy/cycle_b`, not only `fixturepy/__init__.py`. |
| `python/src/fixturepy/ring_a/__init__.py` | Ring side A: `from fixturepy import ring_b` — wide-form import of a subpackage; `--deps` records only `fixturepy`, so the A → B edge comes from the member-import light parse (openspec/changes/archive/2026-11-24-wide-form-cycle-edges). |
| `python/src/fixturepy/ring_a/api.py` | Defines `marker`, imported by ring side B. |
| `python/src/fixturepy/ring_b/__init__.py` | Ring side B: `from fixturepy.ring_a.api import marker` (name-resolved edge B → A). Its docstring quotes the ring A import, exercising the light parse's string blanking. |
| `python/src/fixturepy/rel_a/__init__.py` | Relative side A: `from ..rel_b.api import token` — a relative **base** that `--deps` records verbatim as `..rel_b.api`; resolved against the importing package to the edge A → B (openspec/changes/archive/2026-11-24-relative-import-edges). |
| `python/src/fixturepy/rel_b/__init__.py` | Relative side B: `from .. import rel_a` — a relative **member** of the bare base `..`; `--deps` records only `..`, so the B → A edge comes from the member-import light parse. |
| `python/src/fixturepy/rel_b/api.py` | Defines `token`, imported by relative side A. |
| `python/src/fixturepy/alias/__init__.py` | Package marker for the aliased-import scenario. |
| `python/src/fixturepy/alias/consumer.py` | `import fixturepy.cycle_a.api as ca` — aliased import (ripwire#478 gap 2): resolves to an edge into `fixturepy/cycle_a`. |
| `python/src/fixturepy/collide/__init__.py` | Package marker for the stdlib-name collision scenario. |
| `python/src/fixturepy/collide/consumer.py` | `import logging` inside a package (ripwire#478 gap 4): must bind the stdlib module, drawing no edge. |
| `python/src/fixturepy/collide/logging.py` | Sibling whose name collides with the stdlib `logging` module; never imported. |

### Declarations in `python/src/fixturepy/models.py`

| Declaration | Scenario exercised |
|-------------|---------------------|
| `Cohesive` | Three methods all reading/writing `self.shared` → LCOM 1. |
| `SplitField` | Four methods: two touch only `self.x`, two touch only `self.y` → LCOM 2. |
| `AbstractThing` | `abc.ABC` subclass with one `@abstractmethod` → abstract classification. |
| `Concrete` | Plain concrete class: no abstract marker, no underscore prefix. |
| `_Internal` | Underscore-prefixed class → private-name classification. |
| `module_function` | Module-level function → excluded from class LCOM analysis. |

`cycle_a` and `cycle_b` form a two-directory import cycle. Neither
package is meant to be imported at runtime.

`ring_a` and `ring_b` form a two-directory cycle that closes **only** through
a wide-form import: name resolution of `--deps` text alone sees `ring_a →
fixturepy` and `ring_b → ring_a`, which is acyclic. At file grain the ring is
acyclic too (`ring_a/__init__.py` → `ring_b/__init__.py` → `ring_a/api.py`), so
ripwire's own `<cycles>` listing does not report it and detection must come
from the adapter's derivation. The `ring_` prefix keeps the directories out of
the `cycle_a`/`cycle_b` layers in `recorded/arch-rules.txt`.

`rel_a` and `rel_b` form a two-directory cycle that closes **only** through
relative imports: before relative resolution both targets (`..rel_b.api`,
`..`) were classified external and neither edge existed. At file grain the
pair is acyclic (`rel_b/__init__.py` → `rel_a/__init__.py` → `rel_b/api.py`),
so ripwire's `<cycles>` listing does not report it. The `rel_` prefix matches
no layer substring in `recorded/arch-rules.txt`.

`wide`, `alias`, and `collide` each sit in their own directory module so each
upstream import fix (openspec/changes/archive/2026-11-25-bump-ripwire-pin design D7) shows
independently at directory grain. Gap 1 imports a subpackage on purpose: at
directory grain `pkg/__init__.py` and `pkg/mod.py` are one module, so a
from-import of a plain module file would not show the fix.

### Mixed-language variant (recordings only)

`recorded/python-mixed/` is captured from a temporary copy of `python/` with
one Go file added at `src/fixturepy/cycle_a/helper.go`
(`package helper` / `func Help() int { return 1 }`). The copy is never
committed: tests replay the recordings over the standard `python/` tree, the
same pattern `recorded/python-zero-edge/` uses. ripwire marks the module
holding the Go file `zone="unknown" no_file_edges="go"` with no `I`/`D`
attributes — the ARCH-HONEST surface the adapter must parse (design D2/D3).

## TypeScript fixture: `typescript/`

| File | Scenario exercised |
|------|--------------------|
| `typescript/package.json` | Minimal private project identity. |
| `typescript/tsconfig.json` | Minimal compiler config; `include: ["src/**/*"]`. |
| `typescript/src/models.ts` | Interface discovery and LCOM scenarios — see table below. |
| `typescript/src/cycle-a/index.ts` | Cycle side A: imports from `../cycle-b` (import edge A → B). |
| `typescript/src/cycle-b/index.ts` | Cycle side B: imports from `../cycle-a` (import edge B → A). |

### Declarations in `typescript/src/models.ts`

| Declaration | Scenario exercised |
|-------------|---------------------|
| `Point`, `Named`, `Repository` | Top-level exported interfaces. |
| `Shapes.Point` | Interface nested inside an exported namespace → nested-interface discovery. |
| `SplitMethods` | Four methods: two access only `this.x`, two access only `this.y` → LCOM 2. |

`cycle-a` and `cycle-b` form a two-directory import cycle. Neither
module is meant to be compiled or executed.

## Manifest note

This manifest lists every fixture file in this directory (it does not
list itself). Its presence is the completion marker for task 1.1.
