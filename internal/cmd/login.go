package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with sota.io",
	Long: `Authenticates with sota.io using a device code flow.
Opens your browser where you log in and enter the code shown in your terminal.

Alternative for CI/headless: sota auth set-key <api-key>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := auth.DeviceLogin(apiURL)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if err := auth.SaveAPIKey(apiKey); err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Println("\nLogged in successfully!")
		fmt.Printf("API key stored in: %s\n", auth.CredentialsPath())
		fmt.Println("\nTry: sota projects list")
		return nil
	},
}
