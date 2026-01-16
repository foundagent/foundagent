# Data Model: Cross-Repo Commit

**Feature**: 014-cross-repo-commit  
**Date**: 2026-01-13

## Entities

### CommitResult

Represents the outcome of a commit operation for a single repository.

| Field | Type | Description |
|-------|------|-------------|
| `RepoName` | string | Name of the repository |
| `Status` | string | One of: "committed", "skipped", "failed" |
| `CommitSHA` | string | Short SHA of created commit (7 chars), empty if skipped/failed |
| `FilesChanged` | int | Number of files in the commit |
| `Insertions` | int | Lines added |
| `Deletions` | int | Lines removed |
| `Error` | error | Error details if status is "failed", nil otherwise |

**Status Values**:
- `committed`: Commit created successfully
- `skipped`: No staged changes (or filtered by --repo flag)
- `failed`: Commit failed (pre-commit hook, git error, etc.)

### CommitSummary

Aggregates results across all repositories.

| Field | Type | Description |
|-------|------|-------------|
| `Total` | int | Total repos evaluated |
| `Committed` | int | Repos with successful commits |
| `Skipped` | int | Repos with nothing to commit |
| `Failed` | int | Repos where commit failed |

### PushResult

Represents the outcome of a push operation for a single repository.

| Field | Type | Description |
|-------|------|-------------|
| `RepoName` | string | Name of the repository |
| `Status` | string | One of: "pushed", "skipped", "failed" |
| `RefsPushed` | []string | Refs that were pushed (e.g., "main -> origin/main") |
| `CommitsPushed` | int | Number of commits pushed |
| `Error` | error | Error details if status is "failed", nil otherwise |

**Status Values**:
- `pushed`: Successfully pushed commits
- `skipped`: Nothing to push (already up-to-date)
- `failed`: Push failed (remote rejected, network error, etc.)

### PushSummary

Aggregates push results across all repositories.

| Field | Type | Description |
|-------|------|-------------|
| `Total` | int | Total repos evaluated |
| `Pushed` | int | Repos with successful pushes |
| `Skipped` | int | Repos with nothing to push |
| `Failed` | int | Repos where push failed |

### CommitOptions

Configuration for commit operations.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Message` | string | "" | Commit message (required unless --amend) |
| `All` | bool | false | Stage all tracked modifications (-a) |
| `Amend` | bool | false | Amend previous commit |
| `DryRun` | bool | false | Preview without executing |
| `Repos` | []string | nil | Limit to specific repos (nil = all) |
| `Verbose` | bool | false | Show detailed output |
| `AllowDetached` | bool | false | Allow commits in detached HEAD |

### PushOptions

Configuration for push operations.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DryRun` | bool | false | Preview without executing |
| `Repos` | []string | nil | Limit to specific repos (nil = all) |
| `Verbose` | bool | false | Show detailed output |
| `Force` | bool | false | Force push (dangerous) |

## State Transitions

### Commit Flow

```
[Workspace] 
    │
    ▼ LoadState()
[Repo List] 
    │
    ▼ Filter by --repo (if specified)
[Target Repos]
    │
    ▼ Check HasStagedChanges() (parallel)
[Repos with Changes]
    │
    ├──▶ DryRun? ──▶ Report what would commit
    │
    ▼ Execute git commit (parallel)
[CommitResults]
    │
    ▼ Aggregate
[CommitSummary]
```

### Push Flow

```
[Workspace]
    │
    ▼ LoadState()
[Repo List]
    │
    ▼ Filter by --repo (if specified)
[Target Repos]
    │
    ▼ Check GetAheadBehindCount() (parallel)
[Repos with Unpushed]
    │
    ├──▶ DryRun? ──▶ Report what would push
    │
    ▼ Execute git push (parallel)
[PushResults]
    │
    ▼ Aggregate
[PushSummary]
```

## Validation Rules

### Commit Validation
1. Message MUST be non-empty (unless --amend reuses existing)
2. At least one repo MUST have staged changes (unless -a)
3. Repos with detached HEAD MUST raise error E402 (unless --allow-detached)
4. Specified --repo names MUST exist in workspace

### Push Validation
1. Repos MUST have upstream tracking branch configured
2. Repos with nothing to push are skipped (not an error)
3. Specified --repo names MUST exist in workspace

## Relationships

```
Workspace (1) ─────────┬──────── (N) Repository
                       │
                       ▼
              CommitOperation
                       │
           ┌───────────┼───────────┐
           ▼           ▼           ▼
      CommitResult CommitResult CommitResult
           │           │           │
           └───────────┼───────────┘
                       ▼
               CommitSummary
```

The same pattern applies to PushOperation → PushResult → PushSummary.
