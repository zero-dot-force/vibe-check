package metrics

import (
	"encoding/json"
	"fmt"
)

// Validate checks whether the given JSON data conforms to the ModuleGraph schema.
// It verifies required fields, value types, enum constraints, and schema version
// compatibility without relying on an external JSON Schema validation library.
// Returns nil if valid, or an error describing the first validation failure.
func Validate(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("validate: empty input")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := validateTopLevel(raw); err != nil {
		return err
	}

	if err := validateModules(raw); err != nil {
		return err
	}

	if err := validateCycles(raw); err != nil {
		return err
	}

	if err := validateWarnings(raw); err != nil {
		return err
	}

	return validateProvenance(raw)
}

// validateTopLevel checks required top-level fields, schema version, language,
// and status.
func validateTopLevel(raw map[string]interface{}) error {
	requiredFields := []string{"schemaVersion", "language", "modules", "cycles", "warnings", "status"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			return fmt.Errorf("validate: missing required field %q", field)
		}
	}

	// Validate schemaVersion is a supported value.
	// Accept "1.0" (no extensions), "1.1" (with extensions), "1.2" (with
	// provenance), and "1.3" (with duplications) for backward compatibility.
	version, ok := raw["schemaVersion"].(string)
	if !ok {
		return fmt.Errorf("validate: field \"schemaVersion\" must be a string")
	}
	if version != "1.0" && version != "1.1" && version != "1.2" && version != "1.3" {
		return fmt.Errorf("validate: unsupported schema version %q (supported: \"1.0\", \"1.1\", \"1.2\", \"1.3\")", version)
	}

	// Validate language is a non-empty string.
	lang, ok := raw["language"].(string)
	if !ok {
		return fmt.Errorf("validate: field \"language\" must be a string")
	}
	if lang == "" {
		return fmt.Errorf("validate: field \"language\" must be non-empty")
	}

	// Validate status is a valid enum value.
	status, ok := raw["status"].(string)
	if !ok {
		return fmt.Errorf("validate: field \"status\" must be a string")
	}
	if err := validateStatusEnum(status); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

// validateModules checks that the modules field is an array of valid module objects.
func validateModules(raw map[string]interface{}) error {
	modulesRaw, ok := raw["modules"].([]interface{})
	if !ok {
		return fmt.Errorf("validate: field \"modules\" must be an array")
	}
	for i, m := range modulesRaw {
		if err := validateModule(m, i); err != nil {
			return fmt.Errorf("validate: %w", err)
		}
	}
	return nil
}

// validateCycles checks that the cycles field is an array (not null).
func validateCycles(raw map[string]interface{}) error {
	if _, ok := raw["cycles"].([]interface{}); !ok {
		return fmt.Errorf("validate: field \"cycles\" must be an array")
	}
	return nil
}

// validateWarnings checks that the warnings field is an array of valid warning objects.
func validateWarnings(raw map[string]interface{}) error {
	warningsRaw, ok := raw["warnings"].([]interface{})
	if !ok {
		return fmt.Errorf("validate: field \"warnings\" must be an array")
	}
	for i, w := range warningsRaw {
		if err := validateWarning(w, i); err != nil {
			return fmt.Errorf("validate: %w", err)
		}
	}
	return nil
}

// validateStatusEnum checks that status is one of the allowed values.
func validateStatusEnum(s string) error {
	switch s {
	case "complete", "partial", "error":
		return nil
	default:
		return fmt.Errorf("invalid status %q: must be one of \"complete\", \"partial\", \"error\"", s)
	}
}

// validateZoneEnum checks that zone is one of the allowed values.
func validateZoneEnum(z string) error {
	switch z {
	case "main-sequence", "zone-of-pain", "zone-of-uselessness", "normal":
		return nil
	default:
		return fmt.Errorf("invalid zone %q: must be one of \"main-sequence\", \"zone-of-pain\", \"zone-of-uselessness\", \"normal\"", z)
	}
}

