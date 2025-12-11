# Feature Specification: Worktree Switch

**Feature Branch**: `009-worktree-switch`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Switch to a different branch worktree with fa wt switch command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Switch VS Code Workspace to Different Branch (Priority: P1)

A developer is working on `main` and needs to switch to `feature-123` to review or continue work. They run `fa wt switch feature-123` and the VS Code workspace file is updated to point to the `feature-123` worktrees instead of `main`. When they reload VS Code, they see all repos checked out to `feature-123`.

**Why this priority**: This is the core value proposition — seamless context switching between branches across all repos with a single command.

**Independent Test**: With worktrees for `main` and `feature-123`, run `fa wt switch feature-123`, verify `.code-workspace` file now points to `feature-123` worktrees for all repos.

**Acceptance Scenarios**:

1. **Given** worktrees exist for `main` and `feature-123`, **When** I run `fa wt switch feature-123`, **Then** the `.code-workspace` file is updated to include `repos/api/worktrees/feature-123/`, `repos/web/worktrees/feature-123/`, `repos/lib/worktrees/feature-123/`
2. **Given** I switch to `feature-123`, **When** I check the `.code-workspace` file, **Then** the `main` worktree folders are no longer listed
3. **Given** switch completes, **When** I reload VS Code, **Then** I see all repos at the `feature-123` branch

---

### User Story 2 - Warn About Uncommitted Changes (Priority: P1)

A developer tries to switch branches but has uncommitted changes in the current worktrees. The system warns them so they don't lose track of their work.

**Why this priority**: Non-destructive by default. Users need to know about uncommitted work before switching context.

