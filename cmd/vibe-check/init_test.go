package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// initAssetNames lists the category-prefixed asset names that scaffold.Run
// deploys. The scaffold writer reports assets with category prefixes (e.g.
// "agents/divisor-entropy.md"), sorted lexicographically. These appear in
// InitResult slices and in the --json payload.
var initAssetNames = []string{
	"agents/divisor-entropy.md",
	"agents/vibe-check-reporter.md",
	"commands/vibe-check.md",
}

// deployedAgentAssetPath returns the on-disk location of the divisor-entropy
// agent asset for a given project root.
func deployedAgentAssetPath(root string) string {
	return filepath.Join(root, ".opencode", "agents", "divisor-entropy.md")
}

// deployedCommandAssetPath returns the on-disk location of the vibe-check
// command asset for a given project root.
func deployedCommandAssetPath(root string) string {
	return filepath.Join(root, ".opencode", "commands", "vibe-check.md")
}

// initFailingWriter is an io.Writer whose Write always fails, used to exercise
// the stdout-write error path of RunInit.
type initFailingWriter struct{ err error }

func (w initFailingWriter) Write([]byte) (int, error) { return 0, w.err }

// TestRunInit_HumanSummaryLifecycle covers task 8.1: the human-readable summary
// lists written files on the first run, reports skips on a second run, and
// reports forced overwrites under Force.
func TestRunInit_HumanSummaryLifecycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	// First run: the asset is written.
	res, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: dir})
	if err != nil {
		t.Fatalf("first RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("first run ExitCode: got %d, want 0", res.ExitCode)
	}
	if got := res.Written; !slices.Equal(got, initAssetNames) {
		t.Errorf("first run Written: got %v, want %v", got, initAssetNames)
	}
	if len(res.Skipped) != 0 || len(res.Forced) != 0 {
		t.Errorf("first run should have no skips/forced: skipped=%v forced=%v", res.Skipped, res.Forced)
	}
	out := stdout.String()
	if !strings.Contains(out, "Written:") {
		t.Errorf("summary missing Written section; got:\n%s", out)
	}
	if !strings.Contains(out, "agents/divisor-entropy.md") {
		t.Errorf("summary missing agent asset name; got:\n%s", out)
	}
	if !strings.Contains(out, "commands/vibe-check.md") {
		t.Errorf("summary missing command asset name; got:\n%s", out)
	}
	if _, statErr := os.Stat(deployedAgentAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed agent asset on disk: %v", statErr)
	}
	if _, statErr := os.Stat(deployedCommandAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed command asset on disk: %v", statErr)
	}

	// Second run: the existing asset is skipped.
	stdout.Reset()
	res2, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: dir})
	if err != nil {
		t.Fatalf("second RunInit returned error: %v", err)
	}
	if got := res2.Skipped; !slices.Equal(got, initAssetNames) {
		t.Errorf("second run Skipped: got %v, want %v", got, initAssetNames)
	}
	if len(res2.Written) != 0 {
		t.Errorf("second run Written should be empty: got %v", res2.Written)
	}
	if !strings.Contains(stdout.String(), "Skipped:") {
		t.Errorf("second run summary missing Skipped section; got:\n%s", stdout.String())
	}

	// Third run with Force: the asset is overwritten.
	stdout.Reset()
	res3, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: dir, Force: true})
	if err != nil {
		t.Fatalf("force RunInit returned error: %v", err)
	}
	if got := res3.Forced; !slices.Equal(got, initAssetNames) {
		t.Errorf("force run Forced: got %v, want %v", got, initAssetNames)
	}
	if len(res3.Written) != 0 || len(res3.Skipped) != 0 {
		t.Errorf("force run should only report forced: written=%v skipped=%v", res3.Written, res3.Skipped)
	}
	if !strings.Contains(stdout.String(), "Forced:") {
		t.Errorf("force run summary missing Forced section; got:\n%s", stdout.String())
	}
}

