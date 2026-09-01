---
tag: add-divisor-entropy-agent
author: jay-flowers
category: pattern
created_at: 2026-09-01T02:34:37Z
identity: add-divisor-entropy-agent-20260901T023437-jay-flowers
tier: draft
---

On 2026-08-31 in the vibe-check repo (branch add-divisor-entropy-agent), two reusable design patterns landed (category: pattern). First, the metrics base-to-PR entropy engine (metrics/delta.go + metrics/verdict.go): ComputeDelta matches modules by Module.Path and classifies cycles by their SORTED MEMBER SET so identity is rotation-invariant; DecideVerdict applies INCLUSIVE gates (ΔInstability≥0.15, ΔDistance≥0.20, ΔLCOM≥2, or any new circular dependency → REQUEST_CHANGES; smaller non-zero shifts → COMMENT; improving/stable → APPROVE), rounds floats to 4 decimals round-half-away-from-zero via math.Round(x*1e4)/1e4, and forces COMMENT (never a false APPROVE) whenever input is unreliable — a nil graph, Status != complete, or a load-error warning (partial build). This deliberately contrasts with the analyze command's threshold checks which use strict > comparison. Second, internal/scaffold/scaffold.go is a symlink-safe embedded-asset writer: Run delegates to an unexported run(assets fs.FS, opts) seam (so fstest.MapFS drives sort-order tests), validates the target with metrics.ValidateProjectPath, then does a deepest-existing-ancestor walk that Lstat-checks each path component (rejecting symlink or non-directory components) and verifies containment via filepath.EvalSymlinks + filepath.Rel (rejecting '..' escapes), writing dirs 0o755 / files 0o644 and normalizing mode with O_TRUNC+Chmod on force-overwrite; Result slices are asset basenames, stable-sorted. Finally, dogfooding works: running `go run ./cmd/vibe-check init .` regenerates .opencode/agents/divisor-entropy.md byte-identical to the embedded source of truth, so the deployed Review Council agent and the embedded asset never drift.
