# Feature Specification: Add Repository

**Feature Branch**: `002-repo-add`  
**Created**: 2025-12-05  
**Status**: Draft  
**Input**: User description: "Add repositories to a Foundagent workspace with fa add command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add Single Repository (Priority: P1)

A developer has initialized a Foundagent workspace and wants to add their first repository. They run `fa add git@github.com:org/my-repo.git` to clone the repository as a bare clone and automatically create a worktree for the default branch so they can start working immediately.

**Why this priority**: This is the core functionality — adding repos is the primary purpose of the command. Without this, the workspace is empty and unusable.

**Independent Test**: Run `fa add <public-repo-url>` in an initialized workspace, verify the bare clone exists in `repos/<repo-name>/.bare/`, verify a worktree for the default branch is created at `repos/worktrees/<name>/main/`, and confirm the repo appears in both `.foundagent.yaml` and `.foundagent/state.json`.

**Acceptance Scenarios**:

1. **Given** I am in a Foundagent workspace, **When** I run `fa add git@github.com:org/my-repo.git`, **Then** the repo is cloned as a bare clone to `repos/.bare/my-repo.git/`
2. **Given** I run `fa add <url>`, **When** the clone completes, **Then** a worktree for the default branch is created at `repos/my-repo/worktrees/main/`
3. **Given** I run `fa add <url>`, **When** the command succeeds, **Then** the repo is registered in `.foundagent/state.json` and `.foundagent.yaml`
4. **Given** I run `fa add <url>`, **When** the command completes, **Then** the `.code-workspace` file is updated to include the new worktree folder

---

### User Story 2 - Add Multiple Repositories (Priority: P2)

A developer setting up a multi-repo project wants to add several repositories at once. They run `fa add <url1> <url2> <url3>` to clone all repos in parallel, creating worktrees for each.

**Why this priority**: Efficiency for multi-repo setups. Single repo add works first; batch add is a usability improvement.

**Independent Test**: Run `fa add <url1> <url2>` with two public repos, verify both are cloned and have worktrees, verify parallel execution completes faster than sequential.

**Acceptance Scenarios**:

1. **Given** I am in a Foundagent workspace, **When** I run `fa add <url1> <url2> <url3>`, **Then** all three repos are cloned in parallel
2. **Given** I add multiple repos, **When** one fails and others succeed, **Then** successful repos are kept and failures are reported with specific errors
3. **Given** I add multiple repos, **When** the command completes, **Then** I see a summary showing success/failure status for each repo

---

### User Story 3 - Add Repository with Custom Name (Priority: P2)

A developer wants to add a repository but use a different local name than the repo's default name. They run `fa add git@github.com:org/my-repo.git custom-name` to clone with a custom identifier.

**Why this priority**: Supports cases where repo names conflict or where a more meaningful local name is desired. Common in multi-repo setups with similarly named repos from different orgs.

**Independent Test**: Run `fa add <url> custom-name`, verify the bare clone is at `.foundagent/repos/custom-name.git/`, verify the worktree uses the custom name.

**Acceptance Scenarios**:

1. **Given** I run `fa add git@github.com:org/my-repo.git api-service`, **When** the clone completes, **Then** the bare clone is stored as `repos/.bare/api-service.git/`
2. **Given** I provide a custom name, **When** the worktree is created, **Then** it is created at `repos/worktrees/api-service/<default-branch>/`

---

### User Story 4 - Handle Already-Added Repository (Priority: P3)

A developer accidentally tries to add a repository that already exists in the workspace. The system should handle this gracefully without re-cloning.

**Why this priority**: Error handling and idempotency. Important for scripts and automation, but not the primary use case.

**Independent Test**: Add a repo, run the same `fa add <url>` again, verify no re-clone occurs and appropriate message is shown.

**Acceptance Scenarios**:

1. **Given** a repo "my-repo" already exists in the workspace, **When** I run `fa add <same-url>`, **Then** the command reports "Repository 'my-repo' already exists" and skips
2. **Given** a repo already exists, **When** I run `fa add <same-url> --force`, **Then** the repo is re-cloned (bare clone replaced, worktrees preserved if possible)
3. **Given** I add multiple repos where some already exist, **When** the command completes, **Then** existing repos are skipped and new repos are added

---

### User Story 5 - JSON Output for Automation (Priority: P3)

A developer using AI agents or automation scripts needs machine-readable output from the add command.

