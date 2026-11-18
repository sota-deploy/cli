package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
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
	Long: "Show database status and connection details. The password (also inside DATABASE_URL)\n" +
		"is masked unless --show-secrets is given, so the output is safe to paste or screen-share.",
	RunE: runDBInfo,
}

// dbShowSecrets prints the database password and the unmasked DATABASE_URL.
var dbShowSecrets bool

// maskedSecret replaces a secret in `sota db info` output.
const maskedSecret = "********"

func init() {
	dbInfoCmd.Flags().BoolVar(&dbShowSecrets, "show-secrets", false, "print the database password and the full DATABASE_URL")
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
			dim.Println("Provision one from the dashboard at https://sota.io/dashboard")
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

	printConnectionDetails(os.Stdout, details, dbShowSecrets)
	return nil
}

// printConnectionDetails writes the connection block of `sota db info`.
// Without showSecrets the password is masked, both on its own line and
// inside DATABASE_URL.
func printConnectionDetails(w io.Writer, details *api.DatabaseConnectionDetails, showSecrets bool) {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	password, databaseURL := details.Password, details.DatabaseURL
	if !showSecrets {
		password = maskedSecret
		databaseURL = maskDatabaseURL(details.DatabaseURL, details.Password)
	}
	cyan.Fprintln(w, "Connection Details:")
	bold.Fprint(w, "  Password:     ")
	fmt.Fprintln(w, password)
	bold.Fprint(w, "  DATABASE_URL: ")
	fmt.Fprintln(w, databaseURL)
	if !showSecrets {
		color.New(color.Faint).Fprintln(w, "  (password masked; run `sota db info --show-secrets` to print it)")
	}
}

// maskDatabaseURL replaces the password of a postgres:// URL with
// maskedSecret. Any other occurrence of the known password (a malformed URL,
// a ?password= query parameter, a percent-encoded form) is replaced as well,
// so the secret never leaks through an unexpected URL shape.
func maskDatabaseURL(raw, password string) string {
	masked := raw
	if u, err := url.Parse(raw); err == nil && u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			// url.URL.String() percent-encodes '*', so set an alphanumeric
			// sentinel and swap in the readable mask afterwards.
			const sentinel = "SOTAMASKEDPASSWORD"
			u.User = url.UserPassword(u.User.Username(), sentinel)
			masked = strings.Replace(u.String(), sentinel, maskedSecret, 1)
		}
	}
	if password == "" {
		return masked
	}
	for _, form := range []string{password, url.QueryEscape(password), url.PathEscape(password)} {
		masked = strings.ReplaceAll(masked, form, maskedSecret)
	}
	return masked
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
