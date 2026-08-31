---
tag: go-analyze
author: jay-flowers
category: decision
created_at: 2026-08-31T16:33:21Z
identity: go-analyze-20260831T163321-jay-flowers
tier: draft
---

Adversary council re-review (iteration 1) of vibe-check branch opsx/go-analyze: prior MEDIUM (ExternalAdapter trust boundary under-enforced schema) is RESOLVED via two-layer enforcement. metrics/validate.go now enforces numeric ranges through a moduleNumber helper (instability/abstractness/distance in [0,1]; ca/ce/lcom/exportedTypes/abstractTypes >= 0, plus number-type checks), and metrics/external.go decodes the subprocess response with json.Decoder.DisallowUnknownFields() enforcing additionalProperties:false. Key insight: integer-ness (schema says ca/ce/lcom are integer) is NOT enforced by Validate() alone (moduleNumber accepts any float64), but the strict decoder rejects fractional values because Module.Ca etc. are Go int fields — so the ExternalAdapter boundary (Validate THEN Decode) is airtight, while the standalone public metrics.Validate() has a minor integer-fidelity gap for external callers (LOW). DisallowUnknownFields correctly flattens the embedded Module struct (promoted fields path/name/ca/ce/exportedTypes/abstractTypes) and does NOT restrict the extensions map[string]any (matches open-object schema). packageEnvAllowlist (resolve.go) excludes GOFLAGS with regression test TestPackageEnvAllowlist_ExcludesInjectionVectors that locks exact set+size. Remaining advisories all LOW/non-blocking: (1) binaryPath unvalidated in NewExternalAdapter but caller-supplied trusted config and not wired to CLI, (2) go/packages type-checking can execute code via cgo/-toolexec on untrusted input — undocumented but intended use is self-analysis of trusted code, (3) recursive Tarjan strongConnect has no depth guard (stack overflow only on pathological graphs). VERDICT: APPROVE.
