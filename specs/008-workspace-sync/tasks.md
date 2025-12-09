# Tasks: Workspace Sync

**Input**: Design documents from `/specs/008-workspace-sync/`
**Prerequisites**: spec.md ✅

**Tests**: Not explicitly requested - include minimal validation tests.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US6)
- Include exact file paths in descriptions

## Path Conventions

- **CLI commands**: `internal/cli/`
- **Workspace logic**: `internal/workspace/`
- **Git operations**: `internal/git/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Create command skeleton

- [ ] T001 Create sync command skeleton in internal/cli/sync.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core git fetch/pull/push infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T002 Implement git fetch function (update remote refs) in internal/git/remote.go
- [ ] T003 [P] Implement git pull function (fast-forward only) in internal/git/remote.go
- [ ] T004 [P] Implement git push function in internal/git/remote.go
- [ ] T005 [P] Implement ahead/behind commit count detection in internal/git/remote.go
- [ ] T006 Implement parallel network operation executor in internal/workspace/parallel.go
- [ ] T007 Add `sync` command to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - git network operations available

---

## Phase 3: User Story 1 - Fetch Updates from All Remotes (Priority: P1) 🎯 MVP

**Goal**: Fetch from all repos with `fa sync`

**Independent Test**: Run `fa sync` with repos that have remote changes, verify refs updated

### Implementation for User Story 1

- [ ] T008 [US1] Get list of all repos from workspace in internal/cli/sync.go
- [ ] T009 [US1] Fetch from origin remote for each repo in parallel in internal/cli/sync.go
- [ ] T010 [US1] Display progress for each repo during fetch in internal/cli/sync.go
- [ ] T011 [US1] Detect repos with available updates (commits behind) in internal/cli/sync.go
- [ ] T012 [US1] Display summary showing which repos have updates in internal/cli/sync.go
- [ ] T013 [US1] Display "already up-to-date" for repos with no changes in internal/cli/sync.go

**Checkpoint**: `fa sync` fetches from all remotes and shows update summary

---

## Phase 4: User Story 2 - Pull Updates into Current Worktree (Priority: P1)

**Goal**: Fetch and pull with `fa sync --pull`

**Independent Test**: Run `fa sync --pull` when remotes have new commits, verify worktrees updated

### Implementation for User Story 2

- [ ] T014 [US2] Add `--pull` flag to sync command in internal/cli/sync.go
- [ ] T015 [US2] Determine current branch from CWD or state in internal/cli/sync.go
- [ ] T016 [US2] Pull all worktrees for current branch after fetch in internal/cli/sync.go
- [ ] T017 [US2] Skip worktrees with uncommitted changes (warn and list) in internal/cli/sync.go
- [ ] T018 [US2] Fail gracefully on non-fast-forward (suggest merge/rebase) in internal/cli/sync.go
- [ ] T019 [US2] Add `--stash` flag to stash changes before pull in internal/cli/sync.go
- [ ] T020 [US2] Implement stash before pull, pop after in internal/git/stash.go

**Checkpoint**: `fa sync --pull` updates current branch worktrees

---

## Phase 5: User Story 3 - Handle Network Failures Gracefully (Priority: P1)

**Goal**: Partial success with clear failure reporting

**Independent Test**: Simulate network failure for one repo, verify others sync and failure reported

### Implementation for User Story 3

- [ ] T021 [US3] Collect per-repo results (success/failure) during sync in internal/cli/sync.go
- [ ] T022 [US3] Continue syncing other repos when one fails in internal/cli/sync.go
- [ ] T023 [US3] Display clear error for each failed repo with suggestion in internal/cli/sync.go
- [ ] T024 [US3] Include auth failure hints (SSH keys, credentials) in internal/cli/sync.go
- [ ] T025 [US3] Exit with non-zero code if any repo fails in internal/cli/sync.go

**Checkpoint**: Partial failures don't block successful repos

---

## Phase 6: User Story 4 - Sync Specific Branch (Priority: P2)

**Goal**: Sync a specific branch with `fa sync feature-123`

**Independent Test**: Run `fa sync feature-123 --pull` while on main, verify feature-123 updated

### Implementation for User Story 4

- [ ] T026 [US4] Accept optional branch argument in internal/cli/sync.go
- [ ] T027 [US4] Sync worktrees for specified branch instead of current in internal/cli/sync.go
- [ ] T028 [US4] Skip repos where branch doesn't exist with message in internal/cli/sync.go

**Checkpoint**: `fa sync feature-123` syncs specific branch

---

## Phase 7: User Story 5 - JSON Output (Priority: P2)

**Goal**: Machine-readable output with `fa sync --json`

**Independent Test**: Run `fa sync --json`, parse JSON, verify per-repo status

### Implementation for User Story 5

- [ ] T029 [US5] Add `--json` flag to sync command in internal/cli/sync.go
- [ ] T030 [US5] Define JSON schema with repos array in internal/cli/sync.go
- [ ] T031 [US5] Include name, status, error, refs_updated for each repo in internal/cli/sync.go
- [ ] T032 [US5] Include summary with counts (synced, updated, failed, skipped) in internal/cli/sync.go

**Checkpoint**: `fa sync --json` produces valid, parseable JSON

---

## Phase 8: User Story 6 - Push Local Changes (Priority: P3)

**Goal**: Push all repos with `fa sync --push`

**Independent Test**: Commit in 2 repos, run `fa sync --push`, verify both pushed

### Implementation for User Story 6

- [ ] T033 [US6] Add `--push` flag to sync command in internal/cli/sync.go
- [ ] T034 [US6] Detect repos with local commits ahead of remote in internal/cli/sync.go
- [ ] T035 [US6] Push only repos with unpushed commits in internal/cli/sync.go
- [ ] T036 [US6] Fail gracefully if remote has new commits (suggest pull first) in internal/cli/sync.go
- [ ] T037 [US6] Display "Nothing to push" when no repos have unpushed commits in internal/cli/sync.go

**Checkpoint**: `fa sync --push` pushes repos with local changes

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and polish

- [ ] T038 [P] Handle empty workspace (no repos) with helpful message in internal/cli/sync.go
- [ ] T039 [P] Add `-v` / `--verbose` flag for detailed progress in internal/cli/sync.go
- [ ] T040 [P] Handle detached HEAD worktrees with warning in internal/cli/sync.go
- [ ] T041 [P] Validate command is run inside Foundagent workspace in internal/cli/sync.go
- [ ] T042 Add help text with examples to sync command in internal/cli/sync.go
- [ ] T043 Write integration test for sync command in internal/cli/sync_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational
  - US1 (Fetch) is P1 - implement first (MVP)
  - US2 (Pull) is P1 - depends on US1
  - US3 (Network Failures) is P1 - integrate with US1/US2
  - US4 (Specific Branch) and US5 (JSON) are P2
  - US6 (Push) is P3
- **Polish (Phase 9)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Fetch) | P1 | Foundational | - |
| US2 (Pull) | P1 | US1 | US3 |
| US3 (Network Failures) | P1 | US1 | US2 |
| US4 (Specific Branch) | P2 | US2 | US5 |
| US5 (JSON) | P2 | US1 | US4 |
| US6 (Push) | P3 | US1 | - |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                           Fetch
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
           [US2]           [US3]           [US5]
           Pull        Network Failures    JSON
              │               
              ▼               
           [US4]                           [US6]
        Specific Branch                    Push
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002-T007)
3. Complete Phase 3: US1 Fetch (T008-T013)
4. Complete Phase 4: US2 Pull (T014-T020)
5. Complete Phase 5: US3 Network Failures (T021-T025)
6. **STOP and VALIDATE**: Sync fetch/pull works with graceful failures
7. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US4 (Specific Branch) + US5 (JSON) → Test independently
3. Add US6 (Push) → Test independently
4. Polish phase

---

## Notes

- Uses standard `origin` remote (future: support multiple remotes)
- Bare clones at `repos/.bare/<repo>.git/` store fetched refs
- Worktrees at `repos/worktrees/<repo>/<branch>/` are updated by pull
- Fast-forward only for `--pull` - no auto-merge
- Auth handled by system Git credentials
- Parallel operations with error collection
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
