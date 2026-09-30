## 1. Verify analyze-side documentation (issue #34 verification items)

- [x] 1.1 Verify README Output section documents the `provenance` object with `producer` (constant `"vibe-check"`), `version`, `generatedAt` (RFC 3339 UTC), and `input{path,modulePath}`; correct wording if inaccurate
- [x] 1.2 Verify README flag table lists `--no-provenance` with its byte-reproducibility purpose
- [x] 1.3 Verify README JSON example sets `schemaVersion: "1.2"` and includes a `provenance` object with `producer`, `version`, `generatedAt`, and `input{path,modulePath}`

## 2. Document diff and init provenance (the actual gap, issue #36)

- [x] 2.1 Add a sentence to the README `vibe-check diff` section stating `diff --json` emits a `provenance` object (`producer` as the constant `"vibe-check"`, `version`, RFC 3339 UTC `generatedAt`) and accepts `--no-provenance`; do not mention `input`
- [x] 2.2 Add a sentence to the README `vibe-check init` section stating `init --json` emits a `provenance` object (`producer` as the constant `"vibe-check"`, `version`, RFC 3339 UTC `generatedAt`) and accepts `--no-provenance`; do not mention `input`

## 3. Verify CHANGELOG coverage

- [x] 3.1 Verify CHANGELOG `[Unreleased]` references `emit-provenance-metadata` for the `analyze` provenance feature
- [x] 3.2 Verify CHANGELOG `[Unreleased]` references `add-diff-init-provenance` for the `diff`/`init` provenance feature

## 4. Verification

- [x] 4.1 Cross-check README wording against actual `analyze --json`, `diff --json`, and `init --json` output from the current `main` tree, confirming: (a) the Output section names all four `analyze` fields; (b) the flag table lists `--no-provenance`; (c) the JSON example shows `schemaVersion` `"1.2"`; (d) the `diff`/`init` sections name `producer` as the constant `"vibe-check"`, `version`, and RFC 3339 UTC `generatedAt`, and do not describe an `input` field
- [x] 4.2 Run `openspec validate document-provenance-output` and ensure the spec passes

<!-- spec-review: passed -->
<!-- code-review: passed -->
