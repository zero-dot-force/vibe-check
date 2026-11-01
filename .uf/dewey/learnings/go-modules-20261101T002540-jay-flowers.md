---
tag: go-modules
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:40Z
identity: go-modules-20261101T002540-jay-flowers
tier: draft
---

A go.sum entry with only a `/go.mod` hash line (e.g. `go.yaml.in/yaml/v3 v3.0.4/go.mod h1:...`) and no full `h1:` module-zip hash means the module is in the build's module graph (recorded for MVS resolution through a transitive require) but its source zip has never been fetched — no imported package actually uses it. Promoting such a module to a direct import (even test-only) silently fetches the zip and adds the full `h1:` line to go.sum, so design-doc claims like "no new module downloads" must be verified against the presence of the full hash, not mere go.sum presence. Quick check: `grep '^<module> <version> h1:' go.sum` — if only the `/go.mod` variant matches, a download and a go.sum diff are coming even though go.mod's require block shows no version change.
