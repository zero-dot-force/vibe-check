package metrics

import "encoding/json"

// JSONRPCVersion is the JSON-RPC protocol version used by the external analyzer
// protocol. All requests and responses MUST include this version string.
const JSONRPCVersion = "2.0"

// JSONRPCRequest represents a JSON-RPC 2.0 request message sent to an external
// analyzer subprocess. When ID is nil, the message is a notification (e.g.,
// shutdown) and no response is expected.
type JSONRPCRequest struct {
	// JSONRPC is the protocol version string, always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// Method is the name of the method to invoke (e.g., "analyze", "capabilities", "shutdown").
	Method string `json:"method"`
	// Params contains the method parameters. Omitted when the method takes no parameters.
	Params any `json:"params,omitempty"`
	// ID is the request identifier. Nil for notifications (no response expected).
	ID *int `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response message received from an
// external analyzer subprocess. Exactly one of Result or Error is non-nil.
type JSONRPCResponse struct {
	// JSONRPC is the protocol version string, always "2.0".
	JSONRPC string `json:"jsonrpc"`
	// Result contains the method return value as raw JSON. Nil when Error is set.
	Result json.RawMessage `json:"result,omitempty"`
	// Error contains the error object when the method invocation failed. Nil on success.
	Error *JSONRPCError `json:"error,omitempty"`
	// ID is the request identifier that this response corresponds to.
	ID *int `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object returned when a method
// invocation fails.
type JSONRPCError struct {
	// Code is a machine-readable error code. Standard JSON-RPC codes apply
	// (e.g., -32600 for invalid request, -32601 for method not found).
	Code int `json:"code"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
}

// Error implements the error interface for JSONRPCError, allowing it to be
// used directly as a Go error value.
func (e *JSONRPCError) Error() string {
	return e.Message
}

// AnalyzeParams contains the parameters for the "analyze" JSON-RPC method.
type AnalyzeParams struct {
	// ProjectPath is the absolute filesystem path to the project to analyze.
	ProjectPath string `json:"projectPath"`
}

// CapabilitiesResult contains the response for the "capabilities" JSON-RPC method.
// It describes what metrics the external analyzer can compute.
type CapabilitiesResult struct {
	// Language is the lowercase language identifier (e.g., "python", "typescript").
	Language string `json:"language"`
	// ProtocolVersion is the version of the analyzer protocol supported.
	ProtocolVersion string `json:"protocolVersion"`
	// Metrics lists the metric identifiers this analyzer can compute
	// (e.g., "ca", "ce", "instability").
	Metrics []string `json:"metrics"`
}