**Independent Test**: Make uncommitted changes in one worktree, run `fa wt switch <branch>`, verify warning is shown but switch still proceeds (it's just updating the workspace file, not touching files).

**Acceptance Scenarios**:

1. **Given** current worktree has uncommitted changes, **When** I run `fa wt switch feature-123`, **Then** I see a warning listing dirty worktrees
2. **Given** warning is shown, **When** switch completes, **Then** the switch still happens (workspace file updated) — changes remain in the original worktree
3. **Given** I use `--quiet` flag, **When** there are uncommitted changes, **Then** warning is suppressed

---

### User Story 3 - Create Worktree If Doesn't Exist (Priority: P2)

A developer wants to switch to a branch that doesn't have worktrees yet. With `--create` flag, the system creates the worktrees first, then switches to them.

**Why this priority**: Convenience for common workflow of "switch to this branch, create it if needed." Basic switch comes first.

**Independent Test**: Run `fa wt switch new-feature --create` when no worktrees exist for `new-feature`, verify worktrees are created and workspace is switched.

**Acceptance Scenarios**:

1. **Given** no worktrees exist for `new-feature`, **When** I run `fa wt switch new-feature --create`, **Then** worktrees are created for all repos and workspace is switched
2. **Given** no worktrees exist and `--create` is NOT used, **When** I run `fa wt switch new-feature`, **Then** error with hint: "No worktrees found for 'new-feature'. Use `--create` to create them."
3. **Given** `--create` is used with `--from release-1.0`, **When** switch runs, **Then** new worktrees are based on `release-1.0` branch

---

### User Story 4 - JSON Output for Automation (Priority: P3)

An AI agent or script needs structured output from switch operations.

**Why this priority**: Agent-friendly design principle. Important for integration but human output is primary.

**Independent Test**: Run `fa wt switch feature-123 --json`, parse JSON output, verify it contains switch status and workspace file path.

**Acceptance Scenarios**:

1. **Given** I run `fa wt switch feature-123 --json`, **When** switch completes, **Then** output is valid JSON with switched branch and workspace file path
2. **Given** switch fails, **When** using `--json`, **Then** JSON includes error details and any warnings

---

### User Story 5 - List Available Branches to Switch To (Priority: P3)

A developer isn't sure which branches have worktrees. Running `fa wt switch` with no arguments shows available options.

**Why this priority**: Discoverability feature. Helps users explore available worktrees.

**Independent Test**: Run `fa wt switch` with no arguments, verify list of available branches is shown.

**Acceptance Scenarios**:

1. **Given** worktrees exist for `main`, `feature-x`, `bugfix-y`, **When** I run `fa wt switch` (no args), **Then** I see a list of branches I can switch to
2. **Given** list is shown, **When** current branch is `main`, **Then** `main` is marked as current

---

### Edge Cases

- **Already on target branch**: `fa wt switch main` when already on `main` — message "Already on 'main'" (no-op, exit 0)
- **Branch doesn't exist**: No worktrees for branch — error with hint to use `--create`
- **Partial worktrees**: Branch exists in some repos but not all — switch to what exists, warn about missing repos
- **Invalid branch name**: Branch name with invalid characters — error before attempting anything
- **No worktrees at all**: Workspace has no worktrees — error with hint to run `fa wt create`
- **Not in workspace**: Run outside Foundagent workspace — error with hint
- **Workspace file missing**: `.code-workspace` file was deleted — recreate it during switch
- **Concurrent modification**: Another process modifying workspace file — handle gracefully with lock

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa worktree switch <branch>` command
- **FR-002**: System MUST support `fa wt switch <branch>` as alias
- **FR-003**: System MUST support running with no arguments to list available branches
- **FR-004**: System MUST support `--create` flag to create worktrees if they don't exist
- **FR-005**: System MUST support `--from <source-branch>` flag (only with `--create`)
- **FR-006**: System MUST support `--quiet` flag to suppress warnings
- **FR-007**: System MUST support `--json` flag for machine-readable output

#### Workspace File Update
- **FR-008**: System MUST update `.code-workspace` file to include target branch worktrees
- **FR-009**: System MUST remove current branch worktrees from `.code-workspace` file
- **FR-010**: Updated workspace file MUST include all repos' worktrees for target branch
- **FR-011**: Workspace file MUST preserve non-worktree folders (e.g., workspace root)
- **FR-012**: Workspace file MUST preserve workspace settings
- **FR-013**: If workspace file doesn't exist, system MUST create it

#### Validation
- **FR-014**: System MUST verify target branch worktrees exist (unless `--create`)
- **FR-015**: If worktrees don't exist, system MUST error with hint to use `--create`
- **FR-016**: System MUST detect and warn about uncommitted changes in current worktrees
- **FR-017**: System MUST handle partial worktrees (branch in some repos but not all)
- **FR-018**: System MUST detect if already on target branch (no-op with message)

#### Create Integration (--create)
- **FR-019**: With `--create`, system MUST invoke worktree creation logic (FR from 004-worktree-create)
- **FR-020**: With `--create --from`, system MUST create worktrees from specified source branch
- **FR-021**: With `--create`, system MUST validate source branch exists in all repos before creating
- **FR-022**: If creation fails, system MUST NOT update workspace file

#### Output and Feedback
- **FR-023**: System MUST display confirmation of successful switch
- **FR-024**: System MUST display path to updated workspace file
- **FR-025**: With no arguments, system MUST list available branches with worktrees
- **FR-026**: Branch list MUST indicate current branch
- **FR-027**: JSON output MUST include: switched_to, previous_branch, workspace_file, warnings

#### State Management
- **FR-028**: System MUST update `.foundagent/state.json` to track current branch
- **FR-029**: System MUST NOT modify any worktree files (only workspace file and state)

### Key Entities

- **Current Branch**: The branch whose worktrees are currently in the VS Code workspace file.
- **Target Branch**: The branch to switch to — its worktrees will be added to workspace file.
- **Workspace File**: The `.code-workspace` JSON file updated to reflect active worktrees.
- **Available Branches**: Branches that have worktrees and can be switched to.

### Assumptions

- Worktrees exist at `repos/<repo>/worktrees/<branch>/` following canonical structure
- Switching only updates the `.code-workspace` file — actual git worktrees remain unchanged
- Users reload VS Code after switch to see the new folders (or VS Code auto-detects)
- Current branch determined from workspace file contents or `.foundagent/state.json`

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can switch between branches in under 1 second (only file I/O)
- **SC-002**: 100% of switch operations correctly update the workspace file
- **SC-003**: Users never lose uncommitted changes (switching doesn't touch worktree files)
- **SC-004**: VS Code correctly loads the new worktrees after reload in 100% of cases
- **SC-005**: `--create` flag successfully creates and switches in a single command
