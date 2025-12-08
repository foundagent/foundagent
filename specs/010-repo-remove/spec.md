# Feature Specification: Repo Remove

**Feature Branch**: `010-repo-remove`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Remove a repository from workspace with fa remove command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Repo from Workspace (Priority: P1)

A developer no longer needs a repo in their workspace (project restructured, repo deprecated, etc.). They run `fa remove api` to remove the `api` repo. The system removes the repo from config, deletes the bare clone, removes all worktrees for that repo, and updates the workspace file.

**Why this priority**: This is the core functionality — the inverse of `fa add`. Users need a clean way to remove repos.

**Independent Test**: Add a repo with `fa add`, run `fa remove <name>`, verify repo is removed from config, bare clone deleted, worktrees deleted, and workspace file updated.

**Acceptance Scenarios**:

1. **Given** repo `api` exists in workspace, **When** I run `fa remove api`, **Then** the repo is removed from `.foundagent.yaml`
2. **Given** repo is removed, **When** I check the filesystem, **Then** `repos/.bare/api.git/` no longer exists
3. **Given** repo had worktrees, **When** removal completes, **Then** all worktrees under `repos/worktrees/api/` are deleted
4. **Given** repo worktrees were in `.code-workspace`, **When** removal completes, **Then** those folders are removed from workspace file

---

### User Story 2 - Prevent Removal with Uncommitted Changes (Priority: P1)

A developer tries to remove a repo that has uncommitted changes in one or more worktrees. The system warns them and blocks removal unless they use `--force`.

**Why this priority**: Non-destructive by default. Preventing accidental data loss is critical.

**Independent Test**: Make uncommitted changes in a worktree, run `fa remove <repo>`, verify command fails with list of dirty worktrees.

**Acceptance Scenarios**:

1. **Given** repo `api` has uncommitted changes in `repos/worktrees/api/feature-x/`, **When** I run `fa remove api`, **Then** command fails with error listing dirty worktrees
2. **Given** removal is blocked, **When** I see the error, **Then** it includes hint: "Use `--force` to remove anyway, or commit/stash changes first"
3. **Given** dirty worktrees exist, **When** I run `fa remove api --force`, **Then** repo and all worktrees are removed including dirty ones

---

### User Story 3 - Remove Only from Config (Priority: P2)

A developer wants to remove a repo from config without deleting the cloned data. They run `fa remove api --config-only` to remove from config but keep the bare clone and worktrees.

**Why this priority**: Useful for temporarily removing a repo or when user wants to manage files manually.

**Independent Test**: Run `fa remove api --config-only`, verify repo removed from config but files remain.

**Acceptance Scenarios**:

1. **Given** repo `api` exists, **When** I run `fa remove api --config-only`, **Then** repo is removed from `.foundagent.yaml`
2. **Given** `--config-only` is used, **When** removal completes, **Then** `repos/.bare/api.git/` still exists
3. **Given** `--config-only` is used, **When** removal completes, **Then** worktrees under `repos/worktrees/api/` still exist
4. **Given** `--config-only` is used, **When** I run `fa status`, **Then** repo shows as "cloned but not in config"

---

### User Story 4 - JSON Output for Automation (Priority: P3)

An AI agent or script needs structured output from remove operations.

**Why this priority**: Agent-friendly design principle. Important for integration but human output is primary.

**Independent Test**: Run `fa remove api --json`, parse JSON output, verify it contains removal status.

**Acceptance Scenarios**:

1. **Given** I run `fa remove api --json`, **When** removal completes, **Then** output is valid JSON with removal status
2. **Given** removal fails, **When** using `--json`, **Then** JSON includes error details

---

### User Story 5 - Remove Multiple Repos (Priority: P3)

A developer wants to remove several repos at once. They run `fa remove api web` to remove both repos in one command.

**Why this priority**: Convenience for bulk operations. Single repo removal comes first.

**Independent Test**: Run `fa remove api web`, verify both repos are removed.

**Acceptance Scenarios**:

