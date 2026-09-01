# Entropy Divisor Validation Fixtures

This directory holds schema-valid `ModuleGraph` JSON fixture pairs
(`*-base.json` / `*-pr.json`) that exercise the structural-entropy
verdict engine (`metrics.ComputeDelta` + `metrics.DecideVerdict`)
end to end through the `vibe-check diff` command. They back the
`divisor-entropy` Review Council agent (deployed via
`vibe-check init`).

Each pair is pinned to one specific verdict path. Every fixture is a
schema-valid `ModuleGraph` (v1.1); `vibe-check diff` runs
`metrics.Validate` on both inputs and exits `2` on any schema-invalid
input, so the checklist below doubles as schema validation.

## Verdict thresholds (protected gates)

`DecideVerdict` uses inclusive `>=` comparisons against the protected
defaults, with per-module float deltas rounded to 4 decimal places
(round-half-away-from-zero):

- new circular dependency -> `REQUEST_CHANGES`
- instability delta `>= 0.15` -> `REQUEST_CHANGES`
- distance-from-main-sequence delta `>= 0.20` -> `REQUEST_CHANGES`
- LCOM delta `>= 2` -> `REQUEST_CHANGES`
- any non-zero worsening below those gates -> `COMMENT`
- improving or stable -> `APPROVE`
- unreliable (partial build) -> forced `COMMENT` (never a false `APPROVE`)

The numeric verdict logic itself is covered by the automated
table-driven tests in `metrics/delta_test.go` and
`metrics/verdict_test.go` (task 2.4). This checklist validates the
CLI-level, fixture-driven behavior (acceptance criterion #7 from
issue #3).

## Validation checklist

Run from this directory (`metrics/testdata/entropy/`):

```
go build -o /tmp/vibe-check ./../../../cmd/vibe-check
```

### Improvement -> APPROVE (acceptance #7)

`improvement-pr.json` resolves the `a <-> b` circular dependency
present in `improvement-base.json` (no other metric changes).

```
$ vibe-check diff improvement-base.json improvement-pr.json ; echo "exit=$?"
...
Resolved cycles:
  - example.com/app/a example.com/app/b
Entropy direction: improving
Verdict: APPROVE
Reasons:
  (none)
exit=0
```

### Degradation -> REQUEST CHANGES (acceptance #7)

`degradation-pr.json` introduces a new `a <-> b` circular dependency
not present in `degradation-base.json`.

```
$ vibe-check diff degradation-base.json degradation-pr.json ; echo "exit=$?"
...
New cycles:
  - example.com/app/a example.com/app/b
Entropy direction: degrading
Verdict: REQUEST_CHANGES
Reasons:
  - new-cycle: example.com/app/a example.com/app/b
exit=0
```

Note: `REQUEST_CHANGES` is data in the payload; `vibe-check diff`
still exits `0` when both inputs are valid. Exit `2` is reserved for
missing/unreadable/schema-invalid input and tighten-only threshold
violations.

### Comment band -> COMMENT

`comment-band-pr.json` raises module `a`'s instability by `0.05`
(below the `0.15` `REQUEST_CHANGES` gate but non-zero).

```
$ vibe-check diff comment-band-base.json comment-band-pr.json ; echo "exit=$?"
Per-module deltas (PR minus base):
  example.com/app/a  ... Instability 0.0500 ...
Entropy direction: stable
Verdict: COMMENT
Reasons:
  - materiality: non-zero metric shifts below REQUEST_CHANGES thresholds
exit=0
```

### Partial build -> COMMENT (unreliable)

`partial-build-pr.json` reports `status: "partial"` with a
`load-error` warning, so the measurement is unreliable.

```
$ vibe-check diff partial-build-base.json partial-build-pr.json ; echo "exit=$?"
partial build — measurement unreliable
...
Entropy direction: stable
Verdict: COMMENT
Reasons:
  - partial-build: measurement unreliable, verdict downgraded to COMMENT
exit=0
```

When a build is unreliable, added/removed package sections are
suppressed and the verdict is forced to `COMMENT` so an incomplete
analysis can never produce a false `APPROVE`.

## Manual-only runtime behaviors

The following `divisor-entropy` agent behaviors depend on git
worktree isolation and live tooling, so they are not reachable by the
Go unit tests and MUST be verified manually. In every case the agent
returns `COMMENT` (never a false `APPROVE`) and never crashes:

- **Binary missing**: if the `vibe-check` binary cannot be found or
  run, the agent reports `COMMENT` with a degradation note.
- **PR won't build**: if `vibe-check analyze` on the PR checkout
  fails (compilation error), the agent reports `COMMENT`.
- **Partial build**: if either analysis yields `status: "partial"`
  or `load-error` warnings, the agent reports `COMMENT` (matches the
  `partial-build` fixture above).
- **Worktree cleanup on failure**: the isolated git worktree created
  at the resolved base SHA is removed (`git worktree remove --force`)
  even when analysis or diffing fails partway through.
