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

	return validateWarnings(raw)
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
	// Accept both "1.0" (no extensions) and "1.1" (with extensions) for backward compatibility.
	version, ok := raw["schemaVersion"].(string)
	if !ok {
		return fmt.Errorf("validate: field \"schemaVersion\" must be a string")
	}
	if version != "1.0" && version != "1.1" {
		return fmt.Errorf("validate: unsupported schema version %q (supported: \"1.0\", \"1.1\")", version)
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
