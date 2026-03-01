package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// deviceCodeResponse is the API response from POST /v1/auth/device-code.
type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// deviceTokenResponse is the API response from POST /v1/auth/device-token on success.
type deviceTokenResponse struct {
	APIKey string `json:"api_key"`
}

// DeviceLogin performs the device code flow for CLI login.
// It requests a device code, displays it to the user, opens the browser,
// and polls until the code is approved or expires.
func DeviceLogin(apiBaseURL string) (string, error) {
	// 1. Request device code
	resp, err := http.Post(apiBaseURL+"/v1/auth/device-code", "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		return "", fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("device code request failed (HTTP %d)", resp.StatusCode)
	}

	var dcResp deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&dcResp); err != nil {
		return "", fmt.Errorf("parsing device code response: %w", err)
	}

	// 2. Display instructions
	fmt.Println()
	fmt.Println("To authenticate, open this URL in your browser:")
	fmt.Println()
	fmt.Printf("  %s\n", dcResp.VerificationURL)
	fmt.Println()
	fmt.Printf("And enter this code: %s\n", dcResp.UserCode)
	fmt.Println()
	fmt.Print("Waiting for authorization...")

	// 3. Open browser
	OpenBrowser(dcResp.VerificationURL)

	// 4. Poll for approval
	interval := dcResp.Interval
	if interval <= 0 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-timeout:
			fmt.Println()
			return "", fmt.Errorf("login timed out (10 minutes). Run 'sota login' to try again")
		case <-ticker.C:
			apiKey, done, pollErr := pollDeviceToken(apiBaseURL, dcResp.DeviceCode)
			if pollErr != nil {
				fmt.Println()
				return "", pollErr
			}
			if done {
				return apiKey, nil
			}
			// Still pending -- print dot for liveness
			fmt.Print(".")
		}
	}
}

// pollDeviceToken makes a single poll request to the device-token endpoint.
// Returns (apiKey, done, error).
func pollDeviceToken(apiBaseURL, deviceCode string) (string, bool, error) {
	body, _ := json.Marshal(map[string]string{"device_code": deviceCode})
	resp, err := http.Post(apiBaseURL+"/v1/auth/device-token", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", false, fmt.Errorf("polling device token: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var tokenResp deviceTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return "", false, fmt.Errorf("parsing token response: %w", err)
		}
		return tokenResp.APIKey, true, nil
	case 428:
		// Authorization pending -- keep polling
		return "", false, nil
	case 410:
		return "", false, fmt.Errorf("code expired. Run 'sota login' again")
	case 404:
		return "", false, fmt.Errorf("device code not found. Run 'sota login' again")
	default:
		return "", false, fmt.Errorf("unexpected response (HTTP %d)", resp.StatusCode)
	}
}
