````markdown
# Implementation Plan: Worktree Switch

**Branch**: `009-worktree-switch` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-worktree-switch/spec.md`

## Summary

Implement `fa wt switch <branch>` command that helps developers switch to a different branch's worktrees. Updates the VS Code workspace file to show the target branch's worktrees. Optionally opens VS Code with the updated workspace. Warns about uncommitted changes in current worktrees.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), os/exec (code command)  
**Storage**: Updates `.code-workspace`  
**Testing**: `go test` with `testify/assert`, table-driven tests  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Switch operation under 2 seconds  
**Constraints**: VS Code `code` command must be in PATH for --open  
**Scale/Scope**: Support workspaces with 20+ repos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Output to stdout/stderr, `--json` support |
| II. Git-Native Operations | ✅ PASS | Reads git worktree state |
| III. Non-Destructive by Default | ✅ PASS | Warns on dirty worktrees, no data loss |
| IV. Multi-Repository Awareness | ✅ PASS | Switches all repos to same branch |
| V. Simplicity and Discoverability | ✅ PASS | `fa wt switch feature-x` intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for switch scenarios |
| VII. Agent-Friendly Design | ✅ PASS | `--json` with switch result |
| VIII. Agent-Agnostic Design | ✅ PASS | VS Code is IDE infrastructure, not agent-specific |

**Post-Design Re-check**: ✅ All principles satisfied. VS Code workspace update is IDE infrastructure per constitution.

## Project Structure

### Documentation (this feature)

```text
specs/009-worktree-switch/
├── plan.md              # This file
├── spec.md              # Feature specification
├── tasks.md             # Implementation tasks
└── checklists/
    └── requirements.md  # Spec validation checklist
```

### Source Code (repository root)

```text
internal/
├── cli/
│   ├── wt_switch.go         # NEW: switch subcommand
│   └── wt_switch_test.go    # NEW: switch command tests
├── workspace/
│   ├── vscode.go            # MODIFY: replace workspace folders for switch
│   └── worktree.go          # REUSE: worktree discovery from 005
├── git/
│   └── status.go            # REUSE: dirty detection from 005
└── ide/
    └── vscode.go            # NEW: VS Code command invocation (code --wait)
```

**Structure Decision**: New `internal/ide/` package for VS Code command invocation. Switch primarily updates the workspace file and optionally opens VS Code.

## Complexity Tracking

**Complexity: Cross-Platform VS Code Detection**

Finding the `code` command differs across platforms:
- macOS: `/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code` or PATH
- Linux: PATH
- Windows: `code.cmd` in PATH or typical install locations

**Justification**: Constitution cross-platform compatibility requirements. Graceful fallback if `code` not found.

## Dependencies

**Depends on**:
- spec 004-worktree-create: worktree command group
- spec 005-worktree-list: worktree discovery, dirty detection

**Provides to other specs**:
- `internal/ide/vscode.go` (VS Code invocation) → general utility
- Workspace file replacement pattern → distinct from add/remove patterns
````