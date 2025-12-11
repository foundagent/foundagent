<div align="center">

# <img src="assets/foundagent_logo.png" width="40" alt="Foundagent Logo" style="vertical-align: middle;"> &nbsp; Foundagent

The foundational workspace manager for agentic development

</div>

## Overview

Foundagent is a CLI tool for managing multi-repository development environments using git worktrees. It creates a unified workspace that integrates seamlessly with VS Code, allowing you to work with multiple repositories and branches simultaneously.

## Installation

### From GitHub Releases (Recommended)

Download the latest pre-built binary for your platform from the [GitHub Releases](https://github.com/foundagent/foundagent/releases) page.

**macOS/Linux:**
```bash
# Download the binary (replace VERSION and PLATFORM with appropriate values)
curl -L -o foundagent https://github.com/foundagent/foundagent/releases/download/VERSION/foundagent-PLATFORM

# Make it executable
chmod +x foundagent

# Move to a location in your PATH
sudo mv foundagent /usr/local/bin/fa
```

**Prerequisites:**
- Git

### Build from Source

If you prefer to build from source:

**Prerequisites:**
- Go 1.25 or later
- Git

**Build:**
```bash
git clone https://github.com/foundagent/foundagent.git
cd foundagent
go build -o foundagent ./cmd/foundagent
```

Move the binary to a location in your PATH:

```bash
sudo mv foundagent /usr/local/bin/fa
```

## Features

### Workspace Initialization

Create a new Foundagent workspace with a single command:

```bash
fa init my-project
```

This creates a self-contained project structure:

```
my-project/
├── .foundagent/              # Machine-managed state
│   └── state.json            # Runtime state (JSON)
├── .foundagent.yaml          # User-editable configuration
├── repos/                    # Repository storage
│   ├── .bare/                # Bare repository clones
│   └── worktrees/            # Working directories by repo/branch
└── my-project.code-workspace # VS Code workspace file
```

### Machine-Readable Output

Get JSON output for automation and AI tools:

```bash
fa init my-project --json
```

Output example:
```json
{
  "status": "success",
  "data": {
    "name": "my-project",
    "path": "/path/to/my-project",
    "action": "created"
  }
}
```

### Workspace Recovery

Reinitialize a corrupted workspace while preserving repositories:

```bash
fa init my-project --force
```

This recreates the workspace configuration and state files while preserving the `repos/` directory contents.

## Usage

### Initialize a Workspace

```bash
# Create a new workspace
fa init my-project

# Get JSON output
fa init my-project --json

# Force reinitialize existing workspace
fa init my-project --force
```

### Add Repositories

```bash
# Add a single repository
fa add git@github.com:org/my-repo.git

# Add with custom name
fa add git@github.com:org/my-repo.git api-service

# Add multiple repositories in parallel
fa add git@github.com:org/repo1.git git@github.com:org/repo2.git

# Get JSON output
fa add git@github.com:org/my-repo.git --json

# Force re-clone existing repository
fa add git@github.com:org/my-repo.git --force
```

### Manage Worktrees

```bash
# Create worktrees across all repos
fa wt create feature-123

# Create from specific branch
fa wt create hotfix-1 --from release-2.0

# List all worktrees
fa wt list

# List worktrees for specific branch
fa wt list feature-123

# Switch to different branch's worktrees
fa wt switch feature-123

# Switch and create if doesn't exist
fa wt switch new-feature --create

# Remove worktrees
fa wt remove feature-123

# Force removal with uncommitted changes
fa wt remove feature-123 --force
```

### Remove Repositories

```bash
# Remove a repository
fa remove api

# Remove multiple repositories
fa remove api web

# Force removal with uncommitted changes
fa remove api --force

# Remove from config but keep files
fa remove api --config-only
```

### Check Workspace Status

```bash
# Show workspace status overview
fa status

# Show detailed status with file changes
fa status -v

# Get JSON output
fa status --json

# Use short alias
fa st
```

### Sync with Remotes

```bash
# Fetch all repos
fa sync

# Fetch and pull current branch
fa sync --pull

# Fetch and pull specific branch
fa sync feature-123 --pull

# Push all repos with unpushed commits
fa sync --push

# Stash uncommitted changes before pull
fa sync --pull --stash
```

### Health Checks

```bash
# Run diagnostic checks
fa doctor

# Get detailed output
fa doctor --verbose

# Auto-fix fixable issues
fa doctor --fix

# JSON output
fa doctor --json
```

### Version Information

```bash
# Show version
fa version

# Show detailed build information
fa version --full

# Check for updates
fa version --check

# JSON output
fa version --json
```

### Shell Completion

```bash
# Generate Bash completion
source <(fa completion bash)

# Generate Zsh completion
fa completion zsh > ~/.zsh/completion/_fa

# Generate Fish completion
fa completion fish > ~/.config/fish/completions/fa.fish

# Generate PowerShell completion
fa completion powershell > fa_completion.ps1
```

### Workspace Structure

- **`.foundagent.yaml`**: User-editable YAML configuration containing workspace name and repository list
- **`.foundagent/state.json`**: Machine-managed JSON state for runtime tracking
- **`repos/.bare/`**: Hidden directory for bare repository clones
- **`repos/worktrees/`**: Visible working directories organized by repository and branch
- **`<name>.code-workspace`**: VS Code workspace file for multi-root workspace support

## Design Principles

1. **CLI-First**: All operations are CLI-driven with meaningful exit codes
2. **Git-Native**: Uses standard git worktree operations
3. **Non-Destructive**: Requires explicit flags for potentially destructive operations
4. **Multi-Repository Aware**: Designed from the ground up for multiple repositories
5. **Agent-Friendly**: JSON output mode for AI tools and automation

## Commands Reference

### Workspace Commands
- `fa init <name>` - Initialize a new workspace
- `fa add <url> [name]` - Add repository to workspace
- `fa remove <repo>...` - Remove repositories from workspace
- `fa status` (alias: `fa st`) - Show workspace status
- `fa sync [branch]` - Sync workspace with remotes

### Worktree Commands
- `fa wt create <branch>` - Create worktrees across all repos
- `fa wt list [branch]` (alias: `fa wt ls`) - List all worktrees
- `fa wt switch [branch]` - Switch to different branch's worktrees
- `fa wt remove <branch>` (alias: `fa wt rm`) - Remove worktrees

### Utility Commands
- `fa doctor` - Run workspace health checks
- `fa version` - Show version information
- `fa completion <shell>` - Generate shell completion script

### Global Flags
- `--json` - Output in JSON format (available on most commands)
- `--force` - Force operation (skip safety checks)
- `--verbose` / `-v` - Show detailed output
- `--help` / `-h` - Show help information

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run tests with coverage:

```bash
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
go tool cover -html=coverage.txt
```

## Project Status

### Implemented Features

All core features are fully implemented and tested (Specs 001-013):

#### Workspace Management
- ✅ Workspace initialization (`fa init`)
- ✅ Add repositories (`fa add`)
- ✅ Remove repositories (`fa remove`)
- ✅ Workspace status overview (`fa status`)
- ✅ Sync with remotes (`fa sync`)

#### Worktree Operations
- ✅ Create worktrees across all repos (`fa wt create`)
- ✅ List worktrees (`fa wt list`)
- ✅ Switch between worktrees (`fa wt switch`)
- ✅ Remove worktrees (`fa wt remove`)

#### Developer Tools
- ✅ Health checks and diagnostics (`fa doctor`)
- ✅ Version information and update checking (`fa version`)
- ✅ Shell completion for bash, zsh, fish, and PowerShell (`fa completion`)

#### Features
- ✅ JSON output mode for all commands (`--json`)
- ✅ Parallel operations for performance
- ✅ VS Code workspace integration
- ✅ Comprehensive error handling with remediation hints
- ✅ Cross-platform support (macOS, Linux, Windows)
- ✅ Git worktree-native operations

## Development

### Project Structure

```
foundagent/
├── cmd/
│   └── foundagent/          # Main entry point
│       └── main.go
├── internal/
│   ├── cli/                 # CLI commands
│   │   ├── root.go          # Root command setup
│   │   ├── init.go          # Workspace initialization
│   │   ├── add.go           # Add repositories
│   │   ├── remove.go        # Remove repositories
│   │   ├── status.go        # Workspace status
│   │   ├── sync.go          # Sync operations
│   │   ├── worktree.go      # Worktree parent command
│   │   ├── wt_create.go     # Create worktrees
│   │   ├── wt_list.go       # List worktrees
│   │   ├── wt_switch.go     # Switch worktrees
│   │   ├── wt_remove.go     # Remove worktrees
│   │   ├── doctor.go        # Health checks
│   │   ├── version.go       # Version information
│   │   ├── completion.go    # Shell completion
│   │   └── *_test.go        # Comprehensive tests
│   ├── workspace/           # Workspace management
│   │   ├── workspace.go     # Core operations
│   │   ├── config.go        # Configuration handling
│   │   ├── state.go         # State management
│   │   ├── status.go        # Status collection
│   │   ├── sync.go          # Sync operations
│   │   ├── worktree.go      # Worktree operations
│   │   ├── repository.go    # Repository management
│   │   ├── removal.go       # Removal operations
│   │   ├── vscode.go        # VS Code integration
│   │   ├── validation.go    # Name/path validation
│   │   └── *_test.go        # Tests
│   ├── git/                 # Git operations
│   │   ├── clone.go         # Clone operations
│   │   ├── worktree.go      # Worktree operations
│   │   ├── branch.go        # Branch operations
│   │   ├── status.go        # Status operations
│   │   ├── remote.go        # Remote operations
│   │   ├── stash.go         # Stash operations
│   │   ├── url.go           # URL parsing
│   │   └── validation.go    # Git validation
│   ├── config/              # Configuration management
│   │   ├── schema.go        # Config schema
│   │   ├── loader.go        # Config loading
│   │   ├── yaml.go          # YAML format
│   │   ├── json.go          # JSON format
│   │   ├── toml.go          # TOML format
│   │   ├── template.go      # Config templates
│   │   └── validate.go      # Config validation
│   ├── doctor/              # Health checks
│   │   ├── check.go         # Check interface
│   │   ├── runner.go        # Check runner
│   │   ├── fix.go           # Auto-fix operations
│   │   ├── git.go           # Git checks
│   │   ├── structure.go     # Structure checks
│   │   ├── repos.go         # Repository checks
│   │   ├── worktrees.go     # Worktree checks
│   │   └── consistency.go   # Consistency checks
│   ├── version/             # Version management
│   │   ├── version.go       # Version info
│   │   └── update.go        # Update checking
│   ├── errors/              # Error handling
│   │   ├── codes.go         # Error codes
│   │   └── error.go         # Error types
│   └── output/              # Output formatting
│       └── json.go          # JSON utilities
└── specs/                   # Feature specifications (001-013)
    ├── 001-workspace-init/
    ├── 002-repo-add/
    ├── 003-workspace-config/
    ├── 004-worktree-create/
    ├── 005-worktree-list/
    ├── 006-worktree-remove/
    ├── 007-workspace-status/
    ├── 008-workspace-sync/
    ├── 009-worktree-switch/
    ├── 010-repo-remove/
    ├── 011-version/
    ├── 012-doctor/
    └── 013-completion/
```

### Development Quick Start

```bash
# Clone the repository
git clone https://github.com/foundagent/foundagent.git
cd foundagent

# Build the project
make build

# Run tests
make test

# Run tests with coverage
make coverage

# Format code and run all checks
make check

# Build for all platforms
make release
```

Available make targets:
- `make build` - Build the binary
- `make test` - Run tests
- `make test-v` - Run tests with verbose output
- `make coverage` - Generate test coverage report
- `make lint` - Run linters (requires golangci-lint)
- `make fmt` - Format code
- `make vet` - Run go vet
- `make install` - Install to $GOPATH/bin
- `make clean` - Remove build artifacts
- `make check` - Run format, vet, lint, and test
- `make dev` - Quick development build and test
- `make release` - Build binaries for all platforms

### Contributing

1. Each feature follows the spec-driven development process:
   - Specification (`spec.md`)
   - Implementation plan (`plan.md`)
   - Task breakdown (`tasks.md`)
   - Quality checklist (`checklists/`)

2. All features must include:
   - Comprehensive tests with table-driven test cases
   - Error handling with remediation hints
   - JSON output support where applicable
   - Documentation updates
   - Shell completion support where relevant

3. Development guidelines:
   - Follow Go best practices and idioms
   - Maintain test coverage above 80%
   - Write clear commit messages
   - Add tests before fixing bugs
   - Keep functions small and focused
   - Document exported functions and types

## CI/CD

Foundagent uses GitHub Actions for continuous integration:

- **Test Job**: Runs tests on Ubuntu, macOS, and Windows with Go 1.25
  - Executes full test suite with race detection
  - Generates coverage reports
  - Displays coverage summary in GitHub UI
- **Lint Job**: Runs golangci-lint for code quality checks
- **Build Job**: Verifies builds for all supported platforms
  - Linux (amd64)
  - macOS (amd64, arm64)
  - Windows (amd64)

All checks must pass before code can be merged to main.

## License

See [LICENSE](LICENSE) file for details.

## Error Codes

Foundagent uses structured error codes for better debuggability:

- **E0xx**: Configuration errors (E001-E005)
- **E1xx**: Filesystem errors (E101-E104)
- **E2xx**: Git errors (E201-E203)
- **E9xx**: General errors (E999)

All errors include actionable remediation hints when possible.
