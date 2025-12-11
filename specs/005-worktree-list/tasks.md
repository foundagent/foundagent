# Tasks: Worktree List

**Input**: Design documents from `/specs/005-worktree-list/`
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

- [X] T001 Create list subcommand skeleton in internal/cli/wt_list.go
- [X] T002 [P] Add `ls` alias for list subcommand in internal/cli/worktree.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core worktree discovery and status detection

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Implement worktree discovery from filesystem in internal/workspace/worktree.go
- [X] T004 [P] Implement current worktree detection (based on CWD) in internal/workspace/worktree.go
- [X] T005 [P] Implement git status detection (clean/modified/untracked/conflict) in internal/git/status.go
- [X] T006 Implement parallel status detection across worktrees in internal/workspace/worktree.go
- [X] T007 Add `worktree list`, `wt list`, and `wt ls` commands to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - worktrees can be discovered with status

---

## Phase 3: User Story 1 - List All Worktrees (Priority: P1) 🎯 MVP

**Goal**: Display all worktrees grouped by branch with `fa wt list`

**Independent Test**: Create worktrees for 2 branches across 3 repos, run `fa wt list`, verify all 6 shown

### Implementation for User Story 1

- [X] T008 [US1] Discover all worktrees from `repos/<repo>/worktrees/<branch>/` structure in internal/cli/wt_list.go
- [X] T009 [US1] Group worktrees by branch name in internal/cli/wt_list.go
- [X] T010 [US1] Display branch name as group header in internal/cli/wt_list.go
- [X] T011 [US1] Display repo name and path for each worktree in internal/cli/wt_list.go
- [X] T012 [US1] Indicate current/active worktree with marker (e.g., `*`) in internal/cli/wt_list.go
- [X] T013 [US1] Sort branches alphabetically, then repos alphabetically in internal/cli/wt_list.go

**Checkpoint**: `fa wt list` shows all worktrees organized by branch

---

## Phase 4: User Story 2 - JSON Output (Priority: P2)

**Goal**: Machine-readable output with `fa wt list --json`

**Independent Test**: Run `fa wt list --json`, parse JSON, verify structure with all worktree details

### Implementation for User Story 2

- [X] T014 [US2] Add `--json` flag to list command in internal/cli/wt_list.go
- [X] T015 [US2] Define JSON schema with worktrees array in internal/cli/wt_list.go
- [X] T016 [US2] Include branch, repo, path, is_current, status for each worktree in internal/cli/wt_list.go
- [X] T017 [US2] Include workspace metadata (name, total_worktrees, total_branches) in internal/cli/wt_list.go

**Checkpoint**: `fa wt list --json` produces valid, parseable JSON

---

## Phase 5: User Story 3 - Show Worktree Status (Priority: P2)

**Goal**: Display status indicators for dirty worktrees

**Independent Test**: Make changes in one worktree, run `fa wt list`, verify status indicator shown

### Implementation for User Story 3

- [X] T018 [US3] Detect uncommitted changes in each worktree in internal/git/status.go
- [X] T019 [US3] Detect untracked files in each worktree in internal/git/status.go
- [X] T020 [US3] Detect merge conflicts in each worktree in internal/git/status.go
- [X] T021 [US3] Display status indicator (e.g., `[modified]`, `[untracked]`) for dirty worktrees in internal/cli/wt_list.go
- [X] T022 [US3] Run status detection in parallel for performance in internal/cli/wt_list.go

**Checkpoint**: Dirty worktrees clearly marked with status indicators

---

## Phase 6: User Story 4 - Filter by Branch (Priority: P3)

**Goal**: Filter output to specific branch with `fa wt list feature-x`

**Independent Test**: Run `fa wt list feature-x`, verify only feature-x worktrees shown

### Implementation for User Story 4

- [X] T023 [US4] Accept optional branch argument for filtering in internal/cli/wt_list.go
- [X] T024 [US4] Filter worktrees to match specified branch in internal/cli/wt_list.go
- [X] T025 [US4] Display message when no worktrees found for branch in internal/cli/wt_list.go

**Checkpoint**: `fa wt list feature-x` filters to specific branch

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and error handling

- [X] T026 [P] Handle empty workspace (no repos) with helpful message in internal/cli/wt_list.go
- [X] T027 [P] Handle repos with no worktrees with helpful message in internal/cli/wt_list.go
- [X] T028 [P] Handle corrupted worktrees gracefully (show error, continue) in internal/cli/wt_list.go
- [X] T029 [P] Validate command is run inside Foundagent workspace in internal/cli/wt_list.go
- [X] T030 Add help text with examples to wt list command in internal/cli/wt_list.go
- [X] T031 Write integration test for wt list command in internal/cli/wt_list_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational
  - US1 (List All) is P1 - implement first (MVP)
  - US2 (JSON) and US3 (Status) are P2
  - US4 (Filter) is P3
- **Polish (Phase 7)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (List All) | P1 | Foundational | - |
| US2 (JSON) | P2 | US1 | US3, US4 |
| US3 (Status) | P2 | US1 | US2, US4 |
| US4 (Filter) | P3 | US1 | US2, US3 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                        List All
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
           [US2]           [US3]           [US4]
           JSON           Status          Filter
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T007)
3. Complete Phase 3: US1 List All (T008-T013)
4. **STOP and VALIDATE**: `fa wt list` shows worktrees
5. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US2 (JSON) + US3 (Status) → Test independently
3. Add US4 (Filter) → Test independently
4. Polish phase

---

## Notes

- Worktrees at `repos/<repo>/worktrees/<branch>/`
- Current worktree detected by checking if CWD is inside a worktree path
- Status detection runs in parallel for performance (up to 50 worktrees in <5s)
- Clean is implicit (no indicator) - only dirty states shown
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
