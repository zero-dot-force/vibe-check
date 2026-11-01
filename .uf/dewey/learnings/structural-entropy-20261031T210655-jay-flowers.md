---
tag: structural-entropy
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:06:55Z
identity: structural-entropy-20261031T210655-jay-flowers
tier: draft
---

Security-relevant output-bounding helpers duplicated inside a newly added package are invisible to every entropy-divisor gate, twice over: vibe-check diff treats added packages as informational-only (no delta-Instability/delta-Distance/delta-LCOM gate fires for them), and intra-package duplication of unexported free functions moves no Martin metric anyway (Ca/Ce are import-graph metrics; LCOM4 counts method-field connected components). Verified on vibe-check PR #75 (ripwireadapter): previewVersionOutput (version.go — cap 120, TrimSpace, GoDoc'd hostile-binary rationale) and previewBytes (runner.go — cap 512, no trim, no rationale) drifted apart in the same PR that also drifted installFakeRipwireWrapper across two packages — per-file local helpers are the path of least resistance at authoring time, so the pattern recurs. Reviewer rule for newly added packages: grep the package for security primitives implemented more than once (bounding, quoting, sanitization, path validation) and require single-sourcing at introduction; the extraction itself is structurally free — zero metric delta for unexported method-less functions within one package (learning/structural-entropy-20261029T155410).
