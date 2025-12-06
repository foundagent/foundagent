# Feature Specification: Worktree Remove

**Feature Branch**: `006-worktree-remove`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Remove worktrees across all repos with fa wt remove command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Worktree Across All Repos (Priority: P1)

A developer has finished working on a feature and wants to clean up the worktrees for that branch. They run `fa wt remove feature-123` and Foundagent removes the worktree from every repo in the workspace, updating the VS Code workspace file to remove those folders.

**Why this priority**: This is the core functionality — removing worktrees atomically across all repos is essential for workspace hygiene and follows the same all-or-nothing pattern as `fa wt create`.

**Independent Test**: Create worktrees for `feature-123` across 3 repos, run `fa wt remove feature-123`, verify all worktrees are removed and `.code-workspace` no longer includes them.

**Acceptance Scenarios**:

1. **Given** worktrees exist for `feature-123` in repos (api, web, lib), **When** I run `fa wt remove feature-123`, **Then** all 3 worktrees are removed
2. **Given** worktrees are removed, **When** I check the filesystem, **Then** directories `repos/worktrees/api/feature-123/`, `repos/worktrees/web/feature-123/`, and `repos/worktrees/lib/feature-123/` no longer exist
3. **Given** worktrees are removed, **When** I check the `.code-workspace` file, **Then** those folders are no longer listed

---

### User Story 2 - Prevent Removal with Uncommitted Changes (Priority: P1)

A developer accidentally tries to remove worktrees that have uncommitted changes. The system warns them and prevents data loss unless they explicitly confirm with `--force`.

**Why this priority**: Non-destructive by default is a core principle. Preventing accidental data loss is critical for user trust.

**Independent Test**: Make uncommitted changes in one worktree, run `fa wt remove <branch>`, verify command fails with list of dirty worktrees and hint to use `--force`.

**Acceptance Scenarios**:

1. **Given** worktree `api/feature-123` has uncommitted changes, **When** I run `fa wt remove feature-123`, **Then** the command fails with error listing the dirty worktree
2. **Given** removal fails due to dirty worktrees, **When** I see the error, **Then** it includes a hint: "Use `--force` to remove anyway, or commit/stash changes first"
3. **Given** dirty worktrees exist, **When** I run `fa wt remove feature-123 --force`, **Then** all worktrees are removed including dirty ones

---

### User Story 3 - Delete Branch After Removal (Priority: P2)

A developer wants to remove worktrees AND delete the underlying branches (cleanup after merge). They run `fa wt remove feature-123 --delete-branch` to remove worktrees and delete the branch from all repos.

**Why this priority**: Common workflow after merging a PR. Convenience feature that combines two operations but not required for basic cleanup.

**Independent Test**: Create worktrees for `feature-123`, run `fa wt remove feature-123 --delete-branch`, verify worktrees removed AND branches deleted from all repos.

**Acceptance Scenarios**:

1. **Given** worktrees exist for `feature-123`, **When** I run `fa wt remove feature-123 --delete-branch`, **Then** worktrees are removed AND branches `feature-123` are deleted from all repos
2. **Given** branch `feature-123` is not fully merged, **When** I run `fa wt remove feature-123 --delete-branch`, **Then** the command warns about unmerged changes and requires `--force`
3. **Given** `--delete-branch` is used, **When** deletion completes, **Then** output confirms both worktree removal and branch deletion

---

### User Story 4 - JSON Output for Automation (Priority: P3)

A developer using AI agents or automation scripts needs machine-readable output from the remove command.

**Why this priority**: Supports agent-friendly design principle. Important for integration but human output is primary.

**Independent Test**: Run `fa wt remove feature-123 --json`, parse JSON output, verify it contains removal status for each repo.

**Acceptance Scenarios**:

1. **Given** I run `fa wt remove feature-123 --json`, **When** the command completes, **Then** output is valid JSON with removal status for each worktree
2. **Given** some worktrees have issues, **When** using `--json`, **Then** each worktree has individual success/failure status

