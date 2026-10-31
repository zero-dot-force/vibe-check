# Ripwire Adapter Test Fixtures

Static-analysis fixtures for the ripwire adapter — task 1.1 of the
`ripwire-adapter` OpenSpec change (`openspec/changes/ripwire-adapter/`;
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
| `python/src/fixturepy/operations/build/__init__.py` | Real Python package inside a directory named `build` (denylist name). |
| `python/src/fixturepy/operations/build/steps.py` | Source file inside the denylist-named directory; gives the denylist-pruned warning a concrete target. |
| `python/src/fixturepy/cycle_a/__init__.py` | Cycle side A: `from fixturepy.cycle_b.api import helper_b` (import edge A → B). |
| `python/src/fixturepy/cycle_a/api.py` | Defines `helper_a`, imported by cycle side B. |
| `python/src/fixturepy/cycle_b/__init__.py` | Cycle side B: `from fixturepy.cycle_a.api import helper_a` (import edge B → A). |
| `python/src/fixturepy/cycle_b/api.py` | Defines `helper_b`, imported by cycle side A. |

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
