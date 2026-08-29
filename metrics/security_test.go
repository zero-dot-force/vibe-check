package metrics

import (
	"os"
	"strings"
	"testing"
)

func TestValidateProjectPath_ValidDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := ValidateProjectPath(dir); err != nil {
		t.Errorf("ValidateProjectPath(%q) returned unexpected error: %v", dir, err)
	}
}

func TestValidateProjectPath_InvalidPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: "path is empty",
		},
		{
			name:    "path with traversal",
			path:    "/tmp/../etc/passwd",
			wantErr: "\"..\" traversal",
		},
		{
			name:    "nonexistent path",
			path:    "/nonexistent/path/that/does/not/exist",
			wantErr: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateProjectPath(tt.path)
			if err == nil {
				t.Fatalf("ValidateProjectPath(%q) returned nil error, want error containing %q", tt.path, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateProjectPath(%q) error = %v, want error containing %q", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateProjectPath_FileNotDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := dir + "/testfile.txt"
	if err := os.WriteFile(file, []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err := ValidateProjectPath(file)
	if err == nil {
		t.Fatal("ValidateProjectPath(file) returned nil error, want error for non-directory")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error = %v, want error containing \"not a directory\"", err)
	}
}

// TestSanitizeEnvironment_DefaultsIncluded verifies that PATH, HOME, and LANG
// are included in the sanitized environment. Not parallel because t.Setenv
// modifies process-level state.
func TestSanitizeEnvironment_DefaultsIncluded(t *testing.T) {
	t.Setenv("PATH", "/usr/bin:/bin")
	t.Setenv("HOME", "/home/testuser")
	t.Setenv("LANG", "en_US.UTF-8")

	env := SanitizeEnvironment(nil)

	found := make(map[string]string)
	for _, entry := range env {
		key, val, _ := strings.Cut(entry, "=")
		found[key] = val
	}

	for _, required := range []string{"PATH", "HOME", "LANG"} {
		if _, ok := found[required]; !ok {
			t.Errorf("SanitizeEnvironment(nil) missing required variable %q", required)
		}
	}

	// Verify values are preserved correctly.
	expected := map[string]string{
		"PATH": "/usr/bin:/bin",
		"HOME": "/home/testuser",
		"LANG": "en_US.UTF-8",
	}
	for key, wantVal := range expected {
		if gotVal, ok := found[key]; ok && gotVal != wantVal {
			t.Errorf("SanitizeEnvironment variable %q: got %q, want %q", key, gotVal, wantVal)
		}
	}
}

// TestSanitizeEnvironment_CredentialsExcluded verifies that credential-bearing
// variables are excluded from the sanitized environment. Not parallel because
// t.Setenv modifies process-level state.
func TestSanitizeEnvironment_CredentialsExcluded(t *testing.T) {
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	t.Setenv("GITHUB_TOKEN", "ghp_test")
	t.Setenv("GH_TOKEN", "gho_test")
	t.Setenv("NPM_TOKEN", "npm_test")
	t.Setenv("SECRET_TEST_VALUE", "should-not-appear")
	t.Setenv("PATH", "/usr/bin")

	env := SanitizeEnvironment(nil)

	blocked := []string{
		"AWS_SECRET_ACCESS_KEY",
		"GITHUB_TOKEN",
		"GH_TOKEN",
		"NPM_TOKEN",
		"SECRET_TEST_VALUE",
	}

	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		for _, b := range blocked {
			if key == b {
				t.Errorf("SanitizeEnvironment(nil) included blocked variable %q", b)
			}
		}
	}
}

// TestSanitizeEnvironment_AllowlistIncluded verifies that explicitly allowlisted
// variables are included. Not parallel because t.Setenv modifies process-level state.
func TestSanitizeEnvironment_AllowlistIncluded(t *testing.T) {
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("CUSTOM_VAR", "custom_value")

	env := SanitizeEnvironment([]string{"CUSTOM_VAR"})

	found := false
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if key == "CUSTOM_VAR" {
			found = true
			break
		}
	}

	if !found {
		t.Error("SanitizeEnvironment([\"CUSTOM_VAR\"]) did not include allowlisted variable")
	}
}

// TestSanitizeEnvironment_AllowlistBlockedOverride verifies that even explicitly
// allowlisted variables are excluded if they match the blocked list. Not parallel
// because t.Setenv modifies process-level state.
func TestSanitizeEnvironment_AllowlistBlockedOverride(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "ghp_test")
	t.Setenv("PATH", "/usr/bin")

	env := SanitizeEnvironment([]string{"GITHUB_TOKEN"})

	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if key == "GITHUB_TOKEN" {
			t.Error("SanitizeEnvironment allowlisted GITHUB_TOKEN but it should be blocked")
		}
	}
}

func TestIsBlockedEnv_Cases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		envVar  string
		blocked bool
	}{
		{name: "exact match", envVar: "GITHUB_TOKEN", blocked: true},
		{name: "prefix match", envVar: "AWS_SECRET_ACCESS_KEY", blocked: true},
		{name: "prefix match secret", envVar: "SECRET_MY_VALUE", blocked: true},
		{name: "safe variable", envVar: "PATH", blocked: false},
		{name: "safe variable HOME", envVar: "HOME", blocked: false},
		{name: "case insensitive", envVar: "github_token", blocked: true},
		{name: "unrelated variable", envVar: "GOPATH", blocked: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isBlockedEnv(tt.envVar)
			if got != tt.blocked {
				t.Errorf("isBlockedEnv(%q): got %v, want %v", tt.envVar, got, tt.blocked)
			}
		})
	}
}
