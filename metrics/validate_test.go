package metrics

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidate_RoundTrip(t *testing.T) {
	t.Parallel()

	original := ModuleGraph{
		SchemaVersion: "1.0",
		Language:      "go",
		Modules: []ModuleResult{
			{
				Module: Module{
					Path:          "github.com/example/foo",
					Name:          "foo",
					Ca:            3,
					Ce:            2,
					ExportedTypes: 5,
					AbstractTypes: 1,
				},
				Instability:  0.4,
				Abstractness: 0.2,
				Distance:     0.4,
				LCOM:         1,
				Zone:         ZoneMainSequence,
			},
		},
		Cycles:   []Cycle{},
		Warnings: []Warning{},
		Status:   StatusComplete,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	if err := Validate(data); err != nil {
		t.Fatalf("Validate returned error for valid input: %v", err)
	}

	var decoded ModuleGraph
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify top-level fields.
	if got, want := decoded.SchemaVersion, original.SchemaVersion; got != want {
		t.Errorf("SchemaVersion: got %v, want %v", got, want)
	}
	if got, want := decoded.Language, original.Language; got != want {
		t.Errorf("Language: got %v, want %v", got, want)
	}
	if got, want := decoded.Status, original.Status; got != want {
		t.Errorf("Status: got %v, want %v", got, want)
	}
	if got, want := len(decoded.Modules), len(original.Modules); got != want {
		t.Fatalf("len(Modules): got %v, want %v", got, want)
	}
	if got, want := len(decoded.Cycles), len(original.Cycles); got != want {
		t.Errorf("len(Cycles): got %v, want %v", got, want)
	}
	if got, want := len(decoded.Warnings), len(original.Warnings); got != want {
		t.Errorf("len(Warnings): got %v, want %v", got, want)
	}

	// Verify module fields.
	m := decoded.Modules[0]
	om := original.Modules[0]
	if got, want := m.Path, om.Path; got != want {
		t.Errorf("Module.Path: got %v, want %v", got, want)
	}
	if got, want := m.Name, om.Name; got != want {
		t.Errorf("Module.Name: got %v, want %v", got, want)
	}
	if got, want := m.Ca, om.Ca; got != want {
		t.Errorf("Module.Ca: got %v, want %v", got, want)
	}
	if got, want := m.Ce, om.Ce; got != want {
		t.Errorf("Module.Ce: got %v, want %v", got, want)
	}
	if got, want := m.ExportedTypes, om.ExportedTypes; got != want {
		t.Errorf("Module.ExportedTypes: got %v, want %v", got, want)
	}
	if got, want := m.AbstractTypes, om.AbstractTypes; got != want {
		t.Errorf("Module.AbstractTypes: got %v, want %v", got, want)
	}
	if got, want := m.Instability, om.Instability; got != want {
		t.Errorf("Module.Instability: got %v, want %v", got, want)
	}
	if got, want := m.Abstractness, om.Abstractness; got != want {
		t.Errorf("Module.Abstractness: got %v, want %v", got, want)
	}
	if got, want := m.Distance, om.Distance; got != want {
		t.Errorf("Module.Distance: got %v, want %v", got, want)
	}
	if got, want := m.LCOM, om.LCOM; got != want {
		t.Errorf("Module.LCOM: got %v, want %v", got, want)
	}
	if got, want := m.Zone, om.Zone; got != want {
		t.Errorf("Module.Zone: got %v, want %v", got, want)
	}
}

func TestValidate_ZeroMetricsSerialized(t *testing.T) {
	t.Parallel()

	// Verify that zero-value numeric metrics are present in JSON output
	// (no omitempty on metric fields).
	g := ModuleGraph{
		SchemaVersion: "1.0",
		Language:      "go",
		Modules: []ModuleResult{
			{
				Module: Module{
					Path: "github.com/example/empty",
					Name: "empty",
					// All numeric fields are zero.
				},
				Zone: ZoneNormal,
			},
		},
		Cycles:   []Cycle{},
		Warnings: []Warning{},
		Status:   StatusComplete,
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Validate passes for zero-valued metrics.
	if err := Validate(data); err != nil {
		t.Fatalf("Validate returned error for zero-valued metrics: %v", err)
	}

	// Verify zero values are present in JSON (not omitted).
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}
	modules := raw["modules"].([]interface{})
	mod := modules[0].(map[string]interface{})

	zeroFields := []string{"ca", "ce", "instability", "abstractness", "distance", "lcom", "exportedTypes", "abstractTypes"}
	for _, field := range zeroFields {
		if _, ok := mod[field]; !ok {
			t.Errorf("zero-valued field %q was omitted from JSON output", field)
		}
	}
}

func TestValidate_ExtensionsRoundTrip(t *testing.T) {
	t.Parallel()

	original := ModuleGraph{
		SchemaVersion: "1.1",
		Language:      "go",
		Modules: []ModuleResult{
			{
				Module: Module{
					Path:          "github.com/example/ext",
					Name:          "ext",
					Ca:            1,
					Ce:            2,
					ExportedTypes: 3,
					AbstractTypes: 1,
				},
				Instability:  0.666666,
				Abstractness: 0.333333,
				Distance:     0.0,
				LCOM:         1,
				Zone:         ZoneMainSequence,
				Extensions: map[string]any{
					"go.interfaceWidth":     map[string]int{"Reader": 1, "Writer": 2},
					"go.interfaceProximity": map[string]string{"Reader": "producer", "Writer": "consumer"},
				},
			},
		},
		Cycles:   []Cycle{},
		Warnings: []Warning{},
		Status:   StatusComplete,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Validate passes with extensions present.
	if err := Validate(data); err != nil {
		t.Fatalf("Validate returned error for valid input with extensions: %v", err)
	}

	// Verify JSON round-trip type behavior: int becomes float64, nested maps
	// become map[string]interface{}.
	var decoded ModuleGraph
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	ext := decoded.Modules[0].Extensions
	if ext == nil {
		t.Fatal("Extensions is nil after round-trip")
	}

	// After JSON round-trip, map[string]int becomes map[string]interface{} with float64 values.
	widthsRaw, ok := ext["go.interfaceWidth"]
	if !ok {
		t.Fatal("go.interfaceWidth missing from extensions")
	}
	widths, ok := widthsRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("go.interfaceWidth: got type %T, want map[string]interface{}", widthsRaw)
	}
	if got, want := widths["Reader"], float64(1); got != want {
		t.Errorf("go.interfaceWidth[Reader]: got %v (%T), want %v (%T)", got, got, want, want)
	}
	if got, want := widths["Writer"], float64(2); got != want {
		t.Errorf("go.interfaceWidth[Writer]: got %v (%T), want %v (%T)", got, got, want, want)
	}

	// map[string]string round-trips to map[string]interface{} with string values.
	proxRaw, ok := ext["go.interfaceProximity"]
	if !ok {
		t.Fatal("go.interfaceProximity missing from extensions")
	}
	prox, ok := proxRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("go.interfaceProximity: got type %T, want map[string]interface{}", proxRaw)
	}
	if got, want := prox["Reader"], "producer"; got != want {
		t.Errorf("go.interfaceProximity[Reader]: got %v, want %v", got, want)
	}
}

