package ripwireadapter

// Fake-ripwire harness tests — openspec/changes/ripwire-adapter task 1.3.
//
// These tests install the committed testdata/fake-ripwire script on PATH
// and drive it through the same subprocess shape the production runner
// (task 2.3) will use: exec.CommandContext on the PATH-resolved executable
// with metrics.SanitizeEnvironment(nil) as the environment. They verify the
// harness contract the adapter tasks (2.2, 2.3, 8.1) will rely on: recorded
// bytes replay byte-identically by flag, version variants switch exactly,
// and every misconfiguration fails closed with exit 2 and a stderr message.
//
// Naming (the change's own conventions, kept exact to avoid look-alike
// drift): the committed script is testdata/fake-ripwire/ripwire; it is
// installed on PATH as "ripewire" and looked up as "ripewire", the name the
// production adapter resolves (design.md D2); the tool's own version output
// begins "ripwire 0.6.9" (recorded/version.txt).
//
// None of these tests may use t.Parallel(): t.Setenv pins the process-wide
// environment for the duration of each test.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// fakeFixtureLanguages returns the fixture languages that have committed
// ripwire recordings under testdata/recorded/. Every harness behavior test
// runs against both so the replay path is exercised for each recorded corpus.
func fakeFixtureLanguages() []string {
	return []string{"python", "typescript"}
}

// recordingsDir returns the committed recordings directory for a fixture
// language, relative to the package's test working directory.
func recordingsDir(language string) string {
	return filepath.Join("testdata", "recorded", language)
}

// installFakeRipwire installs the committed fake-ripwire script (task 1.3)
// as an executable named "ripewire" in a fresh t.TempDir, prepends that
// directory to PATH so it shadows any real ripewire, and points
// FAKE_RIPWIRE_RECORDINGS at recordings so the fake replays that fixture's
// recorded outputs. It returns the installed binary's path.
//
// The script is copied rather than executed directly from testdata so the
// tests do not depend on the executable bit surviving whatever transport
// delivered the repository, and so every mutation is confined to the test's
// own TempDir. FAKE_RIPWIRE_VARIANT is pinned to empty (the script treats
// that as its default, "valid") so a value inherited from the developer's
// shell cannot leak into a test; variant tests override it with t.Setenv.
//
// Not compatible with t.Parallel(): t.Setenv pins the process-wide
// environment for the duration of the test.
func installFakeRipwire(t *testing.T, recordings string) string {
	t.Helper()

	src := filepath.Join("testdata", "fake-ripwire", "ripwire")
	dir := t.TempDir()
	bin := filepath.Join(dir, "ripewire")

	script, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("installFakeRipwire: read %s: %v", src, err)
	}
	if err := os.WriteFile(bin, script, 0o755); err != nil {
		t.Fatalf("installFakeRipwire: write %s: %v", bin, err)
	}
	// WriteFile's mode is masked by the process umask; Chmod pins the
	// executable bit exactly (Security by Default: 0o755 for executables).
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("installFakeRipwire: chmod %s: %v", bin, err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_RIPWIRE_RECORDINGS", recordings)
	t.Setenv("FAKE_RIPWIRE_VARIANT", "")
	return bin
}

