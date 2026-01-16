# Feature Specification: Cross-Repo Commit

**Feature Branch**: `014-cross-repo-commit`  
**Created**: 2026-01-08  
**Status**: Draft  
**Input**: User description: "synchronized cross-repo git commits and pushes"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Commit Related Changes Across Multiple Repos (Priority: P1)

A developer has been working on a feature that spans multiple repos in their workspace—perhaps adding an API endpoint in the backend, updating the client SDK, and modifying the frontend to use the new functionality. They want to commit all these related changes together with a single commit message that links them logically. They run `fa commit "Add user preferences feature"` and Foundagent creates commits in all repos with staged changes, using the same commit message.

**Why this priority**: This is the core value proposition—developers working across repos need coordinated commits that represent a single logical change. Without this, they must manually commit each repo separately, risking inconsistent messages and forgotten repos.

**Independent Test**: Stage changes in 3 of 5 repos, run `fa commit "Test message"`, verify all 3 repos have commits with the same message and timestamp (within seconds).

**Acceptance Scenarios**:

1. **Given** repos api, web, and lib have staged changes, **When** I run `fa commit "Add feature X"`, **Then** all three repos receive commits with message "Add feature X"
2. **Given** repos api and web have staged changes but lib does not, **When** I run `fa commit "Update Y"`, **Then** only api and web receive commits (lib is skipped)
3. **Given** I run `fa commit` without a message, **When** prompted, **Then** my configured editor opens with a commit template
4. **Given** commit completes successfully, **When** I view the output, **Then** I see a summary showing which repos were committed with short SHA for each

---

### User Story 2 - Push Coordinated Commits Together (Priority: P1)

After committing across repos, a developer wants to push all the changes to their remotes in a coordinated way. They run `fa push` and Foundagent pushes all repos in the current branch worktrees that have unpushed commits. If any push fails, the developer sees exactly which repos succeeded and which failed.

**Why this priority**: Pushing is the natural follow-up to committing. Coordinated pushes ensure that related changes arrive at the remote together, reducing the window where repos are out of sync.

**Independent Test**: After `fa commit`, run `fa push`, verify all repos with new commits are pushed and summary shows success for each.

**Acceptance Scenarios**:

1. **Given** repos api, web, and lib have unpushed commits, **When** I run `fa push`, **Then** all three repos push to their remotes
2. **Given** push succeeds for api and web but fails for lib (remote has new commits), **When** push completes, **Then** I see api and web succeeded, lib failed with reason
3. **Given** no repos have unpushed commits, **When** I run `fa push`, **Then** message says "Nothing to push"
4. **Given** push completes, **When** I view the output, **Then** I see which repos were pushed and the remote refs updated

---

### User Story 3 - Stage and Commit All Changes at Once (Priority: P1)

A developer has been making changes across repos but hasn't staged anything yet. They want to stage all modifications and commit them in one step. They run `fa commit -a "Quick fix"` and Foundagent stages all tracked file changes and commits across repos.

**Why this priority**: The `-a` flag for auto-staging is a common git workflow. Supporting it for cross-repo commits maintains familiar git ergonomics.

**Independent Test**: Make modifications (no staging) in 2 repos, run `fa commit -a "Test"`, verify both repos have commits including all modified files.

**Acceptance Scenarios**:

1. **Given** repos have modified but unstaged tracked files, **When** I run `fa commit -a "Message"`, **Then** all modifications are staged and committed
2. **Given** repos have untracked files, **When** I run `fa commit -a`, **Then** untracked files are NOT included (matching git behavior)
3. **Given** some repos have changes and some don't, **When** I run `fa commit -a "Msg"`, **Then** only repos with changes receive commits

---

### User Story 4 - Dry Run to Preview Commit (Priority: P2)

Before committing, a developer wants to see exactly what will be committed across repos without actually making commits. They run `fa commit --dry-run` to preview the changes.

**Why this priority**: Previewing helps prevent mistakes, especially when working across many repos. Less critical than actual commit/push functionality.

**Independent Test**: Stage changes in 3 repos, run `fa commit --dry-run "Test"`, verify output shows what would be committed without creating actual commits.

