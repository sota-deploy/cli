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

var (
	projectsCreateGitURL    string
	projectsCreateGitBranch string
)

var projectsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project (optionally connected to a git repository)",
	Long: `Create a new sota.io project. The slug is auto-generated from the name.

Pass --git <URL> to record an association with a public repository at the
same time (informational — does NOT enable auto-deploy-on-push). The
dashboard will display the connection on the project detail page.`,
	Args: cobra.ExactArgs(1),
	RunE: runProjectsCreate,
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsCreateCmd)

	projectsCreateCmd.Flags().StringVar(&projectsCreateGitURL, "git", "", "Optional public git repository URL to associate with the project")
	projectsCreateCmd.Flags().StringVar(&projectsCreateGitBranch, "branch", "", "Optional branch name (only meaningful with --git)")
}

func runProjectsCreate(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}
	client := api.NewClient(apiURL, token)

	project, err := client.CreateProject(args[0], projectsCreateGitURL, projectsCreateGitBranch)
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}

	green := color.New(color.FgGreen, color.Bold)
	dim := color.New(color.Faint)
	green.Printf("Project '%s' created\n", project.Name)
	fmt.Printf("  Slug:   %s\n", dim.Sprint(project.Slug))
	fmt.Printf("  ID:     %s\n", dim.Sprint(project.ID))
	fmt.Printf("  URL:    https://%s.sota.io\n", project.Slug)
	if project.GitURL != nil && *project.GitURL != "" {
		branch := ""
		if project.GitBranch != nil && *project.GitBranch != "" {
			branch = " @ " + *project.GitBranch
		}
		fmt.Printf("  Git:    %s%s\n", *project.GitURL, branch)
	}
	return nil
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