**Why this priority**: Supports agent-friendly design principle. Important for integration but human output is primary.

**Independent Test**: Run `fa add <url> --json`, parse output, verify it contains repo name, path, worktree path, and status.

**Acceptance Scenarios**:

1. **Given** I run `fa add <url> --json`, **When** the command completes, **Then** output is valid JSON with repo details
2. **Given** I add multiple repos with `--json`, **When** the command completes, **Then** output is a JSON array with status for each repo

---

### Edge Cases

- **Not in a workspace**: User runs `fa add` outside a Foundagent workspace — error with hint to run `fa init` first
- **Invalid URL**: URL is malformed or not a Git repository — error with clear message
- **Auth failure**: Private repo without valid credentials — error explaining auth requirement, suggest checking SSH keys or credential helper
- **Network failure**: Clone fails due to network — error with retry suggestion
- **Name collision with custom name**: `fa add <url1> api` then `fa add <url2> api` — error: name already in use
- **Empty repo**: Repository exists but has no commits — warn but proceed (bare clone valid, no default worktree created)
- **Repo URL with `.git` suffix vs without**: Both `repo.git` and `repo` URLs should work, name inferred correctly
- **Very long repo names**: Names exceeding filesystem limits — error with max length info
- **SSH vs HTTPS URLs**: Both formats supported, name inference works for both
- **Submodules**: Repository contains submodules — clone them as part of worktree creation (not bare clone)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST clone repositories as bare clones into `repos/<name>/.bare/`
- **FR-002**: System MUST infer repository name from URL if not provided (e.g., `github.com/org/my-repo.git` → `my-repo`)
- **FR-003**: System MUST support optional name argument to override inferred name: `fa add <url> [name]`
- **FR-004**: System MUST automatically create a worktree for the repository's default branch after cloning
- **FR-005**: System MUST create the default branch worktree at `repos/<name>/worktrees/<default-branch>/`
- **FR-006**: System MUST support adding multiple repositories in a single command: `fa add <url1> <url2> ...`
- **FR-007**: System MUST clone multiple repositories in parallel for performance
- **FR-008**: System MUST update `.foundagent/state.json` with the new repository metadata
- **FR-009**: System MUST update the `.code-workspace` file to include the new worktree folder(s)
- **FR-010**: System MUST skip already-existing repositories with a message (not an error)
- **FR-011**: System MUST support `--force` flag to re-clone an existing repository
- **FR-012**: System MUST preserve existing worktrees when using `--force` if possible
- **FR-013**: System MUST rely on system Git credentials for authentication (SSH keys, credential helper)
- **FR-014**: System MUST support `--json` flag for machine-readable output
- **FR-015**: System MUST validate that the command is run inside a Foundagent workspace
- **FR-016**: System MUST display progress during clone operations
- **FR-017**: System MUST report partial success when adding multiple repos (some succeed, some fail)
- **FR-018**: System MUST handle both SSH (`git@...`) and HTTPS (`https://...`) repository URLs
- **FR-019**: System MUST exit with code 0 on full success, non-zero on any failure
- **FR-020**: System MUST update `.foundagent.yaml` to include the added repository in the repos list

### Key Entities

- **Repository**: A Git repository added to the workspace. Stored as a bare clone in `repos/<name>/.bare/`. Key attributes: name, remote URL, default branch, clone status.
- **Worktree**: A checked-out working copy of a repository at a specific branch. Located at `repos/<repo-name>/worktrees/<branch-name>/`. Created automatically for the default branch when a repo is added.
- **Workspace State**: Machine-managed `.foundagent/state.json` tracking clone status, worktree paths, and sync timestamps.
- **Workspace Config**: User-editable `.foundagent.yaml` updated to include added repositories.

## Assumptions

- Worktrees follow the repo → branch hierarchy: `repos/<repo>/worktrees/<branch>/`
- Submodule handling during worktree creation follows standard Git behavior

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Single repository add completes within 30 seconds for typical repos (< 100MB)
- **SC-002**: Multiple repository adds execute in parallel, completing faster than sequential execution
- **SC-003**: 100% of successful adds result in a working worktree that can be opened in VS Code
- **SC-004**: Error messages include actionable remediation steps in 100% of failure cases
- **SC-005**: JSON output parses successfully with standard JSON parsers
- **SC-006**: Re-running `fa add` for existing repos is idempotent (no errors, no side effects)
