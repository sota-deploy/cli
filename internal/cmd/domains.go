package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sota-io/sota-cli/internal/api"
	"github.com/sota-io/sota-cli/internal/auth"
	"github.com/sota-io/sota-cli/internal/config"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage custom domains",
	Long:  "Add, list, get, and remove custom domains for your app.",
}

var domainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all custom domains",
	RunE:  runDomainsList,
}

var domainsAddCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a custom domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainsAdd,
}

var domainsGetCmd = &cobra.Command{
	Use:   "get <domain-id>",
	Short: "Get domain details",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainsGet,
}

var domainsRemoveCmd = &cobra.Command{
	Use:   "remove <domain-id>",
	Short: "Remove a custom domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainsRemove,
}

func init() {
	domainsCmd.AddCommand(domainsListCmd)
	domainsCmd.AddCommand(domainsAddCmd)
	domainsCmd.AddCommand(domainsGetCmd)
	domainsCmd.AddCommand(domainsRemoveCmd)
}

func runDomainsList(cmd *cobra.Command, args []string) error {
	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	domains, err := client.ListDomains(projCfg.ProjectID)
	if err != nil {
		return fmt.Errorf("listing domains: %w", err)
	}

	if len(domains) == 0 {
		fmt.Println("No custom domains configured.")
		return nil
	}

	bold := color.New(color.Bold)
	bold.Printf("%-40s %-12s %s\n", "DOMAIN", "STATUS", "ID")

	for _, d := range domains {
		statusColor := statusColorFor(d.Status)
		fmt.Printf("%-40s %s %s\n", d.Domain, statusColor.Sprint(d.Status), d.ID)
	}
	return nil
}

func runDomainsAdd(cmd *cobra.Command, args []string) error {
	domain := args[0]

	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	result, err := client.AddDomain(projCfg.ProjectID, domain)
	if err != nil {
		return fmt.Errorf("adding domain: %w", err)
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("Domain added: %s\n", result.Domain.Domain)

	if result.DNSInstructions != nil {
		dns := result.DNSInstructions
		fmt.Println()
		cyan := color.New(color.FgCyan, color.Bold)
		cyan.Println("DNS Setup Required:")
		fmt.Printf("  Record Type: %s\n", dns.Type)
		fmt.Printf("  Name:        %s\n", dns.Name)
		fmt.Printf("  Value:       %s\n", dns.Value)
		fmt.Println()
		dim := color.New(color.Faint)
		dim.Println("Add this DNS record at your domain registrar.")
		dim.Println("Once configured, verification and SSL provisioning happen automatically.")
	}

	return nil
}

func runDomainsGet(cmd *cobra.Command, args []string) error {
	domainID := args[0]

	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	result, err := client.GetDomain(projCfg.ProjectID, domainID)
	if err != nil {
		return fmt.Errorf("getting domain: %w", err)
	}

	d := result.Domain
	bold := color.New(color.Bold)
	statusColor := statusColorFor(d.Status)

	bold.Print("Domain:  ")
	fmt.Println(d.Domain)
	bold.Print("ID:      ")
	fmt.Println(d.ID)
	bold.Print("Status:  ")
	statusColor.Println(d.Status)
	bold.Print("Type:    ")
	fmt.Println(d.DNSType)
	bold.Print("Created: ")
	fmt.Println(d.CreatedAt.Format("2006-01-02 15:04:05"))

	if d.VerifiedAt != nil {
		bold.Print("Verified: ")
		fmt.Println(*d.VerifiedAt)
	}

	if d.ErrorMessage != nil {
		errColor := color.New(color.FgRed)
		bold.Print("Error:   ")
		errColor.Println(*d.ErrorMessage)
	}

	if result.DNSInstructions != nil {
		dns := result.DNSInstructions
		fmt.Println()
		cyan := color.New(color.FgCyan, color.Bold)
		cyan.Println("DNS Instructions:")
		fmt.Printf("  Record Type: %s\n", dns.Type)
		fmt.Printf("  Name:        %s\n", dns.Name)
		fmt.Printf("  Value:       %s\n", dns.Value)
	}

	return nil
}

func runDomainsRemove(cmd *cobra.Command, args []string) error {
	domainID := args[0]

	token, _, err := auth.GetToken()
	if err != nil {
		return err
	}

	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(apiURL, token)
	if err := client.RemoveDomain(projCfg.ProjectID, domainID); err != nil {
		return fmt.Errorf("removing domain: %w", err)
	}

	green := color.New(color.FgGreen)
	green.Println("Domain removed successfully.")
	return nil
}

// statusColorFor returns a color based on domain status.
func statusColorFor(status string) *color.Color {
	switch status {
	case "active":
		return color.New(color.FgGreen)
	case "verified", "pending":
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgRed)
	}
}
