package scaffold

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

const (
	embeddedAssetPath = "assets/agents/divisor-entropy.md"
	deployedAssetName = "divisor-entropy.md"
)

// TestRun_DeploysEmbeddedAsset covers task 6.1: assets land in
// .opencode/agents/ with a 0o755 directory tree and 0o644 files, and the
// deployed bytes match the embedded source.
func TestRun_DeploysEmbeddedAsset(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !slices.Equal(res.Written, []string{deployedAssetName}) {
		t.Errorf("Written: got %v, want %v", res.Written, []string{deployedAssetName})
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped: got %v, want empty", res.Skipped)
	}
	if len(res.Forced) != 0 {
		t.Errorf("Forced: got %v, want empty", res.Forced)
	}

	openCodeDir := filepath.Join(dir, ".opencode")
	if oi, err := os.Stat(openCodeDir); err != nil || !oi.IsDir() {
		t.Fatalf(".opencode dir missing or not a directory: %v", err)
	}

	agentsDir := filepath.Join(openCodeDir, "agents")
	di, err := os.Stat(agentsDir)
	if err != nil {
		t.Fatalf("stat agents dir: %v", err)
	}
	if !di.IsDir() {
		t.Fatalf("%s is not a directory", agentsDir)
	}
	if got := di.Mode().Perm(); got != 0o755 {
		t.Errorf("agents dir perm: got %o, want %o", got, 0o755)
	}

	assetPath := filepath.Join(agentsDir, deployedAssetName)
	fi, err := os.Stat(assetPath)
	if err != nil {
		t.Fatalf("stat deployed asset: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("asset perm: got %o, want %o", got, 0o644)
	}

	got, err := os.ReadFile(assetPath)
	if err != nil {
		t.Fatalf("read deployed asset: %v", err)
	}
	want, err := fs.ReadFile(assetsFS, embeddedAssetPath)
	if err != nil {
		t.Fatalf("read embedded asset: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("deployed asset content does not match embedded source")
	}
}

// TestRun_SkipsExistingByDefault covers task 6.1: a second run without Force
// reports the existing asset as skipped and writes nothing.
func TestRun_SkipsExistingByDefault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if _, err := Run(Options{TargetDir: dir}); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if !slices.Equal(res.Skipped, []string{deployedAssetName}) {
		t.Errorf("Skipped: got %v, want %v", res.Skipped, []string{deployedAssetName})
	}
	if len(res.Written) != 0 || len(res.Forced) != 0 {
		t.Errorf("expected only skips; got Written=%v Forced=%v", res.Written, res.Forced)
	}
}

// TestRun_ForceOverwrites covers task 6.1: Force overwrites an existing asset
// and reports it as forced.
func TestRun_ForceOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if _, err := Run(Options{TargetDir: dir}); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	res, err := Run(Options{TargetDir: dir, Force: true})
	if err != nil {
		t.Fatalf("force Run: %v", err)
	}
	if !slices.Equal(res.Forced, []string{deployedAssetName}) {
		t.Errorf("Forced: got %v, want %v", res.Forced, []string{deployedAssetName})
	}
	if len(res.Written) != 0 || len(res.Skipped) != 0 {
		t.Errorf("expected only forced; got Written=%v Skipped=%v", res.Written, res.Skipped)
	}
}

// TestRun_ForceNormalizesPermissions covers task 6.2: a pre-existing file with
// a loose mode is normalized to 0o644 on a forced overwrite.
func TestRun_ForceNormalizesPermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}
	assetPath := filepath.Join(agentsDir, deployedAssetName)
	if err := os.WriteFile(assetPath, []byte("stale"), 0o666); err != nil {
		t.Fatalf("pre-create asset: %v", err)
	}
	// WriteFile is subject to umask; force the loose mode explicitly.
	if err := os.Chmod(assetPath, 0o666); err != nil {
		t.Fatalf("chmod pre-create asset: %v", err)
	}

	res, err := Run(Options{TargetDir: dir, Force: true})
	if err != nil {
		t.Fatalf("force Run: %v", err)
	}
	if !slices.Equal(res.Forced, []string{deployedAssetName}) {
		t.Errorf("Forced: got %v, want %v", res.Forced, []string{deployedAssetName})
	}
	fi, err := os.Stat(assetPath)
	if err != nil {
		t.Fatalf("stat asset: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("perm after forced overwrite: got %o, want %o", got, 0o644)
	}
}

// TestRun_RejectsNonexistentRoot covers task 6.3: a nonexistent TargetDir is
// rejected and nothing is written.
func TestRun_RejectsNonexistentRoot(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	missing := filepath.Join(parent, "does-not-exist")

	res, err := Run(Options{TargetDir: missing})
	if err == nil {
		t.Fatalf("expected error for nonexistent root, got nil (res=%v)", res)
	}
	if _, statErr := os.Stat(filepath.Join(missing, ".opencode")); !os.IsNotExist(statErr) {
		t.Errorf("expected nothing written under a nonexistent root")
	}
}

