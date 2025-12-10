# Tasks: Doctor

**Input**: Design documents from `/specs/012-doctor/`
**Prerequisites**: spec.md ✅

**Tests**: Not explicitly requested - include minimal validation tests.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US4)
- Include exact file paths in descriptions

## Path Conventions

- **CLI commands**: `internal/cli/`
- **Diagnostics**: `internal/doctor/`
- **Workspace logic**: `internal/workspace/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Create doctor infrastructure

- [X] T001 Create doctor package with check framework in internal/doctor/doctor.go
- [X] T002 [P] Create doctor command skeleton in internal/cli/doctor.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core check framework and result collection

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Define Check interface (Name, Run, Remediation) in internal/doctor/check.go
- [X] T004 [P] Define CheckResult struct (Name, Status, Message, Remediation, Fixable) in internal/doctor/check.go
- [X] T005 [P] Define Status enum (Pass, Warn, Fail) in internal/doctor/check.go
- [X] T006 Implement check runner with result collection in internal/doctor/runner.go
- [X] T007 Add `doctor` command to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - check framework available

---

## Phase 3: User Story 1 - Quick Health Check (Priority: P1) 🎯 MVP

**Goal**: Pass/fail summary with `fa doctor`

**Independent Test**: Run `fa doctor` in healthy workspace, verify all checks pass

### Environment Checks

- [X] T008 [US1] Implement Git installation check in internal/doctor/checks/git.go
- [X] T009 [US1] Implement Git version check (minimum version) in internal/doctor/checks/git.go

### Workspace Structure Checks

- [X] T010 [US1] Implement `.foundagent.yaml` exists and valid check in internal/doctor/checks/config.go
- [X] T011 [US1] Implement `.foundagent/` directory exists check in internal/doctor/checks/structure.go
- [X] T012 [US1] Implement `.foundagent/state.json` exists and valid check in internal/doctor/checks/state.go
- [X] T013 [US1] Implement `repos/` directory exists check in internal/doctor/checks/structure.go
- [X] T014 [US1] Implement `repos/.bare/` directory exists check in internal/doctor/checks/structure.go
- [X] T015 [US1] Implement `repos/worktrees/` directory exists check in internal/doctor/checks/structure.go

### Output

- [X] T016 [US1] Display check name, status (✓/✗), and message for each check in internal/cli/doctor.go
- [X] T017 [US1] Display summary line (e.g., "5 checks passed, 1 failed") in internal/cli/doctor.go
- [X] T018 [US1] Exit with code 0 if all pass, non-zero if any fail in internal/cli/doctor.go

**Checkpoint**: `fa doctor` shows health check summary

---

## Phase 4: User Story 2 - Detailed Diagnostics (Priority: P1)

**Goal**: Clear explanations and remediation for failures

**Independent Test**: Break config file, run `fa doctor`, verify error explains issue and fix

### Repository Checks

- [X] T019 [US2] Implement check: each repo in config has bare clone in `repos/.bare/` in internal/doctor/checks/repos.go
- [X] T020 [US2] Implement check: each bare clone is valid Git repository in internal/doctor/checks/repos.go
- [X] T021 [US2] Implement check: detect orphaned bare clones (not in config) in internal/doctor/checks/repos.go

### Worktree Checks

- [X] T022 [US2] Implement check: worktrees in state.json exist on disk in internal/doctor/checks/worktrees.go
- [X] T023 [US2] Implement check: worktrees on disk are tracked by git worktree in internal/doctor/checks/worktrees.go
- [X] T024 [US2] Implement check: detect orphaned worktree directories in internal/doctor/checks/worktrees.go
- [X] T025 [US2] Implement check: worktree paths match expected structure in internal/doctor/checks/worktrees.go

### State Consistency Checks

- [X] T026 [US2] Implement check: config and state.json are in sync in internal/doctor/checks/consistency.go
- [X] T027 [US2] Implement check: `.code-workspace` matches current worktrees in internal/doctor/checks/consistency.go

### Remediation

- [X] T028 [US2] Include clear remediation steps in all check failures in internal/doctor/checks/*.go

**Checkpoint**: All failures include explanation and remediation

---

## Phase 5: User Story 3 - JSON Output (Priority: P2)

**Goal**: Machine-readable output with `fa doctor --json`

**Independent Test**: Run `fa doctor --json`, parse JSON, verify all check results accessible

### Implementation for User Story 3

- [X] T029 [US3] Add `--json` flag to doctor command in internal/cli/doctor.go
- [X] T030 [US3] Define JSON schema with checks array (name, status, message, remediation, fixable) in internal/cli/doctor.go
- [X] T031 [US3] Include summary object (total, passed, warnings, failed) in internal/cli/doctor.go

**Checkpoint**: `fa doctor --json` produces valid, parseable JSON

---

## Phase 6: User Story 4 - Auto-Fix Common Issues (Priority: P3)

**Goal**: Repair fixable issues with `fa doctor --fix`

**Independent Test**: Corrupt state.json, run `fa doctor --fix`, verify repaired

### Implementation for User Story 4

- [X] T032 [US4] Add `--fix` flag to doctor command in internal/cli/doctor.go
- [X] T033 [US4] Mark fixable checks in check definitions in internal/doctor/check.go
- [X] T034 [US4] Implement fix: regenerate missing state.json from filesystem in internal/doctor/fixes/state.go
- [X] T035 [US4] Implement fix: sync workspace file with current worktrees in internal/doctor/fixes/vscode.go
- [X] T036 [US4] Implement fix: remove orphaned state entries in internal/doctor/fixes/state.go
- [X] T037 [US4] Display what was fixed vs what needs manual intervention in internal/cli/doctor.go
- [X] T038 [US4] Never auto-fix destructive operations (e.g., deleting repos) in internal/doctor/fixes/*.go

**Checkpoint**: `fa doctor --fix` repairs fixable issues

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and additional flags

- [X] T039 [P] Handle not in workspace with clear error in internal/cli/doctor.go
- [X] T040 [P] Handle partial workspace (some files missing) gracefully in internal/cli/doctor.go
- [X] T041 [P] Add `-v` / `--verbose` flag for detailed check output in internal/cli/doctor.go
- [X] T042 [P] Group checks by category in verbose output in internal/cli/doctor.go
- [X] T043 Add help text with examples to doctor command in internal/cli/doctor.go
- [X] T044 Write integration test for doctor command in internal/cli/doctor_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational
  - US1 (Health Check) is P1 - implement first (MVP)
  - US2 (Detailed Diagnostics) is P1 - depends on US1
  - US3 (JSON) is P2
  - US4 (Auto-Fix) is P3
- **Polish (Phase 7)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Health Check) | P1 | Foundational | - |
| US2 (Detailed) | P1 | US1 | US3 |
| US3 (JSON) | P2 | US1 | US2 |
| US4 (Auto-Fix) | P3 | US2 | - |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                       Health Check
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           [US2]                           [US3]
          Detailed                         JSON
              │
              ▼
           [US4]
          Auto-Fix
```

---

## Implementation Strategy

### MVP First (User Stories 1-2)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T007)
3. Complete Phase 3: US1 Health Check (T008-T018)
4. Complete Phase 4: US2 Detailed Diagnostics (T019-T028)
5. **STOP and VALIDATE**: Doctor detects common issues with remediation
6. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US3 (JSON) → Test independently
3. Add US4 (Auto-Fix) → Test independently
4. Polish phase

---

## Notes

- Check categories: Environment, Structure, Repos, Worktrees, Consistency
- Workspace structure follows canonical layout
- State can be reconstructed from Git if state.json is corrupted
- Auto-fix never does destructive operations
- All checks complete in <5 seconds for 10 repos
- All error messages must include actionable remediation per constitution
- Commit after each task or logical group
