# Feature Specification: Workspace Configuration

**Feature Branch**: `003-workspace-config`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Configuration file support for Foundagent workspaces"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Config File Generated on Init (Priority: P1)

A developer runs `fa init my-project` and a default configuration file is created at the workspace root. The config file is the source of truth for workspace settings and can be edited directly with any text editor.

**Why this priority**: Config file generation is foundational — all other config features depend on this file existing and being properly structured.

**Independent Test**: Run `fa init test-project`, verify `.foundagent.yaml` is created at workspace root with valid default structure, open in text editor and confirm it's human-readable.

**Acceptance Scenarios**:

1. **Given** I run `fa init my-project`, **When** the workspace is created, **Then** a `.foundagent.yaml` file exists at the workspace root
2. **Given** a new workspace is created, **When** I open `.foundagent.yaml`, **Then** it contains a valid default configuration with workspace name and empty repos list
3. **Given** a new workspace is created, **When** I examine the config file, **Then** it includes helpful comments explaining each section

---

### User Story 2 - Commands Update Config (Priority: P1)

A developer uses `fa add <url>` to add a repository. The config file is automatically updated to include the new repository, keeping config as the source of truth.

**Why this priority**: Bidirectional sync between commands and config is core to the config-driven design. Users should be able to use commands OR edit config and get the same result.

**Independent Test**: Run `fa add <url>`, verify the repo appears in `.foundagent.yaml` under the `repos` section with correct URL and inferred name.

**Acceptance Scenarios**:

1. **Given** I run `fa add git@github.com:org/api.git`, **When** the command completes, **Then** `.foundagent.yaml` contains an entry for the repo under `repos`
2. **Given** I run `fa add <url> custom-name`, **When** the command completes, **Then** the config entry uses the custom name
3. **Given** I add multiple repos, **When** the commands complete, **Then** all repos are listed in the config in the order they were added

---

### User Story 3 - Sync State from Config (Priority: P1)

A developer opens `.foundagent.yaml` in their text editor, modifies the repos list (adds, removes, or renames repos), saves the file, and runs `fa add` (with no arguments). Foundagent reconciles state with config: cloning new repos, warning about removed repos, and updating internal state.

**Why this priority**: Config-driven workflow is a key differentiator. Users can define their entire workspace in a shareable config file and sync with one command.

**Independent Test**: Create a workspace with repos, manually edit config to add one repo and remove another, run `fa add`, verify new repo is cloned and warning shown for removed repo.

**Acceptance Scenarios**:

1. **Given** I manually add a repo URL to `.foundagent.yaml`, **When** I run `fa add` (with no arguments), **Then** Foundagent clones any repos in config that aren't yet cloned
2. **Given** I remove a repo from config, **When** I run `fa add`, **Then** Foundagent warns "Repo 'X' exists locally but is not in config. Run `fa remove X` to clean up, or re-add to config."
3. **Given** I edit a repo's name in the config, **When** I run `fa add`, **Then** Foundagent treats it as a remove + add (warns about old name, clones with new name)
4. **Given** repos are added and removed in config simultaneously, **When** I run `fa add`, **Then** Foundagent processes all changes and shows a summary of actions taken

---

### User Story 4 - Share Config Across Team (Priority: P2)

A team lead creates a workspace config with all the repos needed for the project. They commit `.foundagent.yaml` to a shared repo (or send it to teammates). A new developer clones the config and runs `fa add` to set up their entire environment.

**Why this priority**: Team onboarding is a major use case for config-driven setup, but individual developer workflow comes first.

**Independent Test**: Copy a `.foundagent.yaml` with multiple repos to a new workspace, run `fa add`, verify all repos are cloned and worktrees created.

**Acceptance Scenarios**:

1. **Given** I have a `.foundagent.yaml` with 5 repos defined, **When** I run `fa add` in a fresh workspace, **Then** all 5 repos are cloned in parallel
2. **Given** I share a config file, **When** a teammate uses it, **Then** they get an identical workspace structure

---

### User Story 5 - Config Validation (Priority: P2)

A developer makes a typo in the config file. When they run any `fa` command, Foundagent validates the config and reports errors with line numbers and suggestions.

**Why this priority**: Good error messages are essential for config-driven workflows where users edit files directly.

**Independent Test**: Introduce a syntax error in config, run any `fa` command, verify error message includes line number and clear description.

**Acceptance Scenarios**:

1. **Given** the config has a YAML syntax error, **When** I run any `fa` command, **Then** I see an error with line number and syntax hint
2. **Given** the config has an invalid repo URL format, **When** I run `fa add`, **Then** I see a validation error for that specific entry
3. **Given** the config references an unknown setting, **When** I run a command, **Then** I see a warning (not error) about the unknown key

---

### Edge Cases

