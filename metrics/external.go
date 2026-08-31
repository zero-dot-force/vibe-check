package metrics

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Default limits for ExternalAdapter timeouts and buffer sizes.
const (
	defaultAnalyzeTimeout      = 300 * time.Second
	defaultCapabilitiesTimeout = 10 * time.Second
	defaultShutdownTimeout     = 5 * time.Second
	defaultMaxResponseSize     = 100 * 1024 * 1024 // 100 MB
	defaultMaxStderrSize       = 1 * 1024 * 1024   // 1 MB
)

// ExternalAdapter wraps an external analyzer subprocess that communicates via
// JSON-RPC 2.0 over stdin/stdout. Each JSON-RPC message is framed as a single
// line of JSON terminated by a newline character.
//
// The adapter spawns a new subprocess for each Analyze call, sends a
// capabilities request followed by an analyze request, then shuts down the
// subprocess gracefully. If the subprocess does not exit within the shutdown
// timeout, it is killed with SIGKILL.
type ExternalAdapter struct {
	binaryPath string
	language   string
	env        []string // sanitized environment

	// Configurable limits.
	analyzeTimeout      time.Duration
	capabilitiesTimeout time.Duration
	shutdownTimeout     time.Duration
	maxResponseSize     int64 // bytes
	maxStderrSize       int64 // bytes
}

// ExternalAdapterOption configures an ExternalAdapter.
type ExternalAdapterOption func(*ExternalAdapter)

// WithAnalyzeTimeout sets the timeout for the analyze request.
func WithAnalyzeTimeout(d time.Duration) ExternalAdapterOption {
	return func(a *ExternalAdapter) {
		a.analyzeTimeout = d
	}
}

// WithCapabilitiesTimeout sets the timeout for the capabilities request.
func WithCapabilitiesTimeout(d time.Duration) ExternalAdapterOption {
	return func(a *ExternalAdapter) {
		a.capabilitiesTimeout = d
	}
}

// WithShutdownTimeout sets the grace period for subprocess shutdown before SIGKILL.
func WithShutdownTimeout(d time.Duration) ExternalAdapterOption {
	return func(a *ExternalAdapter) {
		a.shutdownTimeout = d
	}
}

// WithMaxResponseSize sets the maximum response size in bytes.
func WithMaxResponseSize(n int64) ExternalAdapterOption {
	return func(a *ExternalAdapter) {
		a.maxResponseSize = n
	}
}

// WithMaxStderrSize sets the maximum stderr capture size in bytes.
func WithMaxStderrSize(n int64) ExternalAdapterOption {
	return func(a *ExternalAdapter) {
		a.maxStderrSize = n
	}
}

// WithEnvironment sets additional environment variables to include beyond the
// default sanitized set (PATH, HOME, LANG). Credential-bearing variables are
// still excluded.
func WithEnvironment(allowlist []string) ExternalAdapterOption {
	return func(a *ExternalAdapter) {
		a.env = SanitizeEnvironment(allowlist)
	}
}

// NewExternalAdapter creates a new ExternalAdapter for the given binary path
// and language identifier. The binary path must refer to an executable that
// implements the vibe-check external analyzer JSON-RPC protocol.
func NewExternalAdapter(binaryPath, language string, opts ...ExternalAdapterOption) *ExternalAdapter {
	a := &ExternalAdapter{
		binaryPath:          binaryPath,
		language:            language,
		env:                 SanitizeEnvironment(nil),
		analyzeTimeout:      defaultAnalyzeTimeout,
		capabilitiesTimeout: defaultCapabilitiesTimeout,
		shutdownTimeout:     defaultShutdownTimeout,
		maxResponseSize:     defaultMaxResponseSize,
		maxStderrSize:       defaultMaxStderrSize,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Language returns the lowercase language identifier for this adapter.
func (a *ExternalAdapter) Language() string {
	return a.language
}

// Capabilities returns the list of metrics this adapter can compute.
// It spawns the subprocess, sends a capabilities request, and returns the
// result. The subprocess is shut down after the call.
//
// A nil or empty return value may indicate either that the adapter supports
// no capabilities or that a communication error occurred (spawn failure,
// protocol error, etc.). Use [ExternalAdapter.Analyze] for detailed error
// diagnostics when capabilities discovery fails.
func (a *ExternalAdapter) Capabilities() []Capability {
	// Capabilities is defined without error return in the Adapter interface.
	// On failure, return an empty slice — callers can use Analyze to get
	// detailed error information.
	ctx, cancel := context.WithTimeout(context.Background(), a.capabilitiesTimeout)
	defer cancel()

	proc, err := a.spawnProcess(ctx)
	if err != nil {
		return nil
	}
	defer proc.shutdown(a.shutdownTimeout)

	id := 1
	req := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		Method:  "capabilities",
		ID:      &id,
	}

	resp, err := proc.call(req, a.maxResponseSize)
	if err != nil {
		return nil
	}

	if resp.Error != nil {
		return nil
	}

	var result CapabilitiesResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil
	}

	caps := make([]Capability, len(result.Metrics))
	for i, m := range result.Metrics {
		caps[i] = Capability(m)
	}
	return caps
}

