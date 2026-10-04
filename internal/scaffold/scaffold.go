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
	dirPerm  fs.FileMode = 0o755
	filePerm fs.FileMode = 0o644
)

type category struct {
	sourceDir    string
	targetSubdir string
	prefix       string
}

var categories = []category{
	{sourceDir: "assets/agents", targetSubdir: ".opencode/agents", prefix: "agents"},
	{sourceDir: "assets/commands", targetSubdir: ".opencode/commands", prefix: "commands"},
}

// Options configures a scaffold Run.
type Options struct {
	TargetDir string
	Force     bool
	WriteFile func(path string, data []byte, perm fs.FileMode) error
}

// Result reports the outcome of a scaffold Run.
type Result struct {
	Written []string
	Skipped []string
	Forced  []string
}

// Run deploys the embedded agent and command assets into opts.TargetDir.
func Run(opts Options) (*Result, error) {
	return run(agentAssetsFS, commandAssetsFS, opts)
}

func run(agentAssets, commandAssets fs.FS, opts Options) (*Result, error) {
	if err := metrics.ValidateProjectPath(opts.TargetDir); err != nil {
		return nil, fmt.Errorf("scaffold: validate target directory: %w", err)
	}

	root, err := filepath.EvalSymlinks(opts.TargetDir)
	if err != nil {
		return nil, fmt.Errorf("scaffold: resolve target directory %q: %w", opts.TargetDir, err)
	}

	writeFile := opts.WriteFile
	if writeFile == nil {
		writeFile = os.WriteFile
	}

	sources := []fs.FS{agentAssets, commandAssets}
	result := &Result{}
	for i, cat := range categories {
		written, skipped, forced, err := deployCategory(sources[i], cat, root, writeFile, opts.Force)
		if err != nil {
			return nil, err
		}
		result.Written = append(result.Written, written...)
		result.Skipped = append(result.Skipped, skipped...)
		result.Forced = append(result.Forced, forced...)
	}

	sort.Strings(result.Written)
	sort.Strings(result.Skipped)
	sort.Strings(result.Forced)

	return result, nil
}

func deployCategory(assets fs.FS, cat category, root string, writeFile func(string, []byte, fs.FileMode) error, force bool) (written, skipped, forced []string, err error) {
	entries, err := fs.Glob(assets, cat.sourceDir+"/*.md")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("scaffold: enumerate embedded %s assets: %w", cat.prefix, err)
	}

	destDir, err := ensureDir(root, cat.targetSubdir)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, entry := range entries {
		name := path.Base(entry)
		prefixedName := cat.prefix + "/" + name

		data, err := fs.ReadFile(assets, entry)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("scaffold: read embedded asset %q: %w", entry, err)
		}

		destPath := filepath.Join(destDir, name)

		existed, err := regularFileExists(destPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if existed && !force {
			skipped = append(skipped, prefixedName)
			continue
		}

		if err := writeFile(destPath, data, filePerm); err != nil {
			return nil, nil, nil, fmt.Errorf("scaffold: write asset %q: %w", destPath, err)
		}
		if err := os.Chmod(destPath, filePerm); err != nil {
			return nil, nil, nil, fmt.Errorf("scaffold: set mode on %q: %w", destPath, err)
		}

		if existed {
			forced = append(forced, prefixedName)
		} else {
			written = append(written, prefixedName)
		}
	}

	return written, skipped, forced, nil
}

// assetPaths returns a sorted list of relative paths for all embedded assets
// across both agent and command filesystems. Each path is stripped of the
// "assets/" prefix so it maps directly to a .opencode/-relative deployment
// path (e.g. "agents/divisor-entropy.md").
func assetPaths() ([]string, error) {
	var paths []string
	for _, fsys := range []fs.FS{agentAssetsFS, commandAssetsFS} {
		err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel := strings.TrimPrefix(p, "assets/")
			paths = append(paths, rel)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scaffold: walk embedded assets: %w", err)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// assetContent reads and returns the bytes of an embedded asset identified by
// its .opencode/-relative path (without the "assets/" prefix). It dispatches
// to both agentAssetsFS and commandAssetsFS, returning the first successful
// read or an error if the asset is not found in either filesystem.
func assetContent(relPath string) ([]byte, error) {
	fullPath := path.Join("assets", relPath)
	for _, fsys := range []fs.FS{agentAssetsFS, commandAssetsFS} {
		data, err := fs.ReadFile(fsys, fullPath)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("scaffold: asset %q not found in embedded assets", relPath)
}

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
