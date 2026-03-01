package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/auth"
	"github.com/sota-io/sota-cli/internal/config"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to the previous deployment",
	Long:  "Reverts your app to the previous successful deployment by swapping container images.",
	RunE:  runRollback,
}

func runRollback(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)

	yellow := color.New(color.FgYellow, color.Bold)
	yellow.Printf("Rolling back %s...\n", projCfg.ProjectSlug)

	deployment, err := client.Rollback(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	green := color.New(color.FgGreen, color.Bold)
	if deployment.URL != nil {
		green.Printf("Rolled back to https://%s\n", *deployment.URL)
	} else {
		green.Printf("Rollback complete (deployment %s)\n", deployment.ID[:8])
	}
	return nil
}
