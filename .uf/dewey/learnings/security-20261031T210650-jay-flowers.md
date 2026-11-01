---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:06:50Z
identity: security-20261031T210650-jay-flowers
tier: draft
---

Security-relevant bounding of untrusted subprocess output (error-path previews) must be single-sourced per package. In vibe-check's ripwireadapter, previewVersionOutput and previewBytes implemented the same hostile-output defense twice with drifted constants (120 vs 512) and divergent pre-processing (TrimSpace in only one copy), so any hardening of the bounding strategy — rune-aware truncation, control-character handling, trim-before-bound — applied to one copy silently misses the other, and no test notices because indirect error-message Contains assertions with short inputs never exercise the bound. When a defense against untrusted output appears at two call sites, extract it immediately with deliberately chosen, GoDoc-documented parameters, and pin the truncation, ellipsis, and empty-fallback paths with a direct table-driven test.
