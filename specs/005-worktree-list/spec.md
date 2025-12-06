# Feature Specification: Worktree List

**Feature Branch**: `005-worktree-list`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "List all worktrees in the workspace with fa wt list command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List All Worktrees (Priority: P1)

A developer wants to see what worktrees exist across all repos in their workspace. They run `fa wt list` and see a clear summary of all branches with worktrees, showing which repos have worktrees for each branch.

**Why this priority**: This is the core functionality — developers need visibility into their worktree state before switching, creating, or removing worktrees.

**Independent Test**: Run `fa wt list` in a workspace with 3 repos and 2 branches (main, feature-x), verify output shows all 6 worktrees organized by branch.

**Acceptance Scenarios**:

1. **Given** a workspace with repos (api, web, lib) each having worktrees for `main` and `feature-x`, **When** I run `fa wt list`, **Then** I see a list showing both branches with all repos under each
2. **Given** worktrees exist, **When** I run `fa wt list`, **Then** each worktree shows its path (e.g., `repos/worktrees/api/main/`)
3. **Given** I run `fa wt list`, **When** the output is displayed, **Then** the current/active worktree is visually indicated (e.g., with `*` or highlighting)

---

### User Story 2 - JSON Output for Automation (Priority: P2)

A developer using AI agents or automation scripts needs machine-readable output from the list command. They run `fa wt list --json` to get structured data for programmatic consumption.

**Why this priority**: Supports agent-friendly design principle. Essential for integration with AI tools and scripting, but human-readable output is the primary use case.

**Independent Test**: Run `fa wt list --json`, parse the JSON output, verify it contains all worktree details with consistent structure.

**Acceptance Scenarios**:

1. **Given** I run `fa wt list --json`, **When** the command completes, **Then** the output is valid JSON containing an array of worktrees with branch, repo, and path
2. **Given** I run `fa wt list --json`, **When** worktrees have uncommitted changes, **Then** the JSON includes a `dirty` or `status` field for each worktree

---

### User Story 3 - Show Worktree Status (Priority: P2)

A developer wants to quickly see which worktrees have uncommitted changes before switching or syncing. The list command shows status indicators for dirty worktrees.

**Why this priority**: Status visibility prevents accidental data loss when switching worktrees and helps developers track work in progress.

**Independent Test**: Create uncommitted changes in one worktree, run `fa wt list`, verify that worktree is marked as dirty/modified.

**Acceptance Scenarios**:

1. **Given** a worktree has uncommitted changes, **When** I run `fa wt list`, **Then** that worktree shows a status indicator (e.g., `[modified]` or `*`)
2. **Given** a worktree has untracked files, **When** I run `fa wt list`, **Then** that worktree shows an indicator (e.g., `[untracked]` or `?`)
3. **Given** all worktrees are clean, **When** I run `fa wt list`, **Then** no status indicators are shown (clean is the default/implicit state)

---

### User Story 4 - Filter by Branch (Priority: P3)

A developer wants to see worktrees for a specific branch only. They run `fa wt list feature-x` to filter the output.

**Why this priority**: Convenience feature for large workspaces with many branches. Core listing works without filtering.

**Independent Test**: Run `fa wt list feature-x` in a workspace with multiple branches, verify only `feature-x` worktrees are shown.

**Acceptance Scenarios**:

1. **Given** worktrees exist for `main` and `feature-x`, **When** I run `fa wt list feature-x`, **Then** only worktrees for `feature-x` are shown
2. **Given** I filter by a branch that doesn't exist, **When** I run `fa wt list nonexistent`, **Then** I see an empty result with a message "No worktrees found for branch 'nonexistent'"

---

### Edge Cases

- **Empty workspace**: No repos added — show message "No repositories in workspace. Run `fa add` to add repos."
- **Repos but no worktrees**: Repos exist but no worktrees created — show message "No worktrees found. Run `fa wt create <branch>` to create worktrees."
- **Partial worktrees**: Branch exists in some repos but not all — show which repos have the worktree and which don't
- **Corrupted worktree**: Git worktree in bad state — show error indicator for that worktree, don't fail entire command
- **Very long branch names**: Display should truncate or wrap gracefully
- **Many branches**: Dozens of branches — consider pagination or suggest `--json` for full output
- **Not in workspace**: Run outside a Foundagent workspace — error with hint to run `fa init` or navigate to workspace

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa worktree list` command
- **FR-002**: System MUST support `fa wt list` as alias
- **FR-003**: System MUST support `fa wt ls` as short alias
- **FR-004**: System MUST accept optional `[branch]` argument to filter by branch name
- **FR-005**: System MUST support `--json` flag for machine-readable output

#### Output Format (Human-Readable)
- **FR-006**: System MUST display worktrees grouped by branch name
- **FR-007**: System MUST show repo name and path for each worktree
- **FR-008**: System MUST indicate the current/active worktree (based on current working directory)
- **FR-009**: System MUST show status indicators for dirty worktrees (uncommitted changes)
- **FR-010**: System MUST show status indicators for worktrees with untracked files
- **FR-011**: System MUST display worktrees in alphabetical order by branch, then by repo

#### Output Format (JSON)
- **FR-012**: JSON output MUST include array of worktree objects
- **FR-013**: Each worktree object MUST include: `branch`, `repo`, `path`, `is_current`, `status`
- **FR-014**: Status field MUST include: `clean`, `modified`, `untracked`, or `conflict`
- **FR-015**: JSON MUST include workspace-level metadata: `workspace_name`, `total_worktrees`, `total_branches`

#### Status Detection
- **FR-016**: System MUST detect uncommitted changes in worktrees
- **FR-017**: System MUST detect untracked files in worktrees
- **FR-018**: System MUST detect merge conflicts in worktrees
- **FR-019**: Status detection MUST run in parallel across repos for performance

#### Error Handling
- **FR-020**: System MUST validate that command is run inside a Foundagent workspace
- **FR-021**: System MUST handle corrupted worktrees gracefully (show error, continue with others)
- **FR-022**: System MUST show helpful message when no repos exist
- **FR-023**: System MUST show helpful message when no worktrees exist
- **FR-024**: System MUST exit with code 0 on success, non-zero on failure

### Key Entities

- **Worktree**: A working directory at `repos/worktrees/<repo>/<branch>/`. Has associated repo, branch, filesystem path, and status (clean/modified/untracked/conflict).
- **Branch Group**: A logical grouping of worktrees that share the same branch name across repos. Used for display organization.
- **Worktree Status**: The git working tree status — clean (no changes), modified (uncommitted changes), untracked (new files), or conflict (merge conflicts).

### Assumptions

- Worktrees exist at `repos/worktrees/<repo>/<branch>/` following the canonical structure
- Git status can be determined efficiently via git commands
- The current directory determines which worktree (if any) is "active"

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can view all worktrees in under 2 seconds for workspaces with up to 50 worktrees
- **SC-002**: Status detection completes in under 5 seconds for workspaces with up to 50 worktrees
- **SC-003**: 100% of worktree states (clean, modified, untracked, conflict) are correctly detected
- **SC-004**: JSON output parses successfully with standard JSON parsers
- **SC-005**: Users can identify dirty worktrees at a glance without running separate git commands
