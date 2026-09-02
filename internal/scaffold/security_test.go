package scaffold

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
)

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
