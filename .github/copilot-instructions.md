# Foundagent - AI Agent Instructions

## Project Overview

Foundagent is a Go CLI tool for managing multi-repository workspaces using git worktrees. It enables developers to work with multiple repos and branches simultaneously with VS Code integration.

## Architecture

```
internal/
├── cli/         # Cobra commands (one file per command: init.go, add.go, etc.)
├── workspace/   # Core workspace operations (config, state, repos, worktrees)
├── git/         # Git operations (clone, worktree, branch, status)
├── config/      # Config loading/saving (YAML/JSON/TOML formats)
├── doctor/      # Health check system
├── errors/      # Structured errors with codes (E001-E999)
├── output/      # JSON and human-readable output formatting
└── version/     # Version info and update checking
```

**Key design patterns:**
- Commands in `cli/` delegate to `workspace/` and `git/` packages for business logic
- All commands support `--json` flag for machine-readable output via `output.PrintSuccess()`/`output.PrintError()`
- Errors use structured codes (see [internal/errors/codes.go](internal/errors/codes.go)) with remediation hints
- Parallel operations use `sync.WaitGroup` (see `addRepositories()` in [internal/cli/add.go](internal/cli/add.go))

## Development Commands

```bash
make build      # Build binary to ./foundagent
make test       # Run tests with race detector
make test-v     # Verbose test output
make coverage   # Generate coverage.html report
make lint       # Run golangci-lint (must be installed)
make check      # Format + vet + lint + test
make dev        # Quick: format + build + test
```

## Testing Patterns

Tests use **table-driven patterns** with `testify/assert` and `testify/require`:

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        setupFunc   func(t *testing.T, dir string)  // Optional setup
        expectError bool
        validateFunc func(t *testing.T, dir string, output string)
    }{...}
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()  // Always use t.TempDir() for isolation
            // ... test logic
        })
    }
}
```

Reference: [internal/cli/init_test.go](internal/cli/init_test.go)

## Adding a New Command

1. Create `internal/cli/<command>.go` with Cobra command structure
2. Add command to root in `init()`: `rootCmd.AddCommand(<command>Cmd)`
3. Implement both human-readable and JSON output paths
4. Create corresponding `internal/cli/<command>_test.go`
5. Add spec in `specs/NNN-<feature>/spec.md` (follow existing format)

## Error Handling Convention

Always use structured errors from `internal/errors`:

```go
import "github.com/foundagent/foundagent/internal/errors"

// Creating new errors
return errors.New(errors.ErrCodeInvalidName, "message", "remediation hint")

// Wrapping existing errors  
return errors.Wrap(errors.ErrCodeGitOperationFailed, "message", "remediation", err)
```

Error codes: E0xx=config, E1xx=filesystem, E2xx=git, E3xx=worktree, E4xx=network, E5xx=commit/push, E999=unknown

## Workspace Structure (Created by `fa init`)

```
<workspace>/
├── .foundagent.yaml          # User config (repos list)
├── .foundagent/state.json    # Machine state (runtime tracking)
├── repos/<repo>/.bare/       # Bare git clone
├── repos/<repo>/worktrees/   # Working directories by branch
└── <name>.code-workspace     # VS Code multi-root workspace
```

## Key Files to Reference

- **Command structure**: [internal/cli/init.go](internal/cli/init.go) - canonical command pattern
- **Workspace ops**: [internal/workspace/workspace.go](internal/workspace/workspace.go) - constants and core types
- **Git operations**: [internal/git/clone.go](internal/git/clone.go), [internal/git/worktree.go](internal/git/worktree.go)
- **Output formatting**: [internal/output/json.go](internal/output/json.go) - JSON response patterns
- **Test patterns**: [internal/cli/init_test.go](internal/cli/init_test.go) - table-driven tests with setup/validation

## Spec-Driven Development

Features follow specs in `specs/NNN-<feature>/`:
- `spec.md` - User stories, acceptance criteria, requirements
- `plan.md` - Implementation plan
- `tasks.md` - Task breakdown

When implementing features, reference the corresponding spec for acceptance criteria.
