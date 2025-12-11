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

# View help
fa init --help
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

# View help
fa add --help
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

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

## Project Status

### Implemented Features (Specs 001-002)

- ✅ Workspace initialization (`fa init`)
- ✅ JSON output mode (`--json`)
- ✅ Force reinitialize (`--force`)
- ✅ Add repositories (`fa add`)
- ✅ Parallel repository cloning
- ✅ Custom repository names
- ✅ Automatic worktree creation for default branch
- ✅ VS Code workspace integration
- ✅ Comprehensive error handling with remediation hints
- ✅ Cross-platform filesystem validation

### Planned Features

- Worktree management (`fa worktree create`, `fa worktree list`, `fa worktree remove`, `fa worktree switch`)
- Repository removal (`fa repo remove`)
- Workspace operations (`fa workspace status`, `fa workspace sync`, `fa workspace config`)
- Shell completion support
- Doctor command for workspace health checks
- Version command

## Development

### Project Structure

```
foundagent/
├── cmd/
│   └── foundagent/          # Main entry point
│       └── main.go
├── internal/
│   ├── cli/                 # CLI commands
│   │   ├── root.go          # Root command
│   │   ├── init.go          # Init command
│   │   └── init_test.go     # Init command tests
│   ├── workspace/           # Workspace management
│   │   ├── workspace.go     # Core workspace operations
│   │   ├── config.go        # Configuration handling
│   │   ├── state.go         # State management
│   │   ├── vscode.go        # VS Code integration
│   │   ├── validation.go    # Name/path validation
│   │   └── workspace_test.go # Workspace tests
│   ├── errors/              # Error handling
│   │   ├── codes.go         # Error code constants
│   │   └── error.go         # Structured error type
│   └── output/              # Output formatting
│       └── json.go          # JSON output utilities
└── specs/                   # Feature specifications
    └── 001-workspace-init/  # Workspace initialization spec
        ├── spec.md
        ├── plan.md
        ├── tasks.md
        └── checklists/
```

### Contributing

1. Each feature follows the spec-driven development process:
   - Specification (`spec.md`)
   - Implementation plan (`plan.md`)
   - Task breakdown (`tasks.md`)
   - Quality checklist (`checklists/`)

2. All features must include:
   - Comprehensive tests
   - Error handling with remediation hints
   - JSON output support where applicable
   - Documentation updates

3. Follow Go best practices and maintain test coverage

## License

See [LICENSE](LICENSE) file for details.

## Error Codes

Foundagent uses structured error codes for better debuggability:

- **E0xx**: Configuration errors (E001-E005)
- **E1xx**: Filesystem errors (E101-E104)
- **E2xx**: Git errors (E201-E203)
- **E9xx**: General errors (E999)

All errors include actionable remediation hints when possible.