// validateModule checks that a module element has all required fields and valid types.
func validateModule(v interface{}, index int) error {
	m, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("modules[%d]: must be an object", index)
	}

	requiredFields := []string{
		"path", "name", "ca", "ce",
		"instability", "abstractness", "distance", "lcom",
		"exportedTypes", "abstractTypes", "zone",
	}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			return fmt.Errorf("modules[%d]: missing required field %q", index, field)
		}
	}

	// Validate zone enum.
	zone, ok := m["zone"].(string)
	if !ok {
		return fmt.Errorf("modules[%d]: field \"zone\" must be a string", index)
	}
	if err := validateZoneEnum(zone); err != nil {
		return fmt.Errorf("modules[%d]: %w", index, err)
	}

	// Validate extensions field if present: must be a JSON object (not primitive or array).
	if ext, exists := m["extensions"]; exists {
		if _, ok := ext.(map[string]interface{}); !ok {
			return fmt.Errorf("modules[%d]: field \"extensions\" must be a JSON object", index)
		}
	}

	// Enforce numeric ranges matching modulegraph.schema.json. This hardens the
	// validator at the trust boundary against out-of-range values from an
	// untrusted external analyzer. Ratio metrics are bounded to [0, 1]; raw
	// counts must be non-negative.
	for _, field := range []string{"instability", "abstractness", "distance"} {
		val, err := moduleNumber(m, field, index)
		if err != nil {
			return err
		}
		if val < 0.0 || val > 1.0 {
			return fmt.Errorf("modules[%d]: field %q value %g out of range [0, 1]", index, field, val)
		}
	}
	for _, field := range []string{"ca", "ce", "lcom", "exportedTypes", "abstractTypes"} {
		val, err := moduleNumber(m, field, index)
		if err != nil {
			return err
		}
		if val < 0 {
			return fmt.Errorf("modules[%d]: field %q value %g must be >= 0", index, field, val)
		}
	}

	// Validate totalLines if present: must be a non-negative integer.
	if _, exists := m["totalLines"]; exists {
		val, err := moduleNumber(m, "totalLines", index)
		if err != nil {
			return err
		}
		if val < 0 {
			return fmt.Errorf("modules[%d]: field \"totalLines\" value %g must be >= 0", index, val)
		}
	}

	// Validate duplications field if present: must be an array of valid duplication objects.
	if _, exists := m["duplications"]; exists {
		if err := validateDuplications(m, index); err != nil {
			return fmt.Errorf("modules[%d]: %w", index, err)
		}
	}

	return nil
}

// moduleNumber extracts a numeric module field as a float64. JSON unmarshaling
// represents all numbers as float64, so both integer and ratio fields are read
// through this helper. It returns an error if the field is missing or is not a
// JSON number.
func moduleNumber(m map[string]interface{}, field string, index int) (float64, error) {
	raw, ok := m[field]
	if !ok {
		return 0, fmt.Errorf("modules[%d]: missing required field %q", index, field)
	}
	num, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("modules[%d]: field %q must be a number", index, field)
	}
	return num, nil
}

