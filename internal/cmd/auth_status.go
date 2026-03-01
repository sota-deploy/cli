package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/auth"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long:  `Shows which authentication method is active and validates the credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen)
		yellow := color.New(color.FgYellow)
		red := color.New(color.FgRed)

		bold.Println("Authentication Status")
		fmt.Println()

		// Check each auth source
		envKey := os.Getenv("SOTA_API_KEY")
		storedKey, _ := auth.LoadAPIKey()

		// Show available methods
		fmt.Println("Available methods:")
		if envKey != "" {
			truncLen := min(13, len(envKey))
			green.Printf("  + SOTA_API_KEY environment variable (%s...)\n", envKey[:truncLen])
		} else {
			fmt.Println("  - SOTA_API_KEY environment variable: not set")
		}

		if storedKey != "" {
			truncLen := min(13, len(storedKey))
			green.Printf("  + Stored API key (%s...)\n", storedKey[:truncLen])
		} else {
			fmt.Println("  - Stored API key: not configured")
		}

		fmt.Println()

		// Resolve which method would be used
		token, method, err := auth.GetToken()
		if err != nil {
			red.Println("Status: Not authenticated")
			fmt.Println("\nTo authenticate:")
			fmt.Println("  sota login                    Device code login (opens browser)")
			fmt.Println("  sota auth set-key <api-key>   Store an API key")
			fmt.Println("  export SOTA_API_KEY=sota_...  Set environment variable")
			return nil
		}

		bold.Printf("Active method: %s\n", method)

		// Validate by calling the API
		yellow.Println("Validating credentials...")
		req, _ := http.NewRequest("GET", apiURL+"/v1/api-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			red.Printf("Validation failed: %v\n", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			green.Println("Status: Authenticated +")
		} else if resp.StatusCode == 401 {
			red.Println("Status: Invalid credentials (401)")
			fmt.Println("Your stored credentials may have expired or been revoked.")
		} else {
			yellow.Printf("Status: API returned %d (may be a server issue)\n", resp.StatusCode)
		}

		fmt.Println("\nAlternative: sota auth set-key <api-key> (get key at sota.io/dashboard)")

		return nil
	},
}