**Acceptance Scenarios**:

1. **Given** repos have staged changes, **When** I run `fa commit --dry-run "Msg"`, **Then** I see which repos would receive commits and what files
2. **Given** dry-run completes, **When** I check git log in each repo, **Then** no new commits exist
3. **Given** dry-run mode, **When** output is displayed, **Then** it's clearly labeled as "DRY RUN" or similar

---

### User Story 5 - Commit Only Specific Repos (Priority: P2)

A developer has changes in multiple repos but only wants to commit changes in specific repos right now. They run `fa commit --repo api --repo web "Update API"` to limit the commit scope.

**Why this priority**: Fine-grained control is important for complex workflows, but most users will want all-repo commits.

**Independent Test**: Stage changes in 3 repos, run `fa commit --repo api "Test"`, verify only api receives a commit.

**Acceptance Scenarios**:

1. **Given** repos api, web, and lib have staged changes, **When** I run `fa commit --repo api --repo web "Msg"`, **Then** only api and web receive commits
2. **Given** I specify `--repo foo` where foo doesn't exist, **When** command runs, **Then** error reports "repo 'foo' not found"
3. **Given** I specify `--repo api` but api has no staged changes, **When** command runs, **Then** message says "Nothing to commit in api"

---

### User Story 6 - JSON Output for Automation (Priority: P2)

An AI agent or CI script needs structured output from commit operations to track what changed.

**Why this priority**: Agent-friendly design principle. Important for automation but human output is primary use case.

**Independent Test**: Run `fa commit --json "Test"`, parse JSON output, verify it contains per-repo commit details.

**Acceptance Scenarios**:

1. **Given** I run `fa commit --json "Msg"`, **When** commits complete, **Then** output is valid JSON with per-repo status
2. **Given** JSON output, **When** I examine it, **Then** each repo shows: name, commit SHA, files changed count, success/failure
3. **Given** I run `fa push --json`, **When** push completes, **Then** output includes per-repo push status and remote refs

---

### User Story 7 - Amend Previous Cross-Repo Commits (Priority: P3)

A developer realizes they need to add more changes to their last commit across repos. They run `fa commit --amend` to amend the previous commits in repos with new staged changes.

**Why this priority**: Amending is a power-user feature. Core commit/push comes first.

**Independent Test**: Make initial commits across 2 repos, stage new changes, run `fa commit --amend`, verify both repos' HEAD commits are amended.

**Acceptance Scenarios**:

1. **Given** repos have previous commits and new staged changes, **When** I run `fa commit --amend`, **Then** HEAD commits are amended in each repo
2. **Given** I run `fa commit --amend "New msg"`, **When** amend completes, **Then** commit messages are updated to "New msg"
3. **Given** some repos have new changes and some don't, **When** I run `fa commit --amend`, **Then** only repos with staged changes are amended

---

### Edge Cases

**Covered in v1:**
- **No staged changes**: Run `fa commit "Msg"` with nothing staged — show "Nothing to commit" with suggestion to stage or use `-a`
- **Empty workspace**: No repos configured — error with hint to run `fa add`
- **Partial commit failure**: One repo fails to commit (e.g., pre-commit hook fails) — report failure, show which repos succeeded
- **Partial push failure**: Some repos push, some fail — clearly report status for each, suggest resolution for failures
- **Dirty index vs worktree**: Some changes staged, some not — commit only staged (matching git behavior)
- **Detached HEAD**: Repo in detached state — raise error E402, allow with `--allow-detached`
- **Pre-commit hooks**: Respect per-repo pre-commit hooks — if one fails, that repo's commit fails

