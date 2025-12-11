package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for Foundagent.

The completion script will be output to stdout. You can source it directly or
save it to a file and source it from your shell's configuration file.

Supported shells: bash, zsh, fish, powershell

Examples:
  # Generate Bash completion and load it in current session
  source <(fa completion bash)

  # Generate Zsh completion and save to file
  fa completion zsh > ~/.zsh/completion/_fa

  # Generate Fish completion
  fa completion fish > ~/.config/fish/completions/fa.fish

  # Generate PowerShell completion
  fa completion powershell > fa_completion.ps1`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactValidArgs(1),
	RunE:      runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := strings.ToLower(args[0])
	
	// Validate shell
	validShells := []string{"bash", "zsh", "fish", "powershell"}
	isValid := false
	for _, valid := range validShells {
		if shell == valid {
			isValid = true
			break
		}
	}
	
	if !isValid {
		return fmt.Errorf("unsupported shell: %s. Supported shells: %s", shell, strings.Join(validShells, ", "))
	}

	// Generate completion script based on shell
	switch shell {
	case "bash":
		return generateBashCompletion(cmd)
	case "zsh":
		return generateZshCompletion(cmd)
	case "fish":
		return generateFishCompletion(cmd)
	case "powershell":
		return generatePowerShellCompletion(cmd)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func generateBashCompletion(cmd *cobra.Command) error {
	header := `# Bash completion for Foundagent
# 
# Installation:
#
# 1. Current session only:
#    source <(fa completion bash)
#
# 2. Permanent (Linux):
#    fa completion bash > /etc/bash_completion.d/fa
#
# 3. Permanent (macOS with Homebrew):
#    fa completion bash > $(brew --prefix)/etc/bash_completion.d/fa
#
# 4. Permanent (manual):
#    fa completion bash > ~/.bash_completion_fa
#    Add to ~/.bashrc: source ~/.bash_completion_fa
#
# 5. For 'foundagent' alias (add to ~/.bashrc after sourcing completion):
#    complete -F __start_fa foundagent
#

`
	fmt.Fprint(os.Stdout, header)
	return cmd.Root().GenBashCompletionV2(os.Stdout, true)
}

func generateZshCompletion(cmd *cobra.Command) error {
	header := `# Zsh completion for Foundagent
#
# Installation:
#
# 1. Current session only:
#    source <(fa completion zsh)
#
# 2. Permanent (Oh My Zsh):
#    fa completion zsh > ~/.oh-my-zsh/completions/_fa
#
# 3. Permanent (custom fpath):
#    fa completion zsh > /usr/local/share/zsh/site-functions/_fa
#    or:
#    mkdir -p ~/.zsh/completion
#    fa completion zsh > ~/.zsh/completion/_fa
#    Add to ~/.zshrc: fpath=(~/.zsh/completion $fpath)
#
# 4. Then reload shell:
#    exec zsh
#
# 5. For 'foundagent' alias (add to ~/.zshrc after fpath setup):
#    compdef _fa foundagent
#

`
	fmt.Fprint(os.Stdout, header)
	return cmd.Root().GenZshCompletion(os.Stdout)
}

func generateFishCompletion(cmd *cobra.Command) error {
	header := `# Fish completion for Foundagent
#
# Installation:
#
# 1. User-specific:
#    fa completion fish > ~/.config/fish/completions/fa.fish
#
# 2. System-wide:
#    fa completion fish > /usr/share/fish/vendor_completions.d/fa.fish
#
# 3. Then reload Fish:
#    source ~/.config/fish/config.fish
#
# 4. For 'foundagent' command, Fish automatically handles it
#

`
	fmt.Fprint(os.Stdout, header)
	return cmd.Root().GenFishCompletion(os.Stdout, true)
}

func generatePowerShellCompletion(cmd *cobra.Command) error {
	header := `# PowerShell completion for Foundagent
#
# Installation:
#
# 1. Current session only:
#    fa completion powershell | Out-String | Invoke-Expression
#
# 2. Permanent:
#    fa completion powershell > $HOME\Documents\PowerShell\Scripts\fa_completion.ps1
#    Add to $PROFILE:
#    . $HOME\Documents\PowerShell\Scripts\fa_completion.ps1
#
# 3. Reload profile:
#    . $PROFILE
#
# 4. For 'foundagent' alias, add to $PROFILE:
#    Register-ArgumentCompleter -CommandName foundagent -ScriptBlock $__faCompleterBlock
#

`
	fmt.Fprint(os.Stdout, header)
	return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
}
