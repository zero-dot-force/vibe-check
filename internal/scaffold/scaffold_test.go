package scaffold

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
)

const (
	embeddedAgentAssetPath = "assets/agents/divisor-entropy.md"
	deployedAgentAssetName = "divisor-entropy.md"
	prefixedAgentAssetName = "agents/divisor-entropy.md"
	prefixedCmdAssetName   = "commands/vibe-check.md"
	reporterAgentAssetPath = "assets/agents/vibe-check-reporter.md"
	reporterAgentAssetName = "agents/vibe-check-reporter.md"
	commandAssetPath       = "assets/commands/vibe-check.md"
	commandAssetName       = "commands/vibe-check.md"
)

// TestRun_DeploysEmbeddedAssets covers fresh deployment: both agent and command
// assets land in .opencode/agents/ and .opencode/commands/ with a 0o755
// directory tree and 0o644 files, and the deployed bytes match the embedded
// source.
func TestRun_DeploysEmbeddedAssets(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	// All assets should be written, none skipped or forced.
	if len(res.Written) == 0 {
		t.Fatalf("Written is empty, expected assets to be deployed")
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped: got %v, want empty", res.Skipped)
	}
	if len(res.Forced) != 0 {
		t.Errorf("Forced: got %v, want empty", res.Forced)
	}

	// Verify Written contains both agent and command entries with prefixes.
	if !slices.Contains(res.Written, prefixedAgentAssetName) {
		t.Errorf("Written missing %q; got %v", prefixedAgentAssetName, res.Written)
	}
	if !slices.Contains(res.Written, reporterAgentAssetName) {
		t.Errorf("Written missing %q; got %v", reporterAgentAssetName, res.Written)
	}
	if !slices.Contains(res.Written, prefixedCmdAssetName) {
		t.Errorf("Written missing %q; got %v", prefixedCmdAssetName, res.Written)
	}

	// Verify .opencode directory structure.
	openCodeDir := filepath.Join(dir, ".opencode")
	if oi, err := os.Stat(openCodeDir); err != nil || !oi.IsDir() {
		t.Fatalf(".opencode dir missing or not a directory: %v", err)
	}

	// Verify agents dir permissions and content.
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

	// Verify commands dir permissions.
	cmdsDir := filepath.Join(openCodeDir, "commands")
	ci, err := os.Stat(cmdsDir)
	if err != nil {
		t.Fatalf("stat commands dir: %v", err)
	}
	if !ci.IsDir() {
		t.Fatalf("%s is not a directory", cmdsDir)
	}
	if got := ci.Mode().Perm(); got != 0o755 {
		t.Errorf("commands dir perm: got %o, want %o", got, 0o755)
	}

	// Verify agent asset file perm and byte-match.
	agentPath := filepath.Join(agentsDir, deployedAgentAssetName)
	fi, err := os.Stat(agentPath)
	if err != nil {
		t.Fatalf("stat deployed agent asset: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("agent asset perm: got %o, want %o", got, 0o644)
	}
	got, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("read deployed agent asset: %v", err)
	}
	want, err := fs.ReadFile(agentAssetsFS, embeddedAgentAssetPath)
	if err != nil {
		t.Fatalf("read embedded agent asset: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("deployed agent asset content does not match embedded source")
	}

	// Verify command asset file perm and byte-match.
	cmdPath := filepath.Join(cmdsDir, "vibe-check.md")
	cfi, err := os.Stat(cmdPath)
	if err != nil {
		t.Fatalf("stat deployed command asset: %v", err)
	}
	if got := cfi.Mode().Perm(); got != 0o644 {
		t.Errorf("command asset perm: got %o, want %o", got, 0o644)
	}
	gotCmd, err := os.ReadFile(cmdPath)
	if err != nil {
		t.Fatalf("read deployed command asset: %v", err)
	}
	wantCmd, err := fs.ReadFile(commandAssetsFS, commandAssetPath)
	if err != nil {
		t.Fatalf("read embedded command asset: %v", err)
	}
	if !bytes.Equal(gotCmd, wantCmd) {
		t.Errorf("deployed command asset content does not match embedded source")
	}
}

// TestRun_SkipsExistingByDefault covers a second run without Force: all
// existing assets are reported as skipped with category-prefixed names.
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
	if len(res.Written) != 0 || len(res.Forced) != 0 {
		t.Errorf("expected only skips; got Written=%v Forced=%v", res.Written, res.Forced)
	}
	// All assets should be in Skipped.
	if !slices.Contains(res.Skipped, prefixedAgentAssetName) {
		t.Errorf("Skipped missing %q; got %v", prefixedAgentAssetName, res.Skipped)
	}
	if !slices.Contains(res.Skipped, prefixedCmdAssetName) {
		t.Errorf("Skipped missing %q; got %v", prefixedCmdAssetName, res.Skipped)
	}
}

