# Quickstart: Cross-Repo Commit

**Feature**: 014-cross-repo-commit

## Overview

The `fa commit` and `fa push` commands enable coordinated commits and pushes across all repositories in your workspace. When you're working on a feature that spans multiple repos, these commands ensure consistent commit messages and synchronized pushes.

## Basic Workflow

### 1. Make Changes Across Repos

Work on your feature across multiple repos in the workspace:

```bash
# Edit files in api repo
vim repos/worktrees/api/main/src/handlers/user.go

# Edit files in web repo  
vim repos/worktrees/web/main/components/UserProfile.tsx
```

### 2. Stage Changes

Stage your changes in each repo (or use `-a` flag):

```bash
# Stage in api
cd repos/worktrees/api/main && git add .

# Stage in web
cd repos/worktrees/web/main && git add .

# Return to workspace root
cd /path/to/workspace
```

### 3. Commit Across All Repos

Create coordinated commits with the same message:

```bash
fa commit "Add user profile feature"
```

Output:
```
Committing across 2 repositories...

✓ api: committed abc1234 (2 files, +45 -12)
✓ web: committed def5678 (1 file, +30 -5)

Summary: 2 committed, 0 skipped, 0 failed
```

### 4. Push All Commits

Push all repos with unpushed commits:

```bash
fa push
```

Output:
```
Pushing 2 repositories...

✓ api: pushed 1 commit (main -> origin/main)
✓ web: pushed 1 commit (main -> origin/main)

Summary: 2 pushed, 0 skipped, 0 failed
```

## Common Patterns

### Stage and Commit in One Step

Use `-a` to stage all tracked file modifications:

```bash
fa commit -a "Quick fix across repos"
```

### Preview Before Committing

Use `--dry-run` to see what would be committed:

```bash
fa commit --dry-run "Feature X"
```

### Commit Specific Repos Only

Limit the operation to specific repos:

```bash
fa commit --repo api --repo lib "API-only changes"
```

### Amend Previous Commits

Update the most recent commits:

```bash
fa commit --amend "Updated commit message"
```

### JSON Output for Automation

Get structured output for scripts or AI agents:

```bash
fa commit --json "Automated commit" | jq '.summary'
```

## Error Handling

### Nothing to Commit

If no repos have staged changes:

```
Nothing to commit across any repository
Hint: Stage changes first or use -a to stage all modifications
```

### Partial Failures

If some repos fail (e.g., pre-commit hook):

```
Committing across 3 repositories...

✓ api: committed abc1234 (2 files)
✗ web: failed (pre-commit hook: eslint errors)
✓ lib: committed def5678 (1 file)

Summary: 2 committed, 0 skipped, 1 failed
```

The successful commits are preserved. Fix the failing repo and retry:

```bash
# Fix the issues in web repo
cd repos/worktrees/web/main
npm run lint:fix
git add .

# Commit just the fixed repo
fa commit --repo web "Add user profile feature"
```

### Push Rejected

If remote has new commits:

```
✗ api: failed (remote has new commits)

Hint: Run 'fa sync --pull' to fetch and merge, then retry push
```

## Tips

1. **Consistent Messages**: Use descriptive messages that make sense across all repos
2. **Review First**: Use `--dry-run` before large commits
3. **Atomic Features**: Group related changes across repos into single commits
4. **Handle Failures**: Partial successes are preserved; fix failures individually
