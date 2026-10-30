---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T17:33:34Z
identity: coverage-ratchet-20261030T173334-jay-flowers
tier: draft
---

Go coverage ratchets keyed off `go test -coverprofile` can only see packages that have at least one test file — packages with zero tests are absent from the profile entirely, so they bypass unregistered-package checks and entry bars. Deleting all tests from a registered package is still caught (the package vanishes from the profile and trips the stale-floor direction); only brand-new never-tested packages escape. Closing the gap requires cross-referencing `go list ./...` against the profile's package set.
