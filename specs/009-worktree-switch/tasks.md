# Tasks: Worktree Switch

**Input**: Design documents from `/specs/009-worktree-switch/`
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
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Create command skeleton

- [X] T001 Create switch subcommand skeleton in internal/cli/wt_switch.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core workspace file update infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T002 Implement current branch detection from workspace file in internal/workspace/vscode.go
- [X] T003 [P] Implement available branches list from worktrees in internal/workspace/worktree.go
- [X] T004 [P] Implement workspace file folder replacement in internal/workspace/vscode.go
- [X] T005 Add `worktree switch` and `wt switch` commands to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - workspace file manipulation available

---

## Phase 3: User Story 1 - Switch VS Code Workspace (Priority: P1) 🎯 MVP

**Goal**: Update workspace file to point to different branch with `fa wt switch feature-123`

**Independent Test**: Create worktrees for main and feature-123, run `fa wt switch feature-123`, verify workspace file updated

### Implementation for User Story 1

- [X] T006 [US1] Validate target branch has worktrees in internal/cli/wt_switch.go
- [X] T007 [US1] Get list of worktrees for target branch in internal/cli/wt_switch.go
- [X] T008 [US1] Replace current branch worktree folders with target branch folders in internal/workspace/vscode.go
- [X] T009 [US1] Preserve non-worktree folders in workspace file in internal/workspace/vscode.go
- [X] T010 [US1] Preserve workspace settings in workspace file in internal/workspace/vscode.go
- [X] T011 [US1] Update `.foundagent/state.json` to track current branch in internal/workspace/state.go
- [X] T012 [US1] Display confirmation with workspace file path in internal/cli/wt_switch.go
- [X] T013 [US1] Handle already on target branch (no-op with message) in internal/cli/wt_switch.go

**Checkpoint**: `fa wt switch feature-123` updates workspace file

---

## Phase 4: User Story 2 - Warn About Uncommitted Changes (Priority: P1)

**Goal**: Warn (but don't block) when current worktrees have uncommitted changes

**Independent Test**: Make changes, run `fa wt switch`, verify warning shown but switch proceeds

### Implementation for User Story 2

- [X] T014 [US2] Check current worktrees for uncommitted changes in internal/cli/wt_switch.go
- [X] T015 [US2] Display warning listing dirty worktrees in internal/cli/wt_switch.go
- [X] T016 [US2] Proceed with switch despite warning (changes remain in original) in internal/cli/wt_switch.go
- [X] T017 [US2] Add `--quiet` flag to suppress warnings in internal/cli/wt_switch.go

**Checkpoint**: Users warned about uncommitted work before switch

---

## Phase 5: User Story 3 - Create Worktree If Doesn't Exist (Priority: P2)

**Goal**: Create and switch with `fa wt switch new-feature --create`

**Independent Test**: Run `fa wt switch new-feature --create`, verify worktrees created and switched

### Implementation for User Story 3

- [X] T018 [US3] Add `--create` flag to switch command in internal/cli/wt_switch.go
- [X] T019 [US3] Add `--from` flag (only with --create) for source branch in internal/cli/wt_switch.go
- [X] T020 [US3] Invoke worktree creation logic from wt create in internal/cli/wt_switch.go
- [X] T021 [US3] Validate --from branch exists if specified in internal/cli/wt_switch.go
- [X] T022 [US3] Switch to newly created worktrees after creation in internal/cli/wt_switch.go
- [X] T023 [US3] Error without --create if worktrees don't exist in internal/cli/wt_switch.go

**Checkpoint**: `fa wt switch --create` creates and switches in one command

---

## Phase 6: User Story 4 - JSON Output (Priority: P3)

**Goal**: Machine-readable output with `fa wt switch --json`

**Independent Test**: Run `fa wt switch feature-123 --json`, parse JSON, verify structure

### Implementation for User Story 4

- [X] T024 [US4] Add `--json` flag to switch command in internal/cli/wt_switch.go
- [X] T025 [US4] Define JSON schema (switched_to, previous_branch, workspace_file, warnings) in internal/cli/wt_switch.go
- [X] T026 [US4] Include error details in JSON on failure in internal/cli/wt_switch.go

**Checkpoint**: `fa wt switch --json` produces valid, parseable JSON

---

## Phase 7: User Story 5 - List Available Branches (Priority: P3)

**Goal**: Show available branches with `fa wt switch` (no args)

**Independent Test**: Run `fa wt switch` with no args, verify list of branches shown

### Implementation for User Story 5

- [X] T027 [US5] Detect no-args invocation in internal/cli/wt_switch.go
- [X] T028 [US5] List all branches that have worktrees in internal/cli/wt_switch.go
- [X] T029 [US5] Mark current branch in the list in internal/cli/wt_switch.go

**Checkpoint**: `fa wt switch` shows available branches when no args

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and error handling

- [X] T030 [P] Handle partial worktrees (branch in some repos) with warning in internal/cli/wt_switch.go
- [X] T031 [P] Handle missing workspace file - create it in internal/workspace/vscode.go
- [X] T032 [P] Handle no worktrees in workspace with helpful message in internal/cli/wt_switch.go
- [X] T033 [P] Validate command is run inside Foundagent workspace in internal/cli/wt_switch.go
- [X] T034 Add help text with examples to wt switch command in internal/cli/wt_switch.go
- [X] T035 Write integration test for wt switch command in internal/cli/wt_switch_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational
  - US1 (Switch) is P1 - implement first (MVP)
  - US2 (Warn Uncommitted) is P1 - depends on US1
  - US3 (Create) is P2
  - US4 (JSON) and US5 (List) are P3
- **Polish (Phase 8)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Switch) | P1 | Foundational | - |
| US2 (Warn Uncommitted) | P1 | US1 | - |
| US3 (Create) | P2 | US1 | US4, US5 |
| US4 (JSON) | P3 | US1 | US3, US5 |
| US5 (List) | P3 | US1 | US3, US4 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                          Switch
                              │
                              ▼
                           [US2]
                     Warn Uncommitted
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
           [US3]           [US4]           [US5]
          Create           JSON            List
```

---

## Implementation Strategy

### MVP First (User Stories 1-2)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002-T005)
3. Complete Phase 3: US1 Switch (T006-T013)
4. Complete Phase 4: US2 Warn Uncommitted (T014-T017)
5. **STOP and VALIDATE**: Branch switching works with warnings
6. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US3 (Create) → Test independently
3. Add US4 (JSON) + US5 (List) → Test independently
4. Polish phase

---

## Notes

- Switching only updates workspace file - actual worktrees unchanged
- Worktrees at `repos/worktrees/<repo>/<branch>/`
- Users reload VS Code after switch (or VS Code auto-detects)
- Non-destructive - uncommitted changes remain in original worktrees
- Current branch tracked in `.foundagent/state.json`
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