// TestRun_RejectsTraversalPath covers task 6.3: a TargetDir containing a ".."
// segment is rejected before any write.
func TestRun_RejectsTraversalPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sep := string(filepath.Separator)
	traversal := dir + sep + ".." + sep + "escape"

	if _, err := Run(Options{TargetDir: traversal}); err == nil {
		t.Fatalf("expected error for traversal path %q, got nil", traversal)
	}
}

// TestRun_RejectsSymlinkedAgentsDir covers task 6.3: a symlinked
// .opencode/agents that resolves outside the validated root is rejected and no
// asset leaks to the symlink target.
func TestRun_RejectsSymlinkedAgentsDir(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics are unreliable on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()

	if err := os.Mkdir(filepath.Join(root, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, ".opencode", "agents")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	if _, err := Run(Options{TargetDir: root}); err == nil {
		t.Fatalf("expected error when .opencode/agents is a symlink, got nil")
	}
	if _, err := os.Stat(filepath.Join(outside, deployedAssetName)); !os.IsNotExist(err) {
		t.Errorf("asset leaked outside the root via symlink: %v", err)
	}
}

// TestRun_WriteFileError covers the injectable WriteFile seam (task 5.2): an
// I/O failure surfaces as a wrapped error from Run.
func TestRun_WriteFileError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sentinel := errors.New("disk full")

	_, err := Run(Options{
		TargetDir: dir,
		WriteFile: func(string, []byte, fs.FileMode) error { return sentinel },
	})
	if err == nil {
		t.Fatalf("expected error from failing WriteFile, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error chain: got %v, want wrapped %v", err, sentinel)
	}
}

// orderedGlobFS wraps an fs.FS and returns a fixed, caller-controlled Glob
// ordering so tests can prove run sorts its results independent of walk order.
type orderedGlobFS struct {
	fs.FS
	globResult []string
}

func (o orderedGlobFS) Glob(string) ([]string, error) {
	return o.globResult, nil
}

// TestRun_ResultsAreSorted covers task 6.5: over a synthetic multi-entry fs.FS
// presented in non-alphabetical order, run returns Written/Skipped/Forced in
// stable ascending order.
func TestRun_ResultsAreSorted(t *testing.T) {
	t.Parallel()
	base := fstest.MapFS{
		"assets/agents/charlie.md": {Data: []byte("charlie")},
		"assets/agents/alpha.md":   {Data: []byte("alpha")},
		"assets/agents/bravo.md":   {Data: []byte("bravo")},
	}
	assets := orderedGlobFS{
		FS: base,
		globResult: []string{
			"assets/agents/charlie.md",
			"assets/agents/bravo.md",
			"assets/agents/alpha.md",
		},
	}
	want := []string{"alpha.md", "bravo.md", "charlie.md"}
	dir := t.TempDir()

	res, err := run(assets, Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("run (write): %v", err)
	}
	if !slices.Equal(res.Written, want) {
		t.Errorf("Written not sorted: got %v, want %v", res.Written, want)
	}

	res2, err := run(assets, Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("run (skip): %v", err)
	}
	if !slices.Equal(res2.Skipped, want) {
		t.Errorf("Skipped not sorted: got %v, want %v", res2.Skipped, want)
	}

	res3, err := run(assets, Options{TargetDir: dir, Force: true})
	if err != nil {
		t.Fatalf("run (force): %v", err)
	}
	if !slices.Equal(res3.Forced, want) {
		t.Errorf("Forced not sorted: got %v, want %v", res3.Forced, want)
	}
}

// TestEmbeddedAsset_Contract covers task 6.4: the embedded divisor-entropy.md
// carries the required frontmatter, provenance marker, sections, and a bash
// allowlist that is exactly the nine permitted commands.
func TestEmbeddedAsset_Contract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(assetsFS, embeddedAssetPath)
	if err != nil {
		t.Fatalf("read embedded asset: %v", err)
	}
	content := string(data)

	if ok, _ := path.Match("divisor-*.md", deployedAssetName); !ok {
		t.Errorf("asset name %q does not match divisor-*.md glob", deployedAssetName)
	}
	matches, err := fs.Glob(assetsFS, "assets/agents/divisor-*.md")
	if err != nil {
		t.Fatalf("glob embedded assets: %v", err)
	}
	if !slices.Contains(matches, embeddedAssetPath) {
		t.Errorf("embedded assets missing %q; got %v", embeddedAssetPath, matches)
	}

	for _, needle := range []string{
		"mode: subagent",
		"temperature: 0.1",
		"edit: deny",
		"webfetch: deny",
		`"*": "deny"`,
	} {
		if !strings.Contains(content, needle) {
			t.Errorf("frontmatter missing %q", needle)
		}
	}

	if frontmatterDescription(content) == "" {
		t.Errorf("frontmatter description is empty or missing")
	}

	if !strings.Contains(content, "<!-- scaffolded by vibe-check") {
		t.Errorf("missing AP-006 provenance marker prefix")
	}

	for _, section := range []string{
		"## Source Documents",
		"## Code Review Mode",
		"## Output Format",
		"## Decision Criteria",
		"## Security / Operating Constraints",
	} {
		if !strings.Contains(content, "\n"+section) {
			t.Errorf("missing required section heading %q", section)
		}
	}

	allow, denyCatchAll := bashPermissions(content)
	if !denyCatchAll {
		t.Errorf(`bash catch-all "*" must map to "deny"`)
	}
	want := []string{
		"git merge-base *",
		"git rev-parse *",
		"git worktree add *",
		"git worktree remove *",
		"git worktree prune",
		"git fetch origin *",
		"git check-ref-format *",
		"vibe-check analyze *",
		"vibe-check diff *",
	}
	if len(allow) != len(want) {
		t.Errorf("bash allow count: got %d %v, want %d", len(allow), allow, len(want))
	}
	wantSet := make(map[string]bool, len(want))
	for _, k := range want {
		wantSet[k] = true
		if !allow[k] {
			t.Errorf("bash allowlist missing required entry %q", k)
		}
	}
	for k := range allow {
		if !wantSet[k] {
			t.Errorf("bash allowlist has unexpected extra entry %q", k)
		}
	}
}

