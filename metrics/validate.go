package metrics

import (
	"encoding/json"
	"fmt"
)

// Validate checks whether the given JSON data conforms to the ModuleGraph schema.
// It verifies required fields, value types, and enum constraints without relying
// on an external JSON Schema validation library.
// Returns nil if valid, or an error describing the first validation failure.
func Validate(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("validate: empty input")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// Check required top-level fields.
	requiredFields := []string{"schemaVersion", "language", "modules", "cycles", "warnings", "status"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			return fmt.Errorf("validate: missing required field %q", field)
		}
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

	// Validate modules is an array.
	modulesRaw, ok := raw["modules"].([]interface{})
	if !ok {
		return fmt.Errorf("validate: field \"modules\" must be an array")
	}
	for i, m := range modulesRaw {
		if err := validateModule(m, i); err != nil {
			return fmt.Errorf("validate: %w", err)
		}
	}

	// Validate cycles is an array (not null).
	if _, ok := raw["cycles"].([]interface{}); !ok {
		return fmt.Errorf("validate: field \"cycles\" must be an array")
	}

	// Validate warnings is an array (not null).
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
