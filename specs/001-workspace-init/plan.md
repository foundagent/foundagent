````markdown
# Implementation Plan: Workspace Initialization

**Branch**: `001-workspace-init` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-workspace-init/spec.md`

## Summary

Implement `fa init <name>` command that creates a new Foundagent workspace with the canonical directory structure: `.foundagent.yaml` (config), `.foundagent/state.json` (machine state), `repos/<repo-name>/.bare/` (bare clones), `repos/<repo-name>/worktrees/` (working directories), and a VS Code `.code-workspace` file.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI framework), Viper (config management)  
**Storage**: Creates `.foundagent.yaml` (YAML), `.foundagent/state.json` (JSON), `<name>.code-workspace` (JSON)  
**Testing**: `go test` with `testify/assert`, table-driven tests, `t.TempDir()` for filesystem tests  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Complete initialization in under 2 seconds  
**Constraints**: Must validate workspace name for filesystem compatibility across all platforms  
**Scale/Scope**: Creates single workspace; foundational for all other commands

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Command outputs to stdout/stderr, supports `--json`, meaningful exit codes |
| II. Git-Native Operations | ✅ PASS | Creates structure for git worktrees, no custom VCS |
| III. Non-Destructive by Default | ✅ PASS | Errors if workspace exists, `--force` required to reinit |
| IV. Multi-Repository Awareness | ✅ PASS | Creates `repos/` structure for multiple repos |
| V. Simplicity and Discoverability | ✅ PASS | `fa init <name>` follows intuitive pattern |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for validation, filesystem tests with t.TempDir() |
| VII. Agent-Friendly Design | ✅ PASS | `--json` flag for machine-readable output |
| VIII. Agent-Agnostic Design | ✅ PASS | VS Code workspace is IDE infrastructure, not agent-specific |

**Post-Design Re-check**: ✅ All principles satisfied. Foundational command establishes patterns for all subsequent commands.

## Project Structure

### Documentation (this feature)

```text
specs/001-workspace-init/
├── plan.md              # This file
├── spec.md              # Feature specification
├── tasks.md             # Implementation tasks
└── checklists/
    └── requirements.md  # Spec validation checklist
```

### Source Code (repository root)

```text
cmd/
└── foundagent/
    └── main.go              # Entrypoint (existing or created)

internal/
├── cli/
│   ├── root.go              # Root command with global flags
│   ├── init.go              # NEW: init command implementation
│   └── init_test.go         # NEW: init command tests
├── workspace/
│   ├── workspace.go         # NEW: Workspace struct, creation logic
│   ├── config.go            # NEW: .foundagent.yaml generation
│   ├── state.go             # NEW: state.json initialization
│   ├── vscode.go            # NEW: .code-workspace generation
│   ├── validation.go        # NEW: name validation (filesystem-safe)
│   └── workspace_test.go    # NEW: workspace tests
├── errors/
│   ├── codes.go             # NEW: Error code constants (E0xx, E1xx, etc.)
│   └── error.go             # NEW: Structured error type
└── output/
    └── json.go              # NEW: JSON output formatting (shared)
```

**Structure Decision**: Standard Go project layout per constitution. Error infrastructure created here for use by all subsequent specs.

## Complexity Tracking

No constitution violations to justify. Design establishes foundational patterns:
- Error code infrastructure (E0xx for config errors)
- Workspace detection and validation
- JSON output formatting
- VS Code workspace file generation

## Dependencies

This is the foundational spec with no dependencies on other specs.

**Provides to other specs**:
- `internal/workspace/` package (workspace detection, config, state)
- `internal/errors/` package (error codes E001, E002, etc.)
- `internal/output/` package (JSON formatting)
- CLI patterns (flag handling, exit codes)
````