- **Missing config file**: Workspace exists but `.foundagent.yaml` was deleted — regenerate default config on next command, warn user
- **Empty repos list**: Config exists with no repos — valid state, no error
- **Duplicate repo entries**: Same URL listed twice — warn and use first entry
- **Conflicting names**: Two repos with same name — error with clear message
- **Config vs state mismatch**: Repo in state but not in config — warn on commands, suggest removal or re-adding to config
- **Read-only config file**: Permissions prevent writing — error when trying to update, suggest fix
- **Very large config**: Hundreds of repos — should still parse and validate quickly
- **Comments in config**: User adds comments — preserve comments when config is updated by commands
- **Alternate config formats**: User creates `.foundagent.toml` or `.foundagent.json` — support all three, prefer YAML
- **Multiple config files**: Both `.foundagent.yaml` and `.foundagent.toml` exist — use first found per resolution order, warn about multiple

## Requirements *(mandatory)*

### Functional Requirements

#### Config File Structure
- **FR-001**: Config file MUST be located at workspace root as `.foundagent.yaml` (primary), `.foundagent.toml`, or `.foundagent.json`
- **FR-002**: Config file MUST be the source of truth for workspace definition
- **FR-003**: `fa init` MUST generate a default `.foundagent.yaml` with workspace name and empty repos list
- **FR-004**: Generated config MUST include comments explaining each section (for YAML/TOML formats)
- **FR-005**: Config MUST support YAML, TOML, and JSON formats with consistent schema

#### Config Schema
- **FR-006**: Config MUST include `workspace` section with `name` property
- **FR-007**: Config MUST include `repos` section as a list of repository definitions
- **FR-008**: Each repo entry MUST support `url` (required) and `name` (optional, inferred from URL)
- **FR-009**: Each repo entry MAY support `default_branch` (optional, detected from remote)
- **FR-010**: Config MUST include `settings` section for workspace-wide preferences
- **FR-011**: Settings MUST support `auto_create_worktree` (create default branch worktree on add, default: `true`)
- **FR-012**: Worktrees MUST be created at fixed location `repos/worktrees/<repo>/<branch>/`

#### Command-Config Sync
- **FR-013**: `fa add <url>` MUST update config file to include the new repo
- **FR-014**: `fa add` with no arguments MUST reconcile state with config (clone missing repos, warn about stale repos)
- **FR-015**: `fa add` (no args) MUST display a summary of actions: repos added, repos already up-to-date, repos in state but not in config
- **FR-016**: Commands MUST preserve existing config formatting and comments when updating
- **FR-017**: When a repo is in state but not in config, `fa add` MUST warn with remediation hint (remove or re-add to config)
- **FR-018**: When a repo is removed from config, system MUST NOT automatically delete cloned data
- **FR-019**: Stale repo warnings MUST include the repo name and suggest `fa remove <name>` to clean up

#### Validation
- **FR-020**: System MUST validate config on every command execution
- **FR-021**: Validation errors MUST include line numbers (for YAML/TOML) and clear descriptions
- **FR-022**: Unknown config keys MUST produce warnings, not errors (forward compatibility)
- **FR-023**: Invalid repo URLs MUST produce validation errors before attempting clone
- **FR-024**: Duplicate repo names MUST produce validation errors

#### State vs Config
- **FR-025**: Runtime state (clone status, worktree paths, timestamps) MUST be stored in `.foundagent/state.json`
- **FR-026**: User-editable settings (repos, preferences) MUST be stored in `.foundagent.yaml`
- **FR-027**: State file MUST NOT be manually edited by users; system manages it
- **FR-028**: If state and config conflict, config is authoritative

### Key Entities

- **Workspace Configuration**: User-editable YAML/TOML/JSON file at workspace root defining repos and settings. Source of truth for workspace definition.
- **Workspace State**: Machine-managed JSON file in `.foundagent/` tracking runtime state like clone status, worktree locations, sync timestamps.
- **Repository Definition**: Entry in config specifying a repo's URL, optional name override, and optional default branch.
- **Settings**: Workspace-wide preferences like worktree directory location and auto-creation behavior.

### Config Schema Reference

```yaml
# .foundagent.yaml
workspace:
  name: my-project

repos:
  - url: git@github.com:org/api.git
    name: api                      # optional, inferred if omitted
    default_branch: main           # optional, detected if omitted
  - url: git@github.com:org/web.git
  - url: git@github.com:org/shared-lib.git
    name: lib

settings:
  auto_create_worktree: true       # create default branch worktree on add
```

### Workspace Directory Structure

```
my-project/                           # workspace root
├── .foundagent.yaml                  # user-editable config (source of truth)
├── .foundagent/                      # machine-managed state
│   └── state.json
├── my-project.code-workspace         # VS Code workspace file
└── repos/
    ├── .bare/                        # bare clones (hidden)
    │   ├── api.git/
    │   ├── web.git/
    │   └── lib.git/
    └── worktrees/                    # working directories
        ├── api/
        │   ├── main/
        │   └── feature-x/
        ├── web/
        │   └── main/
        └── lib/
            └── main/
```

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Config file parses in under 100ms for workspaces with up to 100 repos
- **SC-002**: 100% of validation errors include actionable remediation information
- **SC-003**: Config updates from commands preserve 100% of user comments and formatting
- **SC-004**: Team members using the same config file get identical workspace structures
- **SC-005**: Users can define and clone a 10-repo workspace via config in under 2 minutes
