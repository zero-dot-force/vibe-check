package metrics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestHelperProcess is not a real test — it is invoked by ExternalAdapter tests
// as a mock subprocess. It reads JSON-RPC requests from stdin and writes
// responses to stdout. The HELPER_MODE environment variable controls behavior.
func TestHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	mode := os.Getenv("HELPER_MODE")
	switch mode {
	case "success":
		helperSuccess()
	case "timeout":
		helperTimeout()
	case "crash":
		os.Exit(1)
	case "stderr":
		helperStderr()
	case "oversized":
		helperOversized()
	case "env_check":
		helperEnvCheck()
	default:
		fmt.Fprintf(os.Stderr, "unknown HELPER_MODE: %s\n", mode)
		os.Exit(2)
	}
}

// helperSuccess reads JSON-RPC requests and returns valid responses.
func helperSuccess() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			fmt.Fprintf(os.Stderr, "unmarshal error: %v\n", err)
			os.Exit(1)
		}

		switch req.Method {
		case "capabilities":
			result := CapabilitiesResult{
				Language:        "test",
				ProtocolVersion: "1.0",
				Metrics:         []string{"ca", "ce", "instability"},
			}
			sendResponse(req.ID, result)

		case "analyze":
			graph := ModuleGraph{
				SchemaVersion: "1.0",
				Language:      "test",
				Modules: []ModuleResult{
					{
						Module: Module{
							Path:          "example/pkg",
							Name:          "pkg",
							Ca:            2,
							Ce:            3,
							ExportedTypes: 4,
							AbstractTypes: 1,
						},
						Instability:  0.6,
						Abstractness: 0.25,
						Distance:     0.15,
						LCOM:         1,
						Zone:         ZoneMainSequence,
					},
				},
				Cycles:   []Cycle{},
				Warnings: []Warning{},
				Status:   StatusComplete,
			}
			sendResponse(req.ID, graph)

		case "shutdown":
			// Notification — no response expected. Exit cleanly.
			os.Exit(0)
		}
	}
}

// helperTimeout sleeps indefinitely to simulate a subprocess that never responds.
func helperTimeout() {
	// Block forever — the test will cancel via context timeout.
	select {}
}

// helperStderr writes to stderr and then processes requests normally.
func helperStderr() {
	fmt.Fprintf(os.Stderr, "analyzer warning: test stderr output")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			os.Exit(1)
		}

		switch req.Method {
		case "capabilities":
			result := CapabilitiesResult{
				Language:        "test",
				ProtocolVersion: "1.0",
				Metrics:         []string{"ca"},
			}
			sendResponse(req.ID, result)

		case "analyze":
			graph := ModuleGraph{
				SchemaVersion: "1.0",
				Language:      "test",
				Modules:       []ModuleResult{},
				Cycles:        []Cycle{},
				Warnings:      []Warning{},
				Status:        StatusComplete,
			}
			sendResponse(req.ID, graph)

		case "shutdown":
			os.Exit(0)
		}
	}
}

// helperOversized writes a response that exceeds the maximum response size.
func helperOversized() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			os.Exit(1)
		}

		switch req.Method {
		case "capabilities":
			// Write a response with a very large result field.
			// The test sets maxResponseSize to a small value.
			bigData := strings.Repeat("x", 2048)
			resp := fmt.Sprintf(`{"jsonrpc":"2.0","result":{"language":"%s","protocolVersion":"1.0","metrics":[]},"id":1}`, bigData)
			_, _ = fmt.Fprintln(os.Stdout, resp)

		case "shutdown":
			os.Exit(0)
		}
	}
}

// helperEnvCheck verifies that credential-bearing environment variables are
// NOT present, then responds normally.
func helperEnvCheck() {
	// Check for blocked variables.
	blockedVars := []string{
		"AWS_SECRET_ACCESS_KEY",
		"GITHUB_TOKEN",
		"GH_TOKEN",
		"NPM_TOKEN",
		"SECRET_TEST_VALUE",
	}
	for _, v := range blockedVars {
		if val := os.Getenv(v); val != "" {
			fmt.Fprintf(os.Stderr, "BLOCKED_VAR_PRESENT:%s=%s", v, val)
			os.Exit(3)
		}
	}

	// Verify PATH is present (required for basic operation).
	if os.Getenv("PATH") == "" {
		fmt.Fprintf(os.Stderr, "PATH_MISSING")
		os.Exit(4)
	}

	// Report success via normal protocol.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			os.Exit(1)
		}

		switch req.Method {
		case "capabilities":
			result := CapabilitiesResult{
				Language:        "test",
				ProtocolVersion: "1.0",
				Metrics:         []string{"ca"},
			}
			sendResponse(req.ID, result)

		case "analyze":
			graph := ModuleGraph{
				SchemaVersion: "1.0",
				Language:      "test",
				Modules:       []ModuleResult{},
				Cycles:        []Cycle{},
				Warnings:      []Warning{},
				Status:        StatusComplete,
			}
			sendResponse(req.ID, graph)

		case "shutdown":
			os.Exit(0)
		}
	}
}

