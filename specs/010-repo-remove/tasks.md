# Tasks: Repo Remove

**Input**: Design documents from `/specs/010-repo-remove/`
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

- [ ] T001 Create remove command skeleton in internal/cli/remove.go
- [ ] T002 [P] Add `rm` alias for remove command in internal/cli/root.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core removal and safety check infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Implement repo existence check by name in internal/workspace/repository.go
- [ ] T004 [P] Reuse dirty worktree detection from wt remove (internal/git/status.go)
- [ ] T005 [P] Reuse CWD-inside-worktree detection from wt remove (internal/workspace/worktree.go)
- [ ] T006 Implement bare clone deletion in internal/git/clone.go
- [ ] T007 Add `remove` and `rm` commands to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - repo removal infrastructure available

---

## Phase 3: User Story 1 - Remove Repo from Workspace (Priority: P1) 🎯 MVP

**Goal**: Remove repo completely with `fa remove api`

**Independent Test**: Add repo, run `fa remove api`, verify repo gone from config, bare clone deleted, worktrees deleted

### Implementation for User Story 1

- [ ] T008 [US1] Require repo name argument in internal/cli/remove.go
- [ ] T009 [US1] Verify repo exists in workspace in internal/cli/remove.go
- [ ] T010 [US1] Remove repo from `.foundagent.yaml` config in internal/config/writer.go
- [ ] T011 [US1] Delete bare clone at `repos/.bare/<name>.git/` in internal/cli/remove.go
- [ ] T012 [US1] Delete all worktrees at `repos/worktrees/<name>/` using git worktree remove in internal/cli/remove.go
- [ ] T013 [US1] Update `.code-workspace` to remove worktree folders in internal/cli/remove.go
- [ ] T014 [US1] Update `.foundagent/state.json` to remove repo entries in internal/workspace/state.go
- [ ] T015 [US1] Display confirmation of what was removed in internal/cli/remove.go

**Checkpoint**: `fa remove api` completely removes repo from workspace

---

## Phase 4: User Story 2 - Prevent Removal with Uncommitted Changes (Priority: P1)

**Goal**: Block removal of repos with dirty worktrees unless `--force`

**Independent Test**: Make changes in repo's worktree, run `fa remove`, verify blocked

### Implementation for User Story 2

- [ ] T016 [US2] Check all repo's worktrees for uncommitted changes in internal/cli/remove.go
- [ ] T017 [US2] Check all repo's worktrees for untracked files in internal/cli/remove.go
- [ ] T018 [US2] Block removal if any worktree is dirty in internal/cli/remove.go
- [ ] T019 [US2] List all dirty worktrees in error message in internal/cli/remove.go
- [ ] T020 [US2] Add `--force` flag to override dirty check in internal/cli/remove.go
- [ ] T021 [US2] Check if CWD is inside repo's worktree and block in internal/cli/remove.go

**Checkpoint**: Dirty worktrees protected, `--force` required to override

---

## Phase 5: User Story 3 - Remove Only from Config (Priority: P2)

**Goal**: Remove from config but keep files with `fa remove api --config-only`

**Independent Test**: Run `fa remove api --config-only`, verify config updated but files remain

### Implementation for User Story 3

- [ ] T022 [US3] Add `--config-only` flag to remove command in internal/cli/remove.go
- [ ] T023 [US3] Only remove from `.foundagent.yaml` when --config-only in internal/cli/remove.go
- [ ] T024 [US3] Skip bare clone and worktree deletion in internal/cli/remove.go
- [ ] T025 [US3] Still update workspace file (remove folders) in internal/cli/remove.go
- [ ] T026 [US3] Display message that files were kept in internal/cli/remove.go

**Checkpoint**: `fa remove --config-only` removes from config but keeps files

---

## Phase 6: User Story 4 - JSON Output (Priority: P3)

**Goal**: Machine-readable output with `fa remove --json`

**Independent Test**: Run `fa remove api --json`, parse JSON, verify removal status

### Implementation for User Story 4

- [ ] T027 [US4] Add `--json` flag to remove command in internal/cli/remove.go
- [ ] T028 [US4] Define JSON schema (repo_name, removed_from_config, files_deleted, worktrees_deleted) in internal/cli/remove.go
- [ ] T029 [US4] Include error details in JSON on failure in internal/cli/remove.go

**Checkpoint**: `fa remove --json` produces valid, parseable JSON

---

## Phase 7: User Story 5 - Remove Multiple Repos (Priority: P3)

**Goal**: Remove several repos at once with `fa remove api web`

**Independent Test**: Run `fa remove api web`, verify both removed

### Implementation for User Story 5

- [ ] T030 [US5] Accept multiple repo name arguments in internal/cli/remove.go
- [ ] T031 [US5] Process each repo and collect results in internal/cli/remove.go
- [ ] T032 [US5] Skip dirty repos, continue with clean (unless --force) in internal/cli/remove.go
- [ ] T033 [US5] Display summary of removed/skipped/failed repos in internal/cli/remove.go

**Checkpoint**: `fa remove api web` removes multiple repos

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and error handling

- [ ] T034 [P] Handle repo not found with helpful message in internal/cli/remove.go
- [ ] T035 [P] Handle orphaned repo (cloned but not in config) with confirmation in internal/cli/remove.go
- [ ] T036 [P] Suggest similar repo names on typo in internal/cli/remove.go
- [ ] T037 [P] Handle locked worktrees with clear error in internal/cli/remove.go
- [ ] T038 [P] Handle permission denied with clear message in internal/cli/remove.go
- [ ] T039 [P] Validate command is run inside Foundagent workspace in internal/cli/remove.go
- [ ] T040 Add help text with examples to remove command in internal/cli/remove.go
- [ ] T041 Write integration test for remove command in internal/cli/remove_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational
  - US1 (Remove) is P1 - implement first (MVP)
  - US2 (Prevent Dirty) is P1 - depends on US1
  - US3 (Config Only) is P2
  - US4 (JSON) and US5 (Multiple) are P3
- **Polish (Phase 8)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Remove) | P1 | Foundational | - |
| US2 (Prevent Dirty) | P1 | US1 | - |
| US3 (Config Only) | P2 | US1 | US4, US5 |
| US4 (JSON) | P3 | US1 | US3, US5 |
| US5 (Multiple) | P3 | US2 | US3, US4 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                          Remove
                              │
                              ▼
                           [US2]
                       Prevent Dirty
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
           [US3]           [US4]           [US5]
        Config Only        JSON           Multiple
```

---

## Implementation Strategy

### MVP First (User Stories 1-2)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T007)
3. Complete Phase 3: US1 Remove (T008-T015)
4. Complete Phase 4: US2 Prevent Dirty (T016-T021)
5. **STOP and VALIDATE**: Safe repo removal works
6. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US3 (Config Only) → Test independently
3. Add US4 (JSON) + US5 (Multiple) → Test independently
4. Polish phase

---

## Notes

- Repos identified by local name (from config), not URL
- Bare clones at `repos/.bare/<name>.git/`
- Worktrees at `repos/worktrees/<name>/<branch>/`
- Non-destructive by default - dirty worktrees require `--force`
- Removal is permanent (no undo)
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
