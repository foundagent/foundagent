````markdown
# Implementation Plan: Workspace Configuration

**Branch**: `003-workspace-config` | **Date**: 2025-12-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-workspace-config/spec.md`

## Summary

Implement configuration file support with multi-format parsing (YAML, TOML, JSON), comment-preserving writes, config validation with line numbers, and config-state reconciliation. The config file (`.foundagent.yaml`) becomes the source of truth for workspace repos, enabling `fa add` (no args) to sync state with config.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: gopkg.in/yaml.v3 (comment-preserving YAML), github.com/BurntSushi/toml, encoding/json  
**Storage**: `.foundagent.yaml` (primary), `.foundagent.toml`, `.foundagent.json` (alternatives)  
**Testing**: `go test` with `testify/assert`, table-driven tests for parsing/validation  
**Target Platform**: macOS, Linux, Windows  
**Project Type**: Single Go CLI application  
**Performance Goals**: Config parsing in <100ms  
**Constraints**: Must preserve user comments in YAML when updating programmatically  
**Scale/Scope**: Support configs with 50+ repos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Design | ✅ PASS | Config updates via commands, validation errors to stderr |
| II. Git-Native Operations | ✅ PASS | Config stores repo URLs, Git handles actual clones |
| III. Non-Destructive by Default | ✅ PASS | Config updates preserve existing content |
| IV. Multi-Repository Awareness | ✅ PASS | Config designed for multiple repos |
| V. Simplicity and Discoverability | ✅ PASS | Human-readable YAML with helpful comments |
| VI. Test-Driven Development | ✅ PASS | Table-driven tests for all parsers |
| VII. Agent-Friendly Design | ✅ PASS | JSON format available, validation errors are structured |
| VIII. Agent-Agnostic Design | ✅ PASS | No agent-specific config keys |

**Post-Design Re-check**: ✅ All principles satisfied. YAML with comments is human-friendly; JSON available for machine generation.

## Project Structure

### Documentation (this feature)

```text
specs/003-workspace-config/
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
│   ├── add.go               # MODIFY: add no-args reconciliation mode
│   └── init.go              # MODIFY: use config template
├── config/
│   ├── schema.go            # NEW: config struct definitions
│   ├── yaml.go              # NEW: YAML parser with yaml.v3
│   ├── toml.go              # NEW: TOML parser
│   ├── json.go              # NEW: JSON parser
│   ├── loader.go            # NEW: format resolution, file discovery
│   ├── validate.go          # NEW: validation with line numbers
│   ├── writer.go            # NEW: comment-preserving YAML writer
│   ├── template.go          # NEW: default config template
│   └── config_test.go       # NEW: parser and validation tests
├── workspace/
│   ├── reconcile.go         # NEW: config-state diff and sync
│   └── reconcile_test.go    # NEW: reconciliation tests
└── errors/
    └── codes.go             # MODIFY: add E0xx config validation errors
```

**Structure Decision**: Dedicated `internal/config/` package for all config parsing/writing. Reconciliation logic in `workspace/` as it bridges config and state.

## Complexity Tracking

**Complexity: Comment-Preserving YAML Writes**

Writing to YAML while preserving user comments requires using yaml.v3's node API instead of simple marshal/unmarshal. This is more complex but essential for user experience.

**Justification**: Constitution Principle V (Simplicity and Discoverability) implies configs should be human-friendly. Losing comments on programmatic updates would frustrate users.

## Dependencies

**Depends on**:
- spec 001-workspace-init: workspace structure, state.json format
- spec 002-repo-add: repo URL format, clone infrastructure

**Provides to other specs**:
- `internal/config/` package (used by all commands that read/write config)
- `internal/workspace/reconcile.go` (config-state sync, used by status/doctor)
````