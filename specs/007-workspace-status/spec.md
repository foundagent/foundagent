# Feature Specification: Workspace Status

**Feature Branch**: `007-workspace-status`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Show workspace status with fa status command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Workspace Overview (Priority: P1)

A developer opens their workspace and wants a quick overview of the current state: which repos are configured, which worktrees exist, and whether there's any uncommitted work. They run `fa status` and see a summary dashboard of the entire workspace.

**Why this priority**: This is the core functionality — developers need a single command to understand their workspace state, especially when returning to work after time away or when AI agents need to understand context.

**Independent Test**: Run `fa status` in a workspace with 3 repos and 2 branches, verify output shows all repos, all worktrees, and highlights any dirty worktrees.

**Acceptance Scenarios**:

1. **Given** a workspace with repos (api, web, lib), **When** I run `fa status`, **Then** I see a list of all configured repos
2. **Given** worktrees exist for `main` and `feature-x`, **When** I run `fa status`, **Then** I see all worktrees grouped by branch
3. **Given** I'm currently in a worktree, **When** I run `fa status`, **Then** the current worktree is highlighted/indicated
4. **Given** some worktrees have uncommitted changes, **When** I run `fa status`, **Then** dirty worktrees are marked with status indicators

---

### User Story 2 - Check for Uncommitted Work (Priority: P1)

A developer wants to quickly see if they have any uncommitted changes across all their worktrees before switching tasks or ending their day. The status command shows a clear summary of dirty worktrees.

**Why this priority**: Visibility into uncommitted work prevents data loss and helps developers track work in progress across multiple repos.

**Independent Test**: Make changes in 2 of 6 worktrees, run `fa status`, verify both dirty worktrees are clearly indicated.

**Acceptance Scenarios**:

1. **Given** worktree `repos/api/worktrees/feature-x/` has uncommitted changes, **When** I run `fa status`, **Then** that worktree shows `[modified]` or similar indicator
2. **Given** worktree `repos/web/worktrees/main/` has untracked files, **When** I run `fa status`, **Then** that worktree shows `[untracked]` or similar indicator
3. **Given** all worktrees are clean, **When** I run `fa status`, **Then** I see "All worktrees clean" or similar message

---

### User Story 3 - JSON Output for AI Agents (Priority: P1)

An AI coding agent needs to understand the complete workspace state to make intelligent suggestions. It runs `fa status --json` to get structured data about repos, worktrees, and their states.

**Why this priority**: Agent-friendly output is a core principle. AI agents need complete context to provide useful assistance.

**Independent Test**: Run `fa status --json`, parse the JSON, verify it contains complete workspace structure with all repos, worktrees, and status.

**Acceptance Scenarios**:

1. **Given** I run `fa status --json`, **When** the command completes, **Then** output is valid JSON with workspace metadata
2. **Given** I run `fa status --json`, **When** I examine the output, **Then** it includes arrays for repos, worktrees, and status for each
3. **Given** worktrees have various states, **When** I run `fa status --json`, **Then** each worktree has `status` field (clean/modified/untracked/conflict)

---

### User Story 4 - Config vs State Sync Check (Priority: P2)

A developer wants to verify their workspace is in sync — that all repos in config are cloned and all cloned repos are in config. The status command highlights any discrepancies.

**Why this priority**: Config is source of truth; knowing when state drifts from config helps maintain consistency.

**Independent Test**: Manually add a repo to config without cloning, run `fa status`, verify it shows "not cloned" for that repo.

**Acceptance Scenarios**:

1. **Given** a repo is in config but not cloned, **When** I run `fa status`, **Then** that repo shows `[not cloned]` with hint to run `fa add`
2. **Given** a repo is cloned but not in config, **When** I run `fa status`, **Then** that repo shows `[not in config]` with hint to update config or run `fa remove`
3. **Given** config and state are in sync, **When** I run `fa status`, **Then** I see "Config in sync" or similar confirmation

---

### User Story 5 - Verbose Status Details (Priority: P3)

A developer wants more details about their workspace, including file counts and specific changed files. They run `fa status -v` for verbose output.

**Why this priority**: Convenience for power users who want details without running separate git commands. Basic status comes first.

**Independent Test**: Make changes in a worktree, run `fa status -v`, verify output includes specific file names and change types.

**Acceptance Scenarios**:

