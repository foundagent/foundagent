package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name           string
		shell          string
		expectedString string
		expectError    bool
	}{
		{
			name:           "bash generates valid script",
			shell:          "bash",
			expectedString: "complete",
			expectError:    false,
		},
		{
			name:           "bash includes header",
			shell:          "bash",
			expectedString: "Installation:",
			expectError:    false,
		},
		{
			name:           "zsh generates valid script",
			shell:          "zsh",
			expectedString: "#compdef",
			expectError:    false,
		},
		{
			name:           "zsh includes header",
			shell:          "zsh",
			expectedString: "Installation:",
			expectError:    false,
		},
		{
			name:           "fish generates valid script",
			shell:          "fish",
			expectedString: "complete -c",
			expectError:    false,
		},
		{
			name:           "fish includes header",
			shell:          "fish",
			expectedString: "Installation:",
			expectError:    false,
		},
		{
			name:           "powershell generates valid script",
			shell:          "powershell",
			expectedString: "Register-ArgumentCompleter",
			expectError:    false,
		},
		{
			name:           "powershell includes header",
			shell:          "powershell",
			expectedString: "Installation:",
			expectError:    false,
		},
		{
			name:        "invalid shell returns error",
			shell:       "invalidshell",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdout
			oldStdout := os.Stdout

			// Create pipe to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create root command
			rootCmd := &cobra.Command{
				Use: "fa",
			}

			// Clone completion command to avoid state issues
			testCompletionCmd := &cobra.Command{
				Use:       "completion",
				Short:     "Generate shell completion scripts",
				ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
				Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
				RunE:      runCompletion,
			}
			rootCmd.AddCommand(testCompletionCmd)

			// Read output in goroutine to prevent pipe deadlock on Windows
			var buf bytes.Buffer
			done := make(chan struct{})
			go func() {
				_, _ = buf.ReadFrom(r)
				close(done)
			}()

			// Execute completion command
			rootCmd.SetArgs([]string{"completion", tt.shell})
			err := rootCmd.Execute()

			// Close writer and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Wait for reader to finish
			<-done
			output := buf.String()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedString, "Expected output to contain '%s'", tt.expectedString)
			}
		})
	}
}

func TestCompletionCommandNoArgs(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{
		Use: "fa",
	}

	// Clone completion command to avoid state issues
	testCompletionCmd := &cobra.Command{
		Use:       "completion",
		Short:     "Generate shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:      runCompletion,
	}
	rootCmd.AddCommand(testCompletionCmd)

	// Execute without arguments should show error
	rootCmd.SetArgs([]string{"completion"})
	err := rootCmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestCompletionShellValidation(t *testing.T) {
	validShells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range validShells {
		t.Run("valid_"+shell, func(t *testing.T) {
			// Save original stdout
			oldStdout := os.Stdout

			// Create pipe to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create root command
			rootCmd := &cobra.Command{
				Use: "fa",
			}

			// Clone completion command to avoid state issues
			testCompletionCmd := &cobra.Command{
				Use:       "completion",
				Short:     "Generate shell completion scripts",
				ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
				Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
				RunE:      runCompletion,
			}
			rootCmd.AddCommand(testCompletionCmd)

			// Read output in goroutine to prevent pipe deadlock on Windows
			var buf bytes.Buffer
			done := make(chan struct{})
			go func() {
				_, _ = buf.ReadFrom(r)
				close(done)
			}()

			rootCmd.SetArgs([]string{"completion", shell})
			err := rootCmd.Execute()

			// Close writer and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Wait for reader to finish
			<-done
			output := buf.String()

			assert.NoError(t, err)
			assert.NotEmpty(t, output, "Output should not be empty for shell: "+shell)
		})
	}
}

func TestCompletionAliasSupport(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create root command
	rootCmd := &cobra.Command{
		Use: "fa",
	}

	// Clone completion command to avoid state issues
	testCompletionCmd := &cobra.Command{
		Use:       "completion",
		Short:     "Generate shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:      runCompletion,
	}
	rootCmd.AddCommand(testCompletionCmd)

	// Read output in goroutine to prevent pipe deadlock on Windows
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(r)
		close(done)
	}()

	// Test bash completion for mention of alias support
	rootCmd.SetArgs([]string{"completion", "bash"})
	err := rootCmd.Execute()

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Wait for reader to finish
	<-done
	output := buf.String()

	assert.NoError(t, err)

	// Check that header mentions alias/foundagent support
	assert.True(t,
		strings.Contains(output, "foundagent") || strings.Contains(output, "alias"),
		"Completion script should mention alias support")
}
