---
tag: go-testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T15:53:30Z
identity: go-testing-20261101T155330-jay-flowers
tier: draft
---

To inject an enumeration failure into Go code that calls fs.Glob, wrap an fs.FS in a value type implementing Glob(string) ([]string, error) — the fs.GlobFS interface. fs.Glob dispatches to the filesystem's Glob method when the interface is implemented, so the injected error flows through the production path and errors.Is can verify the %w error chain end-to-end. Companion pattern: an orderedGlobFS with a canned globResult advertises entries (or none) without touching a real filesystem or embed.FS. Used in vibe-check internal/scaffold to test deployCategory's enumerate-failure branch (globErrFS, TestRun_GlobEnumerationError) and the ghost-entry read-failure branch (orderedGlobFS/emptyFS) hermetically.
