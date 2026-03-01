package cmd

import "github.com/spf13/cobra"

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Manage authentication with sota.io. Use API keys for CI/CD or headless environments.`,
}

func init() {
	authCmd.AddCommand(authSetKeyCmd)
	authCmd.AddCommand(authStatusCmd)
}
