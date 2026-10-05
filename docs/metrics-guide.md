# Vibe-Check Metrics Guide

A plain-language reference for every metric produced by `vibe-check analyze`.
If a report numbers in a table confuse you, start here.

---

## Instability (I)

### What it is
A number from 0.0 to 1.0 that tells you how resistant a package is to change.
It is the ratio of outgoing dependencies to total dependencies:  
`I = Ce / (Ca + Ce)`

### Scale
- **0.0** — Maximally stable. Everyone depends on this package. Changing it
  breaks everything.
- **1.0** — Maximally unstable. This package depends on everyone else.
  Changing it doesn't affect anything upstream.

### Why care
If your core logic packages have high instability, every dependency change
ripples through them. If your leaf packages have low instability, you may have
a dependency inversion problem.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 0.20–0.80 | Well-balanced |
| 🟡 0.10–0.20 or 0.80–0.95 | Mildly imbalanced — acceptable for leaf/cli packages |
| 🔴 < 0.10 or > 0.95 | At an extreme — changes ripple widely or package is isolated |

---

## Abstractness (A)

### What it is
A number from 0.0 to 1.0 that tells you how much of a package is interfaces
and abstract types versus concrete implementations:  
`A = (abstract types) / (total types)`

### Scale
- **0.0** — Fully concrete. No interfaces or abstract types. Everything is a
  struct with methods.
- **1.0** — Pure interfaces. No concrete implementations.

### Why care
Packages at either extreme are difficult to extend. All-concrete packages
can't be swapped out. All-abstract packages need an implementation somewhere
else.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 0.10–0.60 | Reasonable mix of interfaces and concrete code |
| 🟡 0.01–0.10 or 0.60–0.90 | Slightly imbalanced |
| 🔴 0.00 or > 0.90 | At an extreme — no extensibility or all scaffolding |

---

## Distance from Main Sequence (D)

### What it is
A number from 0.0 to 1.0 that measures how far your package is from the
ideal balance of instability and abstractness:  
`D = |A + I - 1.0|`

### Scale
- **0.0** — Perfect balance. The package is on the Main Sequence.
  Its abstractness and instability complement each other.
- **1.0** — Maximum distance. The package is in the Zone of Pain
  (concrete + stable) or Zone of Uselessness (abstract + unstable).

### Why care
This is your single best indicator of package health. A high distance value
means the package's design doesn't match its position in the dependency graph.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 < 0.30 | On or near the Main Sequence |
| 🟡 0.30–0.50 | Moderate distance — worth reviewing |
| 🔴 > 0.50 | Far from ideal — consider refactoring |

---

## Lack of Cohesion of Methods 4 (LCOM4)

### What it is
A count starting at 1 that measures how many disconnected groups of methods
exist in a package. LCOM4 treats method groups connected by shared field
access as one component; separate components mean separate responsibilities.

### Scale
- **1** — Perfectly cohesive. All methods share at least one field. One
  responsibility.
- **2–3** — Slightly fragmented. The package may have 2–3 distinct
  responsibilities.
- **≥ 4** — Low cohesion. The package should probably be split into multiple
  smaller packages.

### Why care
LCOM4 catches the "god package" problem. A package with LCOM4=7 almost
certainly does too many unrelated things. Splitting it simplifies testing and
understanding.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 1 | Single responsibility |
| 🟡 2–3 | Moderate fragmentation — keep an eye on growth |
| 🔴 ≥ 4 | Consider splitting the package |

---

## Afferent Coupling (Ca)

### What it is
How many packages depend on **this** package. Also called "fan-in."

### Scale
- **0** — Nothing depends on this package. It is a leaf.
- **High** — Many packages depend on this one. Changes here affect many
  consumers.

### Why care
Packages with high Ca are stable by definition — changing them breaks
many things. They should have interfaces to insulate consumers. Packages
with Ca=0 are safe to refactor but may be dead code.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 1–10 | Moderate usage |
| 🟡 0 or 11–20 | Unused (dead code?) or heavily depended on |
| 🔴 > 20 | Very high fan-in — ensure interfaces protect consumers |

---

## Efferent Coupling (Ce)

### What it is
How many packages **this** package depends on. Also called "fan-out."

### Scale
- **0** — This package depends on nothing. Highly stable but possibly
  isolated.
- **High** — This package depends on many others. It is fragile — any
  upstream change can break it.

### Why care
High Ce makes a package fragile and hard to test in isolation. Consider
extracting interfaces for the most heavily used dependencies.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 1–10 | Reasonable number of dependencies |
| 🟡 0 or 11–20 | Self-contained or heavily dependent |
| 🔴 > 20 | Too many dependencies — hard to isolate for testing |

---

## Circular Dependencies

### What it is
A cycle in the package dependency graph where package A depends on B,
B depends on C, and C depends on A (or any variant).

### Scale
- **0** — No cycles. The package graph is a directed acyclic graph (DAG).
- **≥ 1** — At least one cycle exists.

### Why care
Cycles mean you can't test any package in the cycle in isolation. They also
make it impossible to determine build order. Every cycle in a Go codebase is a
compiler error waiting to happen.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 0 | No cycles — clean dependency graph |
| 🔴 ≥ 1 | Cycles detected — must be broken before packages can be independently tested |

---

## Code Duplication

### What it is
Blocks of code that appear in multiple places with high similarity.
Vibe-check detects duplicated blocks within and across files in a package.

### Scale
- **0%** — No duplication detected.
- **> 0%** — Percentage of duplicated lines within the package.

### Why care
Duplication is the enemy of refactoring. If you fix a bug in one copy, you
have to remember to fix it in all copies. Even a small duplication percentage
indicates a missed extraction opportunity.

### Emoji ranges
| Range | Meaning |
|-------|---------|
| 🟢 0% | No duplication |
| 🟡 < 5% | Minor duplication — watch for growth |
| 🔴 ≥ 5% | Significant duplication — extract shared logic |

---

## Go-Specific Extensions

The Go adapter also produces language-specific metrics in the `extensions`
field:

- **`go.interfaceWidth`** — Number of exported methods on each interface.
  Wider interfaces are harder to implement and mock.
- **`go.interfaceProximity`** — How close package types are to the interfaces
  they implement (co-location metric).

See `internal/goadapter/extensions.go` for implementation details.

---

## Further Reading

- Robert C. Martin, ["Design Principles and Design Patterns"](https://web.archive.org/web/20150906155800/http://www.objectmentor.com/resources/articles/Principles_and_Patterns.pdf) — the original paper defining Instability, Abstractness, and Distance from Main Sequence
- Hitz & Montazeri, ["Measuring Coupling and Cohesion in Object-Oriented Systems"](http://www.isys.uni-klu.ac.at/PDF/1995-0043-MHBM.pdf) — LCOM4 definition and connected-component analysis
- Tarjan, ["Depth-First Search and Linear Graph Algorithms"](https://epubs.siam.org/doi/10.1137/0201010) — the SCC algorithm used for cycle detection