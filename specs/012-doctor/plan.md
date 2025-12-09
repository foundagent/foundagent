````markdown
# Implementation Plan: Doctor

**Branch**: `012-doctor` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-doctor/spec.md`

## Summary

Implement `fa doctor` command that runs diagnostic checks on the workspace: environment (Git installed), structure (required files/directories), repositories (bare clones valid), worktrees (tracked by Git), and state consistency. Provides actionable remediation for failures. Supports `--fix` for auto-repairing fixable issues.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), os/exec (git commands)  
**Storage**: Reads all workspace files; `--fix` may regenerate state.json  
**Testing**: `go test` with `testify/assert`, table-driven tests  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: All checks complete in under 5 seconds for 10 repos  
**Constraints**: Auto-fix must never be destructive (no deleting repos)  
**Scale/Scope**: 15+ individual checks across categories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Pass/fail output, `--json` support |
| II. Git-Native Operations | ✅ PASS | Validates Git state, uses git commands |
| III. Non-Destructive by Default | ✅ PASS | Read-only by default; `--fix` is non-destructive |
| IV. Multi-Repository Awareness | ✅ PASS | Checks all repos and worktrees |
| V. Simplicity and Discoverability | ✅ PASS | `fa doctor` intuitive (common pattern) |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for each check |
| VII. Agent-Friendly Design | ✅ PASS | Complete diagnostic state in `--json` |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied. Remediation messages satisfy Constitution error message requirements.

## Project Structure

### Documentation (this feature)

```text
specs/012-doctor/
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
│   ├── doctor.go            # NEW: doctor command
│   └── doctor_test.go       # NEW: doctor tests
├── doctor/
│   ├── doctor.go            # NEW: check runner, result collection
│   ├── check.go             # NEW: Check interface, CheckResult struct
│   ├── runner.go            # NEW: parallel check execution
│   ├── checks/
│   │   ├── git.go           # NEW: Git installation/version checks
│   │   ├── config.go        # NEW: config file validity
│   │   ├── structure.go     # NEW: directory structure checks
│   │   ├── state.go         # NEW: state.json validity
│   │   ├── repos.go         # NEW: bare clone checks
│   │   ├── worktrees.go     # NEW: worktree integrity checks
│   │   └── consistency.go   # NEW: config-state-filesystem sync
│   └── fixes/
│       ├── state.go         # NEW: regenerate state.json
│       └── vscode.go        # NEW: sync workspace file
└── errors/
    └── codes.go             # REUSE: error codes for check failures
```

**Structure Decision**: Dedicated `internal/doctor/` package with modular check system. Checks organized by category. Fixes separate from checks.

## Complexity Tracking

**Complexity: Check Framework Design**

Creating an extensible check framework:
```go
type Check interface {
    Name() string
    Category() string
    Run(ctx context.Context) CheckResult
    CanFix() bool
    Fix(ctx context.Context) error
}
```

**Justification**: Constitution mandates many specific checks; extensible framework enables adding checks without refactoring.

## Dependencies

**Depends on**:
- spec 001-workspace-init: workspace structure expectations
- spec 003-workspace-config: config validation
- spec 005-worktree-list: worktree discovery
- spec 007-workspace-status: reconciliation logic

**Provides to other specs**:
- `internal/doctor/` package → diagnostic framework for future checks
- Health check pattern → can be extended for CI/pre-commit hooks
````