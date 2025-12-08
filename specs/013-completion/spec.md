# Feature Specification: Shell Completion

**Feature Branch**: `013-completion`  
**Created**: 2025-12-08  
**Status**: Draft  
**Input**: User description: "Generate shell completion scripts with fa completion command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Bash Completion Script (Priority: P1)

A developer using Bash wants tab completion for Foundagent commands. They run `fa completion bash` and receive a script they can source in their shell.

**Why this priority**: Bash is the most common shell; this is the primary use case.

**Independent Test**: Run `fa completion bash`, source the output, verify tab completion works for commands and flags.

**Acceptance Scenarios**:

1. **Given** I run `fa completion bash`, **Then** I receive a valid Bash completion script
2. **Given** I source the script, **When** I type `fa <TAB>`, **Then** I see available commands
3. **Given** completion is active, **When** I type `fa wt <TAB>`, **Then** I see worktree subcommands

---

### User Story 2 - Generate Zsh Completion Script (Priority: P1)

A developer using Zsh wants tab completion. They run `fa completion zsh` and receive a script for their shell.

**Why this priority**: Zsh is the default shell on macOS; equally important as Bash.

**Independent Test**: Run `fa completion zsh`, install the script, verify tab completion works.

**Acceptance Scenarios**:

1. **Given** I run `fa completion zsh`, **Then** I receive a valid Zsh completion script
2. **Given** I install the script, **When** I type `fa <TAB>`, **Then** I see available commands with descriptions

---

### User Story 3 - Generate Fish Completion Script (Priority: P2)

A developer using Fish shell wants tab completion. They run `fa completion fish` and receive a script.

**Why this priority**: Fish is popular but less common than Bash/Zsh.

**Independent Test**: Run `fa completion fish`, install the script, verify tab completion works.

**Acceptance Scenarios**:

1. **Given** I run `fa completion fish`, **Then** I receive a valid Fish completion script
2. **Given** I install the script, **When** I type `fa <TAB>`, **Then** I see available commands

---

### User Story 4 - Generate PowerShell Completion Script (Priority: P2)

A developer on Windows using PowerShell wants tab completion. They run `fa completion powershell` and receive a script.

**Why this priority**: PowerShell is essential for Windows users but less common overall.

**Independent Test**: Run `fa completion powershell`, install the script, verify tab completion works.

**Acceptance Scenarios**:

1. **Given** I run `fa completion powershell`, **Then** I receive a valid PowerShell completion script
2. **Given** I install the script, **When** I type `fa <TAB>`, **Then** I see available commands

---

### User Story 5 - Dynamic Completions for Workspace Context (Priority: P2)

A developer types `fa wt switch <TAB>` and sees a list of available worktrees from their current workspace, not just static command names.

**Why this priority**: Dynamic completions make the tool significantly more usable but require more complex implementation.

**Independent Test**: Create worktrees, type `fa wt switch <TAB>`, verify worktree names appear as completions.

**Acceptance Scenarios**:

1. **Given** worktrees exist, **When** I type `fa wt switch <TAB>`, **Then** I see worktree/branch names
2. **Given** repos exist, **When** I type `fa remove <TAB>`, **Then** I see repo names
3. **Given** I'm outside a workspace, **When** I type `fa wt switch <TAB>`, **Then** completion gracefully shows no options

---

### Edge Cases

- **Not in workspace**: Dynamic completions should degrade gracefully (show nothing or generic options)
- **Large number of completions**: Handle workspaces with many repos/worktrees without hanging
- **Special characters in names**: Branch names with special chars should be properly escaped
- **Both aliases work**: Completions work for both `fa` and `foundagent`
- **Slow filesystem**: Dynamic completions should have timeout to avoid blocking shell

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa completion <shell>` command
- **FR-002**: System MUST support `bash` as a shell option
- **FR-003**: System MUST support `zsh` as a shell option
- **FR-004**: System MUST support `fish` as a shell option
- **FR-005**: System MUST support `powershell` as a shell option
- **FR-006**: Running `fa completion` without shell argument MUST show usage help
- **FR-007**: Running with unsupported shell MUST show error with list of supported shells

#### Script Output
- **FR-008**: Output MUST be the completion script content (stdout)
- **FR-009**: Scripts MUST be directly sourceable/usable without modification
- **FR-010**: Scripts MUST include installation instructions as comments
- **FR-011**: Scripts MUST be idempotent (safe to source multiple times)

#### Static Completions
- **FR-012**: Completion MUST include all commands (`init`, `add`, `remove`, `wt`, `status`, `sync`, `config`, `version`, `doctor`, `completion`)
- **FR-013**: Completion MUST include all subcommands (`wt create`, `wt list`, `wt remove`, `wt switch`)
- **FR-014**: Completion MUST include all global flags (`--help`, `--version`, `--json`, `-v`)
- **FR-015**: Completion MUST include command-specific flags (e.g., `--force` for `remove`)
- **FR-016**: Completion MUST include flag descriptions where shell supports it

#### Dynamic Completions
- **FR-017**: `fa wt switch` MUST complete with available branch/worktree names
- **FR-018**: `fa wt remove` MUST complete with available worktree names
- **FR-019**: `fa remove` MUST complete with repo names from config
- **FR-020**: Dynamic completions MUST work when inside a Foundagent workspace
- **FR-021**: Dynamic completions MUST gracefully return empty when outside workspace
- **FR-022**: Dynamic completions MUST have reasonable timeout (under 500ms)

#### Alias Support
- **FR-023**: Completions MUST work for `fa` command
- **FR-024**: Completions MUST work for `foundagent` command
- **FR-025**: Both commands MUST share the same completion logic

#### Installation Help
- **FR-026**: `fa completion bash --help` MUST show Bash installation instructions
- **FR-027**: `fa completion zsh --help` MUST show Zsh installation instructions
- **FR-028**: `fa completion fish --help` MUST show Fish installation instructions
- **FR-029**: `fa completion powershell --help` MUST show PowerShell installation instructions

### Key Entities

- **Completion Script**: Shell-specific script that enables tab completion
- **Static Completion**: Fixed completions for commands, subcommands, and flags
- **Dynamic Completion**: Context-aware completions based on current workspace state

### Assumptions

- Users have standard shell configurations (Bash 4+, Zsh 5+, Fish 3+, PowerShell 5+)
- Completion scripts follow each shell's standard completion conventions
- Dynamic completions read from config file, not by executing Foundagent commands

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can install completions for any supported shell in under 2 minutes following the instructions
- **SC-002**: Tab completion shows all commands and flags accurately
- **SC-003**: Dynamic completions respond in under 500ms for workspaces with up to 20 repos
- **SC-004**: Completion scripts work on default shell configurations without additional setup
- **SC-005**: Both `fa` and `foundagent` commands have working completions
