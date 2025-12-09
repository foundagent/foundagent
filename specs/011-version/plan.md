````markdown
# Implementation Plan: Version

**Branch**: `011-version` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-version/spec.md`

## Summary

Implement `fa version` command that displays version information. Supports `--full` for detailed build metadata (commit, date, Go version, OS/arch), `--json` for machine-readable output, and `--check` to query GitHub releases for updates.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Cobra (CLI), net/http (update check)  
**Storage**: None (version embedded at build time via ldflags)  
**Testing**: `go test` with `testify/assert`, table-driven tests  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Version display instant; update check under 5 seconds  
**Constraints**: Version info injected via `-ldflags` at build time  
**Scale/Scope**: Single command, foundational for support/debugging

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Single-line default, `--json` support |
| II. Git-Native Operations | ✅ PASS | N/A (version is metadata) |
| III. Non-Destructive by Default | ✅ PASS | Read-only operation |
| IV. Multi-Repository Awareness | ✅ PASS | N/A (tool-level command) |
| V. Simplicity and Discoverability | ✅ PASS | `fa version` / `fa --version` intuitive |
| VI. Test-Driven Development | ✅ PASS | Tests for version formatting, update check |
| VII. Agent-Friendly Design | ✅ PASS | `--json` for version parsing |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific behavior |

**Post-Design Re-check**: ✅ All principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/011-version/
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
│   ├── version.go           # NEW: version command
│   ├── version_test.go      # NEW: version tests
│   └── root.go              # MODIFY: add --version flag
├── version/
│   ├── version.go           # NEW: version variables, build info
│   ├── update.go            # NEW: GitHub releases API check
│   └── version_test.go      # NEW: version tests
└── Makefile                 # MODIFY: add ldflags for version injection
```

**Structure Decision**: Dedicated `internal/version/` package for version info and update checking. Separates concerns from CLI.

## Complexity Tracking

**Complexity: Build-Time Version Injection**

Version, commit, and build date must be injected at build time via:
```
go build -ldflags "-X internal/version.Version=1.0.0 -X internal/version.Commit=abc1234 ..."
```

**Justification**: Standard Go practice. Must update Makefile and goreleaser config.

## Dependencies

**Depends on**:
- None (foundational utility command)

**Provides to other specs**:
- `internal/version/` package → used by any command needing version info
- Build-time ldflags pattern → documented for goreleaser
````