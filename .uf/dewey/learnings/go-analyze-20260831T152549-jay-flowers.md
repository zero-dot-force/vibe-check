---
tag: go-analyze
author: jay-flowers
category: gotcha
created_at: 2026-08-31T15:25:49Z
identity: go-analyze-20260831T152549-jay-flowers
tier: draft
---

In vibe-check's Go adapter, resolveReceiverType (internal/goadapter/lcom.go:99) has a latent correctness bug with ZERO test coverage: for generic pointer receivers `func (s *S[T]) M()`, the receiver expr is *ast.StarExpr{X: *ast.IndexExpr}, so the `case *ast.StarExpr` branch does `t.X.(*ast.Ident)` which FAILS (X is IndexExpr not Ident), falls through, and returns "" — silently dropping the method from computeLCOM4, corrupting the LCOM metric. The value-receiver `*ast.Ident` branch and the generic `*ast.IndexExpr`/`*ast.IndexListExpr` branches are also untested because every testdata fixture uses only plain pointer receivers `func (s *S)`. Gaze flagged this as CRAP 27.0 / 33.3% line coverage — the highest-CRAP gap in the branch. Fix: add a fixture (or table-driven unit test using go/parser) covering value receiver `func (t T)`, generic value `func (t T[P])`, generic pointer `func (t *T[P])`, and multi-param generic `func (t *T[P1,P2])`; then fix the StarExpr branch to recurse into IndexExpr/IndexListExpr. This is a determinism/correctness-mandate risk (AGENTS.md requires deterministic, correct metric computation).
