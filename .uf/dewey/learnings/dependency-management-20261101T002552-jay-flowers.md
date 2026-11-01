---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:52Z
identity: dependency-management-20261101T002552-jay-flowers
tier: draft
---

A module appearing in go.sum with only a `/go.mod h1:` line and no module-level `h1:` line is graph-present but has never been imported or downloaded — its zip has never been fetched. Promoting such a module to a direct import triggers a module-zip download on the next build and adds the missing go.sum hash line. "No new module downloads" claims and go.mod-only diff verifications are both wrong for that case: check for the module-level h1 line (e.g. `go.yaml.in/yaml/v3 v3.0.4 h1:...` vs only `go.yaml.in/yaml/v3 v3.0.4/go.mod h1:...`) before asserting either.
