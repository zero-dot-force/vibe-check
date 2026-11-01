---
tag: version-pinning
author: jay-flowers
category: gotcha
created_at: 2026-11-01T01:55:52Z
identity: version-pinning-20261101T015552-jay-flowers
tier: draft
---

Resolving a release tag to its commit via `git ls-remote <repo> refs/tags/<tag>^{}` only works for annotated tags — lightweight tags have no peeled ref, so the query returns empty output and a fail-closed resolver reports a misleading "tag not found / rerun with network access" error for a tag that exists. Generators that pin emitted CI `uses:` refs this way (verified on vibe-check's init --ci tag→SHA resolver, cmd/vibe-check/init.go) should either fall back to the unpeeled `refs/tags/<tag>` when the peeled query returns empty (a lightweight tag's ref points directly at the commit) or document annotated-tags-only in the release process. Companion diagnosability rule for the same code path: capture git's stderr when shelling out (`cmd.Stderr = &buf`, or CombinedOutput) — `.Output()` discards it, leaving only "exit status 128" to distinguish DNS failure, missing network, and repository-not-found; and add a context.WithTimeout to the resolution subprocess so a black-holed connection cannot hang the generator (an interactive signal-context alone bounds nothing).
