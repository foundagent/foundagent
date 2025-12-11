# Tasks: Worktree Remove

**Input**: Design documents from `/specs/006-worktree-remove/`
**Prerequisites**: spec.md ✅

**Tests**: Not explicitly requested - include minimal validation tests.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US4)
- Include exact file paths in descriptions

## Path Conventions

- **CLI commands**: `internal/cli/`
- **Workspace logic**: `internal/workspace/`
- **Git operations**: `internal/git/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Create command skeleton

- [X] T001 Create remove subcommand skeleton in internal/cli/wt_remove.go
- [X] T002 [P] Add `rm` alias for remove subcommand in internal/cli/worktree.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core worktree removal and safety check infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Implement git worktree remove function in internal/git/worktree.go
- [X] T004 [P] Implement dirty worktree detection (uncommitted/untracked) in internal/git/status.go
- [X] T005 [P] Implement CWD-inside-worktree detection in internal/workspace/worktree.go
- [X] T006 Implement VS Code workspace file update (remove folders) in internal/workspace/vscode.go
- [X] T007 Add `worktree remove`, `wt remove`, and `wt rm` commands to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - worktree removal with safety checks available

---

## Phase 3: User Story 1 - Remove Worktree Across All Repos (Priority: P1) 🎯 MVP

**Goal**: Remove worktrees from all repos with `fa wt remove feature-123`

**Independent Test**: Create worktrees for feature-123, run `fa wt remove feature-123`, verify all removed

### Implementation for User Story 1

- [X] T008 [US1] Require branch name argument in internal/cli/wt_remove.go
- [X] T009 [US1] Find all worktrees for specified branch across repos in internal/cli/wt_remove.go
- [X] T010 [US1] Remove worktrees using `git worktree remove` in internal/git/worktree.go
- [X] T011 [US1] Delete worktree directories from filesystem in internal/cli/wt_remove.go
- [X] T012 [US1] Update `.code-workspace` to remove worktree folders in internal/cli/wt_remove.go
- [X] T013 [US1] Update `.foundagent/state.json` to remove worktree entries in internal/workspace/state.go
- [X] T014 [US1] Display progress for each repo during removal in internal/cli/wt_remove.go
- [X] T015 [US1] Display summary of removed worktrees in internal/cli/wt_remove.go

**Checkpoint**: `fa wt remove feature-123` removes worktrees from all repos

---

## Phase 4: User Story 2 - Prevent Removal with Uncommitted Changes (Priority: P1)

**Goal**: Block removal of dirty worktrees unless `--force` used

**Independent Test**: Make changes in worktree, run `fa wt remove`, verify blocked with error

### Implementation for User Story 2

- [X] T016 [US2] Check all target worktrees for uncommitted changes in internal/cli/wt_remove.go
- [X] T017 [US2] Check all target worktrees for untracked files in internal/cli/wt_remove.go
- [X] T018 [US2] Block removal if any worktree is dirty in internal/cli/wt_remove.go
- [X] T019 [US2] List all dirty worktrees in error message in internal/cli/wt_remove.go
- [X] T020 [US2] Add `--force` flag to override dirty check in internal/cli/wt_remove.go
- [X] T021 [US2] Check if CWD is inside target worktree and block (even with --force) in internal/cli/wt_remove.go
- [X] T022 [US2] Warn before removing default branch worktrees (require --force) in internal/cli/wt_remove.go

**Checkpoint**: Dirty worktrees protected, `--force` required to override

---

## Phase 5: User Story 3 - Delete Branch After Removal (Priority: P2)

**Goal**: Remove worktrees AND delete branches with `fa wt remove feature-123 --delete-branch`

**Independent Test**: Run `fa wt remove feature-123 --delete-branch`, verify worktrees and branches gone

### Implementation for User Story 3

- [X] T023 [US3] Add `--delete-branch` flag to remove command in internal/cli/wt_remove.go
- [X] T024 [US3] Check if branch is merged before deletion in internal/git/branch.go
- [X] T025 [US3] Warn about unmerged branches and require --force in internal/cli/wt_remove.go
- [X] T026 [US3] Delete branch from each repo after worktree removal in internal/git/branch.go
- [X] T027 [US3] Skip branch deletion if worktree removal failed in internal/cli/wt_remove.go
- [X] T028 [US3] Display confirmation of both worktree removal and branch deletion in internal/cli/wt_remove.go

**Checkpoint**: `fa wt remove --delete-branch` cleans up both worktrees and branches

---

## Phase 6: User Story 4 - JSON Output (Priority: P3)

**Goal**: Machine-readable output with `fa wt remove --json`

**Independent Test**: Run `fa wt remove feature-123 --json`, parse JSON, verify removal status

### Implementation for User Story 4

- [X] T029 [US4] Add `--json` flag to remove command in internal/cli/wt_remove.go
- [X] T030 [US4] Define JSON schema with per-worktree status in internal/cli/wt_remove.go
- [X] T031 [US4] Include status (removed, skipped, failed) for each worktree in internal/cli/wt_remove.go

**Checkpoint**: `fa wt remove --json` produces valid, parseable JSON

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and error handling

- [X] T032 [P] Handle branch not found with helpful message in internal/cli/wt_remove.go
- [X] T033 [P] Handle partial worktrees (branch in some repos) - remove what exists in internal/cli/wt_remove.go
- [X] T034 [P] Handle locked worktrees with clear error in internal/cli/wt_remove.go
- [X] T035 [P] Handle permission denied with clear message in internal/cli/wt_remove.go
- [X] T036 [P] Validate command is run inside Foundagent workspace in internal/cli/wt_remove.go
- [X] T037 Add help text with examples to wt remove command in internal/cli/wt_remove.go
- [X] T038 Write integration test for wt remove command in internal/cli/wt_remove_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational
  - US1 (Remove All) is P1 - implement first (MVP)
  - US2 (Prevent Dirty) is P1 - depends on US1
  - US3 (Delete Branch) is P2
  - US4 (JSON) is P3
- **Polish (Phase 7)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Remove All) | P1 | Foundational | - |
| US2 (Prevent Dirty) | P1 | US1 | - |
| US3 (Delete Branch) | P2 | US1 | US4 |
| US4 (JSON) | P3 | US1 | US3 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                        Remove All
                              │
                              ▼
                           [US2]
                       Prevent Dirty
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           [US3]                           [US4]
        Delete Branch                      JSON
```

---

## Implementation Strategy

### MVP First (User Stories 1-2)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T007)
3. Complete Phase 3: US1 Remove All (T008-T015)
4. Complete Phase 4: US2 Prevent Dirty (T016-T022)
5. **STOP and VALIDATE**: Safe worktree removal works
6. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US3 (Delete Branch) → Test independently
3. Add US4 (JSON) → Test independently
4. Polish phase

---

## Notes

- Worktrees at `repos/<repo>/worktrees/<branch>/`
- Uses `git worktree remove` for proper cleanup
- Non-destructive by default - dirty worktrees require `--force`
- CWD check prevents removing worktree you're standing in
- Branch deletion only affects local branches, not remote
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
