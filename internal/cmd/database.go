package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/auth"
	"github.com/sota-io/sota-cli/internal/config"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage your project database",
	Long:  "View database status and connection details for your managed PostgreSQL database.",
}

var dbInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show database status and connection details",
	RunE:  runDBInfo,
}

func init() {
	dbCmd.AddCommand(dbInfoCmd)
}

func runDBInfo(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)

	db, err := client.GetDatabase(projCfg.ProjectID)
	if err != nil {
		apiErr, ok := err.(*api.APIError)
		if ok && apiErr.StatusCode == 404 {
			fmt.Println("No database provisioned for this project.")
			dim := color.New(color.Faint)
			dim.Println("Provision one from the dashboard at freedom.sota.io")
			return nil
		}
		return fmt.Errorf("fetching database info: %w", err)
	}

	bold := color.New(color.Bold)
	statusColor := dbStatusColor(db.Status)

	bold.Print("Status:   ")
	statusColor.Println(strings.ToUpper(db.Status))
	bold.Print("Host:     ")
	fmt.Println(db.Host)
	bold.Print("Port:     ")
	fmt.Println(db.Port)
	bold.Print("Database: ")
	fmt.Println(db.DBName)
	bold.Print("User:     ")
	fmt.Println(db.DBUser)
	bold.Print("Created:  ")
	fmt.Println(db.CreatedAt.Format("2006-01-02 15:04:05"))

	if db.Status != "running" {
		return nil
	}

	fmt.Println()
	details, err := client.GetDatabaseConnectionDetails(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("fetching connection details: %w", err)
	}

	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Println("Connection Details:")
	bold.Print("  Password:     ")
	fmt.Println(details.Password)
	bold.Print("  DATABASE_URL: ")
	fmt.Println(details.DatabaseURL)

	return nil
}

func dbStatusColor(status string) *color.Color {
	switch status {
	case "running":
		return color.New(color.FgGreen, color.Bold)
	case "provisioning":
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgRed)
	}
}
