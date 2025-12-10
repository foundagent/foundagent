# Tasks: Worktree Create

**Input**: Design documents from `/specs/004-worktree-create/`
**Prerequisites**: spec.md ✅

**Tests**: Not explicitly requested - include minimal validation tests.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Include exact file paths in descriptions

## Path Conventions

- **CLI commands**: `internal/cli/`
- **Workspace logic**: `internal/workspace/`
- **Git operations**: `internal/git/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Create command skeleton and verify dependencies

- [X] T001 Create worktree command group skeleton in internal/cli/worktree.go
- [X] T002 [P] Create `wt` alias for worktree command in internal/cli/root.go
- [X] T003 [P] Create create subcommand skeleton in internal/cli/wt_create.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core validation and git worktree infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Implement branch name validation (valid git branch format) in internal/git/validation.go
- [X] T005 [P] Implement branch existence check across all repos in internal/git/branch.go
- [X] T006 [P] Implement worktree existence check in internal/workspace/worktree.go
- [X] T007 Implement parallel operation executor with error collection in internal/workspace/parallel.go
- [X] T008 Implement VS Code workspace file update (add folders) in internal/workspace/vscode.go
- [X] T009 Add `worktree create` and `wt create` commands to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - branch validation and parallel execution available

---

## Phase 3: User Story 1 - Create Worktree Across All Repos (Priority: P1) 🎯 MVP

**Goal**: Create worktrees in all repos with `fa wt create feature-123`

**Independent Test**: Run `fa wt create feature-123` with 3 repos, verify worktrees at `repos/worktrees/<repo>/feature-123/`

### Implementation for User Story 1

- [X] T010 [US1] Get list of all repos from workspace config in internal/cli/wt_create.go
- [X] T011 [US1] Detect default branch for each repo in internal/git/branch.go
- [X] T012 [US1] Create new branch from default branch in each repo in internal/git/branch.go
- [X] T013 [US1] Create worktree at `repos/worktrees/<repo>/<branch>/` for each repo in internal/git/worktree.go
- [X] T014 [US1] Execute worktree creation in parallel across repos in internal/cli/wt_create.go
- [X] T015 [US1] Update `.code-workspace` to include all new worktree folders in internal/cli/wt_create.go
- [X] T016 [US1] Update `.foundagent/state.json` with worktree info in internal/workspace/state.go
- [X] T017 [US1] Display progress for each repo during creation in internal/cli/wt_create.go
- [X] T018 [US1] Display summary of created worktrees with paths in internal/cli/wt_create.go

**Checkpoint**: `fa wt create feature-123` creates worktrees in all repos

---

## Phase 4: User Story 2 - Create from Specific Branch (Priority: P1)

**Goal**: Create worktrees based on non-default branch with `fa wt create feature-123 --from release-2.0`

**Independent Test**: Run `fa wt create hotfix --from release-2.0`, verify worktrees based on release-2.0

### Implementation for User Story 2

- [X] T019 [US2] Add `--from` flag to accept source branch in internal/cli/wt_create.go
- [X] T020 [US2] Validate `--from` branch exists in ALL repos before starting in internal/cli/wt_create.go
- [X] T021 [US2] Create new branch from specified source branch in internal/git/branch.go
- [X] T022 [US2] Error with list of repos missing `--from` branch if validation fails in internal/cli/wt_create.go

**Checkpoint**: `fa wt create --from release-2.0` creates from specified branch

---

## Phase 5: User Story 3 - Atomic Operation with Pre-validation (Priority: P1)

**Goal**: Validate all repos before any changes - all-or-nothing behavior

**Independent Test**: Try `fa wt create --from nonexistent` where branch missing in 1 repo, verify no worktrees created

### Implementation for User Story 3

- [X] T023 [US3] Implement pre-validation phase before any worktree creation in internal/cli/wt_create.go
- [X] T024 [US3] Check `--from` branch exists in every repo during pre-validation in internal/cli/wt_create.go
- [X] T025 [US3] Check target branch name doesn't already exist (without worktree) in internal/cli/wt_create.go
- [X] T026 [US3] Abort with clear error listing all failing repos if validation fails in internal/cli/wt_create.go
- [X] T027 [US3] Only proceed to creation if all validations pass in internal/cli/wt_create.go

**Checkpoint**: Atomic all-or-nothing - no partial states on validation failure

---

## Phase 6: User Story 4 - Force Recreate (Priority: P2)

**Goal**: Recreate existing worktrees with `fa wt create feature-123 --force`

**Independent Test**: Create worktree, modify, run `fa wt create feature-123 --force`, verify fresh worktree

### Implementation for User Story 4

- [X] T028 [US4] Add `--force` flag to recreate existing worktrees in internal/cli/wt_create.go
- [X] T029 [US4] Check for uncommitted changes before force remove in internal/git/status.go
- [X] T030 [US4] Warn about uncommitted changes and require confirmation or `--force` in internal/cli/wt_create.go
- [X] T031 [US4] Remove existing worktrees before recreation in internal/git/worktree.go
- [X] T032 [US4] Recreate worktrees from source branch in internal/cli/wt_create.go

**Checkpoint**: `fa wt create --force` recreates worktrees fresh

---

## Phase 7: User Story 5 - Handle Existing Branch Gracefully (Priority: P2)

**Goal**: Clear guidance when branch exists but no worktree

**Independent Test**: Create branch manually without worktree, run `fa wt create`, verify hint to use switch

### Implementation for User Story 5

- [X] T033 [US5] Detect if target branch exists in repo without worktree in internal/git/branch.go
- [X] T034 [US5] Error with suggestion to use `fa wt switch` instead in internal/cli/wt_create.go
- [X] T035 [US5] List all repos with existing branches in error message in internal/cli/wt_create.go

**Checkpoint**: Clear guidance to use switch for existing branches

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and error handling

- [X] T036 [P] Handle empty workspace (no repos) with hint to run `fa add` in internal/cli/wt_create.go
- [X] T037 [P] Handle invalid branch name characters with clear error in internal/git/validation.go
- [X] T038 [P] Handle missing `.code-workspace` file - create or warn in internal/workspace/vscode.go
- [X] T039 [P] Add `--json` flag for machine-readable output in internal/cli/wt_create.go
- [X] T040 Add help text with examples to wt create command in internal/cli/wt_create.go
- [X] T041 Write integration test for wt create command in internal/cli/wt_create_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational
  - US1 (Create All) is P1 - implement first (MVP)
  - US2 (From Branch) is P1 - depends on US1
  - US3 (Atomic) is P1 - depends on US1
  - US4 (Force) and US5 (Existing Branch) are P2
- **Polish (Phase 8)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Create All) | P1 | Foundational | - |
| US2 (From Branch) | P1 | US1 | US3 |
| US3 (Atomic) | P1 | US1 | US2 |
| US4 (Force) | P2 | US1 | US5 |
| US5 (Existing Branch) | P2 | US1 | US4 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                      Create All Repos
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
         [US2] [US3]                    [US4] [US5]
        From   Atomic                  Force  Existing
        Branch                                Branch
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T009)
3. Complete Phase 3: US1 Create All Repos (T010-T018)
4. Complete Phase 4: US2 From Branch (T019-T022)
5. Complete Phase 5: US3 Atomic Validation (T023-T027)
6. **STOP and VALIDATE**: Multi-repo worktree creation works atomically
7. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US4 (Force) + US5 (Existing Branch) → Test independently
3. Polish phase

---

## Notes

- Worktrees created at `repos/worktrees/<repo>/<branch>/`
- Bare clones at `repos/.bare/<repo>.git/`
- Parallel creation uses goroutines with error collection
- Pre-validation ensures atomic all-or-nothing behavior
- VS Code workspace file updated to include all new worktree folders
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
