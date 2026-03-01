package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/auth"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE:  runProjectsList,
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
}

func runProjectsList(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)

	projects, _, err := client.ListProjects("", 100)
	if err != nil {
		return fmt.Errorf("listing projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects yet. Run 'sota deploy' to create one.")
		return nil
	}

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	bold.Printf("%-30s %-20s %s\n", "NAME", "SLUG", "CREATED")
	for _, p := range projects {
		fmt.Printf("%-30s %-20s %s\n",
			p.Name,
			dim.Sprint(p.Slug),
			dim.Sprint(p.CreatedAt.Format("2006-01-02")),
		)
	}

	return nil
}
