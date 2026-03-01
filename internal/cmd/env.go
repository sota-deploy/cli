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

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Long:  "Set, get, and list environment variables for your app.",
}

var envSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE",
	Short: "Set an environment variable",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvSet,
}

var envGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get an environment variable",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvGet,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment variables",
	RunE:  runEnvList,
}

func init() {
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envListCmd)
}

func runEnvSet(cmd *cobra.Command, args []string) error {
	parts := strings.SplitN(args[0], "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("format: KEY=VALUE")
	}
	key, value := parts[0], parts[1]

	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	if err := client.SetEnvVar(projCfg.ProjectID, key, value); err != nil {
		return fmt.Errorf("setting env var: %w", err)
	}

	green := color.New(color.FgGreen)
	green.Printf("Set %s\n", key)
	return nil
}

func runEnvGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	envVars, err := client.ListEnvVars(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("getting env vars: %w", err)
	}

	for _, ev := range envVars {
		if ev.Key == key {
			fmt.Println(ev.Value)
			return nil
		}
	}

	return fmt.Errorf("env var %s not found", key)
}

func runEnvList(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	envVars, err := client.ListEnvVars(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("listing env vars: %w", err)
	}

	if len(envVars) == 0 {
		fmt.Println("No environment variables set.")
		return nil
	}

	dim := color.New(color.Faint)
	for _, ev := range envVars {
		// Mask values for security
		masked := maskValue(ev.Value)
		fmt.Printf("%-30s %s\n", ev.Key, dim.Sprint(masked))
	}
	return nil
}

// maskValue shows first 4 chars and masks the rest.
func maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + strings.Repeat("*", min(len(value)-4, 20))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
