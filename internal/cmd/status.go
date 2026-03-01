package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/auth"
	"github.com/sota-io/sota-cli/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current deployment status",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	bold.Printf("Project: %s\n", projCfg.ProjectSlug)
	dim.Printf("URL: https://%s.sota.io\n", projCfg.ProjectSlug)
	fmt.Println()

	// Get latest deployments
	deployments, err := client.GetDeployments(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("getting deployments: %w", err)
	}

	if len(deployments) == 0 {
		fmt.Println("No deployments yet.")
		return nil
	}

	bold.Println("Recent deployments:")
	for i, d := range deployments {
		if i >= 5 {
			break
		}

		statusColor := dim
		switch d.Status {
		case "running":
			statusColor = green
		case "failed":
			statusColor = red
		}

		fmt.Printf("  %s  %s  %s\n",
			dim.Sprint(d.ID[:8]),
			statusColor.Sprint(d.Status),
			dim.Sprint(d.CreatedAt.Format("2006-01-02 15:04")),
		)
	}

	return nil
}
