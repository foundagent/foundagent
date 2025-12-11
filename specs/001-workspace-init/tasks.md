# Tasks: Workspace Initialization

**Input**: Design documents from `/specs/001-workspace-init/`
**Prerequisites**: spec.md ✅

**Tests**: Not explicitly requested - include minimal validation tests.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US3)
- Include exact file paths in descriptions

## Path Conventions

- **CLI commands**: `internal/cli/`
- **Workspace logic**: `internal/workspace/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Verify existing infrastructure and create command skeleton

- [X] T001 Verify Cobra dependency in go.mod supports subcommands
- [X] T002 [P] Create init command skeleton in internal/cli/init.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core workspace structures and validation that all stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Define Workspace struct and constants in internal/workspace/workspace.go
- [X] T004 [P] Define default config schema (YAML structure) in internal/workspace/config.go
- [X] T005 [P] Define state schema (JSON structure) in internal/workspace/state.go
- [X] T006 Implement workspace name validation (filesystem-safe characters) in internal/workspace/validation.go
- [X] T007 [P] Implement VS Code workspace file template generation in internal/workspace/vscode.go
- [X] T008 Add `init` command to rootCmd in internal/cli/root.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Create New Workspace (Priority: P1) 🎯 MVP

**Goal**: Create a new Foundagent workspace with `fa init my-app`

**Independent Test**: Run `fa init test-project` in empty directory, verify folder structure created, open `.code-workspace` in VS Code

### Implementation for User Story 1

- [X] T009 [US1] Implement directory creation with specified name in internal/cli/init.go
- [X] T010 [US1] Create `.foundagent/` subdirectory for machine-managed state in internal/workspace/workspace.go
- [X] T011 [US1] Generate `.foundagent.yaml` with default config (workspace name, empty repos) in internal/workspace/config.go
- [X] T012 [US1] Generate `.foundagent/state.json` initialized as empty object in internal/workspace/state.go
- [X] T013 [US1] Create `repos/` directory structure (`repos/<repo-name>/.bare/`, `repos/<repo-name>/worktrees/`) in internal/workspace/workspace.go
- [X] T014 [US1] Generate `<name>.code-workspace` file with folders array in internal/workspace/vscode.go
- [X] T015 [US1] Display success message with absolute path to created workspace in internal/cli/init.go
- [X] T016 [US1] Implement exit code 0 on success, non-zero on failure in internal/cli/init.go

**Checkpoint**: `fa init my-app` creates complete workspace structure

---

## Phase 4: User Story 2 - JSON Output (Priority: P2)

**Goal**: Machine-readable output with `fa init my-app --json`

**Independent Test**: Run `fa init test-project --json`, parse JSON output, verify structure

### Implementation for User Story 2

- [X] T017 [US2] Add `--json` flag to init command in internal/cli/init.go
- [X] T018 [US2] Define JSON output schema (path, name, status) in internal/cli/init.go
- [X] T019 [US2] Implement JSON error output with error code, message, remediation in internal/cli/init.go
- [X] T020 [US2] Suppress human-readable messages when `--json` flag is set in internal/cli/init.go

**Checkpoint**: `fa init my-app --json` produces valid, parseable JSON

---

## Phase 5: User Story 3 - Force Reinitialize (Priority: P3)

**Goal**: Reinitialize existing workspace with `fa init my-app --force`

**Independent Test**: Create workspace, corrupt config, run `fa init my-app --force`, verify config restored and repos preserved

### Implementation for User Story 3

- [X] T021 [US3] Detect existing `.foundagent/` directory and error without `--force` in internal/cli/init.go
- [X] T022 [US3] Add `--force` flag to init command in internal/cli/init.go
- [X] T023 [US3] Implement workspace regeneration logic (recreate config, state) in internal/workspace/workspace.go
- [X] T024 [US3] Preserve existing `repos/` directory contents during force reinit in internal/workspace/workspace.go
- [X] T025 [US3] Display clear error message suggesting `--force` when workspace exists in internal/cli/init.go

**Checkpoint**: `fa init --force` restores workspace config while preserving repos

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, validation, and final polish

- [X] T026 [P] Handle empty name argument with usage error and example in internal/cli/init.go
- [X] T027 [P] Handle invalid filesystem characters in name with clear error in internal/workspace/validation.go
- [X] T028 [P] Handle path too long error with max path info in internal/workspace/validation.go
- [X] T029 [P] Handle permission denied with clear message in internal/cli/init.go
- [X] T030 [P] Handle `.` and `..` as invalid workspace names in internal/workspace/validation.go
- [X] T031 [P] Trim leading/trailing spaces from name in internal/workspace/validation.go
- [X] T032 Add help text with examples to init command in internal/cli/init.go
- [X] T033 Write integration test for init command in internal/cli/init_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - verify existing infrastructure
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1 (Create Workspace) is P1 - implement first (MVP)
  - US2 (JSON Output) is P2 - can follow US1
  - US3 (Force Reinit) is P3 - can follow US1
- **Polish (Phase 6)**: Depends on user stories being complete

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Create) | P1 | Foundational | US2, US3 (after US1 core) |
| US2 (JSON) | P2 | US1 core | US3 |
| US3 (Force) | P3 | US1 core | US2 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                      Create Workspace
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           [US2]                           [US3]
         JSON Output                   Force Reinit
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T008)
3. Complete Phase 3: User Story 1 (T009-T016)
4. **STOP and VALIDATE**: `fa init my-app` works
5. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US2 (JSON Output) → Test independently
3. Add US3 (Force Reinit) → Test independently
4. Polish phase

---

## Notes

- Workspace structure: `.foundagent.yaml` (config), `.foundagent/state.json` (state), `repos/<repo-name>/.bare/` (bare clones), `repos/<repo-name>/worktrees/` (working directories)
- VS Code workspace file uses JSON format with `folders` array
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
