---
tag: go-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T16:13:30Z
identity: go-testing-20261101T161330-jay-flowers
tier: draft
---

PR CI that compiles the merge of a feature branch into main's tip builds a tree that exists nowhere locally: tests added on the branch that call internal helpers whose signatures drifted on main compile cleanly on the branch but fail typecheck in CI's merge tree. In vibe-check internal/scaffold, run() gained a third ciAssets fs.FS parameter on main after the branch base, so iteration-1's TestRun_GlobEnumerationError (2-arg run call) broke PR CI while passing branch-local checks. Fix pattern: target the stable seam — deployCategory(assets, cat, root, writeFile, force) kept an identical 5-param signature on both trees, and fs.Glob is its first operation with an error short-circuit before writeFile is invoked, so a nil writeFile and a raw (non-EvalSymlinks-resolved) tempdir root are safe for an injected-Glob-failure test (renamed TestDeployCategory_GlobEnumerationError). Prevention: run a git merge-tree simulation (build + vet + test -race in a temp worktree of origin/main + HEAD) before pushing fix-loop iterations; branch-local go build/test passing is insufficient evidence for PR CI. Correction to earlier learning go-testing-20261101T155330-jay-flowers: the pattern holds, the test name is now TestDeployCategory_GlobEnumerationError.
