# Feature Specification: Doctor

**Feature Branch**: `012-doctor`  
**Created**: 2025-12-08  
**Status**: Draft  
**Input**: User description: "Diagnose workspace issues with fa doctor command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quick Health Check (Priority: P1)

A developer suspects something is wrong with their Foundagent workspace. They run `fa doctor` and get a quick pass/fail summary of all critical checks.

**Why this priority**: This is the core functionality — quickly identifying if something is broken is the primary use case.

**Independent Test**: Run `fa doctor` in a healthy workspace, verify all checks pass. Corrupt something, run again, verify failure is detected.

**Acceptance Scenarios**:

1. **Given** a healthy workspace, **When** I run `fa doctor`, **Then** I see all checks passing with green checkmarks
2. **Given** a workspace with issues, **When** I run `fa doctor`, **Then** I see failed checks with red X marks
3. **Given** any workspace state, **When** doctor completes, **Then** I see a summary line (e.g., "5 checks passed, 1 failed")

---

### User Story 2 - Detailed Diagnostics (Priority: P1)

A developer sees a failed check and needs more information. The doctor output includes clear explanations of what's wrong and how to fix it.

**Why this priority**: Detection without remediation is frustrating — users need actionable guidance.

**Independent Test**: Break a specific component, run `fa doctor`, verify the error message explains the issue and suggests a fix.

**Acceptance Scenarios**:

1. **Given** a failed check, **When** I see the output, **Then** it explains what was checked
2. **Given** a failed check, **When** I see the output, **Then** it explains what went wrong
3. **Given** a failed check, **When** I see the output, **Then** it suggests how to fix it

---

### User Story 3 - JSON Output for Automation (Priority: P2)

An AI agent or CI pipeline needs to parse diagnostic results programmatically. They run `fa doctor --json` and receive structured output.

**Why this priority**: Agent-friendly design requires JSON output for all commands.

**Independent Test**: Run `fa doctor --json`, parse as JSON, verify all check results are accessible.

**Acceptance Scenarios**:

1. **Given** I run `fa doctor --json`, **Then** output is valid JSON
2. **Given** JSON output, **When** I parse it, **Then** I can extract each check's name, status, message, and remediation

---

### User Story 4 - Auto-Fix Common Issues (Priority: P3)

A developer has minor fixable issues (e.g., state.json out of sync). They run `fa doctor --fix` and Foundagent automatically repairs what it can.

**Why this priority**: Nice-to-have automation — manual fixes work, but auto-fix saves time.

**Independent Test**: Corrupt state.json, run `fa doctor --fix`, verify state is repaired.

**Acceptance Scenarios**:

1. **Given** a fixable issue, **When** I run `fa doctor --fix`, **Then** the issue is repaired
2. **Given** a fixable issue was repaired, **When** I see output, **Then** it shows what was fixed
3. **Given** an unfixable issue, **When** I run `fa doctor --fix`, **Then** it reports the issue still needs manual intervention

---

### Edge Cases

- **Not in workspace**: Run outside Foundagent workspace — clear error with hint to run `fa init`
- **Partially initialized workspace**: Config exists but repos directory missing — detect and report
- **Git not installed**: System doesn't have Git — critical failure with install instructions
- **Corrupted bare clones**: Bare clone exists but is corrupted — detect and suggest re-clone
- **Orphaned worktrees**: Worktree directories exist but not tracked by Git — detect and report
- **Config/state mismatch**: Repos in config but not cloned, or cloned but not in config — detect both
- **Permission issues**: Can't read/write workspace files — report with permission hints

## Requirements *(mandatory)*

### Functional Requirements

#### Command Interface
- **FR-001**: System MUST support `fa doctor` command
- **FR-002**: System MUST support `--json` flag for machine-readable output
- **FR-003**: System MUST support `--fix` flag to auto-repair fixable issues
- **FR-004**: System MUST support `--verbose` / `-v` flag for detailed check output

#### Environment Checks
- **FR-005**: System MUST check if Git is installed and accessible
- **FR-006**: System MUST check Git version meets minimum requirements
- **FR-007**: System MUST report Git version in verbose mode

#### Workspace Structure Checks
- **FR-008**: System MUST verify `.foundagent.yaml` exists and is valid YAML
- **FR-009**: System MUST verify `.foundagent/` directory exists
- **FR-010**: System MUST verify `.foundagent/state.json` exists and is valid JSON
- **FR-011**: System MUST verify `repos/` directory exists
- **FR-012**: System MUST verify `repos/<repo-name>/.bare/` directory exists
- **FR-013**: System MUST verify `repos/<repo-name>/worktrees/` directory exists

#### Repository Checks
- **FR-014**: System MUST verify each repo in config has a bare clone in `repos/<name>/.bare/`
- **FR-015**: System MUST verify each bare clone is a valid Git repository
- **FR-016**: System MUST detect bare clones not listed in config (orphaned)
- **FR-017**: System MUST verify remotes are configured correctly in bare clones

#### Worktree Checks
- **FR-018**: System MUST verify worktrees listed in state.json exist on disk
- **FR-019**: System MUST verify worktrees on disk are tracked by Git (`git worktree list`)
- **FR-020**: System MUST detect orphaned worktree directories (not tracked by Git)
- **FR-021**: System MUST verify worktree paths match expected structure (`repos/<repo>/worktrees/<branch>/`)

#### State Consistency Checks
- **FR-022**: System MUST verify config and state.json are in sync
- **FR-023**: System MUST verify `.code-workspace` file matches current worktrees
- **FR-024**: System MUST detect and report any inconsistencies between sources of truth

#### Output Format
- **FR-025**: Each check MUST display: check name, status (pass/warn/fail), and message
- **FR-026**: Failed checks MUST include remediation steps
- **FR-027**: Summary MUST show total checks, passed, warnings, and failures
- **FR-028**: Exit code MUST be 0 if all checks pass, non-zero if any fail

#### Auto-Fix (--fix)
- **FR-029**: With `--fix`, system MUST attempt to repair fixable issues
- **FR-030**: Fixable issues include: missing state.json, out-of-sync workspace file, orphaned state entries
- **FR-031**: System MUST NOT auto-fix destructive operations (e.g., deleting repos)
- **FR-032**: System MUST report what was fixed and what still needs manual intervention

#### JSON Output
- **FR-033**: JSON output MUST include array of checks with: name, status, message, remediation
- **FR-034**: JSON output MUST include summary object with: total, passed, warnings, failed
- **FR-035**: JSON output MUST include fixable flag for each issue (true/false)

### Key Entities

- **Check**: A single diagnostic test with name, status (pass/warn/fail), message, and optional remediation
- **Check Category**: Grouping of related checks (environment, structure, repos, worktrees, state)
- **Diagnostic Report**: Collection of all check results with summary

### Assumptions

- Git is expected to be installed (Foundagent requires Git)
- Workspace structure follows canonical layout from constitution v1.6.0+
- State can be reconstructed from Git if state.json is corrupted

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Doctor completes all checks in under 5 seconds for workspaces with up to 10 repos
- **SC-002**: 100% of common issues are detectable (missing files, invalid config, orphaned items)
- **SC-003**: All failure messages include actionable remediation steps
- **SC-004**: `--fix` successfully repairs 100% of designated "fixable" issues
- **SC-005**: JSON output is parseable by standard JSON parsers in 100% of cases
