---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:30:48Z
identity: error-handling-20261031T193048-jay-flowers
tier: draft
---

Go map iteration order is randomized, so a fail-closed validator that ranges over a map to report the first missing key reports a nondeterministic key whenever several are missing — the doc claim "names the first nil pointer" is unimplementable with a map. Use an ordered slice of {label, pointer} structs for required-attribute checks whenever the error message is part of the observable contract (goldens, order-sensitive assertions, deterministic-output conventions).
