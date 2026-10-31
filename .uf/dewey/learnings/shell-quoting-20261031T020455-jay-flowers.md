---
tag: shell-quoting
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:04:55Z
identity: shell-quoting-20261031T020455-jay-flowers
tier: draft
---

Go's fmt %q emits a double-quoted Go string literal, which is not shell-safe quoting: inside shell double quotes, $ and backticks still expand. When generating shell scripts from Go (test wrapper scripts, generated shims), wrap interpolated values in single quotes with the '\'' escape — a shellSingleQuote(s) helper — rather than %q. In vibe-check's ripwire tests, the internal package uses the safe form while the CLI test's %q copy works only because temp/repo paths happen to contain no $ or backticks.