// Analyze spawns the external analyzer subprocess, sends the capabilities and
// analyze requests, validates the response, and returns the resulting ModuleGraph.
func (a *ExternalAdapter) Analyze(ctx context.Context, projectPath string) (*ModuleGraph, error) {
	if err := ValidateProjectPath(projectPath); err != nil {
		return nil, fmt.Errorf("external analyze: %w", err)
	}

	// Use the analyze timeout as the subprocess deadline, but respect the
	// caller's context if it has an earlier deadline.
	analyzeCtx, analyzeCancel := context.WithTimeout(ctx, a.analyzeTimeout)
	defer analyzeCancel()

	proc, err := a.spawnProcess(analyzeCtx)
	if err != nil {
		return nil, fmt.Errorf("external analyze: spawn subprocess: %w", err)
	}
	defer proc.shutdown(a.shutdownTimeout)

	// Step 1: Send capabilities request.
	capID := 1
	capReq := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		Method:  "capabilities",
		ID:      &capID,
	}

	capResp, err := proc.call(capReq, a.maxResponseSize)
	if err != nil {
		return nil, fmt.Errorf("external analyze: capabilities request: %w", err)
	}
	if capResp.Error != nil {
		return nil, fmt.Errorf("external analyze: capabilities error: %s", capResp.Error.Message)
	}

	// Step 2: Send analyze request.
	analyzeID := 2
	analyzeReq := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		Method:  "analyze",
		Params:  AnalyzeParams{ProjectPath: projectPath},
		ID:      &analyzeID,
	}

	analyzeResp, err := proc.call(analyzeReq, a.maxResponseSize)
	if err != nil {
		return nil, fmt.Errorf("external analyze: analyze request: %w", err)
	}
	if analyzeResp.Error != nil {
		return nil, fmt.Errorf("external analyze: analyzer error: %s", analyzeResp.Error.Message)
	}

	// Step 3: Validate response against schema.
	if err := Validate(analyzeResp.Result); err != nil {
		return nil, fmt.Errorf("external analyze: response validation: %w", err)
	}

	// Step 4: Unmarshal into ModuleGraph using a strict decoder. Rejecting
	// unknown fields enforces the schema's additionalProperties:false at the
	// untrusted subprocess boundary: an analyzer that emits unexpected fields is
	// rejected rather than having them silently ignored.
	var graph ModuleGraph
	dec := json.NewDecoder(bytes.NewReader(analyzeResp.Result))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&graph); err != nil {
		return nil, fmt.Errorf("external analyze: unmarshal response: %w", err)
	}

	return &graph, nil
}

// process represents a running analyzer subprocess with stdin/stdout pipes.
type process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	stderr *limitedBuffer
}

// spawnProcess starts the external analyzer binary as a subprocess with
// sanitized environment and connected stdin/stdout/stderr pipes.
func (a *ExternalAdapter) spawnProcess(ctx context.Context) (*process, error) {
	cmd := exec.CommandContext(ctx, a.binaryPath)
	cmd.Env = a.env

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderrBuf := newLimitedBuffer(a.maxStderrSize)
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start subprocess: %w", err)
	}

	return &process{
		cmd:    cmd,
		stdin:  stdinPipe,
		reader: bufio.NewReaderSize(stdoutPipe, 64*1024),
		stderr: stderrBuf,
	}, nil
}

