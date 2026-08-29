package goadapter

import (
	"encoding/json"
	"testing"
)

func TestExtensions_InterfaceWidth(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "extensions", "ifaces")
	ext := computeExtensions(pkg)

	if ext == nil {
		t.Fatal("extensions is nil, want non-nil")
	}

	widths, err := InterfaceWidths(ext)
	if err != nil {
		t.Fatalf("InterfaceWidths: %v", err)
	}

	cases := map[string]int{
		"Reader":    1,
		"Processor": 3,
		"Embedder":  2,
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := widths[name]
			if !ok {
				t.Fatalf("width for %q not found", name)
			}
			if got != want {
				t.Errorf("width for %q: got %d, want %d", name, got, want)
			}
		})
	}
}

func TestExtensions_InterfaceProximity(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "extensions", "ifaces")
	ext := computeExtensions(pkg)

	if ext == nil {
		t.Fatal("extensions is nil, want non-nil")
	}

	proximities, err := InterfaceProximities(ext)
	if err != nil {
		t.Fatalf("InterfaceProximities: %v", err)
	}

	cases := map[string]string{
		"Reader":    "producer",
		"Processor": "consumer",
		"Embedder":  "producer",
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := proximities[name]
			if !ok {
				t.Fatalf("proximity for %q not found", name)
			}
			if got != want {
				t.Errorf("proximity for %q: got %q, want %q", name, got, want)
			}
		})
	}
}

func TestInterfaceWidths_RoundTrip(t *testing.T) {
	t.Parallel()

	// Simulate JSON round-trip: int values become float64.
	original := map[string]any{
		"go.interfaceWidth": map[string]int{
			"Reader": 1,
			"Writer": 2,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	widths, err := InterfaceWidths(roundTripped)
	if err != nil {
		t.Fatalf("InterfaceWidths after round-trip: %v", err)
	}

	if widths["Reader"] != 1 {
		t.Errorf("Reader width: got %d, want %d", widths["Reader"], 1)
	}
	if widths["Writer"] != 2 {
		t.Errorf("Writer width: got %d, want %d", widths["Writer"], 2)
	}
}

func TestInterfaceProximities_RoundTrip(t *testing.T) {
	t.Parallel()

	original := map[string]any{
		"go.interfaceProximity": map[string]string{
			"Reader": "producer",
			"Writer": "consumer",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	proximities, err := InterfaceProximities(roundTripped)
	if err != nil {
		t.Fatalf("InterfaceProximities after round-trip: %v", err)
	}

	if proximities["Reader"] != "producer" {
		t.Errorf("Reader proximity: got %q, want %q", proximities["Reader"], "producer")
	}
	if proximities["Writer"] != "consumer" {
		t.Errorf("Writer proximity: got %q, want %q", proximities["Writer"], "consumer")
	}
}

func TestInterfaceWidths_MissingKey(t *testing.T) {
	t.Parallel()

	ext := map[string]any{
		"other.key": "value",
	}

	_, err := InterfaceWidths(ext)
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestInterfaceProximities_MissingKey(t *testing.T) {
	t.Parallel()

	ext := map[string]any{
		"other.key": "value",
	}

	_, err := InterfaceProximities(ext)
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}