// installFakeRipwireWrapper installs a wrapper script named ripwire on PATH
// that bakes the fake's control variables into the script itself, so the
// fake works even when the caller invokes it through production code that
// sets cmd.Env = metrics.SanitizeEnvironment(nil) — the sanitized
// environment cannot carry FAKE_RIPWIRE_* to the subprocess, and production
// code must never name them. The wrapper sets its own environment before
// exec'ing the committed fake script, which keeps the control variables a
// property of the installed double, not of the parent process.
//
// Each variant gets its own wrapper install (and therefore its own PATH
// directory), so tests select the version-output behavior by choosing the
// install, not by mutating process state. Like installFakeRipwire, this is
// not compatible with t.Parallel().
func installFakeRipwireWrapper(t *testing.T, recordingsDir, variant string) string {
	t.Helper()

	src, err := filepath.Abs(filepath.Join("testdata", "fake-ripwire", "ripwire"))
	if err != nil {
		t.Fatalf("installFakeRipwireWrapper: resolve fake script path: %v", err)
	}
	recordings, err := filepath.Abs(recordingsDir)
	if err != nil {
		t.Fatalf("installFakeRipwireWrapper: resolve recordings path: %v", err)
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "ripwire")
	script := strings.Join([]string{
		"#!/bin/sh",
		"FAKE_RIPWIRE_RECORDINGS=" + shellSingleQuote(recordings),
		"FAKE_RIPWIRE_VARIANT=" + shellSingleQuote(variant),
		"export FAKE_RIPWIRE_RECORDINGS FAKE_RIPWIRE_VARIANT",
		"exec " + shellSingleQuote(src) + " \"$@\"",
		"",
	}, "\n")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("installFakeRipwireWrapper: write %s: %v", bin, err)
	}
	// WriteFile's mode is masked by the process umask; Chmod pins the
	// executable bit exactly.
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatalf("installFakeRipwireWrapper: chmod %s: %v", bin, err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return bin
}

// shellSingleQuote quotes s for safe interpolation into a POSIX shell
// script line.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// runFakeRipwire invokes the installed fake the way the production subprocess
// runner (task 2.3) will invoke the real binary: exec.CommandContext on the
// PATH-resolved executable with metrics.SanitizeEnvironment(nil) as the
// environment, so no host credentials can reach the subprocess. The fake's
// own control variables (FAKE_RIPWIRE_RECORDINGS, FAKE_RIPWIRE_VARIANT) are
// appended on top of the sanitized baseline when present in the test process
// environment; they are test scaffolding, not analyzer input. Tests that
// drive the fake through PRODUCTION code (which must never name those
// variables) use installFakeRipwireWrapper instead, which bakes the control
// variables into the installed double itself.
//
// It returns the subprocess's stdout, stderr, and Run error.
func runFakeRipwire(t *testing.T, args ...string) ([]byte, []byte, error) {
	t.Helper()

	bin, err := exec.LookPath("ripewire")
	if err != nil {
		t.Fatalf("runFakeRipwire: ripewire not on PATH: %v", err)
	}

	env := metrics.SanitizeEnvironment(nil)
	if v, ok := os.LookupEnv("FAKE_RIPWIRE_RECORDINGS"); ok {
		env = append(env, "FAKE_RIPWIRE_RECORDINGS="+v)
	}
	if v, ok := os.LookupEnv("FAKE_RIPWIRE_VARIANT"); ok {
		env = append(env, "FAKE_RIPWIRE_VARIANT="+v)
	}

	cmd := exec.CommandContext(t.Context(), bin, args...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), runErr
}

// readRecording reads a committed recording that the fake must replay
// byte-for-byte.
func readRecording(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read recording %s: %v", path, err)
	}
	return data
}

// assertFakeOutput runs the installed fake with args and asserts that it
// exits 0, writes nothing to stderr, and writes exactly want to stdout.
func assertFakeOutput(t *testing.T, want []byte, args ...string) {
	t.Helper()

	stdout, stderr, runErr := runFakeRipwire(t, args...)
	if runErr != nil {
		t.Fatalf("run fake with %v: %v", args, runErr)
	}
	if len(stderr) > 0 {
		t.Errorf("stderr: got %q, want empty", stderr)
	}
	if !bytes.Equal(stdout, want) {
		t.Errorf("stdout for %v: got %d bytes, want %d bytes\n got: %q\nwant: %q",
			args, len(stdout), len(want), stdout, want)
	}
}

// assertFakeFailClosed asserts that a fake invocation failed closed: exit
// status 2 (the harness's misconfiguration status) with a non-empty stderr
// containing wantInStderr, and no stdout replay.
func assertFakeFailClosed(t *testing.T, stdout, stderr []byte, runErr error, wantInStderr string) {
	t.Helper()
	if runErr == nil {
		t.Fatal("fake exited 0, want a nonzero exit")
	}
	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		t.Fatalf("fake run error = %v, want *exec.ExitError", runErr)
	}
	if code := exitErr.ExitCode(); code != 2 {
		t.Errorf("fake exit code = %d, want 2", code)
	}
	if len(stdout) > 0 {
		t.Errorf("stdout: got %q, want empty on failure", stdout)
	}
	if !bytes.Contains(stderr, []byte(wantInStderr)) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, wantInStderr)
	}
}

