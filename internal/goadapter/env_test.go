package goadapter

import (
	"slices"
	"testing"
)

// TestPackageEnv_ForcesLocalToolchain asserts the constructed subprocess
// environment always contains GOTOOLCHAIN=local, hardening analyze so a target
// module's go.mod "toolchain" directive cannot trigger a toolchain download.
func TestPackageEnv_ForcesLocalToolchain(t *testing.T) {
	t.Parallel()

	env := packageEnv()
	if !slices.Contains(env, "GOTOOLCHAIN=local") {
		t.Errorf("packageEnv() must contain %q, got %v", "GOTOOLCHAIN=local", env)
	}
}

// TestPackageEnv_HostValueDoesNotLeak asserts the forced value wins even when the
// host process sets GOTOOLCHAIN. Because GOTOOLCHAIN is not on the allowlist the
// host value is filtered out by SanitizeEnvironment, and the appended
// GOTOOLCHAIN=local is authoritative (exec resolves duplicate keys last-wins
// regardless). t.Setenv precludes t.Parallel() here.
func TestPackageEnv_HostValueDoesNotLeak(t *testing.T) {
	t.Setenv("GOTOOLCHAIN", "auto")

	env := packageEnv()
	if !slices.Contains(env, "GOTOOLCHAIN=local") {
		t.Errorf("packageEnv() must contain %q even when the host sets GOTOOLCHAIN=auto, got %v",
			"GOTOOLCHAIN=local", env)
	}
	if slices.Contains(env, "GOTOOLCHAIN=auto") {
		t.Errorf("packageEnv() must NOT contain the host value %q, got %v", "GOTOOLCHAIN=auto", env)
	}
}
