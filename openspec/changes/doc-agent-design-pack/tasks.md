## 1. Core Documentation

- [x] 1.1 Create `docs/agent-design-pack.md` with introduction and purpose
- [x] 1.2 Add Rule Reference Table section listing all 10 rules (AD-001–AD-010) with ID, name, threshold, severity, and enforcement tool
- [x] 1.3 Add rule-by-rule detail section for each AD rule covering rationale, compliance examples (PASS/FAIL), and tool-specific flags

## 2. Enforcement Mapping

- [x] 2.1 Add Enforcement Tool Matrix section cross-referencing vibe-check, gaze, golangci-lint, and review agents
- [x] 2.2 Note forward references for AD-002 (`--max-ce`, noting absence of a tracking issue) and AD-008 (`--max-duplication`, with link to tracking issue zero-dot-force/vibe-check#13)

## 3. Override Mechanism

- [x] 3.1 Document `agent-design-custom.md` override mechanism with CR-NNN prefix convention, scoped overrides, and justification requirement

## 4. Semantics Clarification

- [x] 4.1 Document LCOM4 integer semantics for AD-009 (connected components, Hitz & Montazeri 1995, not a 0.0–1.0 float)
- [x] 4.2 Add Authority Statement section identifying `.opencode/uf/packs/agent-design.md` as the authoritative source and this document as a derivative reference

## 5. Verification

- [x] 5.1 Verify the document covers all spec requirements (rule reference, enforcement matrix, overrides, forward references, LCOM4 semantics)
- [x] 5.2 Run all CI-equivalent checks to confirm no accidental code changes: `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...`, `golangci-lint run ./...` — all pass
- [x] 5.3 Update `CHANGELOG.md` with a `docs:` entry for the new agent-design reference documentation

<!-- spec-review: passed -->
<!-- code-review: passed -->