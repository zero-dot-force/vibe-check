package goadapter

import (
	"encoding/json"
	"strings"
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

// TestAdapter_ExtensionCapabilities verifies the exported extension capability
// constants and the accessor that surfaces them. These are separate from the
// seven universal capabilities returned by Capabilities.
func TestAdapter_ExtensionCapabilities(t *testing.T) {
	t.Parallel()

	if CapInterfaceWidth != "go.interfaceWidth" {
		t.Errorf("CapInterfaceWidth: got %q, want %q", CapInterfaceWidth, "go.interfaceWidth")
	}
	if CapInterfaceProximity != "go.interfaceProximity" {
		t.Errorf("CapInterfaceProximity: got %q, want %q", CapInterfaceProximity, "go.interfaceProximity")
	}

	adapter := New()
	caps := adapter.ExtensionCapabilities()
	if len(caps) != 2 {
		t.Fatalf("ExtensionCapabilities count: got %d, want 2", len(caps))
	}

	want := map[string]bool{CapInterfaceWidth: true, CapInterfaceProximity: true}
	for _, c := range caps {
		if !want[c] {
			t.Errorf("unexpected extension capability: %q", c)
		}
		delete(want, c)
	}
	if len(want) != 0 {
		t.Errorf("missing extension capabilities: %v", want)
	}
}

// TestInterfaceWidths_WrongContainerType verifies that a non-map value under
// the interfaceWidth key produces the "expected map" error.
func TestInterfaceWidths_WrongContainerType(t *testing.T) {
	t.Parallel()

	ext := map[string]any{CapInterfaceWidth: "not-a-map"}

	_, err := InterfaceWidths(ext)
	if err == nil {
		t.Fatal("expected error for wrong container type, got nil")
	}
	if !strings.Contains(err.Error(), "expected map") {
		t.Errorf("error %q does not contain %q", err.Error(), "expected map")
	}
}

// TestInterfaceWidths_WrongValueType verifies that a map value whose element is
// neither float64 nor int produces the "unexpected type" error.
func TestInterfaceWidths_WrongValueType(t *testing.T) {
	t.Parallel()

	ext := map[string]any{
		CapInterfaceWidth: map[string]interface{}{"Reader": "not-a-number"},
	}

	_, err := InterfaceWidths(ext)
	if err == nil {
		t.Fatal("expected error for wrong value type, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected type") {
		t.Errorf("error %q does not contain %q", err.Error(), "unexpected type")
	}
}

// TestInterfaceProximities_WrongContainerType verifies that a non-map value
// under the interfaceProximity key produces the "expected map" error.
func TestInterfaceProximities_WrongContainerType(t *testing.T) {
	t.Parallel()

	ext := map[string]any{CapInterfaceProximity: 42}

	_, err := InterfaceProximities(ext)
	if err == nil {
		t.Fatal("expected error for wrong container type, got nil")
	}
	if !strings.Contains(err.Error(), "expected map") {
		t.Errorf("error %q does not contain %q", err.Error(), "expected map")
	}
}

// TestInterfaceProximities_WrongValueType verifies that a map value whose
// element is not a string produces the "unexpected type" error.
func TestInterfaceProximities_WrongValueType(t *testing.T) {
	t.Parallel()

	ext := map[string]any{
		CapInterfaceProximity: map[string]interface{}{"Reader": 123},
	}

	_, err := InterfaceProximities(ext)
	if err == nil {
		t.Fatal("expected error for wrong value type, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected type") {
		t.Errorf("error %q does not contain %q", err.Error(), "unexpected type")
	}
}
