# Research: Shell Completion

**Feature Branch**: `013-completion`  
**Date**: 2025-12-08  
**Status**: Complete

## Overview

This research resolves technical unknowns for implementing `fa completion <shell>` command, focusing on Cobra's built-in completion capabilities and dynamic completion patterns.

---

## Decision 1: Completion Script Generation Method

**Decision**: Use Cobra's built-in `GenXxxCompletion()` methods on `*cobra.Command`

**Rationale**: 
- Cobra provides native completion generation for all four target shells (Bash, Zsh, Fish, PowerShell)
- Script generation is maintained by the Cobra team and follows shell best practices
- Dynamic completions are automatically supported via `ValidArgsFunction`
- No need to hand-write or maintain shell scripts

**Alternatives Considered**:
1. **Hand-written shell scripts**: Rejected — high maintenance burden, error-prone across shells
2. **Third-party completion libraries**: Rejected — Cobra already provides robust solution

**Implementation**:
```go
// Bash (use V2 for dynamic completions)
cmd.GenBashCompletionV2(os.Stdout, true)  // true = include descriptions

// Zsh (with descriptions)
cmd.GenZshCompletion(os.Stdout)

// Fish (with descriptions)
cmd.GenFishCompletion(os.Stdout, true)

// PowerShell (with descriptions)
cmd.GenPowerShellCompletionWithDesc(os.Stdout)
```

---

## Decision 2: Dynamic Completion Pattern

**Decision**: Use `ValidArgsFunction` with workspace state lookup for context-aware completions

**Rationale**:
- `ValidArgsFunction` is Cobra's standard mechanism for runtime completions
- Function receives partial input (`toComplete`) for prefix matching
- Can read local config files without network calls (fast, offline-capable)
- `ShellCompDirective` controls shell behavior (no file fallback, keep order)

**Alternatives Considered**:
1. **Static completion only**: Rejected — loses significant UX value (worktree names, repo names)
2. **External completion script**: Rejected — breaks Cobra's integrated model

**Implementation Pattern**:
```go
var worktreeSwitchCmd = &cobra.Command{
    Use:   "switch <worktree>",
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
        if len(args) != 0 {
            return nil, cobra.ShellCompDirectiveNoFileComp
        }
        
        // Read from workspace config/state (local, fast)
        worktrees := getWorktreeNamesFromWorkspace()
        
        var completions []cobra.Completion
        for _, wt := range worktrees {
            if strings.HasPrefix(wt, toComplete) {
                completions = append(completions, cobra.CompletionWithDesc(wt, "worktree"))
            }
        }
        return completions, cobra.ShellCompDirectiveNoFileComp
    },
}
```

---

## Decision 3: Handling Alias Completions (fa vs foundagent)

**Decision**: Generate completion scripts for `foundagent` binary; document alias setup in help text

**Rationale**:
- Cobra scripts call the binary via `__complete` hidden command
- Aliases work by registering the same completion function for both names
- Each shell has different alias setup requirements

**Shell-Specific Alias Setup**:

| Shell | Alias Completion Setup |
|-------|----------------------|
| Bash | `alias fa=foundagent && complete -o default -F __start_foundagent fa` |
| Zsh | Add `compdef _foundagent fa` after sourcing completion |
| Fish | Aliases automatically inherit completions |
| PowerShell | `Set-Alias fa foundagent; Register-ArgumentCompleter -CommandName fa -ScriptBlock $__foundagentCompleterBlock` |

**Implementation**: Include alias setup in installation instructions (comments in script or `--help`).

---

## Decision 4: Performance and Timeout Handling

**Decision**: Use local file reads only; no external calls during completion

**Rationale**:
- Reading `.foundagent.yaml` and `.foundagent/state.json` is fast (<10ms)
- No network calls = no timeout needed
- Graceful degradation: return empty list if workspace not found

**Performance Requirements (per spec)**:
- Dynamic completions must respond in <500ms (FR-022)
- Support up to 20 repos (SC-003)

**Implementation Strategy**:
1. Attempt to find workspace root (walk up directory tree)
2. If found, parse config file for repo/worktree names
3. If not found or parse fails, return empty completions (no error to user)
4. Use `ShellCompDirectiveNoFileComp` to prevent file fallback

---

## Decision 5: Graceful Degradation Outside Workspace

**Decision**: Return empty completions when not in a Foundagent workspace

