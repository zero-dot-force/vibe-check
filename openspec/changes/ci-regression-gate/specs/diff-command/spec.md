## MODIFIED Requirements

### Requirement: Input Validation and Exit Codes

The command SHALL validate both input documents before computing anything and MUST exit with a non-zero code when an input cannot be read or fails schema validation, reporting success through the payload verdict rather than through the exit code by default. When the `--gate` flag is set, the command SHALL additionally encode a blocking `REQUEST_CHANGES` verdict as exit code `1`. No verdict payload MUST be emitted when exiting with code 2.

Each input path is read and validated against the `ModuleGraph` schema before any delta is computed. When either input is missing or unreadable, the command writes a diagnostic to standard error and exits with code 2. When either input is readable but fails schema validation, the command writes a diagnostic to standard error and exits with code 2. Both missing/unreadable and readable-but-invalid cases produce exit code 2 and no payload. When both inputs are valid and `--gate` is NOT set, the command exits 0 regardless of the verdict; the verdict (including REQUEST_CHANGES) is conveyed in the output payload, not encoded in the exit code, so that a caller can distinguish a tool failure from a legitimate REQUEST_CHANGES verdict. When both inputs are valid and `--gate` IS set, the command exits 1 when the verdict is REQUEST_CHANGES (policy failure) and exits 0 when the verdict is APPROVE or COMMENT; the verdict payload is still written in every case. The tighten-only override rejection (a looser-than-default threshold flag) continues to exit 2 with no payload, unchanged by `--gate`.

#### Scenario: Missing or unreadable input exits non-zero with no payload
- **WHEN** either input file path does not exist or cannot be read
- **THEN** the command writes a diagnostic to standard error, exits with code 2, and emits no verdict payload

#### Scenario: Readable but schema-invalid input exits non-zero with no payload
- **WHEN** either input file is readable but contains JSON that fails `ModuleGraph` schema validation
- **THEN** the command writes a diagnostic to standard error, exits with code 2, and emits no verdict payload

#### Scenario: Valid inputs exit zero even for REQUEST_CHANGES without --gate
- **WHEN** both inputs are valid, `--gate` is NOT set, and the computed verdict is REQUEST_CHANGES
- **THEN** the command exits 0 and conveys REQUEST_CHANGES in the output payload

#### Scenario: Gate mode exits one on REQUEST_CHANGES
- **WHEN** `--gate` is set, both inputs are valid, and the computed verdict is REQUEST_CHANGES
- **THEN** the command exits 1 and still writes the verdict payload to standard output

#### Scenario: Gate mode exits zero on APPROVE and COMMENT
- **WHEN** `--gate` is set, both inputs are valid, and the computed verdict is APPROVE or COMMENT
- **THEN** the command exits 0 and writes the verdict payload to standard output
