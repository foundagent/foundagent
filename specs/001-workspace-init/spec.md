# Feature Specification: Workspace Initialization

**Feature Branch**: `001-workspace-init`  
**Created**: 2025-12-05  
**Status**: Draft  
**Input**: User description: "Initialize a new Foundagent workspace with fa init command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create New Workspace (Priority: P1)

A developer wants to start a new multi-repo project. They navigate to their projects directory (e.g., `~/projects/`) and run `fa init my-app` to create a new Foundagent workspace. The tool creates a self-contained project folder with the necessary structure to begin adding repositories.

**Why this priority**: This is the foundational action — without workspace initialization, no other Foundagent features can be used. It's the entry point to the entire tool.

**Independent Test**: Run `fa init test-project` in an empty directory, verify the folder structure is created correctly, and open the `.code-workspace` file in VS Code to confirm it loads.

**Acceptance Scenarios**:

1. **Given** I am in a directory where I want to create a project, **When** I run `fa init my-app`, **Then** a new `my-app/` directory is created containing the workspace structure
2. **Given** I run `fa init my-app`, **When** the command completes successfully, **Then** I see a confirmation message with the path to the created workspace
3. **Given** I run `fa init my-app`, **When** I examine the created folder, **Then** I find a `my-app.code-workspace` file that can be opened in VS Code

---

### User Story 2 - Initialize with JSON Output (Priority: P2)

A developer using an AI coding agent or automation script needs machine-readable output from the init command. They run `fa init my-app --json` to get structured output for programmatic consumption.

**Why this priority**: Supports the agent-friendly design principle. Essential for integration with AI tools and scripting, but human-readable output is the primary use case.

**Independent Test**: Run `fa init test-project --json`, parse the JSON output, verify it contains the expected fields (path, name, status).

**Acceptance Scenarios**:

1. **Given** I run `fa init my-app --json`, **When** the command completes, **Then** the output is valid JSON containing workspace path, name, and creation status
2. **Given** I run `fa init my-app --json`, **When** an error occurs, **Then** the output is valid JSON containing error code, message, and remediation hint

---

### User Story 3 - Force Reinitialize Existing Workspace (Priority: P3)

A developer has a corrupted or misconfigured workspace and wants to reinitialize it. They run `fa init my-app --force` to recreate the workspace structure while preserving existing repositories.

**Why this priority**: Recovery scenario — less common than initial creation, but important for error recovery. Follows the non-destructive principle with explicit `--force` override.

**Independent Test**: Create a workspace, modify the config file to be invalid, run `fa init my-app --force`, verify the config is restored while any existing repos folder is preserved.

**Acceptance Scenarios**:

1. **Given** a directory `my-app/` exists with a `.foundagent/` folder, **When** I run `fa init my-app`, **Then** I receive an error suggesting `--force` if I want to reinitialize
2. **Given** a directory `my-app/` exists with a `.foundagent/` folder, **When** I run `fa init my-app --force`, **Then** the workspace configuration is regenerated
3. **Given** I run `fa init my-app --force` on an existing workspace, **When** there is a `repos/` directory with content, **Then** the `repos/` directory is preserved

---

### Edge Cases

- **Empty name**: User runs `fa init` without a name — show usage error with example
- **Invalid characters**: Name contains characters invalid for the OS filesystem — show error with list of invalid characters
- **Path too long**: Resulting path exceeds OS limits — show error with max path length
- **Permission denied**: User lacks write permission in current directory — show error with clear message
- **Disk full**: No space available — show error from underlying OS, wrapped with context
- **Directory exists without `.foundagent/`**: Proceed normally (directory may be empty or have unrelated files)
- **Name is `.` or `..`**: Reject as invalid workspace name
- **Name with leading/trailing spaces**: Trim spaces and proceed, or reject if empty after trim

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST create a new directory with the specified name in the current working directory
- **FR-002**: System MUST create a `.foundagent/` subdirectory for machine-managed state
- **FR-003**: System MUST create a `.foundagent.yaml` file at workspace root with default configuration
- **FR-004**: System MUST create a `.foundagent/state.json` file initialized as empty object `{}`
- **FR-005**: System MUST create a `<name>.code-workspace` file in the workspace root
- **FR-006**: The `.code-workspace` file MUST include a `folders` array (initially empty or with workspace root)
- **FR-007**: System MUST reject initialization if `.foundagent/` already exists in target directory (unless `--force`)
- **FR-008**: System MUST support `--force` flag to reinitialize an existing workspace
- **FR-009**: System MUST preserve the `repos/` directory contents when using `--force`
- **FR-010**: System MUST support `--json` flag for machine-readable output
- **FR-011**: System MUST validate that the workspace name is valid for the host operating system's filesystem
- **FR-012**: System MUST display a success message with the absolute path to the created workspace
- **FR-013**: System MUST exit with code 0 on success, non-zero on failure
- **FR-014**: System MUST create a `repos/` directory at workspace root for repository storage
- **FR-015**: System MUST create `repos/.bare/` subdirectory for bare clones
- **FR-016**: System MUST create `repos/worktrees/` subdirectory for working directories

### Key Entities

- **Workspace**: A directory containing `.foundagent.yaml`, `.foundagent/`, `repos/`, and a `.code-workspace` file. Represents a self-contained multi-repo development environment. Key attributes: name, path, configuration, state.
- **Workspace Configuration**: A YAML file (`.foundagent.yaml`) at workspace root storing user-editable settings. Initially contains workspace name and empty repos list.
- **Workspace State**: A JSON file (`.foundagent/state.json`) storing machine-managed runtime state like clone status and worktree tracking. Initially empty.
- **VS Code Workspace File**: A `.code-workspace` JSON file that VS Code uses to open multi-root workspaces. Updated automatically to include worktree folders.
- **Repos Directory**: Top-level `repos/` containing `.bare/` (hidden bare clones) and `worktrees/` (visible working directories organized by repo then branch).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a new workspace in under 2 seconds on typical hardware
- **SC-002**: Created workspace opens successfully in VS Code when double-clicking the `.code-workspace` file
- **SC-003**: 100% of created workspaces pass validation (all required files present, valid syntax)
- **SC-004**: Error messages include actionable remediation steps in 100% of failure cases
- **SC-005**: JSON output mode parses successfully with standard JSON parsers