// call sends a JSON-RPC request and reads the newline-delimited response.
// The response size is limited to maxSize bytes.
func (p *process) call(req JSONRPCRequest, maxSize int64) (*JSONRPCResponse, error) {
	// Marshal and send request with newline framing.
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	reqData = append(reqData, '\n')

	if _, err := p.stdin.Write(reqData); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	// Read one line of response, enforcing size limit.
	// Use a LimitedReader wrapper around the buffered reader to prevent
	// unbounded memory allocation from a malicious subprocess.
	limited := io.LimitReader(p.reader, maxSize+1) // +1 to detect overflow
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 64*1024), int(maxSize)+1)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		// Check if the process exited.
		stderrMsg := p.stderr.String()
		if stderrMsg != "" {
			return nil, fmt.Errorf("read response: subprocess closed stdout (stderr: %s)", stderrMsg)
		}
		return nil, fmt.Errorf("read response: subprocess closed stdout without sending a response")
	}

	line := scanner.Bytes()
	if int64(len(line)) > maxSize {
		return nil, fmt.Errorf("read response: response exceeds maximum size of %d bytes", maxSize)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// shutdown sends a shutdown notification to the subprocess and waits for it to
// exit gracefully. If the subprocess does not exit within the grace period, it
// is killed with SIGKILL.
func (p *process) shutdown(gracePeriod time.Duration) {
	// Send shutdown notification (no ID = notification, no response expected).
	shutdownReq := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		Method:  "shutdown",
	}
	reqData, err := json.Marshal(shutdownReq)
	if err == nil {
		reqData = append(reqData, '\n')
		// Best-effort write — the process may already be dead.
		_, _ = p.stdin.Write(reqData)
	}
	_ = p.stdin.Close()

	// Wait for graceful exit with a timeout.
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited gracefully.
	case <-time.After(gracePeriod):
		// Grace period expired — force kill.
		if p.cmd.Process != nil {
			_ = p.cmd.Process.Kill()
			<-done // Wait for the kill to complete.
		}
	}
}

// limitedBuffer is a bytes.Buffer that stops accepting writes after reaching
// a maximum size. It is used to capture subprocess stderr without unbounded
// memory growth.
type limitedBuffer struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	maxSize  int64
	writeErr bool // set when an internal buffer write fails (e.g., OOM)
}

// newLimitedBuffer creates a limitedBuffer with the given maximum size.
func newLimitedBuffer(maxSize int64) *limitedBuffer {
	return &limitedBuffer{maxSize: maxSize}
}

// Write implements io.Writer. Writes that would exceed the maximum size are
// silently truncated. The full input length is always reported as written to
// avoid breaking the subprocess's stderr pipe.
func (lb *limitedBuffer) Write(p []byte) (int, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	originalLen := len(p)

	remaining := lb.maxSize - int64(lb.buf.Len())
	if remaining <= 0 {
		// Buffer is full — discard silently.
		return originalLen, nil
	}

	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	if _, err := lb.buf.Write(p); err != nil {
		// Return the original length even on error to maintain the contract
		// of never breaking the subprocess stderr pipe. Record the failure
		// for observability in the String() output.
		lb.writeErr = true
		return originalLen, nil
	}
	return originalLen, nil
}

// String returns the captured stderr content. If the buffer was truncated,
// a notice is appended.
func (lb *limitedBuffer) String() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	s := lb.buf.String()
	if int64(lb.buf.Len()) >= lb.maxSize {
		s += "\n[stderr truncated at " + fmt.Sprintf("%d", lb.maxSize) + " bytes]"
	}
	if lb.writeErr {
		s += "\n[stderr capture encountered write error]"
	}
	return s
}

// Stderr returns the captured stderr output from the subprocess.
func (p *process) Stderr() string {
	if p.stderr == nil {
		return ""
	}
	return p.stderr.String()
}

// Compile-time interface satisfaction check.
var _ Adapter = (*ExternalAdapter)(nil)
