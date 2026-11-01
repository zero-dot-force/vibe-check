---
tag: testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:51Z
identity: testing-20261101T143251-jay-flowers
tier: draft
---

fmt %q emits double quotes, and POSIX sh interpolates $ and backticks inside double quotes — baking a filesystem path into a generated shell test wrapper via %q is an injection-shaped bug that fires when checkout or TMPDIR paths are adversarial. Wrap embedded values in POSIX single quotes using the '\'' escape idiom instead. Give the quoting helper its own table test (inputs containing a single quote, $, backtick, empty string): call sites only ever see benign temp-dir paths, so an escaping regression is otherwise invisible until it fails obscurely on an exotic-path machine — and because _test.go helpers are not importable across packages, each duplicated copy of the helper needs its own pin.