1. **Given** a worktree has modified files, **When** I run `fa status -v`, **Then** I see the list of modified file names
2. **Given** verbose mode, **When** viewing worktree status, **Then** I see counts (e.g., "3 modified, 1 untracked")

---

### Edge Cases

- **Empty workspace**: No repos configured — show "No repos configured. Run `fa add <url>` to add a repo."
- **Repos but no worktrees**: Repos cloned but no worktrees created — show repos with `[no worktrees]` hint
- **Not in workspace**: Run outside Foundagent workspace — error with hint to `fa init` or navigate to workspace
- **Corrupted worktree**: Git worktree in bad state — show error indicator, continue with other worktrees
- **Very large workspace**: Many repos/worktrees — output remains readable, consider summary + details format
- **Network status**: Don't check remotes by default (keep it fast/offline); future feature for remote sync status
- **Permission issues**: Can't read some directories — show error for those, continue with others

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa status` command
- **FR-002**: System MUST support `fa st` as short alias
- **FR-003**: System MUST support `-v` / `--verbose` flag for detailed output
- **FR-004**: System MUST support `--json` flag for machine-readable output

#### Workspace Overview
- **FR-005**: System MUST display workspace name from config
- **FR-006**: System MUST display total count of configured repos
- **FR-007**: System MUST display total count of worktrees
- **FR-008**: System MUST display total count of branches with worktrees
- **FR-009**: System MUST indicate current worktree (based on CWD)

#### Repo Status
- **FR-010**: System MUST list all repos from config
- **FR-011**: System MUST show clone status for each repo (bare clone exists in `repos/<name>/.bare/`)
- **FR-012**: System MUST show repos that are cloned but not in config
- **FR-013**: System MUST show repo URL and local name

#### Worktree Status
- **FR-014**: System MUST list all worktrees grouped by branch
- **FR-015**: System MUST show path for each worktree (e.g., `repos/api/worktrees/main/`)
- **FR-016**: System MUST show status for each worktree (clean/modified/untracked/conflict)
- **FR-017**: System MUST highlight the current worktree
- **FR-018**: Status detection MUST run in parallel for performance

#### Verbose Mode
- **FR-019**: Verbose mode MUST show list of changed files per dirty worktree
- **FR-020**: Verbose mode MUST show file counts (modified, added, deleted, untracked)
- **FR-021**: Verbose mode MUST show branch tracking info (ahead/behind) if available

#### JSON Output
- **FR-022**: JSON MUST include `workspace` object with name and root path
- **FR-023**: JSON MUST include `repos` array with name, url, clone_status, in_config
- **FR-024**: JSON MUST include `worktrees` array with branch, repo, path (e.g., `repos/api/worktrees/main/`), status, is_current
- **FR-025**: JSON MUST include `summary` object with counts and sync_status
- **FR-026**: When dirty worktrees exist, JSON `summary.has_uncommitted_changes` MUST be true

#### Error Handling
- **FR-027**: System MUST validate command is run inside a Foundagent workspace
- **FR-028**: System MUST handle corrupted worktrees gracefully (show error, continue)
- **FR-029**: System MUST show helpful message for empty workspace
- **FR-030**: System MUST exit with code 0 on success, non-zero on failure

### Key Entities

- **Workspace Summary**: Overview of workspace including name, repo count, worktree count, sync status between config and state.
- **Repo Status**: A configured or cloned repo with its clone status (cloned/not cloned) and config status (in config/not in config).
- **Worktree Status**: A worktree's current state including branch, repo, path, git status (clean/modified/untracked/conflict), and whether it's the current working directory.
- **Sync Status**: Whether config and state are in sync — all repos in config are cloned, all cloned repos are in config.

### Assumptions

- Status is a read-only operation; no changes to filesystem or config
- Status checks are local only (no network/remote checks) for speed
- Current worktree detected by checking if CWD is inside `repos/<repo>/worktrees/<branch>/`
- Bare clones located at `repos/<repo>/.bare/`
- Parallel status detection for performance in large workspaces

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can view complete workspace status in under 3 seconds for workspaces with up to 50 worktrees
- **SC-002**: Users can identify all uncommitted work at a glance without running separate git commands
- **SC-003**: 100% of dirty worktrees are correctly identified and displayed
- **SC-004**: JSON output contains complete workspace state for AI agent consumption
- **SC-005**: Config/state sync issues are detected and clearly reported in 100% of cases
