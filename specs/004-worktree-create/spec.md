# Feature Specification: Worktree Create

**Feature Branch**: `004-worktree-create`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Create worktrees across all repos in workspace with fa wt create command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Worktree Across All Repos (Priority: P1)

A developer wants to start working on a new feature that spans multiple repositories. They run `fa wt create feature-123` and Foundagent creates a new worktree based on the default branch in every repository in their workspace simultaneously. The VS Code workspace file is automatically updated to include the new worktree directories.

**Why this priority**: This is the core value proposition — creating worktrees atomically across all repos is the primary reason users choose Foundagent over manual git worktree commands.

**Independent Test**: Run `fa wt create feature-123` in a workspace with 3 repos, verify worktrees are created in all 3 repos, new branch `feature-123` exists in each, and `.code-workspace` file includes the new directories.

**Acceptance Scenarios**:

1. **Given** a workspace with 3 repos (api, web, lib), **When** I run `fa wt create feature-123`, **Then** worktrees are created at `repos/worktrees/api/feature-123/`, `repos/worktrees/web/feature-123/`, and `repos/worktrees/lib/feature-123/`
2. **Given** the worktree creation succeeds, **When** I check the `.code-workspace` file, **Then** all 3 new worktree directories are included as folders
3. **Given** all repos have `main` as default branch, **When** I run `fa wt create feature-123`, **Then** each worktree is based on that repo's `main` branch
4. **Given** repos have different default branches (main, master, develop), **When** I run `fa wt create feature-123`, **Then** each worktree is based on its repo's own default branch

---

### User Story 2 - Create Worktree From Specific Branch (Priority: P1)

A developer needs to create a feature branch based on an existing branch other than the default (e.g., a release branch or another feature branch). They run `fa wt create feature-123 --from release-2.0` and Foundagent creates worktrees based on `release-2.0` in all repos.

**Why this priority**: Branching from non-default branches is a common workflow for hotfixes, stacked features, and release branches.

**Independent Test**: Run `fa wt create hotfix-1 --from release-2.0` where `release-2.0` exists in all repos, verify all worktrees are based on `release-2.0`.

**Acceptance Scenarios**:

1. **Given** branch `release-2.0` exists in all repos, **When** I run `fa wt create hotfix-1 --from release-2.0`, **Then** all worktrees are created from `release-2.0`
2. **Given** the `--from` branch exists in all repos, **When** worktrees are created, **Then** the new branch history starts from the specified source branch
3. **Given** branch `release-2.0` does NOT exist in one repo, **When** I run `fa wt create hotfix-1 --from release-2.0`, **Then** the command fails with clear error before any worktrees are created

---

### User Story 3 - Atomic Operation with Pre-validation (Priority: P1)

A developer runs a worktree create command with a `--from` branch that doesn't exist in all repos. Foundagent validates that the source branch exists in ALL repos before starting any worktree creation. This prevents partial states where some repos have the worktree and others don't.

**Why this priority**: Atomic all-or-nothing behavior is essential for multi-repo consistency and prevents messy cleanup scenarios.

**Independent Test**: Try `fa wt create feature-1 --from nonexistent` where `nonexistent` only exists in 2 of 3 repos, verify command fails immediately with list of repos missing the branch.

**Acceptance Scenarios**:

1. **Given** `--from develop` is specified and `develop` doesn't exist in repo `lib`, **When** I run the command, **Then** error states "Branch 'develop' not found in: lib" before any changes
2. **Given** validation fails, **When** I check the repos, **Then** no worktrees were created in any repo
3. **Given** validation passes, **When** worktree creation fails mid-way for one repo, **Then** completed worktrees are retained and error message lists which repos failed

---

### User Story 4 - Force Recreate Existing Worktree (Priority: P2)

A developer has an existing worktree for `feature-123` but wants to start fresh (e.g., the branch got into a bad state). They run `fa wt create feature-123 --force` and Foundagent removes the existing worktrees and recreates them from the source branch.

**Why this priority**: Recovering from bad worktree states is important but less common than initial creation.

**Independent Test**: Create worktree `feature-123`, make changes, run `fa wt create feature-123 --force`, verify worktrees are recreated fresh from source branch.

**Acceptance Scenarios**:

1. **Given** worktrees for `feature-123` already exist, **When** I run `fa wt create feature-123 --force`, **Then** existing worktrees are removed and recreated
2. **Given** existing worktrees have uncommitted changes, **When** I run `fa wt create feature-123 --force`, **Then** I receive a warning about uncommitted changes and must confirm or use `--force`
3. **Given** `--force` is used with `--from release-1.0`, **When** command runs, **Then** worktrees are recreated from the specified source branch

---

### User Story 5 - Handle Existing Branch Gracefully (Priority: P2)

A developer accidentally runs `fa wt create feature-123` when that branch already exists in one or more repos (but no worktree exists). The system provides clear guidance to switch to the existing branch instead.

**Why this priority**: Clear error messages prevent confusion and guide users to the right command.

