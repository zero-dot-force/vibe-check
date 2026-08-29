package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// blockedEnvPrefixes contains environment variable prefixes that MUST NOT be
// passed to analyzer subprocesses. These prefixes cover credential-bearing
// variables across common CI systems, cloud providers, and secret managers.
var blockedEnvPrefixes = []string{
	"AWS_SECRET",
	"AWS_SESSION",
	"AZURE_",
	"ARM_",
	"GOOGLE_APPLICATION_CREDENTIALS",
	"GCLOUD_",
	"GITHUB_TOKEN",
	"GH_TOKEN",
	"GITLAB_TOKEN",
	"NPM_TOKEN",
	"DOCKER_PASSWORD",
	"SECRET_",
	"TOKEN_",
	"PASSWORD_",
	"PRIVATE_KEY",
	"API_KEY",
	"CREDENTIALS",
}

// blockedEnvExact contains environment variable names that MUST NOT be passed
// to analyzer subprocesses.
var blockedEnvExact = []string{
	"AWS_SECRET_ACCESS_KEY",
	"AWS_SESSION_TOKEN",
	"GITHUB_TOKEN",
	"GH_TOKEN",
	"GITLAB_TOKEN",
	"NPM_TOKEN",
	"DOCKER_PASSWORD",
	"SSH_AUTH_SOCK",
	"SSH_AGENT_PID",
	"DATABASE_URL",
	"REDIS_URL",
}

// allowedEnvDefaults contains the environment variables included by default
// in the sanitized environment. These provide basic system context without
// exposing credentials.
var allowedEnvDefaults = []string{
	"PATH",
	"HOME",
	"LANG",
}

// ValidateProjectPath checks that a project path is safe to pass to an analyzer.
// It rejects paths containing ".." traversal components, resolves symlinks to
// prevent symlink-based traversal, and verifies the path exists and is a directory.
func ValidateProjectPath(path string) error {
	if path == "" {
		return fmt.Errorf("validate project path: path is empty")
	}

	// Check for path traversal components in the original path before cleaning.
	// filepath.Clean resolves ".." components, so we must check the raw input
	// to detect traversal attempts like "/tmp/../etc/passwd".
	//
	// Split on both '/' and the platform separator to handle Windows-style
	// paths on all platforms. On Unix filepath.Separator is '/', so splitParts
	// handles both cases uniformly.
	normalized := strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(normalized, "/")
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("validate project path: path contains \"..\" traversal component: %s", path)
		}
	}

	cleaned := filepath.Clean(path)

	// Resolve symlinks to prevent symlink-based traversal where a symlink
	// at an innocent path points to a sensitive directory.
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("validate project path: path does not exist: %s", path)
		}
		return fmt.Errorf("validate project path: resolve symlinks: %w", err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("validate project path: path does not exist: %s", path)
		}
		return fmt.Errorf("validate project path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("validate project path: path is not a directory: %s", path)
	}

	return nil
}

// SanitizeEnvironment constructs a minimal environment for an analyzer subprocess.
// It includes only PATH, HOME, and LANG from the host environment, plus any
// explicitly allowlisted variables. It never includes credential-bearing variables
// (AWS secrets, tokens, passwords, private keys).
//
// The allowlist parameter specifies additional environment variable names to
// include beyond the defaults. Allowlisted variables are still checked against
// the blocked list — a variable on both lists is excluded.
func SanitizeEnvironment(allowlist []string) []string {
	// Build the set of desired variable names.
	wanted := make(map[string]bool, len(allowedEnvDefaults)+len(allowlist))
	for _, name := range allowedEnvDefaults {
		wanted[name] = true
	}
	for _, name := range allowlist {
		wanted[name] = true
	}

	// Remove any blocked variables from the wanted set.
	for name := range wanted {
		if isBlockedEnv(name) {
			delete(wanted, name)
		}
	}

	// Collect matching variables from the host environment.
	var result []string
	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if wanted[key] {
			result = append(result, entry)
		}
	}

	return result
}

// isBlockedEnv checks whether an environment variable name matches any blocked
// prefix or exact name. This is a security boundary — err on the side of blocking.
func isBlockedEnv(name string) bool {
	upper := strings.ToUpper(name)

	for _, exact := range blockedEnvExact {
		if upper == exact {
			return true
		}
	}

	for _, prefix := range blockedEnvPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}

	return false
}
