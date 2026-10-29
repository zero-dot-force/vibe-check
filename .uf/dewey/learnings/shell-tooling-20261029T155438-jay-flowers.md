---
tag: shell-tooling
author: jay-flowers
category: gotcha
created_at: 2026-10-29T15:54:38Z
identity: shell-tooling-20261029T155438-jay-flowers
tier: draft
---

Bulk sed name replacement collides on shared prefixes: replacing `capabilitiesFixtureMissing` before `capabilitiesFixtureMissingScopes` mangled the longer name into `fixture(t, "capabilities-missing.yaml")Scopes`, producing Go composite-literal syntax errors three functions later. When sed-replacing a family of identifiers, order the patterns longest-first (or match with word-boundary suffixes) and grep the transformed output for leftover fragments of every identifier family before compiling. Also: macOS `cat -A` does not exist — use `tail -c 1 | od -c` to check trailing newlines.
