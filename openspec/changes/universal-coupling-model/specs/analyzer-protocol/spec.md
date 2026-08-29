## ADDED Requirements

### Requirement: JSON-RPC 2.0 transport

The external analyzer protocol SHALL use JSON-RPC 2.0 as the message format. Communication SHALL occur over stdin/stdout of the analyzer subprocess. The host process SHALL spawn the analyzer as a child process and exchange messages via its stdin and stdout streams. Each JSON-RPC message SHALL be framed as a single line of JSON terminated by a newline character (`\n`). The host SHALL read one line at a time from stdout and parse each line as a complete JSON-RPC message.

#### Scenario: Valid JSON-RPC request
- **GIVEN** an analyzer subprocess is running
- **WHEN** the host sends a JSON-RPC 2.0 request with method "analyze" to the analyzer's stdin
- **THEN** the analyzer SHALL respond with a JSON-RPC 2.0 response on its stdout, terminated by a newline

#### Scenario: Invalid JSON-RPC request
- **GIVEN** an analyzer subprocess is running
- **WHEN** the host sends malformed JSON to the analyzer's stdin
- **THEN** the analyzer SHALL respond with a JSON-RPC 2.0 error response with code -32700 (Parse error)

#### Scenario: Unknown method
- **GIVEN** an analyzer subprocess is running
- **WHEN** the host sends a JSON-RPC 2.0 request with an unrecognized method
- **THEN** the analyzer SHALL respond with a JSON-RPC 2.0 error response with code -32601 (Method not found)

### Requirement: Analyze method

The protocol SHALL define an `analyze` method that accepts a project path and returns a `ModuleGraph` in the JSON interchange format. The request params SHALL include `projectPath` (string). The result SHALL conform to the metrics JSON schema.

#### Scenario: Successful analysis
- **GIVEN** an analyzer subprocess is running and the project path is valid
- **WHEN** the host sends `{"jsonrpc":"2.0","method":"analyze","params":{"projectPath":"/path/to/project"},"id":1}`
- **THEN** the analyzer SHALL respond with `{"jsonrpc":"2.0","result":<ModuleGraph JSON>,"id":1}`

#### Scenario: Analysis error
- **GIVEN** an analyzer subprocess is running
- **WHEN** the analyzer cannot analyze the given project path
- **THEN** the analyzer SHALL respond with a JSON-RPC 2.0 error response with a descriptive message and error code -32000 (application error)

### Requirement: Capabilities method

The protocol SHALL define a `capabilities` method that returns the analyzer's supported metrics, language identifier, and protocol version. The method takes no parameters. The response MUST include a `protocolVersion` field (string, initial value `"1.0"`).

#### Scenario: Capabilities response
- **GIVEN** an analyzer subprocess is running
- **WHEN** the host sends `{"jsonrpc":"2.0","method":"capabilities","params":{},"id":1}`
- **THEN** the analyzer SHALL respond with `{"jsonrpc":"2.0","result":{"language":"python","protocolVersion":"1.0","metrics":["ca","ce","instability","abstractness","distance","lcom","circular"]},"id":1}`

### Requirement: Lifecycle management

The host process SHALL manage the lifecycle of external analyzer subprocesses. The host MUST send a `shutdown` notification before terminating the subprocess. The analyzer SHOULD perform cleanup upon receiving the shutdown notification and exit cleanly.

#### Scenario: Clean shutdown
- **GIVEN** an analyzer subprocess is running
- **WHEN** the host sends `{"jsonrpc":"2.0","method":"shutdown"}` (notification, no id)
- **THEN** the analyzer SHALL perform cleanup and exit with status code 0

#### Scenario: Analyzer crash handling
- **GIVEN** an analyzer subprocess is running
- **WHEN** the analyzer subprocess exits unexpectedly during analysis
- **THEN** the host SHALL return an error to the caller with the subprocess exit code and any stderr output

#### Scenario: Analyzer timeout
- **GIVEN** an analyzer subprocess is running
- **WHEN** the analyzer does not respond within the configured timeout (defaults: `analyze` = 300s, `capabilities` = 10s, `shutdown` = 5s; configurable per-adapter)
- **THEN** the host SHALL terminate the subprocess and return a timeout error to the caller

#### Scenario: Shutdown grace period
- **GIVEN** the host sends a shutdown notification
- **WHEN** the analyzer does not exit within the shutdown timeout (default: 5s)
- **THEN** the host SHALL forcefully terminate the subprocess (SIGKILL)

### Requirement: Stderr for diagnostics

The analyzer process SHALL use stderr for diagnostic output (logs, progress, debug info). The host process MUST NOT interpret stderr as protocol messages. The host MAY capture stderr for error reporting.

#### Scenario: Diagnostic output
- **GIVEN** an analyzer subprocess is running
- **WHEN** the analyzer writes diagnostic messages to stderr during analysis
- **THEN** the host SHALL ignore stderr for protocol purposes and MAY log it

#### Scenario: Error context from stderr
- **GIVEN** an analyzer has written diagnostic info to stderr
- **WHEN** the analyzer crashes
- **THEN** the host SHALL include stderr content in the error returned to the caller

### Requirement: Input validation

The host MUST validate the `projectPath` parameter before sending it to the analyzer subprocess. The host SHALL reject paths containing `..` traversal components. The host SHALL verify that the path resolves to an existing directory.

#### Scenario: Path traversal rejection
- **GIVEN** a projectPath containing `..` traversal components
- **WHEN** the host receives the analysis request
- **THEN** the host SHALL reject the request with a validation error without forwarding it to the analyzer

#### Scenario: Non-existent path rejection
- **GIVEN** a projectPath that does not exist on the filesystem
- **WHEN** the host receives the analysis request
- **THEN** the host SHALL reject the request with a descriptive error without forwarding it to the analyzer

### Requirement: Subprocess security

The host MUST treat external analyzer subprocesses as untrusted code and enforce security boundaries.

#### Scenario: Environment sanitization
- **GIVEN** the host spawns an analyzer subprocess
- **WHEN** the subprocess environment is constructed
- **THEN** the host SHALL NOT pass its full environment to the subprocess; it SHALL construct a minimal environment containing only variables required for the analyzer to function (PATH, HOME, LANG at minimum; no credential-bearing variables unless explicitly allowlisted)

#### Scenario: Response size limit
- **GIVEN** an analyzer subprocess is producing a response
- **WHEN** the response exceeds the configured maximum size (default: 100 MB)
- **THEN** the host SHALL terminate the subprocess and return a size limit error

#### Scenario: Stderr buffer limit
- **GIVEN** an analyzer subprocess is writing to stderr
- **WHEN** stderr output exceeds the configured maximum (default: 1 MB)
- **THEN** the host SHALL truncate the captured output and include a truncation notice if the output is later reported

## MODIFIED Requirements

<!-- None — greenfield project -->

## REMOVED Requirements

<!-- None — greenfield project -->
