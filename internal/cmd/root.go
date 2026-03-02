package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Build-time variables injected via ldflags
	version = "dev"
	commit  = "none"

	// Global flags
	noColor bool
	jsonOut bool
	apiURL  string
)

var rootCmd = &cobra.Command{
	Use:   "sota",
	Short: "sota - deploy web apps with a single command",
	Long: `sota is the CLI for sota.io, an EU-native deployment platform.
Deploy, manage, and debug your apps from the terminal.

Get started:
  sota login        Authenticate with sota.io
  sota deploy       Deploy the current directory
  sota logs -f      Follow build and runtime logs`,
	Version: fmt.Sprintf("%s (%s)", version, commit),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.sota.io", "API base URL")

	// Register subcommands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(domainsCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(authCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