// TestRun_ForceOverwrites covers Force: existing assets are overwritten and
// reported as forced with category-prefixed names.
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
	if len(res.Written) != 0 || len(res.Skipped) != 0 {
		t.Errorf("expected only forced; got Written=%v Skipped=%v", res.Written, res.Skipped)
	}
	if !slices.Contains(res.Forced, prefixedAgentAssetName) {
		t.Errorf("Forced missing %q; got %v", prefixedAgentAssetName, res.Forced)
	}
	if !slices.Contains(res.Forced, prefixedCmdAssetName) {
		t.Errorf("Forced missing %q; got %v", prefixedCmdAssetName, res.Forced)
	}
}

// TestRun_ForceNormalizesPermissions covers permission normalization: a
// pre-existing file with loose mode is normalized to 0o644 on forced overwrite.
func TestRun_ForceNormalizesPermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}
	assetPath := filepath.Join(agentsDir, deployedAgentAssetName)
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
	if !slices.Contains(res.Forced, prefixedAgentAssetName) {
		t.Errorf("Forced missing %q; got %v", prefixedAgentAssetName, res.Forced)
	}
	fi, err := os.Stat(assetPath)
	if err != nil {
		t.Fatalf("stat asset: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("perm after forced overwrite: got %o, want %o", got, 0o644)
	}
}

// TestRun_RejectsNonexistentRoot covers error handling: a nonexistent TargetDir
// is rejected and nothing is written.
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

// TestRun_WriteFileError covers the injectable WriteFile seam: an I/O failure
// surfaces as a wrapped error from Run.
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

// emptyFS is an fs.FS with no assets, used as a stub for the category that
// should contribute nothing.
var emptyFS = orderedGlobFS{
	FS:         fstest.MapFS{},
	globResult: nil,
}

// TestRun_ResultsAreSorted covers sorting: over a synthetic multi-entry fs.FS
// presented in non-alphabetical order, run returns Written/Skipped/Forced in
// stable ascending order with category prefixes.
func TestRun_ResultsAreSorted(t *testing.T) {
	t.Parallel()
	base := fstest.MapFS{
		"assets/agents/charlie.md": {Data: []byte("charlie")},
		"assets/agents/alpha.md":   {Data: []byte("alpha")},
		"assets/agents/bravo.md":   {Data: []byte("bravo")},
	}
	agents := orderedGlobFS{
		FS: base,
		globResult: []string{
			"assets/agents/charlie.md",
			"assets/agents/bravo.md",
			"assets/agents/alpha.md",
		},
	}
	want := []string{"agents/alpha.md", "agents/bravo.md", "agents/charlie.md"}
	dir := t.TempDir()

	res, err := run(agents, emptyFS, Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("run (write): %v", err)
	}
	if !slices.Equal(res.Written, want) {
		t.Errorf("Written not sorted: got %v, want %v", res.Written, want)
	}

	res2, err := run(agents, emptyFS, Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("run (skip): %v", err)
	}
	if !slices.Equal(res2.Skipped, want) {
		t.Errorf("Skipped not sorted: got %v, want %v", res2.Skipped, want)
	}

	res3, err := run(agents, emptyFS, Options{TargetDir: dir, Force: true})
	if err != nil {
		t.Fatalf("run (force): %v", err)
	}
	if !slices.Equal(res3.Forced, want) {
		t.Errorf("Forced not sorted: got %v, want %v", res3.Forced, want)
	}
}

// TestRun_MixedAgentAndCommandResults verifies that a single Run() call
// produces category-prefixed Result entries from both agent and command
// asset categories, sorted together.
func TestRun_MixedAgentAndCommandResults(t *testing.T) {
	t.Parallel()
	agents := orderedGlobFS{
		FS: fstest.MapFS{
			"assets/agents/beta.md": {Data: []byte("beta")},
		},
		globResult: []string{"assets/agents/beta.md"},
	}
	cmds := orderedGlobFS{
		FS: fstest.MapFS{
			"assets/commands/alpha.md": {Data: []byte("alpha")},
		},
		globResult: []string{"assets/commands/alpha.md"},
	}
	dir := t.TempDir()

	res, err := run(agents, cmds, Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	wantWritten := []string{"agents/beta.md", "commands/alpha.md"}
	if !slices.Equal(res.Written, wantWritten) {
		t.Errorf("Written: got %v, want %v", res.Written, wantWritten)
	}
}
