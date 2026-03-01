package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/auth"
	"github.com/sota-io/sota-cli/internal/config"
)

var (
	followLogs bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show build and runtime logs",
	Long:  "Displays logs for the latest deployment. Use -f to follow in real-time.",
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "follow log output")
}

func runLogs(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)

	// Get latest deployment
	deployments, err := client.GetDeployments(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("getting deployments: %w", err)
	}

	if len(deployments) == 0 {
		fmt.Println("No deployments found.")
		return nil
	}

	latest := deployments[0]
	dim := color.New(color.Faint)
	dim.Printf("Deployment: %s (%s)\n\n", latest.ID[:8], latest.Status)

	if followLogs {
		// Stream logs via SSE
		stream, err := client.StreamLogs(projCfg.ProjectID, latest.ID)
		if err != nil {
			return fmt.Errorf("connecting to log stream: %w", err)
		}
		defer stream.Close()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				fmt.Println(strings.TrimPrefix(line, "data: "))
			}
		}
		return nil
	}

	// Get static logs
	logs, err := client.GetLogs(projCfg.ProjectID, latest.ID)
	if err != nil {
		return fmt.Errorf("getting logs: %w", err)
	}

	if logs == "" {
		fmt.Println("No logs available yet.")
		return nil
	}

	fmt.Print(logs)
	return nil
}
