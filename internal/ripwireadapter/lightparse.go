package ripwireadapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// lightParseResult carries the per-directory counts a light source scan
// contributes to module metrics.
type lightParseResult struct {
	// AbstractTypes is the number of class declarations in the directory
	// whose bases reference an abstract base (abc.ABC, ABCMeta, or
	// typing.Protocol, in dotted or bare imported form).
	AbstractTypes int

	// ExportedTypes is the number of class declarations whose names are
	// not underscore-prefixed, regardless of abstractness.
	ExportedTypes int
}

// pyClassRegex matches a top-level Python class declaration line. Base
// lists containing nested parentheses (e.g. class aliases produced by
// decorators) are a documented light-parse limitation: the bases are cut
// at the first closing parenthesis. Decorators are ignored — the class
// line itself carries the declaration.
var pyClassRegex = regexp.MustCompile(`^class\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(([^)]*)\))?\s*:`)

// pythonAbstractnessScan walks the crawl root's Python sources and
// counts, per root-relative directory, the abstract and exported class
// declarations (design D8: ripwire reports structurally zero Python
// abstractness — its recorded arch.txt shows abstract="0" for every
// module — so vibe-check's own light parse is the only source of that
// metric for Python).
//
// The scan is deliberately light: a line-based regular-expression pass,
// not an AST. Limitations, all documented rather than silently wrong:
//   - a base list continued on following lines is read from the opening
//     line only;
//   - bases with nested parentheses are cut at the first ")";
//   - string literals that begin a line with class-like syntax inside a
//     triple-quoted block can misreport (no such construct appears in
//     supported idioms at top level).
//
// Directories ripwire prunes from the crawl (the shared denylist names)
// are skipped so pruned code contributes no counts, matching what the
// coupling data sees; the crawl root itself is never skipped.
//
// Degradation (task 5.3): an unreadable file or directory does not abort
// the scan — it produces a parse-error warning naming the path and
// contributes zeroed metrics, and the any-warning rule forces the final
// graph to partial status. Only a missing or non-directory crawl root
// fails closed.
func pythonAbstractnessScan(crawlRoot string) (map[string]lightParseResult, []metrics.Warning, error) {
	info, err := os.Stat(crawlRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("python abstractness scan: stat %s: %w", crawlRoot, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("python abstractness scan: %s: %w", crawlRoot, errNotDirectory)
	}

	deny := make(map[string]bool, len(crawlDenylistNames))
	for _, name := range crawlDenylistNames {
		deny[name] = true
	}

	var warnings []metrics.Warning
	results := make(map[string]lightParseResult)
	err = filepath.WalkDir(crawlRoot, func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			warnings = append(warnings, unreadablePathWarning(crawlRoot, filePath, walkErr))
			return nil // skip the unreadable subtree
		}
		if d.IsDir() {
			if filePath != crawlRoot && deny[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".py" {
			return nil
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			warnings = append(warnings, unreadablePathWarning(crawlRoot, filePath, err))
			return nil // zeroed metrics for this file
		}
		dir := dirRelativeToRoot(crawlRoot, filePath)
		result := results[dir]
		for _, class := range scanPythonClasses(string(data)) {
			if class.abstract {
				result.AbstractTypes++
			}
			if class.exported {
				result.ExportedTypes++
			}
		}
		results[dir] = result
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return results, warnings, nil
}

// warningParseError is the warning code for a source path the light
// parse could not read: the file contributes zeroed metrics and the
// graph is partial.
const warningParseError = "parse-error"

// unreadablePathWarning renders the degradation warning for an
// unreadable file or directory encountered during a light scan.
func unreadablePathWarning(crawlRoot, path string, err error) metrics.Warning {
	rel := path
	if r, relErr := filepath.Rel(crawlRoot, path); relErr == nil {
		rel = filepath.ToSlash(r)
	}
	return metrics.Warning{
		Code: warningParseError,
		Message: fmt.Sprintf(
			"unreadable source path skipped, its metrics are zeroed: %s: %v",
			rel, err,
		),
	}
}

// pyClassDecl is one class declaration extracted by the light scan.
type pyClassDecl struct {
	name     string
	abstract bool
	exported bool
}

// scanPythonClasses extracts the class declarations from Python source
// text. Only declarations starting a line are recognized (top-level or
// indented equally inside the file), which excludes commented-out class
// lines only when unindented — the light-parse contract accepts that.
func scanPythonClasses(source string) []pyClassDecl {
	var classes []pyClassDecl
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		match := pyClassRegex.FindStringSubmatch(trimmed)
		if match == nil {
			continue
		}
		name := match[1]
		classes = append(classes, pyClassDecl{
			name:     name,
			abstract: abstractPythonBases(match[2]),
			exported: !strings.HasPrefix(name, "_"),
		})
	}
	return classes
}

// abstractPythonBases reports whether a raw base list references an
// abstract base: abc.ABC, ABCMeta, or typing.Protocol, in dotted or
// bare (from-imported) form. Keyword arguments such as
// metaclass=ABCMeta are ignored: a metaclass keyword is not itself a
// base, and metaclass=ABCMeta without an ABC base does not make the
// class abstract under the light-parse contract.
func abstractPythonBases(rawBases string) bool {
	if rawBases == "" {
		return false
	}
	for _, token := range strings.Split(rawBases, ",") {
		token = strings.TrimSpace(token)
		if i := strings.IndexByte(token, '['); i >= 0 {
			token = token[:i]
		}
		if token == "" || strings.HasPrefix(token, "metaclass=") {
			continue
		}
		if isAbstractPythonBase(token) {
			return true
		}
	}
	return false
}

// isAbstractPythonBase reports whether a single, trimmed base token
// references an abstract base construct.
func isAbstractPythonBase(token string) bool {
	switch token {
	case "abc.ABC", "ABC", "ABCMeta", "typing.Protocol", "Protocol":
		return true
	}
	return strings.HasSuffix(token, ".ABC") ||
		strings.HasSuffix(token, ".ABCMeta") ||
		strings.HasSuffix(token, ".Protocol")
}

// dirRelativeToRoot renders the root-relative, slash-form directory of
// filePath. Files directly in the root report ".".
func dirRelativeToRoot(root, filePath string) string {
	rel, err := filepath.Rel(root, filepath.Dir(filePath))
	if err != nil {
		return "."
	}
	return filepath.ToSlash(rel)
}