1. **Given** repos `api` and `web` exist, **When** I run `fa remove api web`, **Then** both repos are removed
2. **Given** one repo has uncommitted changes, **When** removing multiple, **Then** clean repos are removed and dirty one fails (unless `--force`)

---

### Edge Cases

- **Repo doesn't exist**: `fa remove nonexistent` — error with message "Repo 'nonexistent' not found in workspace"
- **Already removed from config**: Repo is cloned but not in config — still remove the files (with confirmation)
- **Typo protection**: Suggest similar repo names if no exact match
- **Last repo**: Removing the only repo — allowed, workspace becomes empty
- **Currently in worktree**: CWD is inside a worktree being removed — error with hint to change directory
- **Worktree locked**: Git has worktree locked — surface git error clearly
- **Permission denied**: Can't delete files — error with clear message
- **Not in workspace**: Run outside Foundagent workspace — error with hint

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa remove <name>` command
- **FR-002**: System MUST support `fa rm <name>` as alias
- **FR-003**: System MUST accept multiple repo names: `fa remove <name1> <name2> ...`
- **FR-004**: System MUST support `--force` flag to remove despite uncommitted changes
- **FR-005**: System MUST support `--config-only` flag to remove from config without deleting files
- **FR-006**: System MUST support `--json` flag for machine-readable output

#### Safety Checks
- **FR-007**: System MUST check all worktrees for uncommitted changes before removal
- **FR-008**: System MUST block removal if any worktree is dirty (unless `--force`)
- **FR-009**: System MUST list all dirty worktrees in the error message
- **FR-010**: System MUST check if CWD is inside a worktree being removed
- **FR-011**: System MUST block removal if CWD is inside target worktree (even with `--force`)
- **FR-012**: System MUST verify repo exists before attempting removal

#### Removal Process
- **FR-013**: System MUST remove repo from `.foundagent.yaml` config
- **FR-014**: System MUST delete bare clone at `repos/.bare/<name>.git/`
- **FR-015**: System MUST delete all worktrees at `repos/worktrees/<name>/`
- **FR-016**: System MUST use `git worktree remove` for proper worktree cleanup
- **FR-017**: System MUST update `.code-workspace` to remove worktree folders
- **FR-018**: System MUST update `.foundagent/state.json` to remove repo entries

#### Config-Only Mode (--config-only)
- **FR-019**: With `--config-only`, system MUST only remove from `.foundagent.yaml`
- **FR-020**: With `--config-only`, system MUST NOT delete bare clone or worktrees
- **FR-021**: With `--config-only`, system MUST still update workspace file (remove folders)
- **FR-022**: After `--config-only`, `fa status` MUST show repo as "not in config"

#### Output and Feedback
- **FR-023**: System MUST show confirmation of what will be removed before proceeding
- **FR-024**: System MUST display summary of removed items (config, bare clone, worktrees)
- **FR-025**: JSON output MUST include: repo_name, removed_from_config, files_deleted, worktrees_deleted
- **FR-026**: Error messages MUST include actionable remediation steps
- **FR-027**: System MUST exit with code 0 on success, non-zero on failure

### Key Entities

- **Repository**: A repo to be removed, identified by its local name. Located at `repos/.bare/<name>.git/` with worktrees at `repos/worktrees/<name>/`.
- **Dirty Worktree**: A worktree with uncommitted changes that blocks removal.
- **Removal Scope**: What gets removed — config entry, bare clone, worktrees, or just config (with `--config-only`).

### Assumptions

- Repos identified by local name (as shown in config), not by URL
- Bare clones at `repos/.bare/<name>.git/`
- Worktrees at `repos/worktrees/<name>/<branch>/`
- Removal is permanent (no undo) — hence safety checks

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can remove a repo in under 5 seconds
- **SC-002**: 100% of removal attempts with uncommitted changes are blocked unless `--force`
- **SC-003**: Users never lose uncommitted work accidentally (dirty check catches 100% of cases)
- **SC-004**: Config, bare clone, worktrees, and workspace file are all correctly updated
- **SC-005**: `--config-only` leaves files intact in 100% of cases