// TestFakeRipwire_LookPathFindsInstalledFake verifies harness requirement
// (a): with the fake installed for either fixture's recordings,
// exec.LookPath resolves "ripewire" to the installed fake — the prepended
// directory shadows any real ripewire on the machine running the tests.
func TestFakeRipwire_LookPathFindsInstalledFake(t *testing.T) {
	for _, language := range fakeFixtureLanguages() {
		t.Run(language, func(t *testing.T) {
			want := installFakeRipwire(t, recordingsDir(language))

			got, err := exec.LookPath("ripewire")
			if err != nil {
				t.Fatalf("exec.LookPath(ripewire): %v", err)
			}
			if got != want {
				t.Errorf("exec.LookPath(ripewire) = %q, want the installed fake at %q", got, want)
			}
		})
	}
}

// TestFakeRipwire_ReplaysRecordedOutputsByFlag verifies harness requirement
// (b): the fake, invoked through exec.CommandContext with the sanitized
// environment (the precursor of the task 2.3 production runner), returns
// exit 0, empty stderr, and stdout byte-identical to the corresponding
// recorded file for each output flag, for both fixture corpora.
func TestFakeRipwire_ReplaysRecordedOutputsByFlag(t *testing.T) {
	flags := []struct {
		name      string
		args      []string
		recording string
	}{
		{
			name:      "version",
			args:      []string{"--version"},
			recording: "version.txt",
		},
		{
			// Mirrors the task 1.2 capture shape: a real rules-file
			// path, which the fake accepts and ignores.
			name: "arch",
			args: []string{
				"--arch=" + filepath.Join("testdata", "recorded", "arch-rules.txt"),
			},
			recording: "arch.txt",
		},
		{
			name:      "deps",
			args:      []string{"--deps"},
			recording: "deps.xml",
		},
		{
			name:      "export",
			args:      []string{"--export=cc.json"},
			recording: "cc.json",
		},
	}

	for _, language := range fakeFixtureLanguages() {
		t.Run(language, func(t *testing.T) {
			installFakeRipwire(t, recordingsDir(language))
			for _, flag := range flags {
				t.Run(flag.name, func(t *testing.T) {
					assertFakeOutput(t,
						readRecording(t, filepath.Join(recordingsDir(language), flag.recording)),
						flag.args...)
				})
			}
		})
	}
}

// TestFakeRipwire_VersionVariantsReturnExactBytes verifies harness
// requirement (c): the wrong-version and garbage variants return their exact
// expected bytes via --version. Both outputs are newline-terminated lines,
// matching the real binary's line-oriented version output (the recorded
// version.txt is 91 bytes including its trailing newline).
func TestFakeRipwire_VersionVariantsReturnExactBytes(t *testing.T) {
	variants := []struct {
		name    string
		variant string
		want    string
	}{
		{
			name:    "wrong_version",
			variant: "wrong-version",
			// Well-formed but matching neither the pinned version (0.6.9)
			// nor the pinned commit prefix (bedc0cd): the exact line the
			// fake-ripwire contract prescribes.
			want: "ripwire 0.6.8 (Release, AppleClang 21.0.0.21000101, emit=std::print, built_from=deadbeef00)\n",
		},
		{
			name:    "garbage",
			variant: "garbage",
			want:    "totally not a version string\n",
		},
	}

	for _, language := range fakeFixtureLanguages() {
		t.Run(language, func(t *testing.T) {
			installFakeRipwire(t, recordingsDir(language))
			for _, tc := range variants {
				t.Run(tc.name, func(t *testing.T) {
					t.Setenv("FAKE_RIPWIRE_VARIANT", tc.variant)
					assertFakeOutput(t, []byte(tc.want), "--version")
				})
			}
		})
	}
}

