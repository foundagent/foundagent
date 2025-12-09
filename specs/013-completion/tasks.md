# Tasks: Shell Completion

**Input**: Design documents from `/specs/013-completion/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Tests included per constitution requirement (TDD methodology, table-driven tests).

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Include exact file paths in descriptions

## Path Conventions (from plan.md)

- **CLI commands**: `internal/cli/`
- **Workspace logic**: `internal/workspace/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Verify existing infrastructure supports completion command

- [ ] T001 Verify Cobra dependency in go.mod supports completion (v1.7+)
- [ ] T002 [P] Create completion command skeleton in internal/cli/completion.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core completion infrastructure that all shells depend on

**⚠️ CRITICAL**: No shell-specific work can begin until this phase is complete

- [ ] T003 Implement shell validation logic in internal/cli/completion.go (validate "bash", "zsh", "fish", "powershell")
- [ ] T004 Implement error handling for invalid shell argument (E001 error code per constitution)
- [ ] T005 [P] Create completion_helpers.go with workspace discovery wrapper for dynamic completions in internal/cli/completion_helpers.go
- [ ] T006 Add `completion` command to rootCmd in internal/cli/root.go (or appropriate init location)

**Checkpoint**: Foundation ready - shell-specific and dynamic completion work can proceed

---

## Phase 3: User Story 1 - Bash Completion (Priority: P1) 🎯 MVP

**Goal**: Generate valid Bash completion script with `fa completion bash`

**Independent Test**: Run `fa completion bash`, source output, verify `fa <TAB>` shows commands

### Tests for User Story 1

- [ ] T007 [P] [US1] Table-driven test: bash generates non-empty script containing "complete" in internal/cli/completion_test.go
- [ ] T008 [P] [US1] Test: bash script includes installation instructions header in internal/cli/completion_test.go

### Implementation for User Story 1

- [ ] T009 [US1] Implement GenBashCompletionV2 call with descriptions in internal/cli/completion.go
- [ ] T010 [US1] Add Bash-specific installation header (current session, Linux, macOS with/without Homebrew) in internal/cli/completion.go
- [ ] T011 [US1] Add Bash alias setup instructions (complete -F for fa) in installation header

**Checkpoint**: `fa completion bash` produces sourceable script, static completions work

---

## Phase 4: User Story 2 - Zsh Completion (Priority: P1)

**Goal**: Generate valid Zsh completion script with `fa completion zsh`

**Independent Test**: Run `fa completion zsh`, install script, verify `fa <TAB>` shows commands with descriptions

### Tests for User Story 2

- [ ] T012 [P] [US2] Table-driven test: zsh generates script containing "#compdef" in internal/cli/completion_test.go
- [ ] T013 [P] [US2] Test: zsh script includes installation instructions header in internal/cli/completion_test.go

### Implementation for User Story 2

- [ ] T014 [US2] Implement GenZshCompletion call in internal/cli/completion.go
- [ ] T015 [US2] Add Zsh-specific installation header (fpath, oh-my-zsh, compdef for alias) in internal/cli/completion.go

**Checkpoint**: `fa completion zsh` produces valid script, static completions work

---

## Phase 5: User Story 3 - Fish Completion (Priority: P2)

**Goal**: Generate valid Fish completion script with `fa completion fish`

**Independent Test**: Run `fa completion fish`, save to completions dir, verify `fa <TAB>` works

### Tests for User Story 3

- [ ] T016 [P] [US3] Table-driven test: fish generates script containing "complete -c" in internal/cli/completion_test.go

### Implementation for User Story 3

- [ ] T017 [US3] Implement GenFishCompletion call with descriptions in internal/cli/completion.go
- [ ] T018 [US3] Add Fish-specific installation header in internal/cli/completion.go

**Checkpoint**: `fa completion fish` produces valid script

---

## Phase 6: User Story 4 - PowerShell Completion (Priority: P2)

**Goal**: Generate valid PowerShell completion script with `fa completion powershell`

**Independent Test**: Run `fa completion powershell`, add to profile, verify `fa <TAB>` works

### Tests for User Story 4

- [ ] T019 [P] [US4] Table-driven test: powershell generates script containing "Register-ArgumentCompleter" in internal/cli/completion_test.go

### Implementation for User Story 4

