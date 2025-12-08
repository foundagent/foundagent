# Feature Specification: Version

**Feature Branch**: `011-version`  
**Created**: 2025-12-08  
**Status**: Draft  
**Input**: User description: "Display version information with fa version command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Check Installed Version (Priority: P1)

A developer wants to know which version of Foundagent they have installed. They run `fa version` or `fa --version` and see the version number clearly displayed.

**Why this priority**: This is the core functionality — knowing what version is installed is essential for troubleshooting and ensuring compatibility.

**Independent Test**: Install Foundagent, run `fa version`, verify output shows version number.

**Acceptance Scenarios**:

1. **Given** Foundagent is installed, **When** I run `fa version`, **Then** I see the version number (e.g., "1.0.0")
2. **Given** Foundagent is installed, **When** I run `fa --version`, **Then** I see the same version output
3. **Given** Foundagent is installed, **When** I run `foundagent version`, **Then** I see the same version output

---

### User Story 2 - Get Full Build Information (Priority: P2)

A developer is filing a bug report and needs detailed build information. They run `fa version --full` and see version, commit hash, build date, Go version, and platform.

**Why this priority**: Detailed build info is essential for debugging but not needed for casual version checks.

**Independent Test**: Run `fa version --full`, verify output includes all build metadata.

**Acceptance Scenarios**:

1. **Given** I run `fa version --full`, **Then** I see version, git commit, build date, Go version, and OS/arch
2. **Given** I'm filing a bug report, **When** I copy `fa version --full` output, **Then** it includes all info needed for reproduction

---

### User Story 3 - Machine-Readable Version Output (Priority: P2)

An automation script or AI agent needs to parse version information programmatically. They run `fa version --json` and receive structured JSON output.

**Why this priority**: Agent-friendly design requires JSON output for all commands.

**Independent Test**: Run `fa version --json`, parse output as JSON, verify all fields present.

**Acceptance Scenarios**:

1. **Given** I run `fa version --json`, **Then** output is valid JSON
2. **Given** JSON output, **When** I parse it, **Then** I can extract version, commit, build_date, go_version, os, arch

---

### User Story 4 - Check for Updates (Priority: P3)

A developer wants to know if a newer version is available. They run `fa version --check` and see whether they're up to date or if an update is available.

**Why this priority**: Helpful but not essential — users can check GitHub releases manually.

**Independent Test**: Run `fa version --check`, verify it reports whether update is available.

**Acceptance Scenarios**:

1. **Given** I'm on the latest version, **When** I run `fa version --check`, **Then** I see "You're up to date"
2. **Given** a newer version exists, **When** I run `fa version --check`, **Then** I see "Update available: vX.Y.Z"
3. **Given** network is unavailable, **When** I run `fa version --check`, **Then** I see current version with warning that update check failed

---

### Edge Cases

- **Development build**: If built from source without version tags, show "dev" or commit hash
- **No network for update check**: `--check` gracefully degrades with warning, still shows local version
- **Alias consistency**: `fa` and `foundagent` show identical output
- **Short flag**: `-v` is commonly expected but conflicts with potential verbose flag — use `--version` only

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa version` command
- **FR-002**: System MUST support `fa --version` flag (same output as `fa version`)
- **FR-003**: System MUST support `foundagent version` and `foundagent --version`
- **FR-004**: System MUST support `--full` flag for detailed build information
- **FR-005**: System MUST support `--json` flag for machine-readable output
- **FR-006**: System MUST support `--check` flag to check for updates

#### Basic Output (default)
- **FR-007**: Default output MUST show version number (e.g., "foundagent v1.0.0")
- **FR-008**: Output MUST be a single line for easy parsing
- **FR-009**: Version MUST follow semantic versioning format (MAJOR.MINOR.PATCH)

#### Full Output (--full)
- **FR-010**: Full output MUST include version number
- **FR-011**: Full output MUST include git commit hash (short, 7 characters)
- **FR-012**: Full output MUST include build date (ISO 8601 format)
- **FR-013**: Full output MUST include Go version used to build
- **FR-014**: Full output MUST include OS and architecture (e.g., "darwin/arm64")

#### JSON Output (--json)
- **FR-015**: JSON output MUST include `version` field
- **FR-016**: JSON output MUST include `commit` field
- **FR-017**: JSON output MUST include `build_date` field
- **FR-018**: JSON output MUST include `go_version` field
- **FR-019**: JSON output MUST include `os` field
- **FR-020**: JSON output MUST include `arch` field

#### Update Check (--check)
- **FR-021**: Update check MUST query GitHub releases API
- **FR-022**: Update check MUST compare local version to latest release
- **FR-023**: Update check MUST handle network failures gracefully
- **FR-024**: Update check MUST NOT block; use reasonable timeout (5 seconds)
- **FR-025**: If update available, output MUST include download URL or upgrade instructions

#### Development Builds
- **FR-026**: If version info unavailable (dev build), version MUST show "dev"
- **FR-027**: If commit hash unavailable, commit MUST show "unknown"
- **FR-028**: Development builds MUST still function for all flags

### Key Entities

- **Version Info**: Structured data containing version, commit, build_date, go_version, os, arch
- **Release**: A published version on GitHub with tag and assets

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve version information in under 1 second
- **SC-002**: JSON output is parseable by standard JSON parsers in 100% of cases
- **SC-003**: Update check completes or times out within 5 seconds
- **SC-004**: All build metadata is correctly embedded at compile time
- **SC-005**: Version output is consistent across `fa` and `foundagent` commands
