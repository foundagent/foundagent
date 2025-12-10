# Tasks: Workspace Status

**Input**: Design documents from `/specs/007-workspace-status/`
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

**Purpose**: Create command skeleton

- [X] T001 Create status command skeleton in internal/cli/status.go
- [X] T002 [P] Add `st` alias for status command in internal/cli/root.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core status collection infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Implement repo clone status detection (bare clone exists in `repos/.bare/`) in internal/workspace/repository.go
- [X] T004 [P] Implement config-state comparison (repos in config vs cloned) in internal/workspace/reconcile.go
- [X] T005 [P] Reuse worktree discovery from wt list (internal/workspace/worktree.go)
- [X] T006 [P] Reuse git status detection from wt list (internal/git/status.go)
- [X] T007 Add `status` and `st` commands to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - status data collection available

---

## Phase 3: User Story 1 - View Workspace Overview (Priority: P1) 🎯 MVP

**Goal**: Quick overview with `fa status` showing repos, worktrees, dirty state

**Independent Test**: Run `fa status` with 3 repos and 2 branches, verify overview displayed

### Implementation for User Story 1

- [X] T008 [US1] Display workspace name from config in internal/cli/status.go
- [X] T009 [US1] Display total count of configured repos in internal/cli/status.go
- [X] T010 [US1] Display total count of worktrees in internal/cli/status.go
- [X] T011 [US1] Display total count of branches with worktrees in internal/cli/status.go
- [X] T012 [US1] List all repos with clone status in internal/cli/status.go
- [X] T013 [US1] List all worktrees grouped by branch in internal/cli/status.go
- [X] T014 [US1] Indicate current worktree based on CWD in internal/cli/status.go

**Checkpoint**: `fa status` shows complete workspace overview

---

## Phase 4: User Story 2 - Check for Uncommitted Work (Priority: P1)

**Goal**: Clear visibility into dirty worktrees

**Independent Test**: Make changes in 2 worktrees, run `fa status`, verify both marked

### Implementation for User Story 2

- [X] T015 [US2] Detect uncommitted changes in each worktree in internal/cli/status.go
- [X] T016 [US2] Detect untracked files in each worktree in internal/cli/status.go
- [X] T017 [US2] Display status indicator for dirty worktrees in internal/cli/status.go
- [X] T018 [US2] Display "All worktrees clean" when no dirty worktrees in internal/cli/status.go
- [X] T019 [US2] Run status detection in parallel for performance in internal/cli/status.go

**Checkpoint**: Dirty worktrees clearly visible at a glance

---

## Phase 5: User Story 3 - JSON Output for AI Agents (Priority: P1)

**Goal**: Complete structured output with `fa status --json`

**Independent Test**: Run `fa status --json`, parse JSON, verify complete workspace state

### Implementation for User Story 3

- [X] T020 [US3] Add `--json` flag to status command in internal/cli/status.go
- [X] T021 [US3] Define JSON schema with workspace object in internal/cli/status.go
- [X] T022 [US3] Include repos array with name, url, clone_status, in_config in internal/cli/status.go
- [X] T023 [US3] Include worktrees array with branch, repo, path, status, is_current in internal/cli/status.go
- [X] T024 [US3] Include summary object with counts and has_uncommitted_changes in internal/cli/status.go

**Checkpoint**: `fa status --json` provides complete workspace state for AI agents

---

## Phase 6: User Story 4 - Config vs State Sync Check (Priority: P2)

**Goal**: Show discrepancies between config and actual state

**Independent Test**: Add repo to config without cloning, run `fa status`, verify shown as not cloned

### Implementation for User Story 4

- [X] T025 [US4] Detect repos in config but not cloned (missing bare clone) in internal/cli/status.go
- [X] T026 [US4] Detect repos cloned but not in config (orphaned) in internal/cli/status.go
- [X] T027 [US4] Display `[not cloned]` indicator with hint to run `fa add` in internal/cli/status.go
- [X] T028 [US4] Display `[not in config]` indicator with hint to update config in internal/cli/status.go
- [X] T029 [US4] Display "Config in sync" when no discrepancies in internal/cli/status.go

**Checkpoint**: Config-state sync issues clearly reported

---

## Phase 7: User Story 5 - Verbose Status Details (Priority: P3)

**Goal**: Detailed output with `fa status -v` showing changed files

**Independent Test**: Make changes, run `fa status -v`, verify file names shown

### Implementation for User Story 5

- [X] T030 [US5] Add `-v` / `--verbose` flag to status command in internal/cli/status.go
- [X] T031 [US5] Display list of modified files for each dirty worktree in internal/cli/status.go
- [X] T032 [US5] Display file counts (modified, added, deleted, untracked) in internal/cli/status.go
- [X] T033 [US5] Display branch tracking info (ahead/behind) if available in internal/cli/status.go

**Checkpoint**: `fa status -v` shows detailed file-level information

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and error handling

- [X] T034 [P] Handle empty workspace (no repos) with helpful message in internal/cli/status.go
- [X] T035 [P] Handle repos with no worktrees with indicator in internal/cli/status.go
- [X] T036 [P] Handle corrupted worktrees gracefully in internal/cli/status.go
- [X] T037 [P] Validate command is run inside Foundagent workspace in internal/cli/status.go
- [X] T038 Add help text with examples to status command in internal/cli/status.go
- [X] T039 Write integration test for status command in internal/cli/status_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational
  - US1 (Overview) is P1 - implement first (MVP)
  - US2 (Uncommitted) is P1 - depends on US1
  - US3 (JSON) is P1 - depends on US1/US2
  - US4 (Sync Check) is P2
  - US5 (Verbose) is P3
- **Polish (Phase 8)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Overview) | P1 | Foundational | - |
| US2 (Uncommitted) | P1 | US1 | US3 |
| US3 (JSON) | P1 | US1 | US2 |
| US4 (Sync Check) | P2 | US1 | US5 |
| US5 (Verbose) | P3 | US2 | US4 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                         Overview
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           [US2]                           [US3]
        Uncommitted                        JSON
              │                               
              ├───────────────┬───────────────┐
              ▼               ▼               ▼
           [US4]           [US5]
         Sync Check       Verbose
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T007)
3. Complete Phase 3: US1 Overview (T008-T014)
4. Complete Phase 4: US2 Uncommitted (T015-T019)
5. Complete Phase 5: US3 JSON (T020-T024)
6. **STOP and VALIDATE**: Status command works with JSON for AI agents
7. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US4 (Sync Check) → Test independently
3. Add US5 (Verbose) → Test independently
4. Polish phase

---

## Notes

- Status is read-only - no changes to filesystem or config
- Bare clones at `repos/.bare/<repo>.git/`
- Worktrees at `repos/worktrees/<repo>/<branch>/`
- Status checks are local only (no network) for speed
- Parallel status detection for performance
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