**Deferred to future iteration:**
- **Protected branch**: Push blocked by branch protection — git surfaces error natively; no special handling in v1
- **Large commits**: Many files across many repos — git handles progress natively; no special handling in v1
- **Merge in progress**: Repo has merge in progress — git commit handles natively; no special handling in v1

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa commit` command for cross-repo commits
- **FR-002**: System MUST support `fa push` command for cross-repo pushes
- **FR-003**: System MUST support commit message as positional argument (e.g., `fa commit "message"`)
- **FR-003a**: System MUST also support `-m <message>` / `--message <message>` flag for compatibility
- **FR-004**: System MUST support `-a` / `--all` flag to stage all tracked modifications before commit
- **FR-005**: System MUST support `--amend` flag to amend previous commits
- **FR-006**: System MUST support `--dry-run` flag to preview without executing
- **FR-007**: System MUST support `--repo <name>` flag to limit scope (can be repeated)
- **FR-008**: System MUST support `--json` flag for machine-readable output
- **FR-009**: System MUST support `-v` / `--verbose` flag for detailed progress

#### Commit Operations
- **FR-010**: System MUST identify all repos in current branch worktrees with staged changes
- **FR-011**: System MUST create commits with identical messages across all targeted repos
- **FR-012**: System MUST skip repos with no staged changes (unless `-a` adds changes)
- **FR-013**: System MUST run commits in parallel for performance
- **FR-014**: System MUST respect each repo's pre-commit hooks
- **FR-015**: System MUST report per-repo commit results (SHA, files changed)
- **FR-016**: Without `-m`, system MUST open configured editor for commit message

#### Push Operations
- **FR-017**: System MUST identify all repos with unpushed commits
- **FR-018**: System MUST push all targeted repos to their configured remotes
- **FR-019**: System MUST run pushes in parallel for performance
- **FR-020**: System MUST report per-repo push results (success/failure, refs updated)
- **FR-021**: System MUST handle push failures gracefully (continue other repos, report issues)
- **FR-021a**: System MUST support `--force` flag for force push with interactive confirmation (fails in --json mode)

#### Scope Control
- **FR-022**: By default, commands MUST evaluate all repos in the current branch worktrees, but only act on repos with changes (staged changes for commit, unpushed commits for push)
- **FR-023**: With `--repo`, commands MUST limit scope to specified repos only
- **FR-024**: System MUST validate specified repos exist in workspace

#### Progress and Output
- **FR-025**: System MUST show progress during multi-repo operations
- **FR-026**: System MUST show summary of results: repos committed/pushed, skipped, failed
- **FR-027**: Verbose mode MUST show individual file changes and git output per repo
- **FR-028**: Output MUST clearly indicate which repos were affected

#### Error Handling
- **FR-029**: System MUST handle partial failures (some repos succeed, some fail)
- **FR-030**: System MUST report clear error messages per repo on failure
- **FR-031**: System MUST validate command is run inside a Foundagent workspace
- **FR-032**: System MUST exit with code 0 if all targeted repos succeed, non-zero if any fail
- **FR-033**: Push failures MUST include suggestions (e.g., "pull first" for diverged history)

#### JSON Output
- **FR-034**: JSON MUST include `repos` array with name, status, commit_sha (for commit), error (if any)
- **FR-035**: JSON MUST include `summary` with counts (committed/pushed, skipped, failed)
- **FR-036**: For commits, JSON MUST include files_changed count per repo
- **FR-037**: For pushes, JSON MUST include refs_updated per repo

### Key Entities

- **Cross-Repo Commit**: A logical commit operation that creates commits across multiple repos with the same message, representing a single logical change.
- **Commit Result**: Per-repo outcome including success/failure, commit SHA, files changed count, and any error message.
- **Push Result**: Per-repo outcome including success/failure, refs pushed, and any error message.
- **Operation Summary**: Aggregate counts of repos affected, skipped, and failed for the operation.

### Assumptions

- All repos use standard remote named `origin` (consistent with existing sync behavior)
- Commits operate on the current branch across all repos (repos checked out to matching branches)
- Pre-commit hooks are repo-specific and respected per repo
- Commit message encoding follows each repo's configured encoding
- Push targets the upstream tracking branch or configured default remote

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can commit across 10 repos in under 10 seconds on typical hardware
- **SC-002**: Users can push across 10 repos in under 30 seconds on typical network
- **SC-003**: 100% of cross-repo commits have identical commit messages across all affected repos
- **SC-004**: Partial failures are clearly reported with 100% visibility into which repos succeeded/failed
- **SC-005**: JSON output contains complete commit/push details for AI agent consumption
- **SC-006**: Pre-commit hook failures in one repo do not prevent commits in other repos
