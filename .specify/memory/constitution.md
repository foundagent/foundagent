<!--
SYNC IMPACT REPORT
==================
Version change: 1.5.2 → 1.5.3 (Project identity addition)

Modified principles: None

Added sections:
- Project Identity section (domain, GitHub org, naming)

Removed sections: None

Templates requiring updates: None

Follow-up TODOs: None
-->

# Foundagent Constitution

## Project Identity

| Property | Value |
|----------|-------|
| **Name** | Foundagent |
| **CLI Commands** | `foundagent`, `fa` (alias) |
| **Domain** | [foundagent.dev](https://foundagent.dev) |
| **GitHub Organization** | [github.com/foundagent](https://github.com/foundagent) |
| **Primary Repository** | [github.com/foundagent/foundagent](https://github.com/foundagent/foundagent) |
| **Homebrew Tap** | `foundagent/tap` |
| **License** | MIT |

**Domain Usage**:
- `foundagent.dev` — Documentation site (when deployed)
- `foundagent.dev/docs` — Full documentation
- `foundagent.dev/install` — Installation instructions (optional redirect)

## Core Principles

### I. CLI-First Design
All functionality MUST be accessible via command-line interface. The CLI is the primary
user interface and the authoritative way to interact with Foundagent. Every command MUST:

- Accept arguments and stdin for input
- Write results to stdout, errors to stderr
- Support both human-readable and JSON output formats (via `--json` flag)
- Exit with meaningful status codes (0 = success, non-zero = specific error type)

**Rationale**: AI coding agents interact through text. CLI-first ensures seamless integration
with LLM-based workflows and enables scripting, automation, and composability.

### II. Git-Native Operations
Foundagent MUST leverage Git's native capabilities rather than reimplementing version control
primitives. Specifically:

- Worktrees MUST be created and managed using `git worktree` commands
- Repository state MUST be read through Git porcelain/plumbing commands
- No custom file-locking or state management that conflicts with Git internals
- All operations MUST leave the repository in a valid Git state

**Rationale**: Git is the source of truth. Fighting Git creates fragility; embracing it
ensures predictable behavior and interoperability with existing Git tooling.

### III. Non-Destructive by Default
All Foundagent operations MUST be non-destructive unless explicitly confirmed by the user.
This means:

- No force-deletes without `--force` or interactive confirmation
- Worktree removal MUST warn about uncommitted changes before proceeding
- Branch deletions MUST verify the branch is merged or use `--force`
- Configuration changes MUST be reversible or explicitly acknowledged

**Rationale**: Developers switch contexts frequently. Accidental data loss destroys trust.
Safe defaults with explicit overrides balance safety and power-user needs.

### IV. Multi-Repository Awareness
Foundagent MUST treat multi-repo workspaces as first-class citizens. Core requirements:

- Workspace configuration MUST support multiple repositories
- Commands MUST operate across repos when contextually appropriate
- Status and sync operations MUST aggregate state from all configured repos
- Repository relationships (dependencies, shared branches) MUST be configurable

**Rationale**: Modern development involves microservices, monorepos, and polyrepos.
Foundagent's value proposition centers on managing this complexity.

### V. Simplicity and Discoverability
The CLI MUST be intuitive and self-documenting. Requirements:

- Commands MUST follow a consistent `foundagent <noun> <verb>` pattern
- Every command MUST include `--help` with examples
- Error messages MUST include actionable remediation steps
- Configuration MUST use sensible defaults requiring minimal initial setup

**Rationale**: Developers adopt tools that respect their time. A tool that requires a
manual to use will not be used. Clarity reduces support burden and accelerates adoption.

### VI. Test-Driven Development
All features MUST be developed using TDD methodology:

- Tests MUST be written before implementation code
- Tests MUST fail before implementation begins (red phase)
- Implementation MUST make tests pass with minimal code (green phase)
- Code MUST be refactored only after tests pass (refactor phase)

**Coverage Requirements**:

- **Minimum coverage**: 80% line coverage for all packages
- **Critical paths**: 100% coverage for `internal/worktree/` and `internal/workspace/`
- **Table-driven tests**: MUST use table-driven test patterns for all functions with multiple inputs
- **Test isolation**: Tests MUST NOT depend on external state; use `t.TempDir()` for filesystem tests
- **CI enforcement**: Coverage MUST be checked in CI; PRs below threshold MUST NOT merge

**Rationale**: CLI tools interact with file systems and external processes. Tests ensure
correctness across platforms and prevent regressions in complex state management.

### VII. Agent-Friendly Design
While Foundagent is primarily a developer tool, it MUST be designed for future integration
with AI coding agents. Requirements:

- **Human-readable default**: All output optimized for developer consumption by default
- **Structured output**: Every command MUST support `--json` flag for machine parsing
- **Idempotent operations**: Commands SHOULD be safely re-runnable without side effects
- **Predictable errors**: Error output MUST include:
  - Machine-parseable error code (e.g., `E001`, `E002`)
  - Human-readable message
  - Actionable remediation steps
- **Context awareness**: Status commands MUST provide complete state for agent recovery
- **No interactive prompts in JSON mode**: When `--json` is set, MUST fail rather than prompt

**Rationale**: AI agents interact through text and require predictable, parseable output.
Designing for agent compatibility from day one avoids costly retrofitting later.

## Cross-Platform Compatibility

Foundagent targets macOS, Linux, and Windows. The following constraints apply:

- **File paths**: MUST use platform-agnostic path handling (no hardcoded separators)
- **Shell commands**: MUST NOT assume a specific shell; use Git's cross-platform behavior
- **Line endings**: MUST respect repository `.gitattributes` and user configuration
- **Executables**: MUST provide native binaries for macOS (amd64, arm64), Linux (amd64, arm64), and Windows (amd64)
- **Testing**: CI MUST run tests on all three target platforms before release

All path operations MUST use `filepath` package for separator normalization. Subprocess
invocations MUST avoid shell-specific syntax. Documentation MUST include platform-specific
notes where behavior differs.

## Technology Stack

### Language & Runtime

- **Language**: Go (latest stable, currently 1.21+)
- **CLI Binary Names**: `foundagent` (primary), `fa` (alias) — both point to same binary

### Core Libraries

| Purpose | Library | Rationale |
|---------|---------|-----------|
| CLI Framework | `cobra` | Industry standard (kubectl, gh, hugo); subcommand support, auto-generated help |
| Configuration | `viper` | Multi-format config files, env vars, flags integration |
| Git Operations | `go-git` | Pure Go Git implementation; no external Git dependency for core ops |
| Git CLI Fallback | `os/exec` | For `git worktree` commands not supported by go-git |
| Output Formatting | `lipgloss` / `termenv` | Cross-platform colored terminal output |
| JSON Output | `encoding/json` | Standard library; structured output for `--json` flag |
| Testing | `testing` + `testify` | Standard library with assertions; table-driven tests |

### Build & Distribution

- **Build Tool**: `go build` with `goreleaser` for cross-platform releases
- **Binary Alias**: Both `foundagent` and `fa` binaries produced (symlink or duplicate)
- **Platforms**: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- **Distribution**: GitHub Releases, Homebrew tap, optional install script

### Project Structure

```
foundagent/
├── cmd/
│   └── foundagent/
│       └── main.go           # Entrypoint
├── internal/
│   ├── cli/                  # Cobra command definitions
│   ├── workspace/            # Workspace management logic
│   ├── worktree/             # Git worktree operations
│   ├── config/               # Configuration handling
│   └── output/               # Human/JSON output formatting
├── pkg/                      # Public API (if any)
├── testdata/                 # Test fixtures
├── .goreleaser.yaml          # Release configuration
├── go.mod
└── go.sum
```

### Makefile Targets

A `Makefile` MUST be provided for consistent developer experience:

| Target | Required | Purpose |
|--------|----------|----------|
| `make build` | ✅ Yes | Compile binary to `./bin/foundagent` |
| `make test` | ✅ Yes | Run all tests with race detector |
| `make lint` | ✅ Yes | Run `golangci-lint` |
| `make fmt` | ✅ Yes | Format code with `gofmt` and `goimports` |
| `make coverage` | ✅ Yes | Generate coverage report, open in browser |
| `make install` | ✅ Yes | Install to `$GOPATH/bin` |
| `make clean` | ✅ Yes | Remove build artifacts and test cache |
| `make all` | ✅ Yes | Run `fmt`, `lint`, `test`, `build` in sequence |
| `make release-dry-run` | ✅ Yes | Test `goreleaser` locally without publishing |
| `make docs` | ✅ Yes | Serve documentation locally |
| `make help` | ✅ Yes | Show all available targets with descriptions |

**Contributor Workflow**:
```bash
git clone https://github.com/foundagent/foundagent
cd foundagent
make all          # Format, lint, test, build
make install      # Install locally to test
```

**Makefile Standards**:
- Targets MUST be idempotent (safe to run repeatedly)
- Targets MUST print what they're doing
- `make help` MUST be the default target if no target specified
- Use `.PHONY` for all non-file targets

## Go Ecosystem Standards

Foundagent MUST adhere to established Go community standards and idioms:

### Code Style & Formatting

- **gofmt**: All code MUST be formatted with `gofmt` (enforced in CI)
- **goimports**: Import statements MUST be organized by `goimports`
- **golangci-lint**: MUST pass `golangci-lint` with the following linters enabled:
  - `errcheck` — unchecked errors
  - `gosimple` — simplifications
  - `govet` — suspicious constructs
  - `ineffassign` — ineffectual assignments
  - `staticcheck` — static analysis
  - `unused` — unused code
  - `gocyclo` — cyclomatic complexity (max 15)
  - `misspell` — spelling mistakes
  - `gosec` — security issues

### Go Idioms (Effective Go)

- **Error handling**: Errors MUST be handled explicitly; no ignored errors without comment
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` for error context
- **Naming**: Follow Go naming conventions (MixedCaps, not underscores)
- **Package design**: Packages MUST have a single, clear purpose
- **Interface size**: Interfaces SHOULD be small (1-3 methods); accept interfaces, return structs
- **Zero values**: Types SHOULD be usable with zero values where sensible
- **Context propagation**: Long-running operations MUST accept `context.Context` as first parameter

### Documentation

- **Package comments**: Every package MUST have a package-level doc comment
- **Exported symbols**: Every exported function, type, and constant MUST have a doc comment
- **Examples**: Public APIs SHOULD include `Example` test functions for `go doc`
- **README**: Package directories MAY include a README.md for complex packages

### Module & Dependency Management

- **Go modules**: MUST use Go modules (`go.mod`) for dependency management
- **Dependency hygiene**: Minimize external dependencies; prefer standard library
- **Version pinning**: Dependencies MUST be pinned to specific versions
- **Vulnerability scanning**: `govulncheck` MUST pass in CI

### References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

## Configuration & State Management

### Configuration File Format

Foundagent supports multiple configuration formats for user preference:

| Format | File Name | Use Case |
|--------|-----------|----------|
| YAML | `.foundagent.yaml` or `.foundagent.yml` | Human-friendly, most common |
| TOML | `.foundagent.toml` | Explicit, popular in Go/Rust ecosystems |
| JSON | `.foundagent.json` | Machine-generated configs, universal |

**Resolution Order** (first found wins):
1. `--config` flag (explicit path)
2. `.foundagent.yaml` in current directory
3. `.foundagent.toml` in current directory
4. `.foundagent.json` in current directory
5. Walk up directory tree to find workspace root

**Format Notes**:
- YAML: Familiar to most developers, supports comments, whitespace-sensitive
- TOML: More explicit than YAML, no whitespace ambiguity, supports comments
- JSON: Universal but verbose, no comments allowed

All formats MUST be validated against the same JSON Schema. Invalid configs MUST
produce clear error messages with line numbers and remediation hints.

### State Storage

Foundagent state lives **per-workspace** in a `.foundagent/` directory:

```
workspace-root/
├── .foundagent/
│   ├── config.yaml          # Workspace configuration
│   ├── state.json           # Runtime state (worktree mappings, etc.)
│   ├── repos/               # Bare clones for multi-repo workspaces
│   │   ├── repo-a.git/
│   │   └── repo-b.git/
│   └── worktrees/           # Worktree checkouts
│       ├── feature-123/
│       └── bugfix-456/
├── .foundagent.yaml         # User-facing config (can also be here)
└── ... (user's project files)
```

**Benefits**:
- State travels with the workspace (portable, version-controllable if desired)
- No global state pollution
- Easy cleanup: delete `.foundagent/` to reset
- Multi-user friendly: each clone has its own state

**State File Requirements**:
- `state.json` MUST be machine-managed; users SHOULD NOT edit directly
- State MUST be recoverable from Git state if corrupted/deleted
- Lock files (`.foundagent/.lock`) MUST prevent concurrent modifications

### Configuration Schema

- Schema MUST be published as JSON Schema for editor autocomplete
- Schema version MUST be included in config files
- Migration tool MUST handle schema version upgrades
- Unknown keys MUST warn but not fail (forward compatibility)

## Logging & Observability

### Verbosity Levels

Foundagent uses a simple two-tier verbosity system:

| Flag | Level | Output |
|------|-------|--------|
| (none) | Normal | Essential output only: results, warnings, errors |
| `-v` | Verbose | Detailed progress, intermediate steps, timing info |
| `--debug` | Debug | Full trace: all Git commands, internal state, stack traces |

**Implementation**:
- Verbose/debug output goes to stderr (stdout reserved for results)
- Timestamps included in debug mode
- Colors disabled when stderr is not a TTY or `NO_COLOR` env is set

### Log File

Logs are written to file **only when explicitly requested**:

```bash
foundagent workspace sync --log-file=sync.log
```

- Log files use structured JSON format for machine parsing
- Log files include all debug-level output regardless of terminal verbosity
- No automatic log file creation (avoids disk clutter)

### Structured Error Codes

All errors MUST include a code for programmatic handling:

| Code Range | Category | Example |
|------------|----------|----------|
| `E0xx` | Configuration errors | `E001: Invalid config file` |
| `E1xx` | Git operation errors | `E101: Worktree already exists` |
| `E2xx` | Workspace errors | `E201: Repository not found` |
| `E3xx` | Network/remote errors | `E301: Remote unreachable` |
| `E9xx` | Internal errors | `E999: Unexpected error` |

**Error Output Format**:
```
Error E101: Worktree 'feature-123' already exists
  Location: /path/to/workspace/.foundagent/worktrees/feature-123
  Hint: Use 'fa worktree switch feature-123' to switch to it, or
        'fa worktree remove feature-123 --force' to remove and recreate
```

## Performance & Resilience

### Performance Philosophy

Foundagent prioritizes correctness over speed, with these guidelines:

- **No strict timing targets**: Optimize when profiling reveals bottlenecks
- **Parallel by default**: Multi-repo operations MUST run in parallel where safe
- **Progress feedback**: Long operations MUST show progress indicators
- **No timeouts**: Operations complete or fail based on underlying Git behavior

**Parallelization Rules**:
- Repository sync operations: parallel across repos
- Worktree creation: parallel when creating multiple
- Status checks: parallel across all repos and worktrees
- Use worker pool pattern with sensible concurrency limit (e.g., `runtime.NumCPU()`)

### Network Resilience

Foundagent uses **graceful degradation** for network-dependent operations:

| Operation | Network Required? | Offline Behavior |
|-----------|-------------------|------------------|
| `worktree create` | No (uses local bare clone) | Works fully offline |
| `worktree list` | No | Works fully offline |
| `worktree switch` | No | Works fully offline |
| `workspace sync` | Yes (fetches remotes) | Warns and skips remote sync |
| `workspace status` | No | Shows local state; warns if remote check skipped |
| `repo add` | Yes (initial clone) | Fails with clear error |

**Behavior**:
- Operations that CAN work offline MUST work offline
- Operations that REQUIRE network MUST fail fast with clear error
- Mixed operations (e.g., sync) MUST complete local work and report remote failures
- Warnings MUST clearly indicate what was skipped due to network issues

**Implementation**:
- Detect network availability via quick remote HEAD check with short timeout
- Cache network status for duration of command (avoid repeated checks)
- `--offline` flag MAY be added later if explicit control requested

## Development Workflow

### Code Quality Gates

Before any code is merged:

1. All tests MUST pass on macOS, Linux, and Windows
2. Linting and formatting MUST pass with zero warnings
3. New commands MUST include `--help` documentation
4. Breaking changes MUST be documented in CHANGELOG with migration guidance

### Versioning

Foundagent follows Semantic Versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes to CLI interface or configuration format
- **MINOR**: New commands or features, backward-compatible
- **PATCH**: Bug fixes, documentation updates, performance improvements

Pre-1.0 releases may have breaking changes in MINOR versions, documented clearly.

### Backward Compatibility Policy

Foundagent follows strict SemVer for backward compatibility:

**Deprecation Process**:
1. Feature deprecated in version `v1.x` with warning message
2. Warning MUST include removal version and migration path
3. Feature removed in next MAJOR version (`v2.0`)
4. CHANGELOG MUST document all deprecations prominently

**What Constitutes a Breaking Change**:
- CLI flag/argument removal or rename
- Configuration key removal or rename
- Output format changes (human-readable exempted; JSON is contract)
- Exit code meaning changes
- State file format changes without migration

**What Is NOT a Breaking Change**:
- Adding new commands, flags, or config options
- Adding new fields to JSON output
- Performance improvements
- Bug fixes (even if someone depended on the bug)
- Human-readable output formatting changes

**Migration Support**:
- Major version upgrades MUST include `fa migrate` command
- Migration MUST be non-destructive (backup before modify)
- Migration MUST be re-runnable (idempotent)

### Commit Standards

Commits MUST follow Conventional Commits specification:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions/changes
- `refactor:` for code restructuring
- `chore:` for maintenance tasks

## CI/CD Pipeline

All CI/CD MUST run on **GitHub Actions** using standard Go community patterns.

### Workflows

#### 1. CI Workflow (`.github/workflows/ci.yml`)

**Triggers**: Push to `main`, all pull requests

**Jobs**:

| Job | Runs On | Purpose |
|-----|---------|----------|
| `lint` | ubuntu-latest | `golangci-lint` via `golangci/golangci-lint-action` |
| `test` | ubuntu-latest, macos-latest, windows-latest | `go test -race -coverprofile=coverage.out ./...` |
| `coverage` | ubuntu-latest | Upload to Codecov; enforce 80% threshold |
| `build` | ubuntu-latest | `go build ./...` to verify compilation |
| `vulncheck` | ubuntu-latest | `govulncheck ./...` for security vulnerabilities |

**Matrix Strategy**:
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: ['1.21', '1.22']  # Test on latest two Go versions
```

**Required Checks**: All jobs MUST pass before merge.

#### 2. Release Workflow (`.github/workflows/release.yml`)

**Triggers**: Push of version tag (`v*.*.*`)

**Jobs**:

| Job | Purpose |
|-----|----------|
| `goreleaser` | Build binaries via `goreleaser/goreleaser-action` |
| `checksums` | Generate SHA256 checksums for all artifacts |
| `publish` | Create GitHub Release with binaries, changelog |

**goreleaser Configuration** (`.goreleaser.yaml`):
- Builds: `foundagent` and `fa` binaries for all platforms
- Archives: tar.gz (unix), zip (windows)
- Changelog: Auto-generated from conventional commits
- Homebrew: Publish to `foundagent/homebrew-tap`
- Scoop: Generate manifest for Windows users

#### 3. Dependency Workflow (`.github/workflows/deps.yml`)

**Triggers**: Weekly schedule, manual dispatch

**Jobs**:

| Job | Purpose |
|-----|----------|
| `dependabot` | Enabled via `.github/dependabot.yml` for go modules |
| `update` | `go get -u ./...` + `go mod tidy` via PR |

### GitHub Actions Standards

- **Action versions**: Pin to major version (`@v4`) or SHA for security
- **Go setup**: Use `actions/setup-go@v5` with `go-version-file: go.mod`
- **Caching**: Enable module caching via `cache: true` in setup-go
- **Secrets**: Release tokens stored in repository secrets (`GITHUB_TOKEN`, `HOMEBREW_TAP_TOKEN`)
- **Timeouts**: Set `timeout-minutes` on all jobs to prevent hung workflows

### Branch Protection

`main` branch MUST have:

- Require pull request before merging
- Require status checks: `lint`, `test`, `build`, `vulncheck`
- Require branches to be up to date
- Require conversation resolution

## Open Source Community

Foundagent is an open source project that welcomes community contributions. The repository
MUST be structured to support external contributors from day one.

### Required Community Files

All files MUST live in the repository root or `.github/` directory:

| File | Purpose |
|------|----------|
| `README.md` | Project overview, quick install, basic usage, badges |
| `LICENSE` | MIT License (permissive, contributor-friendly) |
| `CONTRIBUTING.md` | How to contribute: setup, workflow, standards |
| `CODE_OF_CONDUCT.md` | Contributor Covenant v2.1 (industry standard) |
| `SECURITY.md` | Security vulnerability reporting process |
| `CHANGELOG.md` | Keep a Changelog format, auto-updated by releases |
| `.github/FUNDING.yml` | GitHub Sponsors / Open Collective (optional) |

### GitHub Issue & PR Templates

#### Issue Templates (`.github/ISSUE_TEMPLATE/`)

| Template | Purpose |
|----------|----------|
| `bug_report.yml` | Structured bug reports with OS, version, repro steps |
| `feature_request.yml` | Feature proposals with use case, alternatives considered |
| `question.yml` | Redirect to Discussions for support questions |
| `config.yml` | Blank issue option disabled; links to templates |

#### Pull Request Template (`.github/PULL_REQUEST_TEMPLATE.md`)

MUST include:
- Description of changes
- Related issue link (`Fixes #123`)
- Type of change (feat/fix/docs/refactor)
- Checklist: tests added, docs updated, lint passes
- Breaking change disclosure

### GitHub Discussions

Enable GitHub Discussions with categories:

| Category | Purpose |
|----------|----------|
| 📣 Announcements | Release notes, project updates (maintainer-only posts) |
| 💡 Ideas | Feature brainstorming before formal issues |
| 🙏 Q&A | Support questions, how-to guidance |
| 🙌 Show and Tell | Community showcases, integrations |

### Labels

Standardized labels for issue/PR triage:

| Label | Description |
|-------|-------------|
| `good first issue` | Suitable for new contributors |
| `help wanted` | Maintainer seeks community help |
| `bug` | Confirmed bug |
| `enhancement` | New feature or improvement |
| `documentation` | Docs improvements |
| `breaking change` | Requires major version bump |
| `needs triage` | Awaiting maintainer review |
| `wontfix` | Declined with explanation |
| `duplicate` | Already reported |
| `priority: high` | Urgent issues |
| `priority: low` | Backlog items |

### Contributor Workflow

1. **Fork** → Clone → Create feature branch
2. **Develop** → Follow TDD, run `make lint test`
3. **Commit** → Conventional commits, sign-off (`-s`) recommended
4. **Push** → Open PR against `main`
5. **Review** → Address feedback, maintainer approves
6. **Merge** → Squash merge with conventional commit message

### Recognition

- **All Contributors**: Use `all-contributors` bot to recognize non-code contributions
- **Release credits**: Contributors mentioned in CHANGELOG and GitHub Release notes
- **README**: Contributors section with avatars (auto-generated)

### Community Health Workflows (`.github/workflows/`)

#### Stale Issues (`.github/workflows/stale.yml`)

- Mark issues inactive after 60 days with `stale` label
- Close after 14 additional days if no response
- Exempt: `priority: high`, `help wanted`, `good first issue`

#### Welcome Bot (`.github/workflows/welcome.yml`)

- Auto-comment on first-time contributor PRs with thank you + guidance
- Auto-comment on first issues with expected response time

#### Issue Labeler (`.github/workflows/labeler.yml`)

- Auto-apply labels based on file paths changed (e.g., `documentation` for docs/)

## Documentation Requirements

Comprehensive documentation is essential for adoption and contributions.

### Documentation Structure

```
docs/
├── getting-started/
│   ├── installation.md         # All install methods (brew, scoop, binary, source)
│   ├── quickstart.md            # 5-minute intro to core workflow
│   └── configuration.md         # Config file format, env vars, defaults
├── guides/
│   ├── workspaces.md            # Creating and managing workspaces
│   ├── worktrees.md             # Git worktree operations
│   ├── multi-repo.md            # Multi-repository workflows
│   └── ai-agents.md             # Integration with AI coding tools
├── reference/
│   ├── cli.md                   # Complete CLI reference (auto-generated)
│   ├── config-schema.md         # Full configuration reference
│   └── environment-vars.md      # Environment variable reference
├── contributing/
│   ├── development-setup.md     # Local dev environment setup
│   ├── architecture.md          # Codebase structure, key decisions
│   ├── testing.md               # How to write and run tests
│   └── releasing.md             # Release process documentation
└── faq.md                       # Common questions and troubleshooting
```

### README.md Requirements

The root README MUST include:

1. **Header**: Logo, tagline, badges (CI, coverage, release, license)
2. **What is Foundagent**: One-paragraph description
3. **Key Features**: Bullet list of capabilities
4. **Quick Install**: One-liner for each platform
5. **Quick Start**: 3-5 commands showing basic workflow
6. **Documentation**: Link to full docs
7. **Contributing**: Link to CONTRIBUTING.md
8. **License**: MIT with link
9. **Community**: Links to Discussions, issues

### CLI Self-Documentation

- `foundagent --help` MUST show all commands with descriptions
- `foundagent <command> --help` MUST show full usage with examples
- `foundagent docs` command SHOULD open documentation in browser
- CLI help text MUST be kept in sync with docs/reference/cli.md

### Shell Completion

Foundagent MUST provide shell completion for all major shells:

```bash
# Generate completion scripts
foundagent completion bash    # Bash completion
foundagent completion zsh     # Zsh completion
foundagent completion fish    # Fish completion
foundagent completion powershell  # PowerShell completion
```

**Requirements**:
- Completion MUST cover all commands, subcommands, and flags
- Completion MUST include dynamic completions where applicable:
  - Worktree names for `worktree switch`, `worktree remove`
  - Repository names for multi-repo commands
  - Branch names from local/remote refs
- Installation instructions MUST be included in `--help` output
- Documentation MUST include setup instructions for each shell

**Implementation**: Use Cobra's built-in completion generation (`cobra.Command.GenBashCompletion`, etc.)

### Documentation Standards

- **Format**: Markdown with consistent heading hierarchy
- **Code blocks**: Include shell commands with copy-friendly formatting
- **Platform notes**: Call out macOS/Linux/Windows differences
- **Versioning**: Docs versioned with releases (git tags)
- **Link checking**: CI MUST validate internal documentation links
- **Spell checking**: CI SHOULD run spell checker on docs

### Documentation Site (Optional, Future)

When traffic warrants, deploy documentation to `foundagent.dev`:

- Use `mkdocs` with Material theme or similar
- Deploy to GitHub Pages with custom domain
- Include search functionality
- Add versioned documentation dropdown
- Configure `foundagent.dev` CNAME in GitHub Pages settings

## Privacy & Telemetry

**Foundagent collects no telemetry.**

- No usage data is collected or transmitted
- No analytics, crash reports, or phone-home behavior
- No network requests except explicit user-initiated Git operations
- This policy MUST be stated in the README

**Rationale**: Developer tools must be trustworthy. Telemetry creates friction with
enterprise adoption and erodes user trust. GitHub metrics (stars, issues, clones)
provide sufficient signal for project health.

## Governance

This constitution is the authoritative guide for Foundagent development. It supersedes
informal practices, chat discussions, and undocumented conventions.

### Amendment Process

1. Propose amendment via GitHub Issue with `constitution` label
2. Document rationale, impact on existing code, and migration plan
3. Require maintainer approval before merge
4. Update constitution version according to semantic versioning:
   - MAJOR: Principle removal or fundamental redefinition
   - MINOR: New principle or significant guidance expansion
   - PATCH: Clarifications, typo fixes, non-semantic changes

### Compliance

- All pull requests MUST verify compliance with constitution principles
- Code reviews MUST check for principle violations
- Deviations MUST be documented with explicit justification in the PR description

**Version**: 1.5.3 | **Ratified**: 2025-12-03 | **Last Amended**: 2025-12-03