// sendResponse marshals a result and writes a JSON-RPC response to stdout.
func sendResponse(id *int, result any) {
	resultData, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal error: %v\n", err)
		os.Exit(1)
	}

	resp := JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		Result:  resultData,
		ID:      id,
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal response error: %v\n", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintln(os.Stdout, string(respData))
}

func TestExternalAdapter_SuccessfulAnalysis(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(10*time.Second),
		WithCapabilitiesTimeout(5*time.Second),
		WithShutdownTimeout(2*time.Second),
	)
	// Override the binary path and env to use the helper process.
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=success",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	ctx := context.Background()
	graph, err := adapter.Analyze(ctx, projectDir)
	if err != nil {
		t.Fatalf("Analyze() returned unexpected error: %v", err)
	}

	if got, want := graph.Language, "test"; got != want {
		t.Errorf("Language: got %v, want %v", got, want)
	}
	if got, want := graph.Status, StatusComplete; got != want {
		t.Errorf("Status: got %v, want %v", got, want)
	}
	if got, want := len(graph.Modules), 1; got != want {
		t.Fatalf("len(Modules): got %v, want %v", got, want)
	}

	mod := graph.Modules[0]
	if got, want := mod.Path, "example/pkg"; got != want {
		t.Errorf("Module.Path: got %v, want %v", got, want)
	}
	if got, want := mod.Ca, 2; got != want {
		t.Errorf("Module.Ca: got %v, want %v", got, want)
	}
	if got, want := mod.Ce, 3; got != want {
		t.Errorf("Module.Ce: got %v, want %v", got, want)
	}
	if got, want := mod.Zone, ZoneMainSequence; got != want {
		t.Errorf("Module.Zone: got %v, want %v", got, want)
	}
}

func TestExternalAdapter_Timeout(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(200*time.Millisecond),
		WithCapabilitiesTimeout(100*time.Millisecond),
		WithShutdownTimeout(100*time.Millisecond),
	)
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=timeout",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	ctx := context.Background()
	_, err := adapter.Analyze(ctx, projectDir)
	if err == nil {
		t.Fatal("Analyze() returned nil error for timeout scenario, want error")
	}
}

func TestExternalAdapter_Crash(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(5*time.Second),
		WithCapabilitiesTimeout(2*time.Second),
		WithShutdownTimeout(1*time.Second),
	)
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=crash",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	ctx := context.Background()
	_, err := adapter.Analyze(ctx, projectDir)
	if err == nil {
		t.Fatal("Analyze() returned nil error for crash scenario, want error")
	}
}

func TestExternalAdapter_StderrCapture(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(10*time.Second),
		WithCapabilitiesTimeout(5*time.Second),
		WithShutdownTimeout(2*time.Second),
	)
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=stderr",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	ctx := context.Background()
	graph, err := adapter.Analyze(ctx, projectDir)
	if err != nil {
		t.Fatalf("Analyze() returned unexpected error: %v", err)
	}

	// The analysis should still succeed even with stderr output.
	if got, want := graph.Status, StatusComplete; got != want {
		t.Errorf("Status: got %v, want %v", got, want)
	}
}

func TestExternalAdapter_ResponseSizeLimit(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(5*time.Second),
		WithCapabilitiesTimeout(2*time.Second),
		WithShutdownTimeout(1*time.Second),
		WithMaxResponseSize(100), // Very small limit to trigger overflow.
	)
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=oversized",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	ctx := context.Background()
	_, err := adapter.Analyze(ctx, projectDir)
	if err == nil {
		t.Fatal("Analyze() returned nil error for oversized response, want error")
	}
}

// TestExternalAdapter_EnvironmentSanitization verifies that credential-bearing
// environment variables are not passed to the subprocess. Not parallel because
// t.Setenv modifies process-level state.
func TestExternalAdapter_EnvironmentSanitization(t *testing.T) {
	projectDir := t.TempDir()

	// Set credential-bearing env vars that MUST NOT reach the subprocess.
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	t.Setenv("GITHUB_TOKEN", "ghp_test_token")
	t.Setenv("SECRET_TEST_VALUE", "should-not-leak")

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(10*time.Second),
		WithCapabilitiesTimeout(5*time.Second),
		WithShutdownTimeout(2*time.Second),
	)
	adapter.binaryPath = os.Args[0]
	// Build env from scratch using SanitizeEnvironment — this is what
	// NewExternalAdapter does internally.
	sanitized := SanitizeEnvironment(nil)
	adapter.env = append(sanitized,
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=env_check",
	)

	ctx := context.Background()
	graph, err := adapter.Analyze(ctx, projectDir)
	if err != nil {
		t.Fatalf("Analyze() returned unexpected error: %v", err)
	}

	// If the helper process found blocked vars, it would have exited with
	// code 3 and the Analyze call would have failed. A successful result
	// confirms sanitization worked.
	if got, want := graph.Status, StatusComplete; got != want {
		t.Errorf("Status: got %v, want %v", got, want)
	}
}

