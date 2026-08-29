package metrics

import (
	"encoding/json"
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

func TestValidate_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
	}{
		{
			name: "empty input",
			data: "",
		},
		{
			name: "malformed JSON",
			data: `{invalid}`,
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
		},
		{
			name: "missing metric field on module",
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
		},
		{
			name: "invalid zone value",
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
					"lcom": 0,
					"exportedTypes": 0,
					"abstractTypes": 0,
					"zone": "invalid-zone"
				}],
				"cycles": [],
				"warnings": [],
				"status": "complete"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate([]byte(tt.data))
			if err == nil {
				t.Errorf("Validate returned nil error for %q, want non-nil", tt.name)
			}
		})
	}
}
