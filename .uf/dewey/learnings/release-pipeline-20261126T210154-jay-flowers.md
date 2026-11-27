---
tag: release-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-26T21:01:54Z
identity: release-pipeline-20261126T210154-jay-flowers
tier: draft
as_of: "vibe-check@40a2e0eb4a5313ddb15330ed799f48435eca3ad7"
---

When a repo bumps a SHA-pinned reusable workflow and passes new inputs it did not previously use (e.g. ship's go_install / go_install_check), review the caller against three things before approving: (1) the new SHA resolves to the claimed tag (annotated tag must be dereferenced to the commit SHA, not the tag object); (2) each new input name/type/default actually exists in the callee ref at that exact SHA, and the check string's expansion context (e.g. $BIN exported by the callee's bash -c step) matches the callee's documented example; (3) the downstream job's runtime prerequisites — for a `go install module@tag` verification, the module's `go` directive (go.mod) must be satisfiable by the callee runner's toolchain or by GOTOOLCHAIN=auto, otherwise the first release run fails as an environment mismatch rather than a packaging defect. For non-blocking post-release verification jobs, classify such first-run toolchain risk as LOW/advisory, not HIGH.
