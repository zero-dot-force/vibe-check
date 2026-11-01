package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

// minReleaseTag is the minimum vibe-check release tag a scaffolded caller
// workflow may be pinned to: v0.3.0, the first release containing the
// reusable structural-gate workflow. It is a protected gate value (design
// D6) — it may only be raised deliberately, never lowered to make a build
// pass.
const minReleaseTag = "v0.3.0"

// ciTemplatePath is the embedded caller-workflow template's path within the
// CI assets filesystem (design D10).
const ciTemplatePath = "assets/ci/vibe-check-structural-gate.yaml"

// CIWorkflowName is the deployed caller workflow's filename under
// .github/workflows/ in the target repository — the same name this
// repository's own self-gate uses (design D5). It is exported so the init
// command's summary names the deployed file from the same source of truth.
const CIWorkflowName = "vibe-check-structural-gate.yaml"

// ciWorkflowAssetName is the category-prefixed name the deployed caller
// workflow is reported under in Result.Written and Result.Skipped.
const ciWorkflowAssetName = "workflows/vibe-check-structural-gate.yaml"

// The template's closed placeholder set (design D10): the release commit
// SHA substituted into the `uses:` reference and the release tag substituted
// into the `with: version:` pin, each appearing exactly once.
const (
	versionSHAPlaceholder = "__VIBE_CHECK_VERSION_SHA__"
	versionTagPlaceholder = "__VIBE_CHECK_VERSION_TAG__"
)

// manifestAdapters maps target-root manifest filenames to the language whose
// adapter the reusable structural gate currently runs. Detection is a seam
// (design D8): adding Python or TypeScript support extends this table and
// the no-manifest notice, not the deployment code.
var manifestAdapters = []struct {
	manifest string
	language string
}{
	{manifest: "go.mod", language: "Go"},
}

// noManifestNotice is appended to Result.Notices when Options.CI is set and
// no known language manifest exists at the target root (design D8). The
// caller workflow is still emitted — the notice is advisory.
const noManifestNotice = "no Go module manifest (go.mod) at target root — the reusable gate currently analyzes Go modules; Python/TypeScript adapters will reuse this same caller workflow."

// ValidateReleaseTag checks that tag is a v-prefixed semantic version at or
// above minReleaseTag — the floor a scaffolded caller may pin (design D6,
// a protected value). It is exported so the init command can fail closed on
// a dev build or unparseable override before attempting tag→SHA resolution
// (avoiding a wasted network round trip for a tag that can never resolve).
func ValidateReleaseTag(tag string) error {
	if !semver.IsValid(tag) || semver.Compare(tag, minReleaseTag) < 0 {
		return fmt.Errorf("scaffold: cannot resolve a vibe-check release tag >= %s (got %q) — install vibe-check from a release tag or pass --vibe-version", minReleaseTag, tag)
	}
	return nil
}

// validateReleasePins checks the caller-workflow pin pair fail-closed
// (design D6/D12): tag must be a v-prefixed semantic version at or above
// minReleaseTag and commit a full 40-character hex SHA. run executes it
// before any filesystem writes, so an invalid pin leaves no partial state.
func validateReleasePins(tag, commit string) error {
	if err := ValidateReleaseTag(tag); err != nil {
		return err
	}
	if !IsCommitSHA(commit) {
		return fmt.Errorf("scaffold: cannot resolve the release commit SHA for tag %q (got %q) — install vibe-check from a release binary or rerun with network access", tag, commit)
	}
	return nil
}

// IsCommitSHA reports whether s is a full 40-character hexadecimal git
// commit SHA — the pin form Options.VibeCommit requires and the emitted
// caller workflow's `uses:` reference carries (design D3). It is exported so
// the CLI layer applies the same predicate when choosing between the
// embedded release commit and tag resolution at init time.
func IsCommitSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

// manifestNotice returns the no-manifest notice (design D8) when none of the
// manifestAdapters entries names a regular file at the target root, or ""
// when a known language manifest is present.
func manifestNotice(root string) string {
	for _, m := range manifestAdapters {
		info, err := os.Stat(filepath.Join(root, m.manifest))
		if err == nil && info.Mode().IsRegular() {
			return ""
		}
	}
	return noManifestNotice
}

// renderCITemplate substitutes the pin placeholders in the embedded
// caller-workflow template: the release commit SHA into the `uses:`
// reference and the release tag into the `with: version:` pin (design D10).
// Each placeholder appears exactly once, so the two pins cannot drift apart
// in rendering.
func renderCITemplate(tmpl, tag, commit string) string {
	rendered := strings.Replace(tmpl, versionSHAPlaceholder, commit, 1)
	return strings.Replace(rendered, versionTagPlaceholder, tag, 1)
}

// deployCIWorkflow renders the embedded caller-workflow template with the
// validated pin pair and writes it to .github/workflows/vibe-check-structural-gate.yaml
// under root, reusing the shared deploy helpers (ensureDir with symlink
// refusal and containment checks, regularFileExists, the writeFile seam,
// 0o644 file mode — design D11). The emitted file is skip-if-present
// regardless of --force (design D7): it is user property once written, and
// users edit thresholds and pins in it. It is reported under the workflows/
// prefix in result.Written or result.Skipped.
func deployCIWorkflow(ciAssets fs.FS, root, tag, commit string, writeFile func(string, []byte, fs.FileMode) error, result *Result) error {
	data, err := fs.ReadFile(ciAssets, ciTemplatePath)
	if err != nil {
		return fmt.Errorf("scaffold: read embedded CI workflow template: %w", err)
	}
	rendered := renderCITemplate(string(data), tag, commit)

	destDir, err := ensureDir(root, ".github/workflows")
	if err != nil {
		return err
	}
	destPath := filepath.Join(destDir, CIWorkflowName)

	existed, err := regularFileExists(destPath)
	if err != nil {
		return err
	}
	if existed {
		result.Skipped = append(result.Skipped, ciWorkflowAssetName)
		return nil
	}

	if err := writeFile(destPath, []byte(rendered), filePerm); err != nil {
		return fmt.Errorf("scaffold: write CI workflow %q: %w", destPath, err)
	}
	if err := os.Chmod(destPath, filePerm); err != nil {
		return fmt.Errorf("scaffold: set mode on %q: %w", destPath, err)
	}
	result.Written = append(result.Written, ciWorkflowAssetName)
	return nil
}
