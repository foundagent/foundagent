````markdown
# Implementation Plan: Worktree Create

**Branch**: `004-worktree-create` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-worktree-create/spec.md`

## Summary

Implement `fa wt create <branch>` command that creates worktrees across ALL repos in the workspace atomically. Supports `--from <source-branch>` to branch from non-default branches, pre-validates all repos before making changes (all-or-nothing), and updates the VS Code workspace file.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), go-git/os/exec (git worktree operations)  
**Storage**: Updates `.foundagent/state.json`, `.code-workspace`  
**Testing**: `go test` with `testify/assert`, table-driven tests, t.TempDir() for filesystem  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Create worktrees across 5 repos in under 10 seconds  
**Constraints**: Atomic pre-validation; no partial states on validation failure  
**Scale/Scope**: Support workspaces with 20+ repos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Progress to stderr, `--json` support, meaningful exit codes |
| II. Git-Native Operations | ✅ PASS | Uses `git worktree add`, `git branch` |
| III. Non-Destructive by Default | ✅ PASS | Errors on existing worktree, `--force` to recreate |
| IV. Multi-Repository Awareness | ✅ PASS | Core multi-repo worktree creation |
| V. Simplicity and Discoverability | ✅ PASS | `fa wt create feature-x` is intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for validation, multi-repo scenarios |
| VII. Agent-Friendly Design | ✅ PASS | `--json` with per-repo status |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Atomic validation prevents messy partial states.

## Project Structure

### Documentation (this feature)

```text
specs/004-worktree-create/
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
│   ├── worktree.go          # NEW: worktree command group, wt alias
│   ├── wt_create.go         # NEW: create subcommand
│   └── wt_create_test.go    # NEW: create command tests
├── git/
│   ├── branch.go            # NEW: branch operations (create, exists, default)
│   ├── validation.go        # NEW: branch name validation
│   ├── worktree.go          # MODIFY: add worktree creation
│   └── status.go            # NEW: uncommitted changes detection
├── workspace/
│   ├── worktree.go          # NEW: worktree existence check, discovery
│   ├── parallel.go          # REUSE: from 002-repo-add
│   ├── vscode.go            # MODIFY: add new worktree folders
│   └── state.go             # MODIFY: track worktree metadata
└── errors/
    └── codes.go             # MODIFY: add E1xx worktree errors
```

**Structure Decision**: Worktree command group (`worktree`/`wt`) created here, subcommands in separate files. Reuses parallel executor from 002.

## Complexity Tracking

**Complexity: Atomic Pre-Validation**

Pre-validating across all repos before any changes adds complexity but is essential for consistency. Implementation checks:
1. Source branch exists in ALL repos
2. Target branch doesn't already exist
3. No existing worktrees for target branch

**Justification**: Constitution Principle IV (Multi-Repository Awareness) implies repos should stay in sync. Partial states create user confusion and cleanup burden.

## Dependencies

**Depends on**:
- spec 001-workspace-init: workspace detection, state tracking
- spec 002-repo-add: parallel executor, git worktree infrastructure
- spec 003-workspace-config: config reading for repo list

**Provides to other specs**:
- `internal/cli/worktree.go` (worktree command group for list/remove/switch)
- `internal/git/branch.go` (branch operations)
- `internal/git/validation.go` (branch name validation)
- `internal/git/status.go` (dirty detection for remove/switch)
- `internal/workspace/worktree.go` (worktree discovery)
````