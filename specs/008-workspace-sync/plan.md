````markdown
# Implementation Plan: Workspace Sync

**Branch**: `008-workspace-sync` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-workspace-sync/spec.md`

## Summary

Implement `fa sync` command that fetches from all remotes in parallel, with optional `--pull` (fast-forward merge) and `--push` (push local commits). Handles network failures gracefully with partial success reporting. Supports syncing specific branches and `--stash` for dirty worktrees.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), go-git (fetch/push), os/exec (git pull fallback)  
**Storage**: Updates bare clones at `repos/.bare/`, worktrees at `repos/worktrees/`  
**Testing**: `go test` with `testify/assert`, table-driven tests, mocked network  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Fetch 5 repos in under 30 seconds on typical network  
**Constraints**: Auth via system Git; no credential storage; fast-forward only for pull  
**Scale/Scope**: Support workspaces with 20+ repos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Progress to stderr, `--json` for machines |
| II. Git-Native Operations | ✅ PASS | Uses git fetch, git pull, git push |
| III. Non-Destructive by Default | ✅ PASS | Fetch only by default; --pull/--push explicit |
| IV. Multi-Repository Awareness | ✅ PASS | Syncs all repos in parallel |
| V. Simplicity and Discoverability | ✅ PASS | `fa sync` intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests with mocked network |
| VII. Agent-Friendly Design | ✅ PASS | `--json` with per-repo status |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Graceful degradation per constitution network resilience section.

## Project Structure

### Documentation (this feature)

```text
specs/008-workspace-sync/
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
│   ├── sync.go              # NEW: sync command
│   └── sync_test.go         # NEW: sync command tests
├── git/
│   ├── remote.go            # MODIFY: add fetch, ahead/behind detection
│   ├── pull.go              # NEW: fast-forward pull
│   ├── push.go              # NEW: push operations
│   └── stash.go             # NEW: stash/pop operations
├── workspace/
│   ├── parallel.go          # REUSE: from 002
│   └── sync.go              # NEW: sync orchestration, partial failure handling
└── errors/
    └── codes.go             # MODIFY: add E3xx network errors
```

**Structure Decision**: Separate files for pull/push/stash in `git/` package. Sync orchestration in `workspace/sync.go` handles parallel execution and error aggregation.

## Complexity Tracking

**Complexity: Partial Failure Handling**

Sync must continue with remaining repos when one fails. Requires careful error collection and reporting.

**Justification**: Constitution "Network Resilience" section mandates "graceful degradation" — complete partial work and report failures clearly.

## Dependencies

**Depends on**:
- spec 002-repo-add: parallel executor, git remote infrastructure
- spec 005-worktree-list: dirty detection (for --stash decision)

**Provides to other specs**:
- `internal/git/remote.go` (fetch, ahead/behind) → used by 007-status (optional)
- `internal/git/stash.go` (stash operations) → general utility
- E3xx network error codes → used by all network operations
````