// validateDuplications checks that the duplications field is an array of valid
// duplication objects. Each duplication must have modulePath (string), blocks
// (array of objects with file, startLine, endLine, lineCount), and similarity
// (number in [0, 1]).
func validateDuplications(m map[string]interface{}, moduleIndex int) error {
	dupsRaw, ok := m["duplications"]
	if !ok {
		return nil // duplications is optional
	}
	// Accept null as equivalent to an empty array (nil slices serialize as null in Go).
	if dupsRaw == nil {
		return nil
	}
	dups, ok := dupsRaw.([]interface{})
	if !ok {
		return fmt.Errorf("field \"duplications\" must be an array")
	}
	for i, d := range dups {
		dup, ok := d.(map[string]interface{})
		if !ok {
			return fmt.Errorf("duplications[%d]: must be an object", i)
		}
		// Validate required fields.
		for _, field := range []string{"modulePath", "blocks", "similarity"} {
			if _, ok := dup[field]; !ok {
				return fmt.Errorf("duplications[%d]: missing required field %q", i, field)
			}
		}
		// Validate modulePath is a string.
		if _, ok := dup["modulePath"].(string); !ok {
			return fmt.Errorf("duplications[%d]: field \"modulePath\" must be a string", i)
		}
		// Validate similarity is a number in [0, 1].
		sim, ok := dup["similarity"].(float64)
		if !ok {
			return fmt.Errorf("duplications[%d]: field \"similarity\" must be a number", i)
		}
		if sim < 0.0 || sim > 1.0 {
			return fmt.Errorf("duplications[%d]: field \"similarity\" value %g out of range [0, 1]", i, sim)
		}
		// Validate blocks is an array of valid block objects.
		blocksRaw, ok := dup["blocks"].([]interface{})
		if !ok {
			return fmt.Errorf("duplications[%d]: field \"blocks\" must be an array", i)
		}
		for j, b := range blocksRaw {
			block, ok := b.(map[string]interface{})
			if !ok {
				return fmt.Errorf("duplications[%d].blocks[%d]: must be an object", i, j)
			}
			for _, field := range []string{"file", "startLine", "endLine", "lineCount"} {
				if _, ok := block[field]; !ok {
					return fmt.Errorf("duplications[%d].blocks[%d]: missing required field %q", i, j, field)
				}
			}
			// Validate file is a string.
			if _, ok := block["file"].(string); !ok {
				return fmt.Errorf("duplications[%d].blocks[%d]: field \"file\" must be a string", i, j)
			}
			// Validate startLine, endLine, lineCount are numbers.
			for _, field := range []string{"startLine", "endLine", "lineCount"} {
				val, ok := block[field].(float64)
				if !ok {
					return fmt.Errorf("duplications[%d].blocks[%d]: field %q must be a number", i, j, field)
				}
				if val < 0 {
					return fmt.Errorf("duplications[%d].blocks[%d]: field %q value %g must be >= 0", i, j, field, val)
				}
			}
		}
	}
	return nil
}

// validateWarning checks that a warning element has the required code and message fields.
func validateWarning(v interface{}, index int) error {
	w, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("warnings[%d]: must be an object", index)
	}

	if _, ok := w["code"]; !ok {
		return fmt.Errorf("warnings[%d]: missing required field \"code\"", index)
	}
	if _, ok := w["message"]; !ok {
		return fmt.Errorf("warnings[%d]: missing required field \"message\"", index)
	}

	return nil
}

// validateProvenance checks the optional top-level provenance object. When
// present it must contain string producer, version, and generatedAt fields, and
// an input object with string path and modulePath fields. The producer, version,
// and generatedAt fields are required whenever provenance is present, matching
// the required-fields contract in modulegraph.schema.json so the hand-rolled
// validator and the embedded JSON Schema stay in lockstep.
func validateProvenance(raw map[string]interface{}) error {
	provRaw, exists := raw["provenance"]
	if !exists {
		return nil
	}
	p, ok := provRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("validate: field \"provenance\" must be an object")
	}

	for _, field := range []string{"producer", "version", "generatedAt"} {
		v, ok := p[field]
		if !ok {
			return fmt.Errorf("validate: missing required field \"provenance.%s\"", field)
		}
		if _, isString := v.(string); !isString {
			return fmt.Errorf("validate: field \"provenance.%s\" must be a string", field)
		}
	}

	inputRaw, ok := p["input"]
	if !ok {
		return fmt.Errorf("validate: missing required field \"provenance.input\"")
	}
	input, ok := inputRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("validate: field \"provenance.input\" must be an object")
	}
	for _, field := range []string{"path", "modulePath"} {
		if _, ok := input[field]; !ok {
			return fmt.Errorf("validate: missing required field \"provenance.input.%s\"", field)
		}
		if v, ok := input[field].(string); !ok {
			return fmt.Errorf("validate: field \"provenance.input.%s\" must be a string", field)
		} else if v == "" {
			return fmt.Errorf("validate: field \"provenance.input.%s\" must be non-empty", field)
		}
	}

	return nil
}
