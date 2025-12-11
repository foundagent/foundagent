````markdown
# Implementation Plan: Add Repository

**Branch**: `002-repo-add` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-repo-add/spec.md`

## Summary

Implement `fa add <url> [name]` command that clones repositories as bare clones to `repos/<name>/.bare/`, creates a worktree for the default branch at `repos/<name>/worktrees/<branch>/`, updates workspace config and state, and refreshes the VS Code workspace file. Supports adding multiple repos in parallel.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), go-git (bare clone), os/exec (git worktree fallback)  
**Storage**: Updates `.foundagent.yaml`, `.foundagent/state.json`, `.code-workspace`  
**Testing**: `go test` with `testify/assert`, table-driven tests, mock git operations  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Clone typical repo (<100MB) in under 30 seconds; parallel cloning faster than sequential  
**Constraints**: Auth handled by system Git (SSH agent, credential helper); no credential storage  
**Scale/Scope**: Support adding 10+ repos in single command

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Progress output to stderr, results to stdout, `--json` support |
| II. Git-Native Operations | ✅ PASS | Uses `git clone --bare` and `git worktree add` |
| III. Non-Destructive by Default | ✅ PASS | Skips existing repos, `--force` to re-clone |
| IV. Multi-Repository Awareness | ✅ PASS | Core multi-repo functionality |
| V. Simplicity and Discoverability | ✅ PASS | `fa add <url>` is intuitive |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for URL parsing, mocked git ops |
| VII. Agent-Friendly Design | ✅ PASS | `--json` with per-repo status |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Uses native Git for all operations.

## Project Structure

### Documentation (this feature)

```text
specs/002-repo-add/
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
│   ├── add.go               # NEW: add command implementation
│   └── add_test.go          # NEW: add command tests
├── git/
│   ├── clone.go             # NEW: bare clone operations
│   ├── worktree.go          # NEW: git worktree add
│   ├── url.go               # NEW: URL parsing, name inference
│   ├── remote.go            # NEW: default branch detection
│   └── git_test.go          # NEW: git operation tests
├── workspace/
│   ├── repository.go        # NEW: Repository struct, registration
│   ├── discover.go          # NEW: workspace detection (find .foundagent.yaml)
│   ├── parallel.go          # NEW: parallel operation executor
│   ├── config.go            # MODIFY: add repo to config
│   ├── state.go             # MODIFY: add repo to state
│   └── vscode.go            # MODIFY: add worktree folders
└── errors/
    └── codes.go             # MODIFY: add E1xx git operation errors
```

**Structure Decision**: Creates `internal/git/` package for all Git operations. Parallel executor in `workspace/` for reuse by other multi-repo commands.

## Complexity Tracking

No constitution violations to justify. Design follows established patterns:
- Git operations via go-git + os/exec fallback
- Parallel execution with error collection
- Config updates preserve comments (requires yaml.v3 node API)

## Dependencies

**Depends on**:
- spec 001-workspace-init: workspace detection, config/state structures, error codes

**Provides to other specs**:
- `internal/git/` package (clone, worktree, URL parsing)
- `internal/workspace/parallel.go` (parallel operation executor)
- `internal/workspace/repository.go` (Repository struct)
- `internal/workspace/discover.go` (workspace detection)
````