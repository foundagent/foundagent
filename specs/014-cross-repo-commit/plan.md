# Implementation Plan: Cross-Repo Commit

**Branch**: `014-cross-repo-commit` | **Date**: 2026-01-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/014-cross-repo-commit/spec.md`

## Summary

Implement synchronized cross-repo commit and push functionality via two new CLI commands: `fa commit` and `fa push`. These commands enable developers to create coordinated commits with identical messages across all repos in the current branch worktrees, and push unpushed commits across repos in parallel. Follows existing patterns from `fa sync` for parallel execution and result aggregation.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: cobra (CLI), testify (testing), os/exec (git commands)  
**Storage**: N/A (operates on git repos via shell commands)  
**Testing**: Go testing + testify with table-driven tests, t.TempDir() for isolation  
**Target Platform**: macOS, Linux, Windows (cross-platform via filepath)  
**Project Type**: Single CLI application (existing structure)  
**Performance Goals**: Commit 10 repos < 10s, Push 10 repos < 30s (from SC-001, SC-002)  
**Constraints**: Parallel operations, graceful partial failure handling  
**Scale/Scope**: Workspaces with up to ~50 repos (consistent with existing limits)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Commands accept args, support --json, exit codes |
| II. Git-Native Operations | ✅ PASS | Uses `git commit`, `git push` via os/exec |
| III. Non-Destructive by Default | ✅ PASS | No --force required for basic operations; --dry-run available |
| IV. Multi-Repository Awareness | ✅ PASS | Core feature: operates across all repos |
| V. Simplicity and Discoverability | ✅ PASS | Follows existing command patterns |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests planned |
| VII. Agent-Friendly Design | ✅ PASS | --json flag for structured output |
| VIII. Agent-Agnostic Design | ✅ PASS | Pure CLI, no agent-specific features |

**Result**: All gates PASS. No violations requiring justification.

## Project Structure

### Documentation (this feature)

```text
specs/014-cross-repo-commit/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI interface spec)
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── cli/
│   ├── commit.go         # NEW: fa commit command
│   ├── commit_test.go    # NEW: commit command tests
│   ├── push.go           # NEW: fa push command  
│   └── push_test.go      # NEW: push command tests
├── git/
│   ├── commit.go         # NEW: git commit operations
│   ├── commit_test.go    # NEW: commit operation tests
│   ├── push.go           # EXTEND: may need additional push helpers
│   └── push_test.go      # EXTEND: additional tests
└── workspace/
    ├── commit.go         # NEW: cross-repo commit orchestration
    ├── commit_test.go    # NEW: orchestration tests
    └── sync.go           # REFERENCE: existing parallel patterns
```

**Structure Decision**: Extends existing single-project Go CLI structure. New files follow established patterns in `internal/cli/`, `internal/git/`, and `internal/workspace/`. Uses existing `workspace.ExecuteParallel()` pattern for parallel operations.

## Complexity Tracking

> No constitution violations requiring justification. Design follows established patterns.
