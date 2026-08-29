package metrics

import (
	"context"
	"testing"
)

// Compile-time interface satisfaction check (Task 5.5).
var _ Adapter = (*mockAdapter)(nil)

// mockAdapter is a minimal Adapter implementation for testing Registry behavior.
type mockAdapter struct {
	language     string
	capabilities []Capability
}

func (m *mockAdapter) Analyze(_ context.Context, _ string) (*ModuleGraph, error) {
	return &ModuleGraph{Language: m.language, Status: StatusComplete}, nil
}

func (m *mockAdapter) Language() string {
	return m.language
}

func (m *mockAdapter) Capabilities() []Capability {
	return m.capabilities
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	adapter := &mockAdapter{
		language:     "go",
		capabilities: []Capability{CapAfferentCoupling, CapEfferentCoupling},
	}

	if err := reg.Register(adapter); err != nil {
		t.Fatalf("Register() returned unexpected error: %v", err)
	}

	got, err := reg.Get("go")
	if err != nil {
		t.Fatalf("Get(%q) returned unexpected error: %v", "go", err)
	}
	if got != adapter {
		t.Errorf("Get(%q) = %v, want %v", "go", got, adapter)
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	adapter1 := &mockAdapter{language: "go"}
	adapter2 := &mockAdapter{language: "go"}

	if err := reg.Register(adapter1); err != nil {
		t.Fatalf("Register() first call returned unexpected error: %v", err)
	}

	err := reg.Register(adapter2)
	if err == nil {
		t.Fatal("Register() duplicate language returned nil error, want error")
	}
}

func TestRegistry_UnknownLanguage(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()

	_, err := reg.Get("rust")
	if err == nil {
		t.Fatalf("Get(%q) returned nil error for unregistered language, want error", "rust")
	}
}

func TestRegistry_MultipleLanguages(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	goAdapter := &mockAdapter{language: "go"}
	pyAdapter := &mockAdapter{language: "python"}

	if err := reg.Register(goAdapter); err != nil {
		t.Fatalf("Register(go) returned unexpected error: %v", err)
	}
	if err := reg.Register(pyAdapter); err != nil {
		t.Fatalf("Register(python) returned unexpected error: %v", err)
	}

	gotGo, err := reg.Get("go")
	if err != nil {
		t.Fatalf("Get(%q) returned unexpected error: %v", "go", err)
	}
	if gotGo != goAdapter {
		t.Errorf("Get(%q) = %v, want %v", "go", gotGo, goAdapter)
	}

	gotPy, err := reg.Get("python")
	if err != nil {
		t.Fatalf("Get(%q) returned unexpected error: %v", "python", err)
	}
	if gotPy != pyAdapter {
		t.Errorf("Get(%q) = %v, want %v", "python", gotPy, pyAdapter)
	}
}
