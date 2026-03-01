package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/auth"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.ClearCredentials(); err != nil {
			return fmt.Errorf("clearing credentials: %w", err)
		}
		fmt.Println("Logged out.")
		return nil
	},
}
