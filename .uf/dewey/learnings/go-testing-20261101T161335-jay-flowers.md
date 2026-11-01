---
tag: go-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T16:13:35Z
identity: go-testing-20261101T161335-jay-flowers
tier: draft
---

Long-lived feature branches whose CI compiles the merge of the branch into main's tip need test seams chosen for cross-tree signature stability, not just branch-local correctness. In vibe-check's internal/scaffold (PR #75 fix loop), the iteration-1 test TestRun_GlobEnumerationError called the package-private run() with the branch tree's signature (agentAssets, commandAssets fs.FS, Options). Main's tip independently added a third ciAssets fs.FS parameter after the PR base, so the call compiled and passed on the branch but failed vet/lint typecheck in CI's merge tree. The durable fix re-aimed the test at the stable inner seam — deployCategory(assets, cat, root, writeFile, force) — whose signature was identical on both trees and which owns the fs.Glob call under test (a struct embedding fs.FS and overriding Glob implements fs.GlobFS, so fs.Glob dispatches to the injected failure; the %w wrap makes errors.Is chain through the sentinel). The verification that matters is a git merge-tree simulation of origin/main + HEAD (build, vet, -race tests, coverage ratchet), not just a branch-local go test. Pattern: when CI merge-checks against a moving main, test the innermost function that owns the behavior under test rather than an outer wrapper whose parameter list is still evolving, and record the reason in the test's comment so a future contributor does not simplify it back to the drifting call.
