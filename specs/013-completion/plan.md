# Implementation Plan: Shell Completion

**Branch**: `013-completion` | **Date**: 2025-12-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/013-completion/spec.md`

## Summary

Implement `fa completion <shell>` command that generates shell completion scripts for Bash, Zsh, Fish, and PowerShell. Uses Cobra's built-in completion generation with custom `ValidArgsFunction` handlers for dynamic workspace-aware completions (worktree names, repo names).

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI framework with native completion support)  
**Storage**: N/A (reads existing `.foundagent.yaml` and `.foundagent/state.json`)  
**Testing**: `go test` with `testify/assert`, table-driven tests  
**Target Platform**: macOS, Linux, Windows (all supported shells)  
**Project Type**: Single Go CLI application  
**Performance Goals**: Dynamic completions respond in <500ms  
**Constraints**: Local file reads only (no network calls during completion)  
**Scale/Scope**: Support workspaces with up to 20 repos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Command outputs to stdout, supports `--help` |
| II. Git-Native Operations | ✅ PASS | Reads Git worktree state, no custom state |
| III. Non-Destructive by Default | ✅ PASS | Read-only operation |
| IV. Multi-Repository Awareness | ✅ PASS | Dynamic completions include repo names |
| V. Simplicity and Discoverability | ✅ PASS | `fa completion <shell>` follows `noun arg` pattern |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for all shells |
| VII. Agent-Friendly Design | ✅ PASS | Script output is deterministic |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |
| Shell Completion (docs section) | ✅ PASS | Implements mandated shell completion requirement |

**Post-Design Re-check**: ✅ All principles satisfied. Design uses Cobra's standard patterns, reads local files only, and provides graceful degradation outside workspaces.

## Project Structure

### Documentation (this feature)

```text
specs/013-completion/
├── plan.md              # This file
├── research.md          # Technical decisions (Cobra patterns, shell handling)
├── data-model.md        # Entities (CompletionScript, Static/Dynamic completions)
├── quickstart.md        # Implementation guide
├── contracts/           # CLI contract definitions
│   └── cli-completion.md
└── checklists/
    └── requirements.md  # Spec validation checklist
```

### Source Code (repository root)

```text
cmd/
└── foundagent/
    └── main.go              # Entrypoint (existing)

internal/
├── cli/
│   ├── root.go              # Root command (existing)
│   ├── completion.go        # NEW: completion command
│   ├── completion_helpers.go # NEW: dynamic completion functions
│   └── completion_test.go   # NEW: table-driven tests
├── workspace/
│   └── workspace.go         # Existing: Discover(), ListWorktrees(), ListRepos()
└── config/
    └── config.go            # Existing: config file parsing

testdata/
└── completion/              # NEW: test fixtures if needed
```

**Structure Decision**: Standard Go project layout per constitution. Completion command lives in `internal/cli/` alongside other commands. Uses existing `workspace` package for dynamic completions.

## Complexity Tracking

No constitution violations to justify. Design follows all principles:
- Uses Cobra's native completion (no custom shell scripts)
- Reads existing workspace state (no new state management)
- Graceful degradation (no errors when outside workspace)
