````markdown
# Implementation Plan: Worktree List

**Branch**: `005-worktree-list` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-worktree-list/spec.md`

## Summary

Implement `fa wt list` command that displays all worktrees across all repos, grouped by branch. Shows worktree paths, current/active worktree indicator, and status (clean/modified/untracked/conflict). Supports `--json` for machine-readable output and optional branch filter.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), go-git/os/exec (git status)  
**Storage**: Reads `.foundagent/state.json`, filesystem at `repos/worktrees/`  
**Testing**: `go test` with `testify/assert`, table-driven tests  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: List 50 worktrees in under 2 seconds; status detection under 5 seconds  
**Constraints**: Status detection must run in parallel for performance  
**Scale/Scope**: Support workspaces with 50+ worktrees

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Human-readable default, `--json` for machines |
| II. Git-Native Operations | ✅ PASS | Uses `git status` for dirty detection |
| III. Non-Destructive by Default | ✅ PASS | Read-only operation |
| IV. Multi-Repository Awareness | ✅ PASS | Lists worktrees across all repos |
| V. Simplicity and Discoverability | ✅ PASS | `fa wt list` / `fa wt ls` intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for grouping, status |
| VII. Agent-Friendly Design | ✅ PASS | Complete state in `--json` output |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Provides complete workspace visibility.

## Project Structure

### Documentation (this feature)

```text
specs/005-worktree-list/
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
│   ├── wt_list.go           # NEW: list subcommand
│   └── wt_list_test.go      # NEW: list command tests
├── git/
│   ├── status.go            # MODIFY: add detailed status (modified/untracked/conflict)
│   └── status_test.go       # NEW: status detection tests
├── workspace/
│   ├── worktree.go          # MODIFY: add discovery, current detection, grouping
│   └── worktree_test.go     # NEW: worktree discovery tests
└── cli/
    └── worktree.go          # MODIFY: add ls alias
```

**Structure Decision**: List command is read-only, reuses worktree discovery and status detection that will be shared with status command (007).

## Complexity Tracking

No constitution violations to justify. Design is straightforward:
- Filesystem discovery for worktrees
- Parallel git status checks
- Current worktree detection via CWD comparison

## Dependencies

**Depends on**:
- spec 001-workspace-init: workspace detection
- spec 004-worktree-create: worktree command group, worktree discovery base

**Provides to other specs**:
- `internal/workspace/worktree.go` (worktree discovery, grouping) → used by 007-status
- `internal/git/status.go` (dirty detection) → used by 006-remove, 007-status, 010-remove
````