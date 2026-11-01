---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:09:52Z
identity: review-dispatch-20261031T200952-jay-flowers
tier: draft
---

When verifying dead-code or zero-caller claims for a PR delivered in manifest mode (diff not inlined), grep the saved patch artifact directly: point the grep tool at the diff file's directory with include=<patch-filename> and the identifier as the pattern. The review session's working tree may be checked out on a different branch (e.g. main) than the PR head, so grepping repository files returns a false "no matches" and can wrongly confirm or refute the claim. In the PR #75 ripwire-adapter feedback review, grepping repo files for countSourceFiles found nothing, while grepping the patch file found all 7 true occurrences (definition, GoDoc references, and test calls only) — definitive confirmation that the function has zero production callers.
