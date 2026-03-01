package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sota-io/sota-cli/internal/config"
)

// Credentials holds the authentication tokens.
type Credentials struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
	APIKey       string `json:"api_key,omitempty"`
}

// CredentialsPath returns the path to the credentials file.
func CredentialsPath() string {
	return filepath.Join(config.ConfigDir(), "credentials.json")
}

// LoadCredentials reads stored credentials from disk.
func LoadCredentials() (*Credentials, error) {
	data, err := os.ReadFile(CredentialsPath())
	if err != nil {
		return nil, fmt.Errorf("not logged in (run 'sota login')")
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentials writes credentials to disk.
func SaveCredentials(creds *Credentials) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(CredentialsPath(), data, 0o600)
}

// ClearCredentials removes stored credentials.
func ClearCredentials() error {
	err := os.Remove(CredentialsPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// SaveAPIKey persists an API key to the credentials file.
// This clears any existing JWT credentials (API key takes priority).
func SaveAPIKey(apiKey string) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	creds := &Credentials{APIKey: apiKey}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CredentialsPath(), data, 0o600)
}

// LoadAPIKey reads a stored API key from the credentials file.
func LoadAPIKey() (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}
	if creds.APIKey == "" {
		return "", fmt.Errorf("no API key stored")
	}
	return creds.APIKey, nil
}

// GetToken resolves authentication using the priority chain:
// 1. SOTA_API_KEY environment variable
// 2. Stored API key in credentials file
func GetToken() (token string, method string, err error) {
	// Priority 1: Environment variable
	if envKey := os.Getenv("SOTA_API_KEY"); envKey != "" {
		return envKey, "env:SOTA_API_KEY", nil
	}

	// Priority 2: Stored API key
	if apiKey, loadErr := LoadAPIKey(); loadErr == nil && apiKey != "" {
		return apiKey, "stored:api-key", nil
	}

	return "", "", fmt.Errorf("not authenticated. Run 'sota login' or 'sota auth set-key <key>'")
}

// OpenBrowser opens a URL in the default browser.
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	cmd.Start()
}
