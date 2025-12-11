# Feature Specification: Workspace Sync

**Feature Branch**: `008-workspace-sync`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Sync workspace with remotes using fa sync command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fetch Updates from All Remotes (Priority: P1)

A developer starts their day and wants to get the latest changes from all remotes across all repos in their workspace. They run `fa sync` and Foundagent fetches from all remotes in parallel, updating their local references so they can see what's changed.

**Why this priority**: Fetching is the foundation of sync — developers need to know what's available before deciding to pull or merge.

**Independent Test**: Run `fa sync` in a workspace with 3 repos that have remote changes, verify all repos fetch successfully and show summary of new commits/branches.

**Acceptance Scenarios**:

1. **Given** a workspace with repos (api, web, lib), **When** I run `fa sync`, **Then** all repos fetch from their remotes in parallel
2. **Given** remotes have new commits, **When** sync completes, **Then** I see a summary showing which repos have updates available
3. **Given** some repos are up-to-date, **When** sync completes, **Then** those repos show "already up-to-date"
4. **Given** sync completes, **When** I check local refs, **Then** `origin/main` and other remote branches are updated

---

### User Story 2 - Pull Updates into Current Worktree (Priority: P1)

A developer wants to update their current worktree with the latest changes from the remote. They run `fa sync --pull` to fetch and pull (fast-forward) across all repos in the current branch.

**Why this priority**: Pulling is the most common sync operation — getting latest changes into the working directory.

**Independent Test**: Run `fa sync --pull` when remotes have new commits, verify all worktrees for the current branch are updated.

**Acceptance Scenarios**:

1. **Given** I'm in worktree `repos/api/worktrees/main/`, **When** I run `fa sync --pull`, **Then** all worktrees for `main` branch are pulled (api, web, lib)
2. **Given** remote has new commits that fast-forward, **When** pull completes, **Then** worktree is updated to latest
3. **Given** remote has diverged (non-fast-forward), **When** pull would fail, **Then** sync reports conflict and suggests merge/rebase
4. **Given** worktree has uncommitted changes, **When** I run `fa sync --pull`, **Then** sync warns and skips that worktree (or stashes with `--stash`)

---

### User Story 3 - Handle Network Failures Gracefully (Priority: P1)

A developer is on spotty wifi and runs `fa sync`. Some repos succeed, some fail due to network issues. The command completes partial work and clearly reports what failed.

**Why this priority**: Graceful degradation is a core principle. Users need visibility into what worked and what didn't.

**Independent Test**: Simulate network failure for one repo's remote, run `fa sync`, verify successful repos are synced and failed repo is clearly reported.

**Acceptance Scenarios**:

1. **Given** network fails for one repo during sync, **When** sync completes, **Then** successful repos are synced and failure is reported
2. **Given** partial failure, **When** I see the error, **Then** it includes which repo failed, the error, and suggestion to retry
3. **Given** complete network failure, **When** sync fails, **Then** error clearly states "Network unavailable" with offline suggestions

---

### User Story 4 - Sync Specific Branch (Priority: P2)

A developer wants to sync a specific branch across all repos, not just the current one. They run `fa sync feature-123` to fetch/pull that branch in all repos.

**Why this priority**: Useful for updating a feature branch before switching to it. Current branch sync comes first.

**Independent Test**: Run `fa sync feature-123 --pull` while on `main`, verify `feature-123` worktrees are updated in all repos.

**Acceptance Scenarios**:

1. **Given** worktrees exist for `feature-123`, **When** I run `fa sync feature-123`, **Then** those worktrees are synced (not the current branch)
2. **Given** branch doesn't exist in some repos, **When** syncing, **Then** those repos are skipped with message

---

### User Story 5 - JSON Output for Automation (Priority: P2)

An AI agent or script needs structured output from sync operations to understand what changed.

**Why this priority**: Agent-friendly design principle. Important for integration but human output is primary.

**Independent Test**: Run `fa sync --json`, parse JSON output, verify it contains per-repo sync status.

**Acceptance Scenarios**:

1. **Given** I run `fa sync --json`, **When** sync completes, **Then** output is valid JSON with per-repo status
2. **Given** some repos failed, **When** using `--json`, **Then** each repo has success/failure status with error details

---

### User Story 6 - Push Local Changes (Priority: P3)

A developer has committed changes across multiple repos and wants to push them all at once. They run `fa sync --push` to push all repos with local commits ahead of remote.

**Why this priority**: Push is less common than fetch/pull and requires more care. Basic sync comes first.

**Independent Test**: Make commits in 2 of 3 repos, run `fa sync --push`, verify both repos push successfully.

**Acceptance Scenarios**:

1. **Given** repos have local commits ahead of remote, **When** I run `fa sync --push`, **Then** all repos with unpushed commits are pushed
2. **Given** push would fail (remote has new commits), **When** push fails, **Then** error suggests fetch/pull first
3. **Given** no repos have unpushed commits, **When** I run `fa sync --push`, **Then** message says "Nothing to push"