func TestValidate_ExtensionsOmitted(t *testing.T) {
	t.Parallel()

	// ModuleResult without extensions should serialize without the extensions field
	// (omitempty) and still pass validation.
	g := ModuleGraph{
		SchemaVersion: "1.1",
		Language:      "go",
		Modules: []ModuleResult{
			{
				Module: Module{
					Path:          "github.com/example/noext",
					Name:          "noext",
					Ca:            0,
					Ce:            0,
					ExportedTypes: 0,
					AbstractTypes: 0,
				},
				Zone: ZoneNormal,
			},
		},
		Cycles:   []Cycle{},
		Warnings: []Warning{},
		Status:   StatusComplete,
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	if err := Validate(data); err != nil {
		t.Fatalf("Validate returned error for input without extensions: %v", err)
	}

	// Verify extensions field is not present in JSON.
	if strings.Contains(string(data), `"extensions"`) {
		t.Error("extensions field should be omitted from JSON when nil")
	}
}

func TestValidate_SchemaVersion11(t *testing.T) {
	t.Parallel()

	// Validate accepts schema version "1.1".
	data := []byte(`{
		"schemaVersion": "1.1",
		"language": "go",
		"modules": [],
		"cycles": [],
		"warnings": [],
		"status": "complete"
	}`)
	if err := Validate(data); err != nil {
		t.Fatalf("Validate rejected schema version 1.1: %v", err)
	}
}

func TestValidate_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		wantErr string
	}{
		{
			name:    "empty input",
			data:    "",
			wantErr: "empty input",
		},
		{
			name:    "malformed JSON",
			data:    `{invalid}`,
			wantErr: "validate:",
		},
		{
			name: "missing language field",
			data: `{
				"schemaVersion": "1.0",
				"modules": [],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "missing required field \"language\"",
		},
		{
			name: "null warnings",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": null,
				"status": "complete"
			}`,
			wantErr: "\"warnings\" must be an array",
		},
		{
			name: "unsupported schema version",
			data: `{
				"schemaVersion": "2.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "unsupported schema version",
		},
		{
			name: "invalid status value",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": [],
				"status": "unknown"
			}`,
			wantErr: "invalid status",
		},
		{
			name: "missing lcom field on module",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0,
					"ce": 0,
					"instability": 0,
					"abstractness": 0,
					"distance": 0,
					"exportedTypes": 0,
					"abstractTypes": 0,
					"zone": "normal"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "missing required field \"lcom\"",
		},
		{
			name: "missing path field on module",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [{
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "missing required field \"path\"",
		},
		{
			name: "invalid zone value",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "invalid-zone"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "invalid zone",
		},
		{
			name: "extensions is a string (invalid)",
			data: `{
				"schemaVersion": "1.1",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal",
					"extensions": "not-an-object"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "\"extensions\" must be a JSON object",
		},
		{
			name: "extensions is an array (invalid)",
			data: `{
				"schemaVersion": "1.1",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal",
					"extensions": [1, 2, 3]
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "\"extensions\" must be a JSON object",
		},
		{
			name: "extensions is a valid object (passes)",
			data: `{
				"schemaVersion": "1.1",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal",
					"extensions": {"go.interfaceWidth": {"Foo": 1}}
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "",
		},
		{
			name: "warning missing code field",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": [{"message": "test warning"}],
				"status": "complete"
			}`,
			wantErr: "missing required field \"code\"",
		},
		{
			name: "warning missing message field",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": [{"code": "W001"}],
				"status": "complete"
			}`,
			wantErr: "missing required field \"message\"",
		},
		{
			name: "warning is not an object",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": ["not-an-object"],
				"status": "complete"
			}`,
			wantErr: "must be an object",
		},
		{
			name: "valid warning passes",
			data: `{
				"schemaVersion": "1.0",
				"language": "go",
				"modules": [],
				"cycles": [],
				"warnings": [{"code": "W001", "message": "test"}],
				"status": "complete"
			}`,
			wantErr: "",
		},
		{
			name: "instability above 1 (out of range)",
			data: `{
				"schemaVersion": "1.1",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 1.5, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "out of range",
		},
		{
			name: "distance below 0 (out of range)",
			data: `{
				"schemaVersion": "1.1",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": 0, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": -0.5, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "out of range",
		},
		{
			name: "negative ca (must be >= 0)",
			data: `{
				"schemaVersion": "1.1",
				"language": "go",
				"modules": [{
					"path": "foo",
					"name": "foo",
					"ca": -1, "ce": 0,
					"instability": 0, "abstractness": 0, "distance": 0, "lcom": 0,
					"exportedTypes": 0, "abstractTypes": 0,
					"zone": "normal"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
			wantErr: "must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate([]byte(tt.data))
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate returned unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate returned nil error, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