- [ ] T020 [US4] Implement GenPowerShellCompletionWithDesc call in internal/cli/completion.go
- [ ] T021 [US4] Add PowerShell-specific installation header in internal/cli/completion.go

**Checkpoint**: `fa completion powershell` produces valid script

---

## Phase 7: User Story 5 - Dynamic Completions (Priority: P2)

**Goal**: Context-aware completions for worktree names, repo names based on workspace state

**Independent Test**: Create worktrees, type `fa wt switch <TAB>`, verify worktree names appear

### Tests for User Story 5

- [ ] T022 [P] [US5] Test: getWorktreeCompletions returns names from mock workspace in internal/cli/completion_helpers_test.go
- [ ] T023 [P] [US5] Test: getRepoCompletions returns repo names from mock workspace in internal/cli/completion_helpers_test.go
- [ ] T024 [P] [US5] Test: completion helpers return empty list when outside workspace (graceful degradation) in internal/cli/completion_helpers_test.go

### Implementation for User Story 5

- [ ] T025 [US5] Implement getWorktreeCompletions() with workspace.Discover() and prefix filtering in internal/cli/completion_helpers.go
- [ ] T026 [US5] Implement getRepoCompletions() with workspace.Discover() and prefix filtering in internal/cli/completion_helpers.go
- [ ] T027 [US5] Add ValidArgsFunction to worktree switch command (existing cmd file)
- [ ] T028 [US5] Add ValidArgsFunction to worktree remove command (existing cmd file)
- [ ] T029 [US5] Add ValidArgsFunction to repo remove command (existing cmd file)
- [ ] T030 [US5] Ensure ShellCompDirectiveNoFileComp is returned to prevent file fallback

**Checkpoint**: Dynamic completions work for worktree/repo names when in workspace, gracefully empty when outside

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [ ] T031 [P] Add --help examples showing installation for each shell in internal/cli/completion.go
- [ ] T032 [P] Verify both `fa` and `foundagent` commands work with generated scripts
- [ ] T033 Run all completion tests with `go test ./internal/cli/... -run Completion`
- [ ] T034 Run quickstart.md validation steps manually
- [ ] T035 [P] Update any CLI reference documentation if maintained separately

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - verify existing infrastructure
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 (Bash) and US2 (Zsh) are P1 - implement first
  - US3 (Fish) and US4 (PowerShell) are P2 - can be parallel or after P1
  - US5 (Dynamic) is P2 - can start after Foundational
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Bash) | P1 | Foundational | US2, US3, US4, US5 |
| US2 (Zsh) | P1 | Foundational | US1, US3, US4, US5 |
| US3 (Fish) | P2 | Foundational | US1, US2, US4, US5 |
| US4 (PowerShell) | P2 | Foundational | US1, US2, US3, US5 |
| US5 (Dynamic) | P2 | Foundational | US1, US2, US3, US4 |

### Within Each User Story

1. Tests MUST be written and FAIL before implementation (TDD)
2. Core implementation before extras (installation header)
3. Verify independently at checkpoint

### Parallel Opportunities

```
After Foundational (Phase 2) completes, ALL user stories can start in parallel:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
    ┌─────────┬─────────┬─────────┬─────────┬─────────┐
    ▼         ▼         ▼         ▼         ▼         
  [US1]     [US2]     [US3]     [US4]     [US5]
  Bash      Zsh       Fish      PS        Dynamic
```

---

## Parallel Example: Tests (within story)

```bash
# All tests for a user story can be written in parallel:
T007 + T008  # US1 tests - both in same file but independent test functions
T012 + T013  # US2 tests
T022 + T023 + T024  # US5 tests - all helper tests
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T006)
3. Complete Phase 3: User Story 1 - Bash (T007-T011)
4. Complete Phase 4: User Story 2 - Zsh (T012-T015)
5. **STOP and VALIDATE**: Both P1 shells work
6. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US3 (Fish) → Test independently
3. Add US4 (PowerShell) → Test independently
4. Add US5 (Dynamic) → Test independently
5. Polish phase

---

## Notes

- All completion tests are table-driven per constitution
- Cobra handles most complexity - we wrap with headers and dynamic helpers
- Dynamic completions read local files only (<500ms per spec)
- Graceful degradation: empty completions, not errors, when outside workspace
- Commit after each task or logical group
