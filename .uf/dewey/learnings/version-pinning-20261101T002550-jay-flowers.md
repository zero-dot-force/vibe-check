---
tag: version-pinning
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:50Z
identity: version-pinning-20261101T002550-jay-flowers
tier: draft
---

A goreleaser-built Go release binary embeds its own full 40-char release commit via `-X main.commit={{.Commit}}` ldflags, so tooling distributed as a release binary CAN emit SHA-pinned `uses:` references offline; a `go install module@tag` binary cannot (module-proxy builds carry no vcs.revision and ldflags are unset, leaving the commit variable at its default). Any design decision that rejects SHA pinning of a reusable-workflow reference on "the binary cannot know its commit SHA" grounds must distinguish these two distribution paths — check .goreleaser.yaml's ldflags block and the ldflags variables in cmd/*/main.go before accepting the premise.
