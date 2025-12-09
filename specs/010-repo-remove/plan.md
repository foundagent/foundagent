````markdown
# Implementation Plan: Repo Remove

**Branch**: `010-repo-remove` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-repo-remove/spec.md`

## Summary

Implement `fa remove <repo>` command that completely removes a repository from the workspace: deletes bare clone, removes all worktrees, updates config and state, and updates VS Code workspace file. Blocks on dirty worktrees (requires `--force`). Supports `--config-only` to unregister without deleting files.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), os (file deletion)  
**Storage**: Updates `.foundagent.yaml`, `.foundagent/state.json`, `.code-workspace`  
**Testing**: `go test` with `testify/assert`, table-driven tests, t.TempDir()  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Remove repo with 5 worktrees in under 5 seconds  
**Constraints**: Removal is permanent; dirty check mandatory  
**Scale/Scope**: Support removing multiple repos in single command

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Confirmation to stderr, `--json` support |
| II. Git-Native Operations | ✅ PASS | Uses git worktree remove |
| III. Non-Destructive by Default | ✅ PASS | Blocks on dirty, `--force` required |
| IV. Multi-Repository Awareness | ✅ PASS | Manages multi-repo workspace membership |
| V. Simplicity and Discoverability | ✅ PASS | `fa remove api` intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for safety checks |
| VII. Agent-Friendly Design | ✅ PASS | `--json` with removal details |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Constitution Principle III is core to this command.

## Project Structure

### Documentation (this feature)

```text
specs/010-repo-remove/
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
│   ├── remove.go            # NEW: remove command
│   ├── remove_test.go       # NEW: remove command tests
│   └── root.go              # MODIFY: add rm alias
├── workspace/
│   ├── repository.go        # MODIFY: add repo removal
│   ├── config.go            # REUSE: from 003
│   ├── state.go             # MODIFY: remove repo entries
│   └── vscode.go            # MODIFY: remove all repo's worktree folders
├── git/
│   ├── clone.go             # MODIFY: add bare clone deletion
│   ├── worktree.go          # REUSE: worktree removal from 006
│   └── status.go            # REUSE: dirty detection from 005
└── errors/
    └── codes.go             # MODIFY: add E2xx workspace errors
```

**Structure Decision**: Remove command reuses worktree removal from 006. Handles multiple worktrees per repo (all branches). CWD check reused from 006.

## Complexity Tracking

**Complexity: Multi-Worktree Removal**

A repo may have worktrees across multiple branches. Remove must:
1. Find all worktrees for the repo
2. Check ALL for dirty status
3. Remove all in order (git worktree remove)
4. Delete bare clone last

**Justification**: Constitution Principle III requires checking all potential data loss points before any destructive action.

## Dependencies

**Depends on**:
- spec 003-workspace-config: config modification
- spec 005-worktree-list: dirty detection
- spec 006-worktree-remove: worktree removal, CWD detection

**Provides to other specs**:
- Complete repo removal pattern → used by cleanup/reset operations
````