# Tasks: Cross-Repo Commit

**Input**: Design documents from `/specs/014-cross-repo-commit/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on existing Go project structure:
- `internal/cli/` - Cobra command definitions
- `internal/git/` - Git operations via os/exec
- `internal/workspace/` - Cross-repo orchestration
- Tests co-located as `*_test.go` files

---

## Phase 1: Setup

**Purpose**: New error codes and shared infrastructure for commit/push operations

- [X] T001 Add commit/push error codes (E401, E402) in internal/errors/codes.go

---

## Phase 2: Foundational (Git Operations Layer)

**Purpose**: Low-level git operations that ALL user stories depend on. MUST complete before any CLI work.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T002 [P] Implement `HasStagedChanges(path)` function in internal/git/commit.go
- [X] T003 [P] Implement `Commit(path, message, opts)` function in internal/git/commit.go
- [X] T004 [P] Implement `GetStagedFiles(path)` function in internal/git/commit.go
- [X] T005 [P] Implement `GetCommitStats(path, sha)` function in internal/git/commit.go
- [X] T006 [P] Write tests for HasStagedChanges in internal/git/commit_test.go
- [X] T007 [P] Write tests for Commit function in internal/git/commit_test.go
- [X] T008 [P] Write tests for GetStagedFiles in internal/git/commit_test.go
- [X] T009 [P] Write tests for GetCommitStats in internal/git/commit_test.go
- [X] T010 [P] Implement `HasUnpushedCommits(path)` function in internal/git/push.go
- [X] T011 [P] Implement `GetUnpushedCount(path)` function in internal/git/push.go
- [X] T012 [P] Write tests for HasUnpushedCommits in internal/git/push_test.go
- [X] T013 [P] Write tests for GetUnpushedCount in internal/git/push_test.go

**Checkpoint**: Git operations layer ready - workspace orchestration can now begin

---

## Phase 3: User Story 1 - Cross-Repo Commit (Priority: P1) 🎯 MVP

**Goal**: Commit staged changes across all repos with identical message using `fa commit "message"`

**Independent Test**: Stage changes in 3 repos, run `fa commit "Test"`, verify all 3 have commits with same message

### Data Types for User Story 1

- [X] T014 [P] [US1] Define CommitResult struct in internal/workspace/commit.go
- [X] T015 [P] [US1] Define CommitSummary struct in internal/workspace/commit.go
- [X] T016 [P] [US1] Define CommitOptions struct in internal/workspace/commit.go

### Workspace Orchestration for User Story 1

- [X] T017 [US1] Implement `CommitAllRepos(opts)` function in internal/workspace/commit.go
- [X] T018 [US1] Implement `CalculateCommitSummary(results)` in internal/workspace/commit.go
- [X] T019 [US1] Implement `FormatCommitResults(results)` for human output in internal/workspace/commit.go
- [X] T020 [US1] Write tests for CommitAllRepos in internal/workspace/commit_test.go
- [X] T021 [US1] Write tests for CalculateCommitSummary in internal/workspace/commit_test.go

### CLI for User Story 1

- [X] T022 [US1] Create commit command skeleton with flags in internal/cli/commit.go
- [X] T023 [US1] Implement runCommit function (basic flow) in internal/cli/commit.go
- [X] T024 [US1] Implement human output formatting in internal/cli/commit.go
- [X] T025 [US1] Wire commit command to rootCmd in internal/cli/commit.go
- [X] T026 [US1] Write basic commit command tests in internal/cli/commit_test.go

**Checkpoint**: `fa commit "message"` works for basic cross-repo commits

---

## Phase 4: User Story 2 - Cross-Repo Push (Priority: P1)

**Goal**: Push unpushed commits across all repos using `fa push`

**Independent Test**: After committing, run `fa push`, verify all repos with commits are pushed

### Data Types for User Story 2

- [X] T027 [P] [US2] Define PushResult struct in internal/workspace/push.go
- [X] T028 [P] [US2] Define PushSummary struct in internal/workspace/push.go
- [X] T029 [P] [US2] Define PushOptions struct in internal/workspace/push.go

### Workspace Orchestration for User Story 2

- [X] T030 [US2] Implement `PushAllRepos(opts)` function in internal/workspace/push.go
- [X] T031 [US2] Implement `CalculatePushSummary(results)` in internal/workspace/push.go
- [X] T032 [US2] Implement `FormatPushResults(results)` for human output in internal/workspace/push.go
- [X] T033 [US2] Write tests for PushAllRepos in internal/workspace/push_test.go
- [X] T034 [US2] Write tests for CalculatePushSummary in internal/workspace/push_test.go

### CLI for User Story 2

- [X] T035 [US2] Create push command skeleton with flags in internal/cli/push.go
- [X] T036 [US2] Implement runPush function in internal/cli/push.go
- [X] T037 [US2] Implement human output formatting in internal/cli/push.go
- [X] T038 [US2] Wire push command to rootCmd in internal/cli/push.go
- [X] T039 [US2] Write basic push command tests in internal/cli/push_test.go

**Checkpoint**: `fa push` works for cross-repo pushes

---

## Phase 5: User Story 3 - Stage and Commit (-a flag) (Priority: P1)

**Goal**: Stage all tracked modifications and commit with `fa commit -a "message"`

**Independent Test**: Make modifications (no staging), run `fa commit -a "Test"`, verify commits include all changes

### Implementation for User Story 3

- [X] T040 [US3] Add `-a/--all` flag handling to commit command in internal/cli/commit.go
- [X] T041 [US3] Implement `StageAllTracked(path)` in internal/git/commit.go
- [X] T042 [US3] Update CommitAllRepos to handle `All` option in internal/workspace/commit.go
- [X] T043 [US3] Write tests for StageAllTracked in internal/git/commit_test.go
- [X] T044 [US3] Write tests for commit with -a flag in internal/git/commit_test.go

**Checkpoint**: `fa commit -a "message"` stages and commits in one step

---

## Phase 6: User Story 4 - Dry Run Preview (Priority: P2)

**Goal**: Preview commits without executing using `fa commit --dry-run`

**Independent Test**: Stage changes, run `fa commit --dry-run "Test"`, verify no commits created

### Implementation for User Story 4

- [X] T045 [US4] Add `--dry-run` flag handling to commit command in internal/cli/commit.go
- [X] T046 [US4] Implement `GetDryRunPreview(opts)` in internal/workspace/commit.go
- [X] T047 [US4] Implement dry-run output formatting in internal/cli/commit.go
- [X] T048 [US4] Add `--dry-run` flag to push command in internal/cli/push.go
- [X] T049 [US4] Implement `GetPushDryRunPreview(opts)` in internal/workspace/push.go
- [X] T050 [US4] Write tests for commit dry-run in internal/cli/commit_test.go
- [X] T051 [US4] Write tests for push dry-run in internal/cli/push_test.go

**Checkpoint**: `--dry-run` shows preview without making changes

---

## Phase 7: User Story 5 - Repo Filtering (Priority: P2)

**Goal**: Limit operations to specific repos using `--repo` flag

**Independent Test**: Stage changes in 3 repos, run `fa commit --repo api "Test"`, verify only api committed

### Implementation for User Story 5

- [X] T052 [US5] Add `--repo` flag (repeatable) to commit command in internal/cli/commit.go
- [X] T053 [US5] Implement repo filtering in CommitAllRepos in internal/workspace/commit.go
- [X] T054 [US5] Add repo validation (check repo exists) in internal/workspace/commit.go
- [X] T055 [US5] Add `--repo` flag to push command in internal/cli/push.go
- [X] T056 [US5] Implement repo filtering in PushAllRepos in internal/workspace/push.go
- [X] T057 [US5] Write tests for repo filtering in internal/cli/commit_test.go
- [X] T058 [US5] Write tests for invalid repo error in internal/cli/commit_test.go

**Checkpoint**: `--repo` flag filters operations to specific repos

---

## Phase 8: User Story 6 - JSON Output (Priority: P2)

**Goal**: Machine-readable JSON output for automation

**Independent Test**: Run `fa commit --json "Test"`, parse output as valid JSON with per-repo details

### Implementation for User Story 6

- [X] T059 [US6] Add `--json` flag to commit command in internal/cli/commit.go
- [X] T060 [US6] Implement outputCommitJSON function in internal/cli/commit.go
- [X] T061 [US6] Add `--json` flag to push command in internal/cli/push.go
- [X] T062 [US6] Implement outputPushJSON function in internal/cli/push.go
- [X] T063 [US6] Write tests for JSON output format in internal/cli/commit_test.go
- [X] T064 [US6] Write tests for JSON error output in internal/cli/commit_test.go

**Checkpoint**: JSON output matches contract spec

---

## Phase 9: User Story 7 - Amend Commits (Priority: P3)

**Goal**: Amend previous commits using `fa commit --amend`

**Independent Test**: Make commits, stage new changes, run `fa commit --amend`, verify HEAD amended

### Implementation for User Story 7

- [X] T065 [US7] Add `--amend` flag to commit command in internal/cli/commit.go
- [X] T066 [US7] Implement amend support in git.Commit function in internal/git/commit.go
- [X] T067 [US7] Update CommitAllRepos to handle Amend option in internal/workspace/commit.go
- [X] T068 [US7] Write tests for amend functionality in internal/git/commit_test.go
- [X] T069 [US7] Write tests for amend command in internal/cli/commit_test.go

**Checkpoint**: `fa commit --amend` updates previous commits

---

## Phase 10: Polish & Edge Cases

**Purpose**: Error handling, edge cases, and documentation

- [X] T070 [P] Add empty workspace handling with helpful error in internal/cli/commit.go
- [X] T071 [P] Add empty workspace handling in internal/cli/push.go
- [X] T072 [P] Add "nothing to commit" message handling in internal/cli/commit.go
- [X] T073 [P] Add "nothing to push" message handling in internal/cli/push.go
- [X] T074 [P] Add detached HEAD detection and warning in internal/workspace/commit.go
- [X] T075 [P] Add `--allow-detached` flag to commit command in internal/cli/commit.go
- [X] T076 [P] Add push failure suggestions (e.g., "pull first") in internal/cli/push.go
- [ ] T077 [P] Implement editor integration for commit without -m in internal/cli/commit.go (DEFERRED - requires interactive terminal)
- [ ] T077.1 [P] Write tests for editor integration in internal/cli/commit_test.go (DEFERRED - depends on T077)
- [X] T078 [P] Add verbose mode output (`-v` flag) in internal/cli/commit.go
- [X] T079 [P] Add verbose mode output in internal/cli/push.go
- [ ] T080 Run quickstart.md validation (manual test)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational completion
  - US1 (commit) and US2 (push) can proceed in parallel
  - US3-US7 build on US1/US2 foundations
- **Polish (Phase 10)**: Depends on all user stories being complete

### User Story Dependencies

| Story | Depends On | Can Parallel With |
|-------|------------|-------------------|
| US1 (commit) | Foundational | US2 |
| US2 (push) | Foundational | US1 |
| US3 (-a flag) | US1 | US4, US5, US6 |
| US4 (dry-run) | US1, US2 | US3, US5, US6 |
| US5 (--repo) | US1, US2 | US3, US4, US6 |
| US6 (JSON) | US1, US2 | US3, US4, US5 |
| US7 (amend) | US1 | - |

### Within Each Phase

- Tasks marked [P] can run in parallel
- Data type definitions before implementations
- Git layer before workspace layer
- Workspace layer before CLI layer
- Tests alongside or immediately after implementation

---

## Parallel Example: Foundational Phase

```bash
# All git operations can be implemented in parallel:
T002: HasStagedChanges in internal/git/commit.go
T003: Commit function in internal/git/commit.go
T004: GetStagedFiles in internal/git/commit.go
T005: GetCommitStats in internal/git/commit.go
T010: HasUnpushedCommits in internal/git/push.go
T011: GetUnpushedCount in internal/git/push.go

# All tests can run in parallel with their implementations:
T006-T009: commit_test.go tests
T012-T013: push_test.go tests
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Setup (error codes)
2. Complete Phase 2: Foundational (git operations)
3. Complete Phase 3: US1 - Basic commit
4. Complete Phase 4: US2 - Basic push
5. **STOP and VALIDATE**: Test `fa commit "msg"` and `fa push`
6. Deploy/demo if ready - this is the core functionality

### Incremental Delivery

| Increment | Stories | New Capability |
|-----------|---------|----------------|
| MVP | US1 + US2 | Basic commit and push |
| +Convenience | US3 | Stage-and-commit with `-a` |
| +Safety | US4 | Preview with `--dry-run` |
| +Control | US5 | Filter with `--repo` |
| +Automation | US6 | JSON output |
| +Power User | US7 | Amend commits |

---

## Notes

- All tasks follow existing patterns from `fa sync` implementation
- Uses `workspace.ExecuteParallel()` for parallel operations
- JSON output follows existing contract patterns
- Error codes follow existing E-series pattern
- Tests use table-driven patterns with `t.TempDir()` for isolation