// TestAssets_ReturnsReadableFS covers the exported Assets accessor.
func TestAssets_ReturnsReadableFS(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(Assets(), embeddedAssetPath)
	if err != nil {
		t.Fatalf("read via Assets(): %v", err)
	}
	if len(data) == 0 {
		t.Errorf("Assets() returned an empty asset")
	}
}

// TestRun_RejectsNonDirectoryComponent covers ensureDir when a deploy-path
// component already exists as a regular file rather than a directory.
func TestRun_RejectsNonDirectoryComponent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Plant a regular file where the .opencode directory must be created.
	if err := os.WriteFile(filepath.Join(root, ".opencode"), []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("plant file: %v", err)
	}
	if _, err := Run(Options{TargetDir: root}); err == nil {
		t.Fatalf("expected error when .opencode is a regular file, got nil")
	}
}

// TestRun_RefusesToOverwriteSymlinkDestination covers regularFileExists when
// the destination asset path is a symlink: it must be refused, not followed.
func TestRun_RefusesToOverwriteSymlinkDestination(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics are unreliable on Windows")
	}
	root := t.TempDir()
	agentsDir := filepath.Join(root, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "target")
	if err := os.Symlink(outside, filepath.Join(agentsDir, deployedAssetName)); err != nil {
		t.Fatalf("create dest symlink: %v", err)
	}
	if _, err := Run(Options{TargetDir: root, Force: true}); err == nil {
		t.Fatalf("expected error when destination is a symlink, got nil")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Errorf("write followed symlink to %q: %v", outside, err)
	}
}

// TestRun_ReadEmbeddedAssetError covers the asset-read failure path via a
// synthetic fs.FS whose Glob advertises an entry that cannot be read.
func TestRun_ReadEmbeddedAssetError(t *testing.T) {
	t.Parallel()
	assets := orderedGlobFS{
		FS:         fstest.MapFS{},
		globResult: []string{"assets/agents/ghost.md"},
	}
	if _, err := run(assets, Options{TargetDir: t.TempDir()}); err == nil {
		t.Fatalf("expected error reading a missing embedded asset, got nil")
	}
}

// TestVerifyContained covers containment: a path inside the resolved root is
// accepted and a path outside the root is rejected.
func TestVerifyContained(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	inside := filepath.Join(root, "sub")
	if err := os.Mkdir(inside, 0o755); err != nil {
		t.Fatalf("mkdir inside: %v", err)
	}
	if err := verifyContained(root, inside); err != nil {
		t.Errorf("verifyContained(inside): unexpected error %v", err)
	}

	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve outside: %v", err)
	}
	if err := verifyContained(root, outside); err == nil {
		t.Errorf("verifyContained(outside): expected escape error, got nil")
	}
}

// frontmatterDescription returns the unquoted value of the description field,
// or "" if it is absent. Stdlib scanning only; no YAML dependency.
func frontmatterDescription(content string) string {
	for _, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(ln, "description:") {
			v := strings.TrimSpace(strings.TrimPrefix(ln, "description:"))
			return strings.Trim(v, `"`)
		}
	}
	return ""
}

// bashPermissions extracts the permission.bash block from the frontmatter,
// returning the set of keys mapped to "allow" and whether the "*" catch-all is
// "deny". Stdlib scanning only; no YAML dependency.
func bashPermissions(content string) (allow map[string]bool, denyCatchAll bool) {
	allow = map[string]bool{}
	inBash := false
	for _, ln := range strings.Split(content, "\n") {
		if !inBash {
			if strings.TrimSpace(ln) == "bash:" {
				inBash = true
			}
			continue
		}
		// Entries are indented four spaces; any dedent ends the block.
		if !strings.HasPrefix(ln, "    ") {
			break
		}
		entry := strings.TrimSpace(ln)
		idx := strings.Index(entry, ":")
		if idx < 0 {
			continue
		}
		key := strings.Trim(strings.TrimSpace(entry[:idx]), `"`)
		val := strings.Trim(strings.TrimSpace(entry[idx+1:]), `"`)
		switch {
		case key == "*":
			denyCatchAll = val == "deny"
		case val == "allow":
			allow[key] = true
		}
	}
	return allow, denyCatchAll
}
