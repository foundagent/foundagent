````markdown
# Implementation Plan: Workspace Status

**Branch**: `007-workspace-status` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-workspace-status/spec.md`

## Summary

Implement `fa status` command that provides a comprehensive workspace overview: repo list with clone status, worktrees grouped by branch with dirty indicators, config-state sync check, and current worktree highlighting. Supports `--json` for AI agent consumption and `-v` for detailed file-level status.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), go-git/os/exec (git status)  
**Storage**: Reads `.foundagent.yaml`, `.foundagent/state.json`, filesystem  
**Testing**: `go test` with `testify/assert`, table-driven tests  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Status for 50 worktrees in under 3 seconds  
**Constraints**: Local-only (no network calls); parallel status detection  
**Scale/Scope**: Support workspaces with 50+ worktrees

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Human-readable default, `--json` for machines |
| II. Git-Native Operations | ✅ PASS | Uses git status, reads Git state |
| III. Non-Destructive by Default | ✅ PASS | Read-only operation |
| IV. Multi-Repository Awareness | ✅ PASS | Aggregates state across all repos |
| V. Simplicity and Discoverability | ✅ PASS | `fa status` / `fa st` intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for status aggregation |
| VII. Agent-Friendly Design | ✅ PASS | Complete workspace state in `--json` |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Essential for agent context recovery (Principle VII).

## Project Structure

### Documentation (this feature)

```text
specs/007-workspace-status/
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
│   ├── status.go            # NEW: status command
│   ├── status_test.go       # NEW: status command tests
│   └── root.go              # MODIFY: add st alias
├── workspace/
│   ├── repository.go        # MODIFY: add clone status check
│   ├── worktree.go          # REUSE: discovery from 005
│   ├── reconcile.go         # REUSE: config-state diff from 003
│   └── status.go            # NEW: workspace status aggregation
├── git/
│   └── status.go            # REUSE: dirty detection from 005
└── output/
    └── status.go            # NEW: status output formatting
```

**Structure Decision**: Status command reuses most infrastructure from 003 (reconcile) and 005 (worktree discovery, status). New aggregation logic in `workspace/status.go`.

## Complexity Tracking

No constitution violations to justify. Design leverages existing infrastructure:
- Worktree discovery from 005
- Config-state reconciliation from 003
- Parallel status detection from 005

## Dependencies

**Depends on**:
- spec 003-workspace-config: config-state reconciliation
- spec 005-worktree-list: worktree discovery, status detection

**Provides to other specs**:
- `internal/workspace/status.go` (workspace aggregation) → used by 012-doctor
- Complete JSON schema for workspace state → used by AI agents
````