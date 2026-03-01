package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/auth"
)

var authSetKeyCmd = &cobra.Command{
	Use:   "set-key <api-key>",
	Short: "Store an API key for authentication",
	Long: `Stores an API key in ~/.config/sota/credentials.json.
The stored key is used for all CLI commands (overridden by SOTA_API_KEY env var).

Get your API key from the dashboard: https://sota.io/dashboard/settings`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := strings.TrimSpace(args[0])

		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		if !strings.HasPrefix(apiKey, "sota_") {
			return fmt.Errorf("invalid API key format (expected 'sota_' prefix)")
		}

		if err := auth.SaveAPIKey(apiKey); err != nil {
			return fmt.Errorf("saving API key: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Println("API key saved successfully!")
		fmt.Printf("Stored in: %s\n", auth.CredentialsPath())
		fmt.Println("\nYou can now use all sota commands. Try: sota projects list")
		return nil
	},
}
