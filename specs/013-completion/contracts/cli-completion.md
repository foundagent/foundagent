# CLI Contract: Shell Completion

**Feature Branch**: `013-completion`  
**Date**: 2025-12-08

## Command: `fa completion`

### Synopsis

```
fa completion <shell>
fa completion bash
fa completion zsh
fa completion fish
fa completion powershell
foundagent completion <shell>
```

### Description

Generate shell completion script for the specified shell. The script is written to stdout and includes installation instructions as comments.

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `shell` | Yes | Target shell: `bash`, `zsh`, `fish`, or `powershell` |

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--help` | bool | false | Show help with installation instructions |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - script written to stdout |
| 1 | Error - invalid shell argument |

### Examples

```bash
# Generate bash completion and save to file
fa completion bash > ~/.bash_completion.d/fa

# Source directly for current session
source <(fa completion zsh)

# PowerShell
fa completion powershell | Out-String | Invoke-Expression

# Fish
fa completion fish > ~/.config/fish/completions/fa.fish
```

---

## Output Contract: Bash Completion Script

### Format

Shell script (Bash syntax)

### Structure

```bash
# <installation-instructions-header>
# ...

# Cobra-generated completion script
_foundagent_completions() {
    # ... completion logic
}
complete -o default -F _foundagent_completions foundagent
```

### Contract

- MUST be valid Bash syntax (parseable by Bash 4+)
- MUST define completion function for `foundagent` command
- MUST include installation instructions in comments
- MUST call `foundagent __complete` for dynamic completions
- SHOULD include alias setup instructions for `fa`

---

## Output Contract: Zsh Completion Script

### Format

Shell script (Zsh syntax with `#compdef` header)

### Structure

```zsh
#compdef foundagent

_foundagent() {
    # ... completion logic
}

compdef _foundagent foundagent
```

### Contract

- MUST be valid Zsh syntax (parseable by Zsh 5+)
- MUST include `#compdef` directive
- MUST support completion descriptions
- SHOULD include `compdef _foundagent fa` instruction for alias

---

## Output Contract: Fish Completion Script

### Format

Shell script (Fish syntax)

### Structure

```fish
# Installation instructions
# ...

complete -c foundagent -f -n "__fish_seen_subcommand_from completion" -a "bash zsh fish powershell"
# ... more completion definitions
```

### Contract

- MUST be valid Fish syntax (parseable by Fish 3+)
- MUST use `complete -c foundagent` commands
- MUST include `-f` (no file completion) where appropriate

---

## Output Contract: PowerShell Completion Script

### Format

PowerShell script

### Structure

```powershell
# Installation instructions
# ...

$__foundagentCompleterBlock = {
    # ... completion logic
}
Register-ArgumentCompleter -Native -CommandName 'foundagent' -ScriptBlock $__foundagentCompleterBlock
```

### Contract

- MUST be valid PowerShell syntax (PS 5.1+)
- MUST use `Register-ArgumentCompleter`
- SHOULD include `Register-ArgumentCompleter` for `fa` alias

---

## Dynamic Completion Contract

### `__complete` Hidden Command

Cobra generates a hidden `__complete` command that provides completions programmatically.

```bash
# Get completions for "fa worktree switch" with partial input "fea"
foundagent __complete worktree switch fea

# Output (one completion per line, directive on last line):
feature-auth
feature-login
:4  # ShellCompDirective
```

### Directive Values

| Value | Meaning |
|-------|---------|
| `:0` | Default (allow file completion fallback) |
| `:4` | No file completion (`ShellCompDirectiveNoFileComp`) |

### Dynamic Completion Behavior

| Command | Completes | Outside Workspace |
|---------|-----------|-------------------|
| `fa wt switch <TAB>` | Worktree names from state | Empty list |
| `fa wt remove <TAB>` | Worktree names from state | Empty list |
| `fa remove <TAB>` | Repo names from config | Empty list |
| `fa completion <TAB>` | "bash", "zsh", "fish", "powershell" | Same (static) |

---

## Error Contract

### Invalid Shell Error

**Trigger**: `fa completion invalid-shell`

**Exit Code**: 1

**Stderr Output**:
```
Error E001: unsupported shell "invalid-shell"
Supported shells: bash, zsh, fish, powershell
```

### Missing Shell Argument

**Trigger**: `fa completion` (no argument)

**Exit Code**: 1

**Stderr Output**:
```
Error: accepts 1 arg(s), received 0
Usage:
  fa completion <shell> [flags]

Available shells:
  bash        Generate bash completion script
  zsh         Generate zsh completion script
  fish        Generate fish completion script
  powershell  Generate PowerShell completion script

Use "fa completion --help" for more information.
```

---

## Installation Instructions Contract

Each completion script MUST include platform-appropriate installation instructions as comments. Example for Bash:

```bash
# Foundagent completion for Bash
#
# Installation (choose one):
#
#   # Current session only:
#   source <(fa completion bash)
#
#   # Every session (Linux):
#   fa completion bash | sudo tee /etc/bash_completion.d/fa > /dev/null
#
#   # Every session (macOS with Homebrew):
#   fa completion bash > $(brew --prefix)/etc/bash_completion.d/fa
#
#   # Every session (macOS without Homebrew):
#   fa completion bash > ~/.bash_completion.d/fa
#   echo 'source ~/.bash_completion.d/fa' >> ~/.bashrc
#
# Alias support:
#   alias fa=foundagent
#   complete -o default -F __start_foundagent fa
```
