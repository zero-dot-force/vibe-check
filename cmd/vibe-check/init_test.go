package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/internal/scaffold"
)

// initAssetNames lists the category-prefixed asset names that scaffold.Run
// deploys. The scaffold writer reports assets with category prefixes (e.g.
// "agents/divisor-entropy.md"), sorted lexicographically. These appear in
// InitResult slices and in the --json payload.
var initAssetNames = []string{
	"agents/divisor-entropy.md",
	"agents/mx-f-architecture-trend.md",
	"agents/vibe-check-reporter.md",
	"commands/vibe-check.md",
}

// deployedAgentAssetPath returns the on-disk location of the divisor-entropy
// agent asset for a given project root.
func deployedAgentAssetPath(root string) string {
	return filepath.Join(root, ".opencode", "agents", "divisor-entropy.md")
}

// deployedTrendAgentAssetPath returns the on-disk location of the
// mx-f-architecture-trend agent asset for a given project root.
func deployedTrendAgentAssetPath(root string) string {
	return filepath.Join(root, ".opencode", "agents", "mx-f-architecture-trend.md")
}

// deployedCommandAssetPath returns the on-disk location of the vibe-check
// command asset for a given project root.
func deployedCommandAssetPath(root string) string {
	return filepath.Join(root, ".opencode", "commands", "vibe-check.md")
}

// deployedCIWorkflowPath returns the on-disk location of the scaffolded
// structural-gate caller workflow for a given project root.
func deployedCIWorkflowPath(root string) string {
	return filepath.Join(root, ".github", "workflows", scaffold.CIWorkflowName)
}

// ciWorkflowAssetName is the category-prefixed name the scaffolded caller
// workflow is reported under in InitResult slices and the --json payload.
var ciWorkflowAssetName = "workflows/" + scaffold.CIWorkflowName

