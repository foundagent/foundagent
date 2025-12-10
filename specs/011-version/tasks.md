# Tasks: Version

**Input**: Design documents from `/specs/011-version/`
**Prerequisites**: spec.md ✅

**Tests**: Not explicitly requested - include minimal validation tests.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US4)
- Include exact file paths in descriptions

## Path Conventions

- **CLI commands**: `internal/cli/`
- **Version info**: `internal/version/`
- **Tests**: Alongside source files (`*_test.go`)

---

## Phase 1: Setup

**Purpose**: Create version infrastructure

- [X] T001 Create version package with build-time variables in internal/version/version.go
- [X] T002 [P] Create version command skeleton in internal/cli/version.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build-time version injection infrastructure

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Define version variables (Version, Commit, BuildDate, GoVersion) in internal/version/version.go
- [X] T004 [P] Update build process to inject version via ldflags in Makefile or build script
- [X] T005 [P] Add `--version` flag to root command in internal/cli/root.go
- [X] T006 Add `version` command to CLI in internal/cli/root.go

**Checkpoint**: Foundation ready - version info available at build time

---

## Phase 3: User Story 1 - Check Installed Version (Priority: P1) 🎯 MVP

**Goal**: Display version with `fa version` or `fa --version`

**Independent Test**: Run `fa version`, verify version number displayed

### Implementation for User Story 1

- [X] T007 [US1] Implement version output (e.g., "foundagent v1.0.0") in internal/cli/version.go
- [X] T008 [US1] Ensure `fa version` and `fa --version` produce same output in internal/cli/version.go
- [X] T009 [US1] Ensure `foundagent version` works identically in internal/cli/version.go
- [X] T010 [US1] Format version as single line for easy parsing in internal/cli/version.go

**Checkpoint**: `fa version` shows version number

---

## Phase 4: User Story 2 - Get Full Build Information (Priority: P2)

**Goal**: Detailed build info with `fa version --full`

**Independent Test**: Run `fa version --full`, verify all build metadata displayed

### Implementation for User Story 2

- [X] T011 [US2] Add `--full` flag to version command in internal/cli/version.go
- [X] T012 [US2] Display git commit hash (short, 7 chars) in internal/cli/version.go
- [X] T013 [US2] Display build date (ISO 8601 format) in internal/cli/version.go
- [X] T014 [US2] Display Go version used to build in internal/cli/version.go
- [X] T015 [US2] Display OS and architecture (e.g., "darwin/arm64") in internal/cli/version.go

**Checkpoint**: `fa version --full` shows complete build metadata

---

## Phase 5: User Story 3 - Machine-Readable Output (Priority: P2)

**Goal**: JSON output with `fa version --json`

**Independent Test**: Run `fa version --json`, parse JSON, verify all fields present

### Implementation for User Story 3

- [X] T016 [US3] Add `--json` flag to version command in internal/cli/version.go
- [X] T017 [US3] Define JSON schema (version, commit, build_date, go_version, os, arch) in internal/cli/version.go
- [X] T018 [US3] Output valid JSON with all fields in internal/cli/version.go

**Checkpoint**: `fa version --json` produces valid, parseable JSON

---

## Phase 6: User Story 4 - Check for Updates (Priority: P3)

**Goal**: Check for new versions with `fa version --check`

**Independent Test**: Run `fa version --check`, verify update status displayed

### Implementation for User Story 4

- [X] T019 [US4] Add `--check` flag to version command in internal/cli/version.go
- [X] T020 [US4] Query GitHub releases API for latest version in internal/version/update.go
- [X] T021 [US4] Compare local version to latest release in internal/version/update.go
- [X] T022 [US4] Display "You're up to date" if current in internal/cli/version.go
- [X] T023 [US4] Display "Update available: vX.Y.Z" with download URL if newer exists in internal/cli/version.go
- [X] T024 [US4] Handle network failure gracefully (show local version, warn about check failure) in internal/cli/version.go
- [X] T025 [US4] Use timeout (5 seconds) for update check in internal/version/update.go

**Checkpoint**: `fa version --check` reports update availability

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and development builds

- [X] T026 [P] Handle development builds (no version tag) - show "dev" in internal/version/version.go
- [X] T027 [P] Handle missing commit hash - show "unknown" in internal/version/version.go
- [X] T028 [P] Ensure `fa` and `foundagent` show identical output in internal/cli/version.go
- [X] T029 Add help text to version command in internal/cli/version.go
- [X] T030 Write unit test for version command in internal/cli/version_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational
  - US1 (Basic Version) is P1 - implement first (MVP)
  - US2 (Full) and US3 (JSON) are P2
  - US4 (Check) is P3
- **Polish (Phase 7)**: Depends on user stories

### User Story Dependencies

| Story | Priority | Depends On | Can Parallel With |
|-------|----------|------------|-------------------|
| US1 (Basic) | P1 | Foundational | - |
| US2 (Full) | P2 | US1 | US3 |
| US3 (JSON) | P2 | US1 | US2 |
| US4 (Check) | P3 | US1 | - |

### Parallel Opportunities

```
After Foundational (Phase 2) completes:

┌──────────────────────────────────────────────────────────────┐
│                    Foundational Complete                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                           [US1]
                        Basic Version
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
           [US2]                           [US3]
           Full                            JSON
                              │
                              ▼
                           [US4]
                           Check
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T006)
3. Complete Phase 3: US1 Basic Version (T007-T010)
4. **STOP and VALIDATE**: `fa version` works
5. Deploy/demo if ready

### Full Feature

1. MVP above
2. Add US2 (Full) + US3 (JSON) → Test independently
3. Add US4 (Check) → Test independently
4. Polish phase

---

## Notes

- Version follows semantic versioning (MAJOR.MINOR.PATCH)
- Build info injected via Go ldflags at compile time
- Update check uses GitHub releases API
- Network timeout of 5 seconds for update check
- All commands work for both `fa` and `foundagent` aliases
- Commit after each task or logical group
