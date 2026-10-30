package coverratchet

import (
	"fmt"
	"path"
	"strings"

	"golang.org/x/tools/cover"
)

// ModulePath is the module path prefix that every coverage profile file
// name must carry. The gate keys packages by the module-relative path
// between this prefix and the file name.
const ModulePath = "github.com/zero-dot-force/vibe-check/"

// LoadProfiles parses the Go coverage profile stored at path and returns
// one [cover.Profile] per source file, sorted by file name. A missing or
// malformed profile is a tool error.
func LoadProfiles(path string) ([]*cover.Profile, error) {
	profiles, err := cover.ParseProfiles(path)
	if err != nil {
		return nil, fmt.Errorf("coverratchet: parse coverage profile %s: %w", path, err)
	}
	return profiles, nil
}

// PackageKey derives the module-relative package path for a coverage
// profile file name by stripping the module path prefix and the file name
// (e.g. "github.com/zero-dot-force/vibe-check/metrics/values.go" becomes
// "metrics"). A file name that does not carry the module prefix is a tool
// error.
func PackageKey(fileName string) (string, error) {
	after, ok := strings.CutPrefix(fileName, ModulePath)
	if !ok {
		return "", fmt.Errorf("coverratchet: profile file %q does not carry module prefix %q", fileName, ModulePath)
	}
	return path.Dir(after), nil
}