// assertInitNoWrites asserts a failed init left the target untouched: no
// .opencode asset tree and no .github workflow tree were created.
func assertInitNoWrites(t *testing.T, dir string) {
	t.Helper()
	for _, sub := range []string{".opencode", ".github"} {
		if _, err := os.Stat(filepath.Join(dir, sub)); !os.IsNotExist(err) {
			t.Errorf("failed init must not create %s: stat err %v", sub, err)
		}
	}
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
	// The t.TempDir target carries no .uf/ policy files, so the registration
	// pass legitimately reports them as absent via the Notices section.
	if !strings.Contains(out, "Notices:") {
		t.Errorf("summary missing Notices section for absent policy files; got:\n%s", out)
	}
	if !strings.Contains(out, ".uf/reviewer-capabilities.yaml: not present; not created") {
		t.Errorf("summary missing absent-capabilities notice; got:\n%s", out)
	}
	if !strings.Contains(out, "agents/divisor-entropy.md") {
		t.Errorf("summary missing agent asset name; got:\n%s", out)
	}
	if !strings.Contains(out, "agents/mx-f-architecture-trend.md") {
		t.Errorf("summary missing trend agent asset name; got:\n%s", out)
	}
	if !strings.Contains(out, "commands/vibe-check.md") {
		t.Errorf("summary missing command asset name; got:\n%s", out)
	}
	if _, statErr := os.Stat(deployedAgentAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed agent asset on disk: %v", statErr)
	}
	if _, statErr := os.Stat(deployedTrendAgentAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed trend agent asset on disk: %v", statErr)
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
	if !strings.Contains(stdout.String(), "agents/mx-f-architecture-trend.md") {
		t.Errorf("expected trend agent asset name in output; got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "commands/vibe-check.md") {
		t.Errorf("expected command asset name in output; got:\n%s", stdout.String())
	}
	if _, statErr := os.Stat(deployedAgentAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed agent asset on disk: %v", statErr)
	}
	if _, statErr := os.Stat(deployedTrendAgentAssetPath(dir)); statErr != nil {
		t.Errorf("expected deployed trend agent asset on disk: %v", statErr)
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
	if !slices.Contains(p.Written, "agents/mx-f-architecture-trend.md") {
		t.Errorf("expected agents/mx-f-architecture-trend.md in written; got %v", p.Written)
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

// TestRunInit_JSONNoticesListsAbsentPolicyFiles covers the --json notices key:
// a target without .uf/ policy files reports the absent-file caveats as a JSON
// array while the deployment still succeeds with exit code 0.
func TestRunInit_JSONNoticesListsAbsentPolicyFiles(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   t.TempDir(),
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode: got %d, want 0", res.ExitCode)
	}
	wantNotices := []string{
		".uf/reviewer-capabilities.yaml: not present; not created",
		".uf/review-matrix.yaml: not present; not created",
	}
	if got := res.Notices; !slices.Equal(got, wantNotices) {
		t.Errorf("Notices: got %v, want %v", got, wantNotices)
	}
	var payload struct {
		Notices []string `json:"notices"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", stdout.String(), err)
	}
	if !slices.Equal(payload.Notices, wantNotices) {
		t.Errorf("json notices: got %v, want %v", payload.Notices, wantNotices)
	}
}

// TestWriteInitJSON_NoticesNormalizedToEmptyArray verifies the notices key is
// always a JSON array, never null, even when no caveats were reported.
func TestWriteInitJSON_NoticesNormalizedToEmptyArray(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := writeInitJSON(&buf, &scaffold.Result{}, true); err != nil {
		t.Fatalf("writeInitJSON returned error: %v", err)
	}
	if !strings.Contains(buf.String(), `"notices": []`) {
		t.Errorf("JSON payload missing empty notices array; got:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "null") {
		t.Errorf("JSON payload must not contain null arrays; got:\n%s", buf.String())
	}
}

// TestWriteInitSummary_NoticesSectionOnlyWhenPresent verifies the human
// summary renders a Notices section exactly when the registration pass
// reported caveats, and a CI next-steps section exactly when init ran with
// --ci (the two conditional sections render independently).
func TestWriteInitSummary_NoticesSectionOnlyWhenPresent(t *testing.T) {
	t.Parallel()
	var quiet bytes.Buffer
	writeInitSummary(&quiet, ".", &scaffold.Result{}, false)
	if strings.Contains(quiet.String(), "Notices:") {
		t.Errorf("summary without notices must omit the Notices section; got:\n%s", quiet.String())
	}
	if strings.Contains(quiet.String(), "CI workflow next steps") {
		t.Errorf("summary from a plain init must omit the CI next-steps section; got:\n%s", quiet.String())
	}

	var noisy bytes.Buffer
	writeInitSummary(&noisy, ".", &scaffold.Result{
		Notices: []string{".uf/reviewer-capabilities.yaml: not present; not created"},
	}, false)
	if !strings.Contains(noisy.String(), "Notices:") {
		t.Errorf("summary with notices missing the Notices section; got:\n%s", noisy.String())
	}
	if !strings.Contains(noisy.String(), ".uf/reviewer-capabilities.yaml: not present; not created") {
		t.Errorf("summary with notices missing the notice line; got:\n%s", noisy.String())
	}

	var ciOut bytes.Buffer
	writeInitSummary(&ciOut, ".", &scaffold.Result{
		Written: []string{"workflows/vibe-check-structural-gate.yaml"},
	}, true)
	if !strings.Contains(ciOut.String(), "CI workflow next steps") {
		t.Errorf("summary from a --ci run missing the CI next-steps section; got:\n%s", ciOut.String())
	}
	if !strings.Contains(ciOut.String(), ".github/workflows/vibe-check-structural-gate.yaml") {
		t.Errorf("CI next-steps section must name the deployed workflow path; got:\n%s", ciOut.String())
	}
	if !strings.Contains(ciOut.String(), "branch protection") {
		t.Errorf("CI next-steps section must state the gate is advisory until branch protection; got:\n%s", ciOut.String())
	}
	if strings.Contains(ciOut.String(), "Notices:") {
		t.Errorf("summary without notices must omit the Notices section even under --ci; got:\n%s", ciOut.String())
	}
}

// writeUfPolicyFixtures seeds a minimal but valid UF policy pair so the
// CLI-level registration tests exercise the same shapes Run meets in real
// projects.
func writeUfPolicyFixtures(t *testing.T, dir string) {
	t.Helper()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	capabilities := "reviewers:\n" +
		"  - agent: divisor-guard\n" +
		"    capability: review\n" +
		"    scopes:\n" +
		"      - standard\n"
	if err := os.WriteFile(filepath.Join(ufDir, "reviewer-capabilities.yaml"), []byte(capabilities), 0o644); err != nil {
		t.Fatalf("write reviewer-capabilities.yaml: %v", err)
	}
	matrix := "advisor:\n" +
		"  code:\n    - divisor-guard\n" +
		"  spec:\n    - divisor-guard\n" +
		"  triage:\n    - divisor-guard\n" +
		"  feedback:\n    - divisor-guard\n"
	if err := os.WriteFile(filepath.Join(ufDir, "review-matrix.yaml"), []byte(matrix), 0o644); err != nil {
		t.Fatalf("write review-matrix.yaml: %v", err)
	}
}

// TestRunInit_RegistersReviewAgentsInPayload verifies end to end that init
// registers the divisor-entropy reviewer into both UF policy files and
// reports both registrations through the --json payload's written array
// (and the human summary), rather than as skipped or forced assets.
func TestRunInit_RegistersReviewAgentsInPayload(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	dir := t.TempDir()
	writeUfPolicyFixtures(t, dir)

	res, err := RunInit(context.Background(), InitOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   dir,
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode: got %d, want 0", res.ExitCode)
	}

	var payload struct {
		Written []string `json:"written"`
		Skipped []string `json:"skipped"`
		Forced  []string `json:"forced"`
		Notices []string `json:"notices"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", stdout.String(), err)
	}
	for _, name := range []string{"reviewers/divisor-entropy", "advisors/divisor-entropy"} {
		if !slices.Contains(payload.Written, name) {
			t.Errorf("json written missing %q; got %v", name, payload.Written)
		}
		if slices.Contains(payload.Skipped, name) {
			t.Errorf("json skipped must not contain %q; got %v", name, payload.Skipped)
		}
		if slices.Contains(payload.Forced, name) {
			t.Errorf("json forced must not contain %q; got %v", name, payload.Forced)
		}
	}
	if len(payload.Notices) != 0 {
		t.Errorf("json notices: got %v, want none", payload.Notices)
	}

	var human bytes.Buffer
	humanDir := t.TempDir()
	writeUfPolicyFixtures(t, humanDir)
	if _, err := RunInit(context.Background(), InitOptions{
		Stdout: &human,
		Stderr: &stderr,
		Path:   humanDir,
	}); err != nil {
		t.Fatalf("RunInit (human) returned error: %v", err)
	}
	for _, name := range []string{"reviewers/divisor-entropy", "advisors/divisor-entropy"} {
		if !strings.Contains(human.String(), name) {
			t.Errorf("human summary missing %q; got:\n%s", name, human.String())
		}
	}
}

// TestInitFlagParsing verifies the init command registers the --ci and
// --vibe-version flags and parses them alongside a positional target.
func TestInitFlagParsing(t *testing.T) {
	t.Parallel()
	cmd := initCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Flags().Parse([]string{"--ci", "--vibe-version", "v0.4.1", "target"}); err != nil {
		t.Fatalf("Flags().Parse: %v", err)
	}
	ci, err := cmd.Flags().GetBool("ci")
	if err != nil || !ci {
		t.Errorf("--ci: got (%v, %v), want (true, nil)", ci, err)
	}
	v, err := cmd.Flags().GetString("vibe-version")
	if err != nil || v != "v0.4.1" {
		t.Errorf("--vibe-version: got (%q, %v), want (%q, nil)", v, err, "v0.4.1")
	}
}

// TestRunInit_CIOverrideWithoutCIIsUsageError covers the D12 flag-combination
// rule: --vibe-version without --ci is a usage error that exits 2 before any
// work, leaving the target untouched and stdout empty.
func TestRunInit_CIOverrideWithoutCIIsUsageError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout:      &stdout,
		Stderr:      &stderr,
		Path:        dir,
		VibeVersion: "v0.4.1",
	})
	if err == nil {
		t.Fatalf("expected usage error, got success: %+v", res)
	}
	if res.ExitCode != 2 {
		t.Errorf("ExitCode: got %d, want 2", res.ExitCode)
	}
	if !strings.Contains(err.Error(), "--vibe-version requires --ci") {
		t.Errorf("error must name the flag combination; got: %v", err)
	}
	assertInitNoWrites(t, dir)
	if stdout.Len() != 0 {
		t.Errorf("usage error must write nothing to stdout; got:\n%s", stdout.String())
	}
}

// TestRunInit_CIPinResolutionFailsClosed walks design D3's resolution channel
// table end to end through RunInit: a dev build, an unparseable version, and a
// below-floor tag or override all fail closed with the D6 tag error before any
// writes; a release binary pins its embedded commit fully offline; a release
// tag without an embedded commit (or with a --vibe-version override, which
// always forces resolution) resolves the SHA at init time.
func TestRunInit_CIPinResolutionFailsClosed(t *testing.T) {
	// This test mutates the package-level version and commit vars, so it must
	// not run in parallel (TestSemanticVersion precedent).
	origVersion, origCommit := version, commit
	defer func() { version, commit = origVersion, origCommit }()

	embeddedSHA := strings.Repeat("a", 40)
	resolvedSHA := strings.Repeat("b", 40)

	cases := []struct {
		name        string
		version     string
		commit      string
		vibeVersion string
		wantErr     string // substring expected in the error; empty means success
		wantTag     string // tag pin expected in the emitted workflow on success
		wantSHA     string // SHA pin expected in the emitted workflow on success
		wantResolve bool   // whether tag→SHA resolution must run
	}{
		{
			name:    "dev build fails closed",
			version: "dev", commit: "none",
			wantErr: `cannot resolve a vibe-check release tag >= v0.3.0 (got "dev")`,
		},
		{
			name:    "below-floor release fails closed",
			version: "v0.2.9", commit: "none",
			wantErr: `cannot resolve a vibe-check release tag >= v0.3.0 (got "v0.2.9")`,
		},
		{
			name:    "unparseable version fails closed",
			version: "1.2.3", commit: "none",
			wantErr: `cannot resolve a vibe-check release tag >= v0.3.0 (got "1.2.3")`,
		},
		{
			name:    "below-floor override fails closed",
			version: "dev", commit: "none", vibeVersion: "v0.1.0",
			wantErr: `cannot resolve a vibe-check release tag >= v0.3.0 (got "v0.1.0")`,
		},
		{
			name:    "release binary pins embedded commit offline",
			version: "v0.4.2", commit: embeddedSHA,
			wantTag: "v0.4.2", wantSHA: embeddedSHA,
		},
		{
			name:    "release tag without embedded commit resolves remotely",
			version: "v0.3.1", commit: "none",
			wantTag: "v0.3.1", wantSHA: resolvedSHA, wantResolve: true,
		},
		{
			name:    "override forces resolution even with embedded commit",
			version: "v0.4.2", commit: embeddedSHA, vibeVersion: "v0.5.0",
			wantTag: "v0.5.0", wantSHA: resolvedSHA, wantResolve: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			version, commit = tc.version, tc.commit
			dir := t.TempDir()
			var stdout, stderr bytes.Buffer
			resolved := false
			res, err := RunInit(context.Background(), InitOptions{
				Stdout:      &stdout,
				Stderr:      &stderr,
				Path:        dir,
				CI:          true,
				VibeVersion: tc.vibeVersion,
				resolveTagCommit: func(ctx context.Context, tag string) (string, error) {
					resolved = true
					if tag != tc.wantTag {
						t.Errorf("resolver called with tag %q, want %q", tag, tc.wantTag)
					}
					return resolvedSHA, nil
				},
			})
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected failure, got success: %+v", res)
				}
				if res.ExitCode != 2 {
					t.Errorf("ExitCode: got %d, want 2", res.ExitCode)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error missing %q; got: %v", tc.wantErr, err)
				}
				if !strings.Contains(err.Error(), "install vibe-check from a release tag or pass --vibe-version") {
					t.Errorf("error missing the D6 remediation; got: %v", err)
				}
				assertInitNoWrites(t, dir)
				if stdout.Len() != 0 {
					t.Errorf("failed init must write nothing to stdout; got:\n%s", stdout.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.ExitCode != 0 {
				t.Fatalf("ExitCode: got %d, want 0", res.ExitCode)
			}
			if resolved != tc.wantResolve {
				t.Errorf("tag→SHA resolution ran: got %v, want %v", resolved, tc.wantResolve)
			}
			if !slices.Contains(res.Written, ciWorkflowAssetName) {
				t.Errorf("Written missing %q; got %v", ciWorkflowAssetName, res.Written)
			}
			data, readErr := os.ReadFile(deployedCIWorkflowPath(dir))
			if readErr != nil {
				t.Fatalf("read deployed workflow: %v", readErr)
			}
			if !strings.Contains(string(data), "@"+tc.wantSHA) {
				t.Errorf("workflow uses: pin: got file without @%s;\n%s", tc.wantSHA, data)
			}
			if !strings.Contains(string(data), `version: "`+tc.wantTag+`"`) {
				t.Errorf("workflow version pin: got file without version %q;\n%s", tc.wantTag, data)
			}
		})
	}
}

// TestRunInit_CIResolutionFailureExitsTwo covers the offline tag→SHA
// resolution failure path: the error carries the D6 remediation and the run
// exits 2 before any writes, leaving stdout empty.
func TestRunInit_CIResolutionFailureExitsTwo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout:      &stdout,
		Stderr:      &stderr,
		Path:        dir,
		CI:          true,
		VibeVersion: "v0.4.1",
		resolveTagCommit: func(ctx context.Context, tag string) (string, error) {
			return "", errors.New("network unreachable")
		},
	})
	if err == nil {
		t.Fatalf("expected resolution failure, got success: %+v", res)
	}
	if res.ExitCode != 2 {
		t.Errorf("ExitCode: got %d, want 2", res.ExitCode)
	}
	if !strings.Contains(err.Error(), `cannot resolve the release commit SHA for tag "v0.4.1"`) {
		t.Errorf("error must name the tag; got: %v", err)
	}
	if !strings.Contains(err.Error(), "install vibe-check from a release binary or rerun with network access") {
		t.Errorf("error missing the D6 remediation; got: %v", err)
	}
	assertInitNoWrites(t, dir)
	if stdout.Len() != 0 {
		t.Errorf("failed init must write nothing to stdout; got:\n%s", stdout.String())
	}
}

