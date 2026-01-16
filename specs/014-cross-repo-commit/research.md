# Research: Cross-Repo Commit

**Feature**: 014-cross-repo-commit  
**Date**: 2026-01-13  
**Status**: Complete

## Research Questions

### 1. Git Commit CLI Patterns

**Question**: What git commit options need to be supported for cross-repo operations?

**Decision**: Support core commit options that make sense for cross-repo use cases:
- `-m <message>` / `--message` - Required for non-interactive cross-repo commits
- `-a` / `--all` - Stage all tracked modifications
- `--amend` - Amend previous commit (per-repo)
- `--allow-empty` - Not supported (no use case for cross-repo empty commits)

**Rationale**: Cross-repo commits should mirror git's familiar interface but only expose options that make sense when applied across multiple repositories simultaneously. Interactive features (like editing messages in editor) apply to the shared message, not per-repo.

**Alternatives Considered**:
- Full git commit option parity: Rejected. Many options (--fixup, --squash, --signoff) are too repo-specific
- Custom message format with repo tokens: Rejected. Adds complexity without clear benefit

### 2. Parallel Commit Execution Strategy

**Question**: Should commits run in parallel or sequentially?

**Decision**: Parallel execution using existing `workspace.ExecuteParallel()` pattern.

**Rationale**: 
- Commits are local-only operations (no network), so parallelization is safe
- Pre-commit hooks run per-repo and are independent
- Matches existing sync pattern for consistency
- Performance: 10 sequential commits ~5-10s, parallel ~1-2s

**Alternatives Considered**:
- Sequential with early-exit on failure: Rejected. Partial success is more useful than all-or-nothing
- Transactional (rollback on any failure): Rejected. Git doesn't support commit rollback cleanly; adds complexity

### 3. Pre-commit Hook Handling

**Question**: How should pre-commit hook failures be handled?

**Decision**: Fail the individual repo, continue others. Report clearly in summary.

**Rationale**:
- Matches FR-014, FR-029 (partial failure handling)
- Pre-commit hooks are repo-specific; one repo's hook shouldn't block others
- Developers can fix the failing repo and re-run commit

**Implementation**:
```go
// Per-repo commit returns error if pre-commit fails
// Parallel execution collects all results
// Summary shows: 3 committed, 1 failed (pre-commit hook)
```

### 4. Editor Integration for Commit Messages

**Question**: How should `fa commit` (without `-m`) work?

**Decision**: Open editor with template, apply resulting message to all repos.

**Rationale**:
- Matches git behavior for single-repo commit
- Single message makes sense for coordinated cross-repo changes
- Uses `$EDITOR` or `git config core.editor`

**Implementation Flow**:
1. Create temp file with commit template
2. Open editor (blocking)
3. Read message from file
4. Apply to all repos with staged changes
5. Clean up temp file

### 5. Detecting Repos with Staged Changes

**Question**: How to efficiently detect which repos need commits?

**Decision**: Use `git diff --cached --quiet` exit code per worktree.

**Rationale**:
- Exit code 0 = no staged changes, 1 = has staged changes
- Faster than parsing `git status --porcelain`
- Works reliably across git versions

**Implementation**:
```go
func HasStagedChanges(worktreePath string) (bool, error) {
    cmd := exec.Command("git", "-C", worktreePath, "diff", "--cached", "--quiet")
    err := cmd.Run()
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return exitErr.ExitCode() == 1, nil
        }
        return false, err
    }
    return false, nil // exit code 0 = no staged changes
}
```

### 6. Push Strategy for Cross-Repo

**Question**: Should `fa push` push all branches or just current?

**Decision**: Push current branch only, for all repos.

**Rationale**:
- Matches user mental model: "push what I just committed"
- Pushing all branches could have unintended side effects
- Aligns with `fa sync --push` behavior

**Alternatives Considered**:
- Push all branches with unpushed commits: Too aggressive, could push unrelated work
- Push only repos that were just committed: Limiting; user might want to push older commits too

### 7. JSON Output Structure

**Question**: What JSON structure for commit/push results?

**Decision**: Follow existing `fa sync --json` pattern:

```json
{
  "repos": [
    {
      "name": "api",
      "status": "committed",
      "commit_sha": "abc1234",
      "files_changed": 3,
      "error": null
    },
    {
      "name": "web", 
      "status": "skipped",
      "commit_sha": null,
      "files_changed": 0,
      "error": "nothing to commit"
    }
  ],
  "summary": {
    "total": 5,
    "committed": 3,
    "skipped": 2,
    "failed": 0
  }
}
```

**Rationale**: Consistent with existing JSON output patterns. Includes all info needed for automation.

### 8. Handling Detached HEAD State

**Question**: What happens if a repo is in detached HEAD state?

**Decision**: Skip with warning, allow override with `--allow-detached`.

**Rationale**:
- Committing in detached HEAD creates orphan commits (confusing)
- Most users don't want this behavior
- Power users can opt-in with flag

### 9. Commit Message Validation

**Question**: Should message be validated before committing?

**Decision**: Minimal validation - reject empty messages only.

**Rationale**:
- Git itself allows any non-empty message
- Commit message linting is repo-specific (commitlint, etc.)
- Pre-commit hooks handle repo-specific validation

## Implementation Patterns

### From Existing Codebase

| Pattern | Source | Reuse For |
|---------|--------|-----------|
| Parallel execution | `workspace.ExecuteParallel()` | Commit/push parallelization |
| Result aggregation | `workspace.SyncResult`, `SyncSummary` | `CommitResult`, `CommitSummary` |
| JSON output | `cli/sync.go:outputSyncJSON()` | Commit/push JSON output |
| Git command execution | `git/*.go` via `os/exec` | New commit/push operations |
| Worktree discovery | `workspace.LoadState()` | Finding repos for current branch |
| Error handling | `errors.Wrap()` with codes | Commit/push error messages |

### New Patterns Needed

| Pattern | Purpose |
|---------|---------|
| Staged changes detection | `git.HasStagedChanges()` |
| Git commit execution | `git.Commit(path, message, opts)` |
| Unpushed commits detection | Enhance `git.GetAheadBehindCount()` |
| Editor integration | `git.OpenEditor(template)` → message |

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Pre-commit hook timeout | Low | Medium | Document timeout behavior; no custom handling |
| Commit message encoding issues | Low | Low | Use git's default encoding handling |
| Parallel commit race conditions | Very Low | Low | Git worktrees are isolated; no shared state |
| Large file staging with `-a` | Medium | Medium | Document that `-a` stages all tracked changes |

## Open Questions (Resolved)

All research questions resolved. Ready for Phase 1 design.
