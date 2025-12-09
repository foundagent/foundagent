# Tasks: Add Repository

**Input**: Design documents from `/specs/002-repo-add/`
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

**Purpose**: Verify infrastructure and create command skeleton

- [X] T001 Verify workspace detection logic exists (from 001-workspace-init)
- [X] T002 [P] Create add command skeleton in internal/cli/add.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core git and repository infrastructure that all stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Implement bare clone function (`git clone --bare`) in internal/git/clone.go
- [X] T004 [P] Implement repository URL parsing and name inference in internal/git/url.go
- [X] T005 [P] Implement worktree creation function (`git worktree add`) in internal/git/worktree.go
- [X] T006 [P] Implement default branch detection from remote in internal/git/remote.go
- [X] T007 Define Repository struct with name, URL, default branch in internal/workspace/repository.go
- [X] T008 Implement workspace detection (find `.foundagent.yaml` up the tree) in internal/workspace/discover.go
- [X] T009 Add `add` command to rootCmd in internal/cli/root.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Add Single Repository (Priority: P1) 🎯 MVP

**Goal**: Clone repo as bare clone and create default branch worktree with `fa add <url>`

**Independent Test**: Run `fa add <public-repo-url>`, verify bare clone at `repos/.bare/`, worktree at `repos/worktrees/<name>/main/`

### Implementation for User Story 1

- [X] T010 [US1] Implement bare clone to `repos/.bare/<name>.git/` in internal/cli/add.go
- [X] T011 [US1] Create worktree for default branch at `repos/worktrees/<name>/<branch>/` in internal/cli/add.go
- [X] T012 [US1] Register repo in `.foundagent/state.json` with metadata in internal/workspace/state.go
- [X] T013 [US1] Register repo in `.foundagent.yaml` config in internal/workspace/config.go
- [X] T014 [US1] Update `.code-workspace` to include new worktree folder in internal/workspace/vscode.go
- [X] T015 [US1] Display progress during clone operation in internal/cli/add.go
- [X] T016 [US1] Display success message with repo name and worktree path in internal/cli/add.go
- [X] T017 [US1] Implement exit code 0 on success, non-zero on failure in internal/cli/add.go

**Checkpoint**: `fa add <url>` clones repo and creates working worktree

---

## Phase 4: User Story 2 - Add Multiple Repositories (Priority: P2)

**Goal**: Clone multiple repos in parallel with `fa add <url1> <url2> <url3>`

**Independent Test**: Run `fa add <url1> <url2>`, verify both cloned, parallel execution faster than sequential

### Implementation for User Story 2

- [X] T018 [US2] Accept multiple URL arguments in add command in internal/cli/add.go
- [X] T019 [US2] Implement parallel cloning with goroutines in internal/cli/add.go
- [X] T020 [US2] Collect and report partial success/failure for each repo in internal/cli/add.go
- [X] T021 [US2] Display summary showing success/failure status for each repo in internal/cli/add.go

**Checkpoint**: `fa add <url1> <url2>` clones repos in parallel with summary

---

## Phase 5: User Story 3 - Add with Custom Name (Priority: P2)

**Goal**: Clone with custom local name using `fa add <url> custom-name`

**Independent Test**: Run `fa add <url> custom-name`, verify bare clone at `repos/.bare/custom-name.git/`

### Implementation for User Story 3

- [X] T022 [US3] Accept optional name argument after URL in internal/cli/add.go
- [X] T023 [US3] Use custom name for bare clone and worktree paths in internal/cli/add.go
- [X] T024 [US3] Register with custom name in config and state in internal/cli/add.go

**Checkpoint**: `fa add <url> api-service` uses custom name for all paths

---

## Phase 6: User Story 4 - Handle Already-Added Repository (Priority: P3)

**Goal**: Skip repos that already exist, optionally re-clone with `--force`

**Independent Test**: Add repo twice, verify skip message on second add

### Implementation for User Story 4

- [X] T025 [US4] Check if repo name already exists before cloning in internal/cli/add.go
- [X] T026 [US4] Skip existing repos with message (not error) in internal/cli/add.go
- [X] T027 [US4] Add `--force` flag to re-clone existing repos in internal/cli/add.go
- [X] T028 [US4] Implement re-clone logic preserving worktrees when possible in internal/cli/add.go

**Checkpoint**: Idempotent add - running twice is safe

---

## Phase 7: User Story 5 - JSON Output (Priority: P3)

**Goal**: Machine-readable output with `fa add <url> --json`

**Independent Test**: Run `fa add <url> --json`, parse JSON, verify structure

### Implementation for User Story 5

- [X] T029 [US5] Add `--json` flag to add command in internal/cli/add.go
- [X] T030 [US5] Define JSON output schema (repo name, path, worktree path, status) in internal/cli/add.go
- [X] T031 [US5] Output JSON array when adding multiple repos in internal/cli/add.go
- [X] T032 [US5] Include error details in JSON for failed repos in internal/cli/add.go

**Checkpoint**: `fa add --json` produces valid, parseable JSON

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, validation, and error handling

- [X] T033 [P] Validate command is run inside Foundagent workspace in internal/cli/add.go
- [X] T034 [P] Handle invalid/malformed URLs with clear error in internal/git/url.go
- [X] T035 [P] Handle auth failures with SSH/credential hints in internal/git/clone.go
- [X] T036 [P] Handle network failures with retry suggestion in internal/git/clone.go
- [X] T037 [P] Handle name collision when custom name already exists in internal/cli/add.go
- [X] T038 [P] Handle both SSH and HTTPS URL formats in internal/git/url.go
- [X] T039 [P] Handle repos with/without `.git` suffix in URL in internal/git/url.go
- [X] T040 [P] Handle empty repos (no commits) with warning in internal/cli/add.go
- [X] T041 Add help text with examples to add command in internal/cli/add.go
- [X] T042 Write integration test for add command in internal/cli/add_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational
  - US1 (Single Repo) is P1 - implement first (MVP)
  - US2 (Multiple) and US3 (Custom Name) are P2 - can follow US1
  - US4 (Already Added) and US5 (JSON) are P3 - can follow US1
- **Polish (Phase 8)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Single) | P1 | Foundational | - |
| US2 (Multiple) | P2 | US1 | US3 |
| US3 (Custom Name) | P2 | US1 | US2 |
| US4 (Already Added) | P3 | US1 | US5 |
| US5 (JSON) | P3 | US1 | US4 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                       Single Repo Add
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
           [US2]           [US3]         [US4] [US5]
         Multiple       Custom Name    Already  JSON
                                       Added
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T009)
3. Complete Phase 3: User Story 1 (T010-T017)
4. **STOP and VALIDATE**: `fa add <url>` works
5. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US2 (Multiple) + US3 (Custom Name) → Test independently
3. Add US4 (Already Added) + US5 (JSON) → Test independently
4. Polish phase

---

## Notes

- Bare clones stored at `repos/.bare/<name>.git/`
- Worktrees at `repos/worktrees/<repo>/<branch>/`
- URL parsing must handle: `git@github.com:org/repo.git`, `https://github.com/org/repo.git`, with/without `.git`
- Git credentials handled by system (SSH agent, credential helper)
- Parallel cloning uses goroutines with error collection
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