// TestRunInit_JSONLifecycle covers task 8.2: the --json payload unmarshals and
// carries the expected per-stage values, with every key present as an array.
func TestRunInit_JSONLifecycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	type initPayload struct {
		Written []string `json:"written"`
		Skipped []string `json:"skipped"`
		Forced  []string `json:"forced"`
	}

	run := func(force bool) initPayload {
		t.Helper()
		var stdout, stderr bytes.Buffer
		res, err := RunInit(context.Background(), InitOptions{
			Stdout: &stdout,
			Stderr: &stderr,
			Path:   dir,
			Force:  force,
			JSON:   true,
		})
		if err != nil {
			t.Fatalf("RunInit(force=%v) returned error: %v", force, err)
		}
		if res.ExitCode != 0 {
			t.Fatalf("RunInit(force=%v) ExitCode: got %d, want 0", force, res.ExitCode)
		}
		var p initPayload
		if err := json.Unmarshal(stdout.Bytes(), &p); err != nil {
			t.Fatalf("json.Unmarshal(%q): %v", stdout.String(), err)
		}
		return p
	}

	// First run: only written is populated; skipped and forced are empty arrays.
	first := run(false)
	if got := first.Written; !slices.Equal(got, initAssetNames) {
		t.Errorf("first run written: got %v, want %v", got, initAssetNames)
	}
	if len(first.Skipped) != 0 {
		t.Errorf("first run skipped: got %v, want empty", first.Skipped)
	}
	if len(first.Forced) != 0 {
		t.Errorf("first run forced: got %v, want empty", first.Forced)
	}

	// Second run: the asset moves to skipped.
	second := run(false)
	if got := second.Skipped; !slices.Equal(got, initAssetNames) {
		t.Errorf("second run skipped: got %v, want %v", got, initAssetNames)
	}
	if len(second.Written) != 0 {
		t.Errorf("second run written: got %v, want empty", second.Written)
	}

	// Force run: the asset moves to forced.
	forced := run(true)
	if got := forced.Forced; !slices.Equal(got, initAssetNames) {
		t.Errorf("force run forced: got %v, want %v", got, initAssetNames)
	}
	if len(forced.Written) != 0 || len(forced.Skipped) != 0 {
		t.Errorf("force run should only populate forced: written=%v skipped=%v", forced.Written, forced.Skipped)
	}
}

// TestRunInit_JSONArraysAreNeverNull verifies the --json payload emits empty
// arrays rather than null for the unused categories on a first run.
func TestRunInit_JSONArraysAreNeverNull(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	if _, err := RunInit(context.Background(), InitOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   t.TempDir(),
		JSON:   true,
	}); err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if strings.Contains(stdout.String(), "null") {
		t.Errorf("JSON payload must not contain null arrays; got:\n%s", stdout.String())
	}
}

// TestRunInit_SuccessExitCode covers task 8.3: a valid deployment exits 0.
func TestRunInit_SuccessExitCode(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: t.TempDir()})
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode: got %d, want 0", res.ExitCode)
	}
}

// TestRunInit_NonexistentPathExitsTwo covers task 8.3: an invalid (nonexistent)
// target path exits 2, writes nothing to stdout, and creates no files.
func TestRunInit_NonexistentPathExitsTwo(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: missing})
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
	if res == nil || res.ExitCode != 2 {
		t.Fatalf("ExitCode: got %v, want 2", res)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout must be empty on error, got: %q", stdout.String())
	}
	if _, statErr := os.Stat(missing); !os.IsNotExist(statErr) {
		t.Errorf("nonexistent path must not be created: stat err = %v", statErr)
	}
}

// TestRunInit_TraversalPathExitsTwo covers task 8.3: a path-traversal target is
// rejected with exit 2 and nothing is written.
func TestRunInit_TraversalPathExitsTwo(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	traversal := base + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "escape"
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: traversal})
	if err == nil {
		t.Fatal("expected error for traversal path, got nil")
	}
	if res == nil || res.ExitCode != 2 {
		t.Fatalf("ExitCode: got %v, want 2", res)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout must be empty on error, got: %q", stdout.String())
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(base), "escape", ".opencode")); !os.IsNotExist(statErr) {
		t.Errorf(".opencode must not be created via traversal: stat err = %v", statErr)
	}
}

