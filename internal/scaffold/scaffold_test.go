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

// TestRun_RejectsTraversalPath covers error handling: a TargetDir containing a
// ".." segment is rejected before any write.
func TestRun_RejectsTraversalPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sep := string(filepath.Separator)
	traversal := dir + sep + ".." + sep + "escape"

	if _, err := Run(Options{TargetDir: traversal}); err == nil {
		t.Fatalf("expected error for traversal path %q, got nil", traversal)
	}
}

// TestRun_RejectsSymlinkedAgentsDir covers symlink safety: a symlinked
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
	if _, err := os.Stat(filepath.Join(outside, deployedAgentAssetName)); !os.IsNotExist(err) {
		t.Errorf("asset leaked outside the root via symlink: %v", err)
	}
}

// TestRun_RejectsSymlinkedCommandsDir covers symlink safety for the commands
// directory: a symlinked .opencode/commands that resolves outside the validated
// root is rejected.
func TestRun_RejectsSymlinkedCommandsDir(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics are unreliable on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()

	// Create a valid .opencode/agents so agents deploy succeeds, but symlink commands.
	if err := os.MkdirAll(filepath.Join(root, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode/agents: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, ".opencode", "commands")); err != nil {
		t.Fatalf("create commands symlink: %v", err)
	}

	if _, err := Run(Options{TargetDir: root}); err == nil {
		t.Fatalf("expected error when .opencode/commands is a symlink, got nil")
	}
	entries, _ := os.ReadDir(outside)
	if len(entries) > 0 {
		t.Errorf("assets leaked outside the root via commands symlink")
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

// TestEmbeddedAsset_DivisorEntropyContract covers the embedded
// divisor-entropy.md: required frontmatter, provenance marker, sections, and
// the bash allowlist.
func TestEmbeddedAsset_DivisorEntropyContract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(agentAssetsFS, embeddedAgentAssetPath)
	if err != nil {
		t.Fatalf("read embedded asset: %v", err)
	}
	content := string(data)

	if ok, _ := path.Match("divisor-*.md", deployedAgentAssetName); !ok {
		t.Errorf("asset name %q does not match divisor-*.md glob", deployedAgentAssetName)
	}
	matches, err := fs.Glob(agentAssetsFS, "assets/agents/divisor-*.md")
	if err != nil {
		t.Fatalf("glob embedded assets: %v", err)
	}
	if !slices.Contains(matches, embeddedAgentAssetPath) {
		t.Errorf("embedded assets missing %q; got %v", embeddedAgentAssetPath, matches)
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

// TestEmbeddedAsset_ReporterAgentContract covers the embedded
// vibe-check-reporter.md agent asset: required frontmatter fields, provenance
// marker, required sections, and bash allowlist.
func TestEmbeddedAsset_ReporterAgentContract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(agentAssetsFS, reporterAgentAssetPath)
	if err != nil {
		t.Fatalf("read embedded reporter agent: %v", err)
	}
	content := string(data)

	// Frontmatter fields.
	for _, needle := range []string{
		"mode: subagent",
		"temperature: 0.3",
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

	// Provenance marker.
	if !strings.Contains(content, "<!-- scaffolded by vibe-check") {
		t.Errorf("missing AP-006 provenance marker prefix")
	}

	// Required sections.
	for _, section := range []string{
		"## Source Documents",
		"## Mode Parsing",
		"## Summary Mode",
		"## Detailed Mode",
		"## Trending Mode",
		"## Natural Language Interpretation",
		"## Graceful Degradation",
		"## Security / Operating Constraints",
	} {
		if !strings.Contains(content, "\n"+section) {
			t.Errorf("missing required section heading %q", section)
		}
	}

	// Bash allowlist: exactly 2 entries + catch-all deny.
	allow, denyCatchAll := bashPermissions(content)
	if !denyCatchAll {
		t.Errorf(`bash catch-all "*" must map to "deny"`)
	}
	wantAllow := []string{
		"vibe-check analyze *",
		"git rev-parse *",
	}
	if len(allow) != len(wantAllow) {
		t.Errorf("bash allow count: got %d %v, want %d", len(allow), allow, len(wantAllow))
	}
	for _, k := range wantAllow {
		if !allow[k] {
			t.Errorf("bash allowlist missing required entry %q", k)
		}
	}
	wantSet := make(map[string]bool, len(wantAllow))
	for _, k := range wantAllow {
		wantSet[k] = true
	}
	for k := range allow {
		if !wantSet[k] {
			t.Errorf("bash allowlist has unexpected extra entry %q", k)
		}
	}
}

// TestEmbeddedAsset_CommandContract covers the embedded vibe-check.md command
// asset: required frontmatter fields, provenance marker, mode documentation,
// and $ARGUMENTS passthrough instruction.
func TestEmbeddedAsset_CommandContract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(commandAssetsFS, commandAssetPath)
	if err != nil {
		t.Fatalf("read embedded command asset: %v", err)
	}
	content := string(data)

	// Frontmatter: description and agent delegation.
	if frontmatterDescription(content) == "" {
		t.Errorf("frontmatter description is empty or missing")
	}
	if !strings.Contains(content, "agent: vibe-check-reporter") {
		t.Errorf("frontmatter missing agent delegation to vibe-check-reporter")
	}

	// Provenance marker.
	if !strings.Contains(content, "<!-- scaffolded by vibe-check") {
		t.Errorf("missing AP-006 provenance marker prefix")
	}

	// Mode documentation: all three modes must be mentioned.
	for _, mode := range []string{"summary", "detailed", "trending"} {
		if !strings.Contains(content, mode) {
			t.Errorf("command body missing mode documentation for %q", mode)
		}
	}

	// $ARGUMENTS passthrough instruction.
	if !strings.Contains(content, "$ARGUMENTS") {
		t.Errorf("command body missing $ARGUMENTS passthrough instruction")
	}
}

// TestAgentAssets_ReturnsReadableFS covers the exported AgentAssets accessor.
func TestAgentAssets_ReturnsReadableFS(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(AgentAssets(), embeddedAgentAssetPath)
	if err != nil {
		t.Fatalf("read via AgentAssets(): %v", err)
	}
	if len(data) == 0 {
		t.Errorf("AgentAssets() returned an empty asset")
	}
}

// TestCommandAssets_ReturnsReadableFS covers the exported CommandAssets
// accessor.
func TestCommandAssets_ReturnsReadableFS(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(CommandAssets(), commandAssetPath)
	if err != nil {
		t.Fatalf("read via CommandAssets(): %v", err)
	}
	if len(data) == 0 {
		t.Errorf("CommandAssets() returned an empty asset")
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
	if err := os.Symlink(outside, filepath.Join(agentsDir, deployedAgentAssetName)); err != nil {
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
	agents := orderedGlobFS{
		FS:         fstest.MapFS{},
		globResult: []string{"assets/agents/ghost.md"},
	}
	if _, err := run(agents, emptyFS, Options{TargetDir: t.TempDir()}); err == nil {
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
