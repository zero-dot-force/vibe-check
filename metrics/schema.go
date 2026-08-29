package metrics

import _ "embed"

//go:embed modulegraph.schema.json
var schemaJSON []byte

// SchemaJSON returns the raw JSON Schema document for ModuleGraph validation.
// It returns a copy of the embedded schema to prevent callers from mutating
// the package-level data.
func SchemaJSON() []byte {
	cp := make([]byte, len(schemaJSON))
	copy(cp, schemaJSON)
	return cp
}