// TestFakeRipwire_UnknownVariantFailsClosed verifies the fake's variant
// contract beyond the prescribed variants: any other FAKE_RIPWIRE_VARIANT
// value fails closed with exit 2 and a stderr message naming the value, so a
// typo'd variant in a future test fails loudly instead of silently replaying
// the valid version.
func TestFakeRipwire_UnknownVariantFailsClosed(t *testing.T) {
	for _, language := range fakeFixtureLanguages() {
		t.Run(language, func(t *testing.T) {
			installFakeRipwire(t, recordingsDir(language))
			t.Setenv("FAKE_RIPWIRE_VARIANT", "no-such-variant")

			stdout, stderr, runErr := runFakeRipwire(t, "--version")
			assertFakeFailClosed(t, stdout, stderr, runErr, "no-such-variant")
		})
	}
}

// TestFakeRipwire_UnknownFlagFailsClosed verifies harness requirement (d):
// an unknown flag exits nonzero with a non-empty stderr naming the flag.
func TestFakeRipwire_UnknownFlagFailsClosed(t *testing.T) {
	for _, language := range fakeFixtureLanguages() {
		t.Run(language, func(t *testing.T) {
			installFakeRipwire(t, recordingsDir(language))

			stdout, stderr, runErr := runFakeRipwire(t, "--deps", "--frobnicate=1")
			assertFakeFailClosed(t, stdout, stderr, runErr, "--frobnicate=1")
		})
	}
}

// TestFakeRipwire_MissingRecordingsEnvFailsClosed verifies harness
// requirement (e): with FAKE_RIPWIRE_RECORDINGS unset, the fake exits
// nonzero with a stderr message instead of guessing which recordings to
// replay.
func TestFakeRipwire_MissingRecordingsEnvFailsClosed(t *testing.T) {
	for _, language := range fakeFixtureLanguages() {
		t.Run(language, func(t *testing.T) {
			installFakeRipwire(t, recordingsDir(language))
			// Unset rather than t.Setenv(..., ""): the contract under test
			// is the variable being absent from the subprocess environment.
			// t.Setenv's cleanup still restores the pre-install value when
			// the test ends.
			if err := os.Unsetenv("FAKE_RIPWIRE_RECORDINGS"); err != nil {
				t.Fatalf("unset FAKE_RIPWIRE_RECORDINGS: %v", err)
			}

			stdout, stderr, runErr := runFakeRipwire(t, "--version")
			assertFakeFailClosed(t, stdout, stderr, runErr, "FAKE_RIPWIRE_RECORDINGS")
		})
	}
}

// TestFakeRipwire_AcceptsRunnerFlagsInAnyOrder exercises the fake's argument
// contract beyond the minimal per-flag invocations: the positional crawl
// root, --no-cache, --limit, and --offset are accepted and ignored in any
// position, mirroring the argument shapes of the task 1.2 capture commands
// (and therefore of the future task 2.3 runner). The argument scanner is
// recordings-independent, so unlike the requirement (a)-(e) tests this runs
// against a single fixture corpus.
func TestFakeRipwire_AcceptsRunnerFlagsInAnyOrder(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		recording string
	}{
		{
			// The exact --deps invocation shape from the task 1.2 capture
			// commands, crawl root last.
			name:      "deps_with_limit_and_root_last",
			args:      []string{"--deps", "--limit=1000", "--no-cache", "."},
			recording: "deps.xml",
		},
		{
			// The --arch capture shape: crawl root first.
			name: "arch_with_root_first",
			args: []string{
				".",
				"--no-cache",
				"--arch=" + filepath.Join("testdata", "recorded", "arch-rules.txt"),
			},
			recording: "arch.txt",
		},
		{
			name:      "export_with_offset",
			args:      []string{"--no-cache", "--offset=0", ".", "--export=cc.json"},
			recording: "cc.json",
		},
	}

	installFakeRipwire(t, recordingsDir("python"))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFakeOutput(t,
				readRecording(t, filepath.Join(recordingsDir("python"), tc.recording)),
				tc.args...)
		})
	}
}
