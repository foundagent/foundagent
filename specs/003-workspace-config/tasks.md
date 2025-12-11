# Tasks: Workspace Configuration

**Input**: Design documents from `/specs/003-workspace-config/`
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
- **Config parsing**: `internal/config/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Verify infrastructure for config file support

- [X] T001 Add YAML parsing dependency (gopkg.in/yaml.v3) to go.mod
- [X] T002 [P] Add TOML parsing dependency (github.com/BurntSushi/toml) to go.mod

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core config parsing and schema infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Define config file schema struct (workspace, repos, settings) in internal/config/schema.go
- [X] T004 [P] Implement YAML config parser in internal/config/yaml.go
- [X] T005 [P] Implement TOML config parser in internal/config/toml.go
- [X] T006 [P] Implement JSON config parser in internal/config/json.go
- [X] T007 Implement config file resolution order (YAML > TOML > JSON) in internal/config/loader.go
- [X] T008 Implement config validation with line-number errors in internal/config/validate.go
- [X] T009 Implement comment-preserving config writer for YAML in internal/config/writer.go

**Checkpoint**: Foundation ready - config can be read, validated, and written

---

## Phase 3: User Story 1 - Config Generated on Init (Priority: P1) 🎯 MVP

**Goal**: Generate default `.foundagent.yaml` with helpful comments on `fa init`

**Independent Test**: Run `fa init test-project`, verify `.foundagent.yaml` exists with valid structure and comments

### Implementation for User Story 1

- [X] T010 [US1] Create default config template with workspace section in internal/config/template.go
- [X] T011 [US1] Add empty repos list to default template in internal/config/template.go
- [X] T012 [US1] Add settings section with defaults (auto_create_worktree: true) in internal/config/template.go
- [X] T013 [US1] Include helpful comments explaining each section in internal/config/template.go
- [X] T014 [US1] Integrate config generation into init command in internal/cli/init.go

**Checkpoint**: `fa init` creates `.foundagent.yaml` with documentation comments

---

## Phase 4: User Story 2 - Commands Update Config (Priority: P1)

**Goal**: `fa add <url>` automatically updates config file with new repo entry

**Independent Test**: Run `fa add <url>`, verify repo appears in `.foundagent.yaml` under repos section

### Implementation for User Story 2

- [X] T015 [US2] Implement AddRepo function to update config in internal/config/writer.go
- [X] T016 [US2] Preserve existing comments and formatting when adding repo in internal/config/writer.go
- [X] T017 [US2] Integrate config update into add command in internal/cli/add.go
- [X] T018 [US2] Support custom name in config entry in internal/config/writer.go

**Checkpoint**: `fa add <url>` updates config file while preserving comments

---

## Phase 5: User Story 3 - Sync State from Config (Priority: P1)

**Goal**: `fa add` (no args) reconciles state with config - clones missing repos, warns about stale

**Independent Test**: Edit config to add repo, run `fa add`, verify new repo cloned and stale warned

### Implementation for User Story 3

- [X] T019 [US3] Implement config-state diff function in internal/workspace/reconcile.go
- [X] T020 [US3] Detect repos in config but not cloned in internal/workspace/reconcile.go
- [X] T021 [US3] Detect repos cloned but not in config (stale) in internal/workspace/reconcile.go
- [X] T022 [US3] Implement `fa add` no-args mode to trigger reconciliation in internal/cli/add.go
- [X] T023 [US3] Clone missing repos when running `fa add` without arguments in internal/cli/add.go
- [X] T024 [US3] Display warning for stale repos with remediation hint in internal/cli/add.go
- [X] T025 [US3] Display summary of reconciliation actions taken in internal/cli/add.go

**Checkpoint**: `fa add` syncs workspace to match config file

---

## Phase 6: User Story 4 - Share Config Across Team (Priority: P2)

**Goal**: Copy config to new workspace, run `fa add`, get identical setup

**Independent Test**: Copy `.foundagent.yaml` with 3 repos to fresh workspace, run `fa add`, verify all cloned

### Implementation for User Story 4

- [X] T026 [US4] Initialize minimal workspace if running `fa add` with config but no `.foundagent/` in internal/cli/add.go
- [X] T027 [US4] Support parallel cloning of all repos from config in internal/cli/add.go
- [X] T028 [US4] Display progress for multi-repo setup from config in internal/cli/add.go

**Checkpoint**: Team members can bootstrap workspace from shared config file

---

## Phase 7: User Story 5 - Config Validation (Priority: P2)

**Goal**: Clear validation errors with line numbers for config issues

**Independent Test**: Introduce syntax error, run any `fa` command, verify error includes line number

### Implementation for User Story 5

- [X] T029 [US5] Implement YAML syntax error detection with line numbers in internal/config/validate.go
- [X] T030 [US5] Validate repo URL format before clone attempt in internal/config/validate.go
- [X] T031 [US5] Detect duplicate repo names with error in internal/config/validate.go
- [X] T032 [US5] Warn (not error) on unknown config keys for forward compatibility in internal/config/validate.go
- [X] T033 [US5] Run validation on every command that loads config in internal/config/loader.go

**Checkpoint**: Config errors are clear, actionable, and include line numbers

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and robustness

- [X] T034 [P] Handle missing config file - regenerate default with warning in internal/config/loader.go
- [X] T035 [P] Handle empty repos list as valid state in internal/config/validate.go
- [X] T036 [P] Handle duplicate URL entries - warn and use first in internal/config/validate.go
- [X] T037 [P] Handle read-only config file with clear error in internal/config/writer.go
- [X] T038 [P] Preserve user comments when config is updated by commands in internal/config/writer.go
- [X] T039 [P] Handle multiple config formats (warn if both .yaml and .toml exist) in internal/config/loader.go
- [X] T040 Write integration test for config loading in internal/config/loader_test.go
- [X] T041 Write integration test for config-state reconciliation in internal/workspace/reconcile_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational
  - US1 (Config on Init) is P1 - implement first (MVP)
  - US2 (Commands Update) is P1 - depends on US1
  - US3 (Sync from Config) is P1 - depends on US2
  - US4 (Team Share) and US5 (Validation) are P2
- **Polish (Phase 8)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Config on Init) | P1 | Foundational | - |
| US2 (Commands Update) | P1 | US1 | - |
| US3 (Sync from Config) | P1 | US2 | - |
| US4 (Team Share) | P2 | US3 | US5 |
| US5 (Validation) | P2 | Foundational | US4 |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                      Config on Init
                              │
                              ▼
                           [US2]
                      Commands Update
                              │
                              ▼
                           [US3]
                      Sync from Config
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           [US4]                           [US5]
         Team Share                      Validation
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T009)
3. Complete Phase 3: US1 Config on Init (T010-T014)
4. Complete Phase 4: US2 Commands Update (T015-T018)
5. Complete Phase 5: US3 Sync from Config (T019-T025)
6. **STOP and VALIDATE**: Config-driven workflow works
7. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US4 (Team Share) + US5 (Validation) → Test independently
3. Polish phase

---

## Notes

- Config schema: `workspace.name`, `repos[]` with `url`/`name`/`default_branch`, `settings.auto_create_worktree`
- State file (`.foundagent/state.json`) stores runtime state, config is source of truth
- Comment preservation requires careful YAML handling - use yaml.v3 node API
- Worktrees at fixed location: `repos/<repo>/worktrees/<branch>/`
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
