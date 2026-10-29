---
tag: byte-preserving-splice
author: jay-flowers
category: gotcha
created_at: 2026-10-29T15:53:52Z
identity: byte-preserving-splice-20261029T155352-jay-flowers
tier: draft
---

Byte-preserving line splices in Go: `strings.Split(content, "\n")` on a newline-terminated file yields a trailing empty-string sentinel element — an insertion anchored at end-of-file must clamp its index to len(lines)-1 when the final element is that sentinel (see vibe-check internal/scaffold clampInsertIndex), otherwise the inserted block lands after the terminator and the rejoined file gains a spurious blank line while losing its trailing newline. Related sentinel behavior: an insertion after an unterminated final line must let the join append the missing terminator so the inserted block starts on its own line.