**Independent Test**: Create branch `feature-123` manually in a repo without worktree, run `fa wt create feature-123`, verify error suggests using `fa wt switch`.

**Acceptance Scenarios**:

1. **Given** branch `feature-123` exists in repo `api` but has no worktree, **When** I run `fa wt create feature-123`, **Then** error says "Branch 'feature-123' already exists in: api. Use `fa wt switch feature-123` to switch to it."
2. **Given** branch exists in all repos without worktrees, **When** I run the command, **Then** all repos with existing branches are listed in the error message

---

### Edge Cases

- **Empty workspace**: No repos configured — error with hint to run `fa add` first
- **Single repo workspace**: Works identically to multi-repo (no special case needed)
- **Very long branch names**: Git handles limits; pass through any git errors
- **Invalid branch name characters**: Validate branch name format before attempting creation
- **Worktree exists in some repos but not others**: Mixed state from previous failure — error with hint to clean up or use `--force`
- **No default branch detected**: Repo has unusual setup — error with hint to specify `--from`
- **Network issues during creation**: Worktrees are local operations, should work offline
- **VS Code workspace file missing**: Create it if it doesn't exist, or skip update with warning
- **Concurrent worktree operations**: Git handles locking; surface any lock errors clearly
- **Disk space issues**: Surface git errors about disk space clearly

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa worktree create <branch>` command
- **FR-002**: System MUST support `fa wt create <branch>` as alias
- **FR-003**: System MUST accept `--from <source-branch>` flag to specify source branch
- **FR-004**: System MUST default to each repo's default branch when `--from` is not specified
- **FR-005**: System MUST accept `--force` flag to recreate existing worktrees

#### Branch Validation
- **FR-006**: System MUST validate that `--from` branch exists in ALL repos before starting any operation
- **FR-007**: When `--from` branch is missing in any repo, system MUST fail with list of repos missing the branch
- **FR-008**: System MUST validate that target branch name is a valid git branch name
- **FR-009**: When target branch already exists (without worktree), system MUST error with hint to use `fa wt switch`

#### Worktree Creation
- **FR-010**: System MUST create worktrees across ALL repos in the workspace
- **FR-011**: System MUST create each worktree at `repos/worktrees/<repo>/<branch>/`
- **FR-012**: System MUST create worktrees in parallel for performance
- **FR-013**: System MUST create a new branch with the specified name in each repo
- **FR-014**: New branch MUST be based on the source branch (default or `--from`)
- **FR-015**: Each worktree checkout MUST be at the tip of the newly created branch

#### Existing Worktree Handling
- **FR-016**: When worktree already exists and `--force` is NOT specified, system MUST error with hint
- **FR-017**: When worktree already exists and `--force` IS specified, system MUST remove and recreate
- **FR-018**: Before force-removing worktrees with uncommitted changes, system MUST warn user
- **FR-019**: In non-interactive mode (e.g., `--json`), force-remove MUST require explicit `--force` flag

#### VS Code Workspace Integration
- **FR-020**: System MUST update the `.code-workspace` file to include new worktree directories
- **FR-021**: If `.code-workspace` file doesn't exist, system MUST create it or skip with warning
- **FR-022**: Workspace file update MUST preserve existing folders and settings

#### Output and Feedback
- **FR-023**: System MUST show progress for each repo during creation
- **FR-024**: On success, system MUST display summary of created worktrees with paths
- **FR-025**: System MUST support `--json` flag for machine-readable output
- **FR-026**: Error messages MUST include actionable remediation steps

#### State Management
- **FR-027**: System MUST update `.foundagent/state.json` with new worktree information
- **FR-028**: Worktree state MUST include repo name, branch name, and filesystem path (e.g., `repos/worktrees/api/feature-123`)

### Key Entities

- **Worktree**: A working directory linked to a specific branch of a repository. Created via git worktree and managed by Foundagent across all repos in a workspace.
- **Source Branch**: The branch from which a new worktree's branch is created. Either the repo's default branch or explicitly specified via `--from`.
- **Target Branch**: The new branch name that will be created in each repo for the worktree.
- **Workspace File**: VS Code's `.code-workspace` JSON file that defines which folders are included in a multi-root workspace.

### Assumptions

- Users have already initialized a workspace with `fa init` and added repos with `fa add`
- Git is installed and available on the system PATH
- Users have appropriate permissions to create directories in the workspace
- Worktrees are created at `repos/worktrees/<repo>/<branch>/` following repo → branch hierarchy
- Bare clones exist at `repos/.bare/<repo>.git/`

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create worktrees across a 5-repo workspace in under 10 seconds
- **SC-002**: 100% of worktree creation attempts either succeed completely or fail with no partial state
- **SC-003**: Users can start coding in their new worktree within 30 seconds of running the command
- **SC-004**: Error messages enable users to resolve issues on first attempt in 90% of cases
- **SC-005**: Force recreate operation preserves no artifacts from previous worktree state