// TestRunInit_WriteFileErrorExitsTwo covers task 8.3: an I/O failure surfaced
// through the injected writeFile seam exits 2, wraps the underlying error, and
// writes nothing to stdout.
func TestRunInit_WriteFileErrorExitsTwo(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("simulated disk failure")
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout:    &stdout,
		Stderr:    &stderr,
		Path:      t.TempDir(),
		writeFile: func(string, []byte, fs.FileMode) error { return sentinel },
	})
	if err == nil {
		t.Fatal("expected error from failing writeFile, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error chain: got %v, want to wrap %v", err, sentinel)
	}
	if res == nil || res.ExitCode != 2 {
		t.Fatalf("ExitCode: got %v, want 2", res)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout must be empty on error, got: %q", stdout.String())
	}
}

// TestRunInit_StdoutWriteErrorExitsTwo exercises the atomic-output error path:
// when the summary cannot be written to stdout, RunInit exits 2.
func TestRunInit_StdoutWriteErrorExitsTwo(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("broken pipe")
	var stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout: initFailingWriter{err: sentinel},
		Stderr: &stderr,
		Path:   t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error from failing stdout writer, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error chain: got %v, want to wrap %v", err, sentinel)
	}
	if res == nil || res.ExitCode != 2 {
		t.Fatalf("ExitCode: got %v, want 2", res)
	}
}

// TestRunInit_CancelledContextExitsTwo verifies RunInit honors context
// cancellation before doing filesystem work.
func TestRunInit_CancelledContextExitsTwo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	res, err := RunInit(ctx, InitOptions{Stdout: &stdout, Stderr: &stderr, Path: t.TempDir()})
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if res == nil || res.ExitCode != 2 {
		t.Fatalf("ExitCode: got %v, want 2", res)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout must be empty on error, got: %q", stdout.String())
	}
}

// TestInitCmd_Execute verifies the cobra command layer wires flags and args
// through to a successful deployment.
func TestInitCmd_Execute(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := initCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("initCmd Execute returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "agents/divisor-entropy.md") {
		t.Errorf("expected agent asset name in output; got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "commands/vibe-check.md") {
		t.Errorf("expected command asset name in output; got:\n%s", stdout.String())
	}
	if _, statErr := os.Stat(deployedAgentAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed agent asset on disk: %v", statErr)
	}
	if _, statErr := os.Stat(deployedCommandAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed command asset on disk: %v", statErr)
	}
}

// TestInitCmd_JSONFlag verifies the --json flag produces a machine-readable
// payload through the command layer.
func TestInitCmd_JSONFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := initCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--json", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("initCmd --json Execute returned error: %v", err)
	}
	var p struct {
		Written []string `json:"written"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &p); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", stdout.String(), err)
	}
	if !slices.Contains(p.Written, "agents/divisor-entropy.md") {
		t.Errorf("expected agents/divisor-entropy.md in written; got %v", p.Written)
	}
	if !slices.Contains(p.Written, "commands/vibe-check.md") {
		t.Errorf("expected commands/vibe-check.md in written; got %v", p.Written)
	}
}

// TestInitCmd_InvalidPathReturnsExitCodeError verifies the command layer maps a
// RunInit failure to an *exitCodeError carrying exit code 2.
func TestInitCmd_InvalidPathReturnsExitCodeError(t *testing.T) {
	t.Parallel()
	cmd := initCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{filepath.Join(t.TempDir(), "missing")})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
	var ece *exitCodeError
	if !errors.As(err, &ece) {
		t.Fatalf("error type: got %T, want *exitCodeError", err)
	}
	if ece.code != 2 {
		t.Errorf("exit code: got %d, want 2", ece.code)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected diagnostic on stderr; got: %q", stderr.String())
	}
}