**Rationale**:
- Per FR-021: "Dynamic completions MUST gracefully return empty when outside workspace"
- Users may run `fa --help` or `fa completion` outside workspaces
- Static completions (commands, flags) still work everywhere

**Implementation**:
```go
func getWorktreeCompletions(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
    ws, err := workspace.Discover()
    if err != nil {
        // Not in workspace - return empty, no error
        return nil, cobra.ShellCompDirectiveNoFileComp
    }
    // ... proceed with dynamic completions
}
```

---

## Decision 6: Command Structure

**Decision**: Use `completion` as noun with shell name as positional argument

**Rationale**:
- Follows constitution pattern: `fa <noun> <verb>` (completion is noun, shell is argument)
- Matches Cobra's convention and other CLI tools (kubectl, gh, docker)
- Subcommands per shell would add unnecessary nesting

**Command Interface**:
```
fa completion <shell>
fa completion bash
fa completion zsh
fa completion fish
fa completion powershell
```

**Invalid Shell Handling (FR-007)**:
```
Error: unsupported shell "tcsh"
Supported shells: bash, zsh, fish, powershell
```

---

## Decision 7: Include Installation Instructions in Output

**Decision**: Include installation instructions as comments at the top of generated scripts

**Rationale**:
- Per FR-010: "Scripts MUST include installation instructions as comments"
- Users can save script and see instructions without referring to external docs
- Cobra's generated scripts already include some instructions; we can add Foundagent-specific notes

**Implementation**: Wrap Cobra's output with header comments:
```bash
#!/bin/bash
# Foundagent shell completion for Bash
#
# Installation:
#   # For current session only:
#   source <(fa completion bash)
#
#   # To load completions for every session (Linux):
#   fa completion bash > /etc/bash_completion.d/fa
#
#   # To load completions for every session (macOS):
#   fa completion bash > $(brew --prefix)/etc/bash_completion.d/fa
#
#   # Enable for alias:
#   alias fa=foundagent
#   complete -o default -F __start_foundagent fa
#
# Cobra-generated script follows:
...
```

---

## Decision 8: Testing Strategy

**Decision**: Table-driven tests validating script output format and dynamic completion results

**Rationale**:
- Per constitution: use table-driven test patterns
- Cannot easily test actual shell completion behavior in unit tests
- Can validate:
  - Script output is non-empty and contains expected shell syntax
  - Dynamic completion functions return expected values given workspace state
  - Error handling for invalid shell arguments

**Test Categories**:

| Test Type | What to Validate |
|-----------|-----------------|
| Unit | `completion.go` returns valid script content for each shell |
| Unit | Dynamic completion functions return correct worktree/repo names |
| Unit | Outside-workspace returns empty completions |
| Unit | Invalid shell argument produces error with supported list |
| Integration | Generated script is syntactically valid (optional, complex) |

---

## API Reference

### Cobra Completion Methods (v1.7+)

| Shell | Method | Description Control |
|-------|--------|---------------------|
| Bash | `GenBashCompletionV2(w, includeDesc)` | Bool param |
| Zsh | `GenZshCompletion(w)` / `GenZshCompletionNoDesc(w)` | Separate methods |
| Fish | `GenFishCompletion(w, includeDesc)` | Bool param |
| PowerShell | `GenPowerShellCompletionWithDesc(w)` / `GenPowerShellCompletion(w)` | Separate methods |

### ShellCompDirective Flags

| Directive | Use Case |
|-----------|----------|
| `ShellCompDirectiveNoFileComp` | Prevent file completion fallback |
| `ShellCompDirectiveNoSpace` | Don't add space after completion |
| `ShellCompDirectiveKeepOrder` | Preserve completion order |
| `ShellCompDirectiveError` | Signal error occurred |

### Debugging Completions

```bash
# Test completions directly
foundagent __complete worktree switch ""
foundagent __complete worktree switch "fea"

# Enable debug logging (Bash)
export BASH_COMP_DEBUG_FILE=/tmp/cobra-completion.log
```

---

## Summary

All technical unknowns resolved. Key decisions:
1. Use Cobra's native completion generation
2. Implement dynamic completions via `ValidArgsFunction`
3. Read local workspace files only (no network, fast)
4. Gracefully degrade outside workspace
5. Include installation instructions in script comments
6. Test output format and completion function logic
