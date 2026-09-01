package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

const (
	// assetSourceDir is the directory within the embedded filesystem that holds
	// the agent asset templates. The embed glob is assetSourceDir + "/*.md".
	assetSourceDir = "assets/agents"

	// targetSubdir is the directory, relative to the target repository root,
	// into which agent assets are deployed.
	targetSubdir = ".opencode/agents"

	// dirPerm is the mode applied to directories created during deployment.
	dirPerm fs.FileMode = 0o755

	// filePerm is the mode applied to asset files written during deployment.
	filePerm fs.FileMode = 0o644
)

// Options configures a scaffold Run.
type Options struct {
	// TargetDir is the root directory of the repository into which agent assets
	// are deployed. It MUST be an existing directory; traversal components are
	// rejected.
	TargetDir string

	// Force, when true, overwrites existing destination files. When false
	// (the default), pre-existing files are left unchanged and reported as
	// skipped.
	Force bool

	// WriteFile is an injectable seam for writing a destination file. When nil,
	// os.WriteFile is used. It mirrors the os.WriteFile signature so tests can
	// substitute a stub to exercise the I/O-failure path without special
	// privileges.
	WriteFile func(path string, data []byte, perm fs.FileMode) error
}

// Result reports the outcome of a scaffold Run. Each slice holds the base
// filenames of the affected assets in stable, ascending lexicographic order.
type Result struct {
	// Written lists assets that were newly created.
	Written []string

	// Skipped lists assets that already existed and were left unchanged because
	// Force was false.
	Skipped []string

	// Forced lists assets that already existed and were overwritten because
	// Force was true.
	Forced []string
}

// Run deploys the embedded agent assets into opts.TargetDir. It validates the
// target directory, creates the .opencode/agents tree if necessary, and writes
// each embedded asset, skipping or overwriting existing files according to
// opts.Force. It returns a Result describing which assets were written,
// skipped, or forced.
func Run(opts Options) (*Result, error) {
	return run(assetsFS, opts)
}

// run is the core scaffold routine parameterized over the source filesystem so
// tests can inject a synthetic fs.FS. Run calls it with the embedded assetsFS.
func run(assets fs.FS, opts Options) (*Result, error) {
	if err := metrics.ValidateProjectPath(opts.TargetDir); err != nil {
		return nil, fmt.Errorf("scaffold: validate target directory: %w", err)
	}

	// Canonicalize the validated root so containment checks compare against the
	// resolved path. ValidateProjectPath already confirmed it exists and is a
	// directory.
	root, err := filepath.EvalSymlinks(opts.TargetDir)
	if err != nil {
		return nil, fmt.Errorf("scaffold: resolve target directory %q: %w", opts.TargetDir, err)
	}

	writeFile := opts.WriteFile
	if writeFile == nil {
		writeFile = os.WriteFile
	}

	entries, err := fs.Glob(assets, assetSourceDir+"/*.md")
	if err != nil {
		return nil, fmt.Errorf("scaffold: enumerate embedded assets: %w", err)
	}

	destDir, err := ensureDir(root, targetSubdir)
	if err != nil {
		return nil, err
	}

	result := &Result{}
	for _, entry := range entries {
		name := path.Base(entry)

		data, err := fs.ReadFile(assets, entry)
		if err != nil {
			return nil, fmt.Errorf("scaffold: read embedded asset %q: %w", entry, err)
		}

		destPath := filepath.Join(destDir, name)

		existed, err := regularFileExists(destPath)
		if err != nil {
			return nil, err
		}
		if existed && !opts.Force {
			result.Skipped = append(result.Skipped, name)
			continue
		}

		if err := writeFile(destPath, data, filePerm); err != nil {
			return nil, fmt.Errorf("scaffold: write asset %q: %w", destPath, err)
		}
		// os.WriteFile preserves an existing file's mode on overwrite and is
		// subject to umask on create, so normalize the mode explicitly to keep
		// deployment deterministic.
		if err := os.Chmod(destPath, filePerm); err != nil {
			return nil, fmt.Errorf("scaffold: set mode on %q: %w", destPath, err)
		}

		if existed {
			result.Forced = append(result.Forced, name)
		} else {
			result.Written = append(result.Written, name)
		}
	}

	sort.Strings(result.Written)
	sort.Strings(result.Skipped)
	sort.Strings(result.Forced)

	return result, nil
}

// ensureDir creates rel (a slash-separated path relative to root) one component
// at a time, verifying after each step that the component is not a symlink and
// still resolves inside root. This prevents a symlink planted at an
// intermediate component (for example .opencode or .opencode/agents) from
// redirecting writes outside the validated root. It returns the absolute path
// of the deepest directory.
func ensureDir(root, rel string) (string, error) {
	current := root
	for _, part := range strings.Split(rel, "/") {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)

		info, err := os.Lstat(current)
		switch {
		case err == nil:
			if info.Mode()&fs.ModeSymlink != 0 {
				return "", fmt.Errorf("scaffold: refusing to follow symlink in deploy path: %s", current)
			}
			if !info.IsDir() {
				return "", fmt.Errorf("scaffold: deploy path component is not a directory: %s", current)
			}
		case os.IsNotExist(err):
			if mkErr := os.Mkdir(current, dirPerm); mkErr != nil {
				return "", fmt.Errorf("scaffold: create directory %q: %w", current, mkErr)
			}
			// Mkdir is subject to umask; normalize to the intended mode.
			if chErr := os.Chmod(current, dirPerm); chErr != nil {
				return "", fmt.Errorf("scaffold: set mode on directory %q: %w", current, chErr)
			}
		default:
			return "", fmt.Errorf("scaffold: inspect %q: %w", current, err)
		}

		if err := verifyContained(root, current); err != nil {
			return "", err
		}
	}

	return current, nil
}

// verifyContained resolves target and confirms it lies within root. root MUST
// already be a symlink-resolved absolute path.
func verifyContained(root, target string) error {
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return fmt.Errorf("scaffold: resolve %q: %w", target, err)
	}

	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return fmt.Errorf("scaffold: compute relative path for %q: %w", resolved, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("scaffold: deploy path escapes target root: %s", target)
	}

	return nil
}

// regularFileExists reports whether p exists as a regular file. It uses Lstat so
// a symlink at the destination is detected rather than followed, and returns an
// error if the destination is a symlink.
func regularFileExists(p string) (bool, error) {
	info, err := os.Lstat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("scaffold: inspect %q: %w", p, err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return false, fmt.Errorf("scaffold: refusing to overwrite symlink: %s", p)
	}

	return true, nil
}
