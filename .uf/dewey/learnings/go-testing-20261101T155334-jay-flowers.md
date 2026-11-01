---
tag: go-testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T15:53:34Z
identity: go-testing-20261101T155334-jay-flowers
tier: draft
---

io/fs.Glob dispatches to the FS's own Glob method when the FS implements io/fs.GlobFS, before any pattern walking — so a minimal struct that embeds fs.FS and overrides Glob with a value receiver (like internal/scaffold's globErrFS for error injection and orderedGlobFS for unsorted-result injection) is the reliable seam for exercising fs.Glob-based enumeration paths in tests. Confirmed against io/fs.GlobFS documentation; deployCategory (scaffold.go) calls fs.Glob(assets, ...) and wraps the result with %w, so errors.Is chains through the injected sentinel.