---

### Edge Cases

- **Branch doesn't exist**: No worktrees for that branch — error with message "No worktrees found for branch 'feature-x'"
- **Partial worktrees**: Branch exists in some repos but not all — remove what exists, report what was skipped
- **Currently in worktree**: User's CWD is inside a worktree being removed — error with hint to change directory first
- **Default branch**: Attempt to remove `main` or default branch — warn that this will leave repo without worktree, require `--force`
- **Locked worktree**: Git has the worktree locked — surface git error clearly
- **Already removed**: Worktree directory doesn't exist but git thinks it does — clean up git worktree reference
- **Permission denied**: Can't delete directory — error with clear message
- **Not in workspace**: Run outside a Foundagent workspace — error with hint

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa worktree remove <branch>` command
- **FR-002**: System MUST support `fa wt remove <branch>` as alias
- **FR-003**: System MUST support `fa wt rm <branch>` as short alias
- **FR-004**: System MUST require branch name argument (not optional)
- **FR-005**: System MUST support `--force` flag to remove despite uncommitted changes
- **FR-006**: System MUST support `--delete-branch` flag to delete branches after worktree removal
- **FR-007**: System MUST support `--json` flag for machine-readable output

#### Safety Checks
- **FR-008**: System MUST check all worktrees for uncommitted changes before removal
- **FR-009**: System MUST check all worktrees for untracked files before removal
- **FR-010**: System MUST block removal if any worktree is dirty (unless `--force`)
- **FR-011**: System MUST list all dirty worktrees in the error message
- **FR-012**: System MUST check if current working directory is inside a worktree being removed
- **FR-013**: System MUST block removal if CWD is inside target worktree (even with `--force`)
- **FR-014**: System MUST warn before removing default branch worktrees (require `--force`)

#### Removal Process
- **FR-015**: System MUST remove worktrees from ALL repos that have them
- **FR-016**: System MUST use `git worktree remove` for proper cleanup
- **FR-017**: System MUST delete the worktree directory from filesystem
- **FR-018**: System MUST handle partial existence (branch in some repos but not all)
- **FR-019**: System MUST update `.code-workspace` file to remove worktree folders
- **FR-020**: System MUST update `.foundagent/state.json` to remove worktree entries

#### Branch Deletion (--delete-branch)
- **FR-021**: With `--delete-branch`, system MUST delete branch from all repos after worktree removal
- **FR-022**: System MUST check if branch is merged before deletion
- **FR-023**: System MUST warn about unmerged branches and require `--force` to delete
- **FR-024**: System MUST NOT delete branches if worktree removal failed

#### Output and Feedback
- **FR-025**: System MUST show progress for each repo during removal
- **FR-026**: System MUST display summary of removed worktrees with paths
- **FR-027**: JSON output MUST include per-worktree status (removed, skipped, failed)
- **FR-028**: Error messages MUST include actionable remediation steps
- **FR-029**: System MUST exit with code 0 on success, non-zero on failure

### Key Entities

- **Worktree**: A working directory at `repos/worktrees/<repo>/<branch>/` to be removed.
- **Dirty Worktree**: A worktree with uncommitted changes or untracked files that blocks removal.
- **Branch**: The git branch associated with worktrees; optionally deleted with `--delete-branch`.

### Assumptions

- Worktrees exist at `repos/worktrees/<repo>/<branch>/` following canonical structure
- Bare clones at `repos/.bare/<repo>.git/` are NOT affected by worktree removal
- Branch deletion only affects local branches, not remote branches

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can remove worktrees across a 5-repo workspace in under 5 seconds
- **SC-002**: 100% of removal attempts with uncommitted changes are blocked unless `--force` is used
- **SC-003**: Users never lose uncommitted work accidentally (dirty worktree check catches 100% of cases)
- **SC-004**: VS Code workspace file is correctly updated in 100% of removal operations
- **SC-005**: JSON output parses successfully with standard JSON parsers