func TestExternalAdapter_InvalidProjectPath(t *testing.T) {
	t.Parallel()

	adapter := NewExternalAdapter("/bin/echo", "test")

	tests := []struct {
		name string
		path string
	}{
		{
			name: "path with traversal",
			path: "/tmp/../etc/passwd",
		},
		{
			name: "nonexistent path",
			path: "/nonexistent/path/that/does/not/exist",
		},
		{
			name: "empty path",
			path: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			_, err := adapter.Analyze(ctx, tt.path)
			if err == nil {
				t.Errorf("Analyze(%q) returned nil error, want error", tt.path)
			}
		})
	}
}

func TestExternalAdapter_Language(t *testing.T) {
	t.Parallel()

	adapter := NewExternalAdapter("/bin/echo", "python")
	if got, want := adapter.Language(), "python"; got != want {
		t.Errorf("Language(): got %v, want %v", got, want)
	}
}

func TestExternalAdapter_ShutdownLifecycle(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithAnalyzeTimeout(10*time.Second),
		WithCapabilitiesTimeout(5*time.Second),
		WithShutdownTimeout(2*time.Second),
	)
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=success",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	// Run Analyze — the helper process exits on shutdown notification.
	// If shutdown is not sent, the process would hang and the test would
	// time out.
	ctx := context.Background()
	graph, err := adapter.Analyze(ctx, projectDir)
	if err != nil {
		t.Fatalf("Analyze() returned unexpected error: %v", err)
	}

	if got, want := graph.Language, "test"; got != want {
		t.Errorf("Language: got %v, want %v", got, want)
	}
}

func TestNewExternalAdapter_Defaults(t *testing.T) {
	t.Parallel()

	adapter := NewExternalAdapter("/bin/echo", "go")
	if adapter == nil {
		t.Fatal("NewExternalAdapter returned nil")
	}
	if got, want := adapter.Language(), "go"; got != want {
		t.Errorf("Language(): got %v, want %v", got, want)
	}
	if got, want := adapter.analyzeTimeout, defaultAnalyzeTimeout; got != want {
		t.Errorf("analyzeTimeout: got %v, want %v", got, want)
	}
	if got, want := adapter.capabilitiesTimeout, defaultCapabilitiesTimeout; got != want {
		t.Errorf("capabilitiesTimeout: got %v, want %v", got, want)
	}
	if got, want := adapter.shutdownTimeout, defaultShutdownTimeout; got != want {
		t.Errorf("shutdownTimeout: got %v, want %v", got, want)
	}
	if got, want := adapter.maxResponseSize, int64(defaultMaxResponseSize); got != want {
		t.Errorf("maxResponseSize: got %v, want %v", got, want)
	}
	if got, want := adapter.maxStderrSize, int64(defaultMaxStderrSize); got != want {
		t.Errorf("maxStderrSize: got %v, want %v", got, want)
	}
	if adapter.env == nil {
		t.Error("env should be populated by SanitizeEnvironment(nil), got nil")
	}
}

func TestNewExternalAdapter_WithOptions(t *testing.T) {
	t.Parallel()

	adapter := NewExternalAdapter("/bin/echo", "python",
		WithAnalyzeTimeout(42*time.Second),
		WithCapabilitiesTimeout(7*time.Second),
		WithShutdownTimeout(3*time.Second),
		WithMaxResponseSize(512),
		WithMaxStderrSize(256),
	)
	if adapter == nil {
		t.Fatal("NewExternalAdapter returned nil")
	}
	if got, want := adapter.Language(), "python"; got != want {
		t.Errorf("Language(): got %v, want %v", got, want)
	}
	if got, want := adapter.analyzeTimeout, 42*time.Second; got != want {
		t.Errorf("analyzeTimeout: got %v, want %v", got, want)
	}
	if got, want := adapter.capabilitiesTimeout, 7*time.Second; got != want {
		t.Errorf("capabilitiesTimeout: got %v, want %v", got, want)
	}
	if got, want := adapter.shutdownTimeout, 3*time.Second; got != want {
		t.Errorf("shutdownTimeout: got %v, want %v", got, want)
	}
	if got, want := adapter.maxResponseSize, int64(512); got != want {
		t.Errorf("maxResponseSize: got %v, want %v", got, want)
	}
	if got, want := adapter.maxStderrSize, int64(256); got != want {
		t.Errorf("maxStderrSize: got %v, want %v", got, want)
	}
}

func TestExternalAdapter_Capabilities(t *testing.T) {
	t.Parallel()

	adapter := NewExternalAdapter(os.Args[0], "test",
		WithCapabilitiesTimeout(5*time.Second),
		WithShutdownTimeout(2*time.Second),
	)
	adapter.binaryPath = os.Args[0]
	adapter.env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_MODE=success",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	caps := adapter.Capabilities()
	if got, want := len(caps), 3; got != want {
		t.Fatalf("len(Capabilities()): got %v, want %v", got, want)
	}

	expected := []Capability{CapAfferentCoupling, CapEfferentCoupling, CapInstability}
	for i, cap := range caps {
		if got, want := cap, expected[i]; got != want {
			t.Errorf("Capabilities()[%d]: got %v, want %v", i, got, want)
		}
	}
}
