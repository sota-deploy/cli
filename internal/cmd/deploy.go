package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/archive"
	"github.com/sota-io/sota-cli/internal/auth"
	"github.com/sota-io/sota-cli/internal/config"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current directory",
	Long:  "Packages the current directory as a tar.gz archive, uploads it to sota.io, and streams build logs.",
	RunE:  runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	// 1. Get auth token
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)

	// 2. Load or create project config
	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		yellow.Println("No .sota.json found. Let's link this directory to a project.")
		fmt.Println()

		// List projects
		projects, _, listErr := client.ListProjects("", 20)
		if listErr != nil {
			return fmt.Errorf("listing projects: %w", listErr)
		}

		if len(projects) > 0 {
			fmt.Println("Your projects:")
			for i, p := range projects {
				fmt.Printf("  %d. %s (%s)\n", i+1, p.Name, p.Slug)
			}
			fmt.Printf("  %d. Create new project\n", len(projects)+1)
			fmt.Println()
		}

		// For now, create a new project
		fmt.Print("Project name: ")
		reader := bufio.NewReader(os.Stdin)
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		if name == "" {
			return fmt.Errorf("project name required")
		}

		project, createErr := client.CreateProject(name)
		if createErr != nil {
			return fmt.Errorf("creating project: %w", createErr)
		}

		projCfg = &config.ProjectConfig{
			ProjectID:   project.ID,
			ProjectSlug: project.Slug,
		}
		if err := config.SaveProjectConfig(projCfg); err != nil {
			return fmt.Errorf("saving project config: %w", err)
		}

		green.Printf("Project '%s' created (%s.sota.io)\n", project.Name, project.Slug)
		fmt.Println()
	}

	bold.Printf("Deploying %s...\n", projCfg.ProjectSlug)
	fmt.Println()

	// 3. Create archive
	cyan.Print("Packaging... ")
	dir, _ := os.Getwd()
	archiveFile, size, err := archive.Create(dir)
	if err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	defer func() {
		archiveFile.Close()
		os.Remove(archiveFile.Name())
	}()
	fmt.Printf("%s\n", archive.FormatSize(size))

	// 4. Upload
	cyan.Print("Uploading... ")
	deployment, err := client.Deploy(projCfg.ProjectID, archiveFile)
	if err != nil {
		return fmt.Errorf("uploading: %w", err)
	}
	fmt.Println("done")
	fmt.Println()

	// 5. Stream build logs
	cyan.Println("Build logs:")
	stream, err := client.StreamLogs(projCfg.ProjectID, deployment.ID)
	if err != nil {
		// Fall back to polling if SSE not available
		yellow.Println("(streaming not available, check logs with 'sota logs')")
	} else {
		defer stream.Close()
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				fmt.Println(strings.TrimPrefix(line, "data: "))
			}
		}
	}

	// 6. Final result
	fmt.Println()
	if deployment.URL != nil {
		green.Printf("Deployed to https://%s\n", *deployment.URL)
	} else {
		green.Printf("Deployed to https://%s.sota.io\n", projCfg.ProjectSlug)
	}

	return nil
}
