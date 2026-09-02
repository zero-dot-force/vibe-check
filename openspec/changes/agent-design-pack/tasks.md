<!-- Tasks marked [P] can be executed in parallel within their phase -->

## 1. Directory and File Setup

- [x] 1.1 Create `.opencode/uf/packs/` directory if it does not exist
- [x] [P] 1.2 Create `.opencode/uf/packs/agent-design.md` with pack header and metadata
- [x] [P] 1.3 Create `.opencode/uf/packs/agent-design-custom.md` placeholder for project-level overrides (include YAML frontmatter with `pack_id: agent-design-custom` and a description of the override mechanism)

## 2. Coupling Rules (vibe-check)

- [x] 2.1 Write AD-002 Package Fan-out rule (Ce < 10, enforcement: `vibe-check analyze` JSON output inspection, note `--max-ce` as forward reference)
- [x] 2.2 Write AD-003 Instability Threshold rule (I < 0.7 non-leaf where leaf = Ca = 0, enforcement: `vibe-check analyze --max-instability=0.7`)
- [x] 2.3 Write AD-004 No Circular Dependencies rule (enforcement: `vibe-check analyze --no-circular-deps`)

## 3. Cohesion and Duplication Rules (vibe-check)

- [x] 3.1 Write AD-008 DRY No Duplication rule (blocks ≥ 6 consecutive lines, enforcement: `vibe-check analyze --max-duplication=5`, mark as forward reference)
- [x] 3.2 Write AD-009 Package Cohesion rule (LCOM4 ≤ 3, enforcement: `vibe-check analyze --max-lcom=3`)

## 4. Complexity and Coverage Rules (gaze)

- [x] 4.1 Write AD-001 Cognitive Complexity rule (< 15/function, enforcement: gaze)
- [x] 4.2 Write AD-006 Contract Coverage rule (every exported function has ≥ 1 test, enforcement: gaze)
- [x] 4.3 Write AD-010 Behavior-Asserting Tests rule (every test function has ≥ 1 behavioral assertion, enforcement: gaze)

## 5. Naming and Structure Rules

- [x] 5.1 Write AD-005 Naming Conventions rule (enforcement: `golangci-lint run` with revive)
- [x] 5.2 Write AD-007 File Size rule (< 400 lines, enforcement: review agent)

## 6. Verification

- [x] [P] 6.1 Verify all 10 rules (AD-001 through AD-010) are present with required fields (ID, Name, Severity, Rationale, Threshold, Enforcement, Example)
- [x] [P] 6.2 Verify AD-002, AD-003, AD-004, AD-008, AD-009 reference vibe-check
- [x] [P] 6.3 Verify AD-001, AD-006, AD-010 reference gaze
- [x] [P] 6.4 Verify AD-008 references `vibe-check analyze --max-duplication`
- [x] 6.5 Update AGENTS.md convention packs list to include `.opencode/uf/packs/agent-design.md` and `.opencode/uf/packs/agent-design-custom.md`
- [x] 6.6 Add CHANGELOG.md entry documenting the new agent-design convention pack
<!-- spec-review: passed -->
<!-- code-review: passed -->
