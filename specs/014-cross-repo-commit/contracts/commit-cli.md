# CLI Contract: fa commit

**Command**: `fa commit`  
**Purpose**: Create coordinated commits across all repos in the workspace

## Synopsis

```
fa commit [message] [flags]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `message` | No | Commit message. If omitted, opens editor. |

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--message` | `-m` | string | "" | Commit message (alternative to positional arg) |
| `--all` | `-a` | bool | false | Stage all tracked modifications |
| `--amend` | | bool | false | Amend the previous commit |
| `--dry-run` | | bool | false | Preview without committing |
| `--repo` | | []string | [] | Limit to specific repos (repeatable) |
| `--json` | | bool | false | Output as JSON |
| `--verbose` | `-v` | bool | false | Show detailed progress |
| `--allow-detached` | | bool | false | Allow commits in detached HEAD |

## Examples

```bash
# Commit staged changes across all repos (simplest form)
fa commit "Add user preferences feature"

# Alternative: use -m flag (git muscle memory)
fa commit -m "Add user preferences feature"

# Stage all changes and commit
fa commit -a "Fix typo in documentation"

# Commit only specific repos
fa commit --repo api --repo lib "Update API handlers"

# Preview what would be committed
fa commit --dry-run "Test commit"

# Amend previous commits
fa commit --amend "Updated commit message"

# JSON output for automation
fa commit --json "Automated commit"

# Open editor for message (no argument)
fa commit
```

## Behavior

### Without Message Argument
If no message is provided (either as positional arg or via `-m`), opens configured editor (`$EDITOR` or `git config core.editor`) with commit template. The entered message is applied to all repos.

### Message Precedence
If both positional argument and `-m` flag are provided, the positional argument takes precedence.

### Default Scope
Evaluates all repos in the current branch worktrees, but only commits to repos with staged changes. Repos without staged changes are skipped (not an error).

### Staged Changes Detection
Only repos with staged changes receive commits. Repos without staged changes are reported as "skipped".

### With `-a` Flag
Stages all tracked file modifications before committing (matches `git commit -a` behavior). Untracked files are NOT staged.

### With `--repo` Flag
Limits operation to specified repos only. Can be repeated for multiple repos.

### With `--amend` Flag
Amends the HEAD commit in each repo with staged changes. Without `-m`, reuses existing commit message.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All targeted repos committed successfully (or nothing to commit) |
| 1 | One or more repos failed to commit |
| 2 | Invalid arguments or configuration error |

## Human Output

```
Committing across 3 repositories...

✓ api: committed abc1234 (3 files, +45 -12)
✓ web: committed def5678 (1 file, +10 -2)
⊘ lib: skipped (nothing to commit)

Summary: 2 committed, 1 skipped, 0 failed
```

### Dry Run Output

```
DRY RUN - No commits will be created

Would commit in 2 repositories:

api:
  M src/handlers/user.go
  M src/handlers/auth.go
  A src/handlers/preferences.go

web:
  M components/Settings.tsx

Summary: 2 repos would be committed, 1 skipped
```

## JSON Output

```json
{
  "repos": [
    {
      "name": "api",
      "status": "committed",
      "commit_sha": "abc1234",
      "files_changed": 3,
      "insertions": 45,
      "deletions": 12,
      "error": null
    },
    {
      "name": "web",
      "status": "committed", 
      "commit_sha": "def5678",
      "files_changed": 1,
      "insertions": 10,
      "deletions": 2,
      "error": null
    },
    {
      "name": "lib",
      "status": "skipped",
      "commit_sha": null,
      "files_changed": 0,
      "insertions": 0,
      "deletions": 0,
      "error": "nothing to commit"
    }
  ],
  "summary": {
    "total": 3,
    "committed": 2,
    "skipped": 1,
    "failed": 0
  },
  "message": "Add user preferences feature"
}
```

### JSON Error Output

```json
{
  "repos": [
    {
      "name": "api",
      "status": "failed",
      "commit_sha": null,
      "files_changed": 0,
      "insertions": 0,
      "deletions": 0,
      "error": "pre-commit hook failed: eslint found 3 errors"
    }
  ],
  "summary": {
    "total": 1,
    "committed": 0,
    "skipped": 0,
    "failed": 1
  },
  "message": "Add feature"
}
```

## Error Messages

| Error | Code | Message | Remediation |
|-------|------|---------|-------------|
| Empty message | E501 | "Commit message cannot be empty" | "Provide a message with -m or let editor open" |
| No repos | E005 | "No repositories configured" | "Add repos with: fa add <url>" |
| Repo not found | E507 | "Repository 'foo' not found" | "Check available repos with: fa status" |
| Nothing to commit | E503 | "Nothing to commit" | "Stage changes first or use -a flag" |
| Not in workspace | E005 | "Not in a Foundagent workspace" | "Run from workspace root or use: fa init" |
| Detached HEAD | E502 | "Repository 'api' in detached HEAD" | "Use --allow-detached to commit anyway" |
