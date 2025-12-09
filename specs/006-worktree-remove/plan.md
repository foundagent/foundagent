````markdown
# Implementation Plan: Worktree Remove

**Branch**: `006-worktree-remove` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-worktree-remove/spec.md`

## Summary

Implement `fa wt remove <branch>` command that removes worktrees for a branch across ALL repos. Blocks removal if worktrees have uncommitted changes (requires `--force`), prevents removal of worktree you're standing in, and optionally deletes the branch with `--delete-branch`. Updates VS Code workspace file.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), go-git/os/exec (git worktree remove)  
**Storage**: Updates `.foundagent/state.json`, `.code-workspace`  
**Testing**: `go test` with `testify/assert`, table-driven tests, t.TempDir()  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Remove worktrees across 5 repos in under 5 seconds  
**Constraints**: Non-destructive by default; dirty check mandatory  
**Scale/Scope**: Support removing worktrees from 20+ repos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Progress to stderr, `--json` support |
| II. Git-Native Operations | ✅ PASS | Uses `git worktree remove`, `git branch -d` |
| III. Non-Destructive by Default | ✅ PASS | Blocks on dirty, `--force` required |
| IV. Multi-Repository Awareness | ✅ PASS | Removes across all repos |
| V. Simplicity and Discoverability | ✅ PASS | `fa wt remove feature-x` intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for safety checks |
| VII. Agent-Friendly Design | ✅ PASS | `--json` with per-repo status |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Constitution Principle III is core to this command's design.

## Project Structure

### Documentation (this feature)

```text
specs/006-worktree-remove/
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
│   ├── wt_remove.go         # NEW: remove subcommand
│   ├── wt_remove_test.go    # NEW: remove command tests
│   └── worktree.go          # MODIFY: add rm alias
├── git/
│   ├── worktree.go          # MODIFY: add worktree removal
│   ├── branch.go            # MODIFY: add branch deletion, merge check
│   └── status.go            # REUSE: dirty detection from 005
├── workspace/
│   ├── worktree.go          # MODIFY: add CWD-inside detection
│   ├── vscode.go            # MODIFY: remove worktree folders
│   └── state.go             # MODIFY: remove worktree entries
└── errors/
    └── codes.go             # MODIFY: add E1xx worktree removal errors
```

**Structure Decision**: Remove command extends worktree group. Reuses dirty detection from 005. CWD check prevents accidental self-removal.

## Complexity Tracking

**Complexity: CWD-Inside-Worktree Detection**

Must detect if user's current working directory is inside a worktree being removed. This prevents confusing shell state where the directory disappears.

**Justification**: Constitution Principle III (Non-Destructive) extends to preventing user from breaking their shell session.

## Dependencies

**Depends on**:
- spec 004-worktree-create: worktree command group
- spec 005-worktree-list: dirty detection, worktree discovery

**Provides to other specs**:
- `internal/git/worktree.go` (worktree removal) → used by 010-repo-remove
- CWD-inside detection pattern → used by 010-repo-remove
````