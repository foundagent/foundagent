# CLI Contract: fa push

**Command**: `fa push`  
**Purpose**: Push unpushed commits across all repos in the workspace

## Synopsis

```
fa push [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--dry-run` | | bool | false | Preview without pushing |
| `--repo` | | []string | [] | Limit to specific repos (repeatable) |
| `--json` | | bool | false | Output as JSON |
| `--verbose` | `-v` | bool | false | Show detailed progress |
| `--force` | `-f` | bool | false | Force push (dangerous, requires confirmation) |

## Examples

```bash
# Push all repos with unpushed commits
fa push

# Push specific repos only
fa push --repo api --repo web

# Preview what would be pushed
fa push --dry-run

# JSON output for automation
fa push --json

# Force push (with confirmation)
fa push --force
```

## Behavior

### Default Scope
Evaluates all repos in the current branch worktrees, but only pushes repos with unpushed commits. Repos already up-to-date are skipped (not an error).

### Unpushed Detection
Uses `git rev-list @{u}..HEAD` to detect commits ahead of upstream. Repos without unpushed commits are reported as "skipped".

### With `--repo` Flag
Limits operation to specified repos only. Can be repeated for multiple repos.

### With `--force` Flag
Force pushes (overwrites remote history). Requires interactive confirmation unless in JSON mode (where it fails).

### Parallel Execution
Pushes run in parallel across repos for performance. Each push is independent.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All targeted repos pushed successfully (or nothing to push) |
| 1 | One or more repos failed to push |
| 2 | Invalid arguments or configuration error |

## Human Output

```
Pushing 3 repositories...

✓ api: pushed 2 commits (main -> origin/main)
✓ web: pushed 1 commit (main -> origin/main)
⊘ lib: skipped (nothing to push)

Summary: 2 pushed, 1 skipped, 0 failed
```

### Dry Run Output

```
DRY RUN - No pushes will be executed

Would push from 2 repositories:

api: 2 commits ahead of origin/main
  abc1234 Add user preferences handler
  def5678 Add preferences tests

web: 1 commit ahead of origin/main
  ghi9012 Update Settings component

Summary: 2 repos would be pushed, 1 already up-to-date
```

### Failure Output

```
Pushing 3 repositories...

✓ api: pushed 2 commits (main -> origin/main)
✗ web: failed (remote has new commits)
⊘ lib: skipped (nothing to push)

Summary: 1 pushed, 1 skipped, 1 failed

Hint: Run 'fa sync --pull' to fetch and merge remote changes, then retry push
```

## JSON Output

```json
{
  "repos": [
    {
      "name": "api",
      "status": "pushed",
      "refs_pushed": ["main -> origin/main"],
      "commits_pushed": 2,
      "error": null
    },
    {
      "name": "web",
      "status": "pushed",
      "refs_pushed": ["main -> origin/main"],
      "commits_pushed": 1,
      "error": null
    },
    {
      "name": "lib",
      "status": "skipped",
      "refs_pushed": [],
      "commits_pushed": 0,
      "error": "nothing to push"
    }
  ],
  "summary": {
    "total": 3,
    "pushed": 2,
    "skipped": 1,
    "failed": 0
  }
}
```

### JSON Error Output

```json
{
  "repos": [
    {
      "name": "api",
      "status": "pushed",
      "refs_pushed": ["main -> origin/main"],
      "commits_pushed": 2,
      "error": null
    },
    {
      "name": "web",
      "status": "failed",
      "refs_pushed": [],
      "commits_pushed": 0,
      "error": "rejected: remote contains commits not present locally"
    }
  ],
  "summary": {
    "total": 2,
    "pushed": 1,
    "skipped": 0,
    "failed": 1
  }
}
```

## Error Messages

| Error | Code | Message | Remediation |
|-------|------|---------|-------------|
| No repos | E201 | "No repositories configured" | "Add repos with: fa add <url>" |
| Repo not found | E202 | "Repository 'foo' not found" | "Check available repos with: fa status" |
| Nothing to push | - | "Nothing to push" | (informational, not an error) |
| Not in workspace | E200 | "Not in a Foundagent workspace" | "Run from workspace root or use: fa init" |
| Remote rejected | E301 | "Push rejected by remote" | "Pull latest changes first: fa sync --pull" |
| No upstream | E302 | "No upstream branch configured" | "Set upstream with: git push -u origin <branch>" |
| Auth failed | E303 | "Authentication failed" | "Check SSH keys or credentials" |
| Force denied | E304 | "Force push requires confirmation" | "Cannot force push in --json mode" |

## Relationship to fa sync --push

`fa push` is functionally similar to `fa sync --push` but:
- `fa push` is a standalone command (more discoverable)
- `fa push` supports `--dry-run` 
- `fa push` supports `--repo` filtering
- `fa sync --push` fetches first, then pushes

For most use cases, `fa push` is preferred after `fa commit`.
