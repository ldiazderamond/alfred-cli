package main

import (
	"fmt"
	"os"
	"strings"

	"alfred-cli/cmd"
	"alfred-cli/config"
	"alfred-cli/ollama"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "alfred-cli [query...]",
	Short: "A CLI tool that generates terminal commands using Ollama",
	Long: `alfred-cli is a command-line tool that uses Ollama to generate terminal commands
from natural language queries. It will prompt you for confirmation before executing
any generated commands.

Examples:
  alfred-cli "list all files in current directory"
  alfred-cli create a new directory called myproject
  alfred-cli "give me a command to initialise a new quarkus project called 'kkpo'"`,
	RunE:    runAlfred,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

func init() {
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	// Add -v as an alias for --version
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")
}

func runAlfred(cobraCmd *cobra.Command, args []string) error {
	// Check if version flag was set
	if v, _ := cobraCmd.Flags().GetBool("version"); v {
		cobraCmd.Println(cobraCmd.Version)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("please provide a query")
	}

	// Join all arguments as the query string
	query := strings.Join(args, " ")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Gather environment context
	envCtx, err := cmd.GatherContext()
	if err != nil {
		return fmt.Errorf("failed to gather environment context: %w", err)
	}
	contextStr := envCtx.Format()

	// Create Ollama client
	ollamaURL := config.GetOllamaURL(cfg)
	client := ollama.NewClient(ollamaURL, cfg.Ollama.Model)

	// Generate command
	fmt.Println("Generating command...")
	generatedCommand, err := client.GenerateCommand(cfg.SystemPrompt, query, contextStr)
	if err != nil {
		return fmt.Errorf("failed to generate command: %w", err)
	}

	if generatedCommand == "" {
		return fmt.Errorf("generated command is empty")
	}

	// Parse commands (support multiple commands separated by newlines)
	commands := cmd.ParseCommands(generatedCommand)

	// Confirm and execute
	confirmed, err := cmd.ConfirmExecution(commands)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		fmt.Println("Command execution cancelled.")
		return nil
	}

	// Execute the commands with retry on error
	maxRetries := 2
	retryCount := 0

	for {
		err := cmd.ExecuteCommands(commands)
		if err == nil {
			return nil
		}

		// Check if it's an ExecutionError that we can try to fix
		execErr, ok := err.(*cmd.ExecutionError)
		if !ok {
			return fmt.Errorf("execution failed: %w", err)
		}

		// Check retry limit
		if retryCount >= maxRetries {
			return fmt.Errorf("execution failed after %d retries: %w", maxRetries, err)
		}

		retryCount++
		fmt.Printf("\n⚠️  Command failed. Attempting to generate a fix (attempt %d/%d)...\n", retryCount, maxRetries)

		// Generate a fix using Ollama
		fixCommand, fixErr := client.GenerateFix(query, execErr.Command, execErr.Combined)
		if fixErr != nil {
			return fmt.Errorf("failed to generate fix: %w\nOriginal error: %w", fixErr, err)
		}

		if fixCommand == "" {
			return fmt.Errorf("generated fix is empty\nOriginal error: %w", err)
		}

		// Parse the fix command
		fixCommands := cmd.ParseCommands(fixCommand)

		// Show the fix and ask for confirmation
		fmt.Printf("\nSuggested fix:\n")
		if len(fixCommands) == 1 {
			fmt.Printf("%s\n\n", fixCommands[0])
		} else {
			for i, cmd := range fixCommands {
				fmt.Printf("%d. %s\n", i+1, cmd)
			}
			fmt.Println()
		}

		confirmed, confirmErr := cmd.ConfirmExecution(fixCommands)
		if confirmErr != nil {
			return fmt.Errorf("failed to get confirmation: %w\nOriginal error: %w", confirmErr, err)
		}

		if !confirmed {
			return fmt.Errorf("fix execution cancelled\nOriginal error: %w", err)
		}

		// Use the fix commands for the next iteration
		commands = fixCommands
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
