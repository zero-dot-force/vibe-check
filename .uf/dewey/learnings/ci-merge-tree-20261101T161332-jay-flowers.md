---
tag: ci-merge-tree
author: jay-flowers
category: gotcha
created_at: 2026-11-01T16:13:32Z
identity: ci-merge-tree-20261101T161332-jay-flowers
tier: draft
---

PR CI compiles the merge of the feature branch into the base branch's current tip, not the feature branch's own tree. A test added on a long-lived branch that calls an unexported function with the branch-tree's signature (vibe-check scaffold's `run` with two fs.FS params) compiles and passes locally but fails vet/lint typecheck in CI when the base branch evolved that signature after the merge base (main's `run` gained a third ciAssets fs.FS param). Symptom: PR checks red while the branch builds clean locally. Fix pattern: retarget the test at the function that owns the behavior under test and whose signature is stable across both trees — `deployCategory` owns the fs.Glob enumeration seam and its 5-param signature is identical on branch and main — preserving the original behavioral assertion (errors.Is sentinel chain check) rather than deleting or weakening the test, and documenting the retargeting rationale in the test comment. Verify before pushing by building and testing a `git merge-tree <base> HEAD` simulation in a temp worktree, including the coverage ratchet, because CI evaluates the merged tree.
