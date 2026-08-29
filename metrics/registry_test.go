package metrics

import (
	"context"
	"sync"
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
	if reg == nil {
		t.Fatal("NewRegistry() returned nil")
	}

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

// TestRegistry_ConcurrentReads verifies that concurrent Get calls on a
// pre-configured Registry are safe. This validates the documented contract
// that Registry is safe for concurrent reads after initial configuration.
func TestRegistry_ConcurrentReads(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	goAdapter := &mockAdapter{language: "go"}
	pyAdapter := &mockAdapter{language: "python"}

	if err := reg.Register(goAdapter); err != nil {
		t.Fatalf("Register(go): %v", err)
	}
	if err := reg.Register(pyAdapter); err != nil {
		t.Fatalf("Register(python): %v", err)
	}

	// Launch concurrent readers after setup is complete.
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				got, err := reg.Get("go")
				if err != nil {
					t.Errorf("concurrent Get(go): %v", err)
					return
				}
				if got != goAdapter {
					t.Errorf("concurrent Get(go) returned wrong adapter")
					return
				}
				got2, err := reg.Get("python")
				if err != nil {
					t.Errorf("concurrent Get(python): %v", err)
					return
				}
				if got2 != pyAdapter {
					t.Errorf("concurrent Get(python) returned wrong adapter")
					return
				}
			}
		}()
	}
	wg.Wait()
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
