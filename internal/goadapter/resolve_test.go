package goadapter

import "testing"

// TestPackageEnvAllowlist_ExcludesInjectionVectors locks the go/packages
// environment allowlist against command-injection. GOFLAGS must never be
// present (it can inject arbitrary build flags into the subprocess that
// go/packages spawns), while the variables required for module resolution and
// build-cache location must be. Changing this set requires consciously updating
// this test.
func TestPackageEnvAllowlist_ExcludesInjectionVectors(t *testing.T) {
	t.Parallel()

	present := make(map[string]bool, len(packageEnvAllowlist))
	for _, v := range packageEnvAllowlist {
		present[v] = true
	}

	// Must NOT contain the command-injection vector.
	if present["GOFLAGS"] {
		t.Error(`packageEnvAllowlist must not contain "GOFLAGS" (command-injection vector)`)
	}

	// Must contain exactly the expected module-resolution variables.
	expected := []string{"GOPATH", "GOROOT", "GOMODCACHE", "GOPROXY", "GONOSUMCHECK", "GOMOD"}
	for _, want := range expected {
		if !present[want] {
			t.Errorf("packageEnvAllowlist missing required entry %q", want)
		}
	}

	if len(packageEnvAllowlist) != len(expected) {
		t.Errorf("packageEnvAllowlist size: got %d, want %d (unexpected entries lock out review)",
			len(packageEnvAllowlist), len(expected))
	}
}