// TestRunInit_CILifecycle covers the --ci output contract end to end (tasks
// 3.1 and 3.3): the first run writes the caller workflow (reported in the
// JSON payload's written array) pinned to the override tag and the stubbed
// SHA; a rerun reports it skipped and leaves the file byte-identical; and a
// forced rerun rewrites the .opencode assets but still skips the workflow
// (design D7: the emitted file is user property), with the human summary
// carrying the CI next-steps section.
func TestRunInit_CILifecycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	const overrideTag = "v0.4.1"
	stubSHA := strings.Repeat("c", 40)
	resolve := func(ctx context.Context, tag string) (string, error) { return stubSHA, nil }

	// First run (JSON): assets plus caller workflow written; the payload
	// carries the workflow path through the existing written array.
	var jsonOut, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout:           &jsonOut,
		Stderr:           &stderr,
		Path:             dir,
		JSON:             true,
		CI:               true,
		VibeVersion:      overrideTag,
		resolveTagCommit: resolve,
	})
	if err != nil {
		t.Fatalf("first RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("first run ExitCode: got %d, want 0", res.ExitCode)
	}
	var payload struct {
		Written []string `json:"written"`
		Skipped []string `json:"skipped"`
		Forced  []string `json:"forced"`
		Notices []string `json:"notices"`
	}
	if err := json.Unmarshal(jsonOut.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", jsonOut.String(), err)
	}
	if !slices.Contains(payload.Written, ciWorkflowAssetName) {
		t.Errorf("json written missing %q; got %v", ciWorkflowAssetName, payload.Written)
	}
	if !slices.Contains(payload.Written, "agents/divisor-entropy.md") {
		t.Errorf("json written missing agent asset; got %v", payload.Written)
	}
	first, readErr := os.ReadFile(deployedCIWorkflowPath(dir))
	if readErr != nil {
		t.Fatalf("read deployed workflow: %v", readErr)
	}
	if !strings.Contains(string(first), "@"+stubSHA) {
		t.Errorf("workflow uses: pin: got file without @%s;\n%s", stubSHA, first)
	}
	if !strings.Contains(string(first), `version: "`+overrideTag+`"`) {
		t.Errorf("workflow version pin: got file without version %q;\n%s", overrideTag, first)
	}

	// Second run (JSON): the caller workflow is skipped and byte-identical.
	var rerunOut bytes.Buffer
	res2, err := RunInit(context.Background(), InitOptions{
		Stdout:           &rerunOut,
		Stderr:           &stderr,
		Path:             dir,
		JSON:             true,
		CI:               true,
		VibeVersion:      overrideTag,
		resolveTagCommit: resolve,
	})
	if err != nil {
		t.Fatalf("second RunInit returned error: %v", err)
	}
	if res2.ExitCode != 0 {
		t.Fatalf("second run ExitCode: got %d, want 0", res2.ExitCode)
	}
	if err := json.Unmarshal(rerunOut.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", rerunOut.String(), err)
	}
	if !slices.Contains(payload.Skipped, ciWorkflowAssetName) {
		t.Errorf("rerun json skipped missing %q; got %v", ciWorkflowAssetName, payload.Skipped)
	}
	if slices.Contains(payload.Written, ciWorkflowAssetName) {
		t.Errorf("rerun json written must not contain %q; got %v", ciWorkflowAssetName, payload.Written)
	}
	second, readErr := os.ReadFile(deployedCIWorkflowPath(dir))
	if readErr != nil {
		t.Fatalf("read deployed workflow after rerun: %v", readErr)
	}
	if string(first) != string(second) {
		t.Errorf("rerun must leave the workflow byte-identical; first %d bytes vs second %d bytes", len(first), len(second))
	}

	// Third run (human, --force): .opencode assets are forced, the workflow
	// is still skipped (design D7), and the summary closes with the CI
	// next-steps section.
	var humanOut bytes.Buffer
	res3, err := RunInit(context.Background(), InitOptions{
		Stdout:           &humanOut,
		Stderr:           &stderr,
		Path:             dir,
		Force:            true,
		CI:               true,
		VibeVersion:      overrideTag,
		resolveTagCommit: resolve,
	})
	if err != nil {
		t.Fatalf("third RunInit returned error: %v", err)
	}
	if !slices.Contains(res3.Forced, "agents/divisor-entropy.md") {
		t.Errorf("forced run must rewrite agent assets; got %v", res3.Forced)
	}
	if !slices.Contains(res3.Skipped, ciWorkflowAssetName) {
		t.Errorf("forced run must still skip the workflow; got skipped=%v", res3.Skipped)
	}
	if slices.Contains(res3.Forced, ciWorkflowAssetName) {
		t.Errorf("forced run must never force the workflow; got forced=%v", res3.Forced)
	}
	if !strings.Contains(humanOut.String(), "CI workflow next steps") {
		t.Errorf("summary from a --ci run missing the CI next-steps section; got:\n%s", humanOut.String())
	}
}

