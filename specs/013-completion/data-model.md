# Data Model: Shell Completion

**Feature Branch**: `013-completion`  
**Date**: 2025-12-08  
**Status**: Complete

## Overview

This document defines the data entities involved in the shell completion feature. The completion command is primarily read-only, generating scripts to stdout and reading workspace state for dynamic completions.

---

## Entities

### 1. Completion Script

**Description**: Shell-specific script that enables tab completion for Foundagent commands.

**Attributes**:

| Field | Type | Description |
|-------|------|-------------|
| shell | string | Target shell identifier ("bash", "zsh", "fish", "powershell") |
| content | string | Complete script content ready for sourcing |
| includesDescriptions | bool | Whether completions include help text |

**Notes**:
- Not persisted; generated on-demand to stdout
- Content includes installation instructions as comments

---

### 2. Static Completion

**Description**: Fixed completions for commands, subcommands, and flags. These are known at compile time.

**Attributes**:

| Field | Type | Description |
|-------|------|-------------|
| name | string | Command/flag name (e.g., "worktree", "--force") |
| description | string | Short help text for completion UI |
| aliases | []string | Alternative names (e.g., "wt" for "worktree") |

**Static Completion Set** (per FR-012 through FR-016):

| Command | Subcommands | Global Flags | Command Flags |
|---------|-------------|--------------|---------------|
| `init` | - | `--help`, `--version`, `--json`, `-v` | `--force` |
| `add` | - | (global) | - |
| `remove` | - | (global) | `--force` |
| `worktree` / `wt` | `create`, `list`, `remove`, `switch` | (global) | per-subcommand |
| `status` | - | (global) | - |
| `sync` | - | (global) | - |
| `config` | - | (global) | - |
| `version` | - | (global) | - |
| `doctor` | - | (global) | - |
| `completion` | - | (global) | - |

---

### 3. Dynamic Completion

**Description**: Context-aware completions based on current workspace state. Resolved at runtime.

**Attributes**:

| Field | Type | Description |
|-------|------|-------------|
| name | string | Completion value (e.g., "feature-auth") |
| description | string | Optional context (e.g., "worktree", "repo") |
| source | string | Where value comes from (config file, state file, git) |

**Dynamic Completion Sources**:

| Command | Completes | Source |
|---------|-----------|--------|
| `fa wt switch <TAB>` | Worktree/branch names | `.foundagent/state.json` → worktrees list |
| `fa wt remove <TAB>` | Worktree names | `.foundagent/state.json` → worktrees list |
| `fa remove <TAB>` | Repository names | `.foundagent.yaml` → repos list |

---

### 4. Workspace (read-only reference)

**Description**: Existing entity from workspace-init spec. Completion reads but does not modify.

**Relevant Fields for Completion**:

| Field | Source File | Used For |
|-------|-------------|----------|
| repos | `.foundagent.yaml` | `fa remove <TAB>` repo name completion |
| worktrees | `.foundagent/state.json` | `fa wt switch/remove <TAB>` completion |

**Discovery**: Walk up from CWD looking for `.foundagent.yaml` marker file.

---

## Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                      Shell Completion                        │
│  (command generates script to stdout)                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Completion Script                         │
│  - Shell-specific syntax                                     │
│  - Calls `foundagent __complete` for dynamic values          │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────────┐
│   Static Completions     │     │   Dynamic Completions       │
│   (compile-time)         │     │   (runtime)                 │
│   - commands             │     │   - worktree names          │
│   - subcommands          │     │   - repo names              │
│   - flags                │     │                             │
└─────────────────────────┘     └─────────────────────────────┘
                                              │
                                              ▼
                              ┌───────────────────────────────┐
                              │        Workspace State         │
                              │  .foundagent.yaml (repos)     │
                              │  .foundagent/state.json (wts) │
                              └───────────────────────────────┘
```

---

## Validation Rules

### Shell Name Validation

| Rule | Constraint |
|------|------------|
| Allowed values | "bash", "zsh", "fish", "powershell" |
| Case sensitivity | Case-insensitive input, normalized to lowercase |
| Error on invalid | Return error with list of valid options |

### Dynamic Completion Validation

| Rule | Constraint |
|------|------------|
| Workspace not found | Return empty completions (no error) |
| Config parse error | Return empty completions (log debug) |
| Empty repo/worktree list | Return empty completions |
| Timeout (target: 500ms) | Not enforced programmatically; design ensures fast local reads |

---

## State Transitions

The completion command is **stateless** — it reads but never modifies workspace state.

| Action | State Change |
|--------|--------------|
| `fa completion <shell>` | None (writes to stdout only) |
| Dynamic completion lookup | None (reads config/state files) |

---

## Data Access Patterns

### Read Patterns

| Operation | File | Frequency |
|-----------|------|-----------|
| Generate script | None | Once per command |
| Find workspace root | Walk up looking for `.foundagent.yaml` | Per completion request |
| Load repo names | `.foundagent.yaml` | Per completion request |
| Load worktree names | `.foundagent/state.json` | Per completion request |

### Write Patterns

None. The completion feature is read-only.

---

## Error States

| Error Condition | Behavior |
|-----------------|----------|
| Invalid shell argument | Error with code, message, valid options list |
| Not in workspace (dynamic) | Return empty completions, no error |
| Corrupt config file (dynamic) | Return empty completions, no error |
| No repos configured | Return empty completions for `fa remove` |
| No worktrees exist | Return empty completions for `fa wt switch/remove` |