---

### Edge Cases

- **Empty workspace**: No repos configured — error with hint to run `fa add`
- **No network**: Complete network failure — fail fast with clear message
- **Auth failure**: Credentials invalid — error with auth troubleshooting hints
- **Dirty worktree on pull**: Uncommitted changes block pull — skip with warning, suggest `--stash`
- **Merge conflict**: Non-fast-forward pull — report conflict, don't auto-merge
- **Protected branch**: Push blocked by branch protection — surface the error clearly
- **Large fetch**: Many commits to fetch — show progress indicator
- **Detached HEAD**: Worktree in detached state — warn and skip sync for that worktree
- **Shallow clone**: Repo is shallow — fetch may behave differently, warn if needed

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa sync` command
- **FR-002**: System MUST accept optional `[branch]` argument to sync specific branch
- **FR-003**: System MUST support `--pull` flag to fetch and pull (fast-forward merge)
- **FR-004**: System MUST support `--push` flag to push local commits
- **FR-005**: System MUST support `--stash` flag to stash uncommitted changes before pull
- **FR-006**: System MUST support `--json` flag for machine-readable output
- **FR-007**: System MUST support `-v` / `--verbose` flag for detailed progress

#### Fetch Operations (Default)
- **FR-008**: By default (no flags), sync MUST fetch from all remotes for all repos
- **FR-009**: Fetch MUST run in parallel across repos for performance
- **FR-010**: Fetch MUST update all remote-tracking branches (e.g., `origin/main`)
- **FR-011**: After fetch, system MUST show summary of repos with available updates
- **FR-012**: Summary MUST show commits behind for each branch with updates

#### Pull Operations (--pull)
- **FR-013**: With `--pull`, system MUST fetch then fast-forward merge for each worktree
- **FR-014**: Pull MUST operate on current branch worktrees (all repos) unless branch specified
- **FR-015**: Pull MUST skip worktrees with uncommitted changes (unless `--stash`)
- **FR-016**: Pull MUST fail gracefully if fast-forward not possible (diverged history)
- **FR-017**: With `--stash`, system MUST stash changes before pull, then pop after

#### Push Operations (--push)
- **FR-018**: With `--push`, system MUST push all repos with local commits ahead of remote
- **FR-019**: Push MUST only push repos/branches with unpushed commits
- **FR-020**: Push MUST fail gracefully if remote has new commits (suggest pull first)
- **FR-021**: Push MUST report which repos were pushed and which had nothing to push

#### Progress and Output
- **FR-022**: System MUST show progress indicator during network operations
- **FR-023**: System MUST show per-repo status (fetching/pulling/pushing/done/failed)
- **FR-024**: Verbose mode MUST show individual ref updates and commit counts
- **FR-025**: Summary MUST show total repos synced, updated, failed, and skipped

#### Error Handling
- **FR-026**: System MUST handle network failures gracefully (complete partial work)
- **FR-027**: System MUST report per-repo errors without stopping other repos
- **FR-028**: Auth failures MUST include troubleshooting hints (SSH keys, credentials)
- **FR-029**: Diverged history MUST suggest merge/rebase resolution
- **FR-030**: System MUST validate command is run inside a Foundagent workspace
- **FR-031**: System MUST exit with code 0 if all repos succeed, non-zero if any fail

#### JSON Output
- **FR-032**: JSON MUST include `repos` array with name, status, error (if any)
- **FR-033**: JSON MUST include `summary` with counts (synced, updated, failed, skipped)
- **FR-034**: For each repo, JSON MUST include refs updated and commit counts

### Key Entities

- **Sync Operation**: A fetch, pull, or push operation across all repos in the workspace.
- **Repo Sync Status**: Per-repo result including success/failure, refs updated, commits fetched/pulled/pushed.
- **Sync Summary**: Aggregate counts of repos synced, updated, failed, and skipped.
- **Remote Tracking Branch**: A ref like `origin/main` updated by fetch to reflect remote state.

### Assumptions

- All repos use standard remote named `origin` (future: support multiple remotes)
- Bare clones at `repos/<repo>/.bare/` store fetched refs
- Worktrees at `repos/<repo>/worktrees/<branch>/` are updated by pull
- Auth handled by system Git credentials (SSH agent, credential helper)
- Fast-forward only for `--pull` (no auto-merge to avoid conflicts)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can sync a 5-repo workspace in under 30 seconds on typical network
- **SC-002**: Network failures for one repo do not block sync of other repos
- **SC-003**: 100% of sync operations report clear status for each repo
- **SC-004**: Uncommitted changes are never lost (dirty worktrees skipped or stashed)
- **SC-005**: JSON output contains complete sync status for AI agent consumption