// TestLsRemoteTagCommitAt_LocalRepository exercises the real tag→SHA resolver
// code path (git ls-remote invocation, ^{} peeling, output parsing) against a
// local throwaway repository: an annotated tag must peel to the commit SHA it
// marks, an absent tag must error rather than return a zero value, and a
// cancelled context must abort the subprocess.
func TestLsRemoteTagCommitAt_LocalRepository(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	dir := t.TempDir()
	gitCmd := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	gitCmd("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("fixture"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	gitCmd("add", "README")
	gitCmd("commit", "-q", "-m", "fixture commit")
	gitCmd("tag", "-a", "v0.3.0", "-m", "fixture release")
	headSHA := strings.TrimSpace(gitCmd("rev-parse", "HEAD"))

	got, err := lsRemoteTagCommitAt(context.Background(), dir, "v0.3.0")
	if err != nil {
		t.Fatalf("lsRemoteTagCommitAt returned error: %v", err)
	}
	if got != headSHA {
		t.Errorf("peeled SHA: got %q, want %q (HEAD)", got, headSHA)
	}

	if _, err := lsRemoteTagCommitAt(context.Background(), dir, "v9.9.9"); err == nil {
		t.Errorf("absent tag must return an error, not a zero value")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := lsRemoteTagCommitAt(ctx, dir, "v0.3.0"); err == nil {
		t.Errorf("cancelled context must abort resolution")
	}
}

// TestRunInit_PlainInitWritesNoGithubTree covers task 3.2: a plain init (no
// --ci) creates no .github tree and its summary carries no CI section — the
// default behavior is byte-for-byte what it was before the flag existed.
func TestRunInit_PlainInitWritesNoGithubTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{Stdout: &stdout, Stderr: &stderr, Path: dir})
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode: got %d, want 0", res.ExitCode)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".github")); !os.IsNotExist(statErr) {
		t.Errorf("plain init must not create a .github tree; stat err: %v", statErr)
	}
	if strings.Contains(stdout.String(), "CI workflow next steps") {
		t.Errorf("plain init summary must omit the CI next-steps section; got:\n%s", stdout.String())
	}
}
