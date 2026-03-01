package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Client is the HTTP client for the sota.io API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for deploy uploads
		},
	}
}

// Response envelope types matching the API format.

type dataResponse struct {
	Data json.RawMessage `json:"data"`
}

type listResponse struct {
	Data       json.RawMessage `json:"data"`
	Pagination *Pagination     `json:"pagination,omitempty"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Pagination holds cursor-based pagination info.
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// API response types

// Project represents a project returned from the API.
type Project struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Deployment represents a deployment returned from the API.
type Deployment struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Status      string    `json:"status"`
	URL         *string   `json:"url,omitempty"`
	ImageTag    string    `json:"image_tag,omitempty"`
	BuildMethod *string   `json:"build_method,omitempty"`
	Framework   *string   `json:"framework,omitempty"`
	Error       *string   `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvVar represents an environment variable returned from the API.
type EnvVar struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIError represents a structured error from the API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// do executes an HTTP request with auth header and returns the response.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

// parseResponse reads a data envelope response and unmarshal into target.
func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return &APIError{StatusCode: resp.StatusCode, Code: "unknown", Message: "request failed"}
		}
		return &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	var envelope dataResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return json.Unmarshal(envelope.Data, target)
}

var reNonAlnum = regexp.MustCompile(`[^a-z0-9-]+`)
var reMultiDash = regexp.MustCompile(`-{2,}`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = reNonAlnum.ReplaceAllString(s, "-")
	s = reMultiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 63 {
		s = s[:63]
		s = strings.TrimRight(s, "-")
	}
	return s
}

// CreateProject creates a new project.
func (c *Client) CreateProject(name string) (*Project, error) {
	body, _ := json.Marshal(map[string]string{"name": name, "slug": slugify(name)})
	req, err := http.NewRequest("POST", c.baseURL+"/v1/projects", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := c.parseResponse(resp, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects returns the user's projects.
func (c *Client) ListProjects(cursor string, limit int) ([]Project, *Pagination, error) {
	url := fmt.Sprintf("%s/v1/projects?limit=%d", c.baseURL, limit)
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, nil, &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	var envelope listResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, nil, err
	}

	var projects []Project
	if err := json.Unmarshal(envelope.Data, &projects); err != nil {
		return nil, nil, err
	}

	return projects, envelope.Pagination, nil
}

// Deploy uploads a code archive to deploy.
func (c *Client) Deploy(projectID string, archive io.Reader) (*Deployment, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("archive", "archive.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, archive); err != nil {
		return nil, fmt.Errorf("copying archive: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/projects/%s/deploy", c.baseURL, projectID), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := c.parseResponse(resp, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

// Rollback triggers a rollback for a project.
func (c *Client) Rollback(projectID string) (*Deployment, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/projects/%s/rollback", c.baseURL, projectID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := c.parseResponse(resp, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetDeployments returns deployments for a project.
func (c *Client) GetDeployments(projectID string) ([]Deployment, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/projects/%s/deployments", c.baseURL, projectID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	var envelope listResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	var deployments []Deployment
	if err := json.Unmarshal(envelope.Data, &deployments); err != nil {
		return nil, err
	}
	return deployments, nil
}

// GetLogs retrieves logs for a deployment.
func (c *Client) GetLogs(projectID, deployID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/projects/%s/deployments/%s/logs", c.baseURL, projectID, deployID), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try to parse as data envelope, fallback to raw
	var envelope dataResponse
	if err := json.Unmarshal(body, &envelope); err == nil {
		var logs string
		if err := json.Unmarshal(envelope.Data, &logs); err == nil {
			return logs, nil
		}
	}

	return string(body), nil
}

// StreamLogs connects to the SSE log stream for a deployment.
func (c *Client) StreamLogs(projectID, deployID string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/projects/%s/deployments/%s/logs/stream", c.baseURL, projectID, deployID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, &APIError{StatusCode: resp.StatusCode, Code: "stream_error", Message: "failed to connect to log stream"}
	}

	return resp.Body, nil
}

// SetEnvVar sets an environment variable.
func (c *Client) SetEnvVar(projectID, key, value string) error {
	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/projects/%s/envs", c.baseURL, projectID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	return nil
}

// ListEnvVars returns all environment variables for a project.
func (c *Client) ListEnvVars(projectID string) ([]EnvVar, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/projects/%s/envs", c.baseURL, projectID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	var envelope dataResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	var envVars []EnvVar
	if err := json.Unmarshal(envelope.Data, &envVars); err != nil {
		return nil, err
	}
	return envVars, nil
}

// DeleteEnvVar deletes an environment variable.
func (c *Client) DeleteEnvVar(projectID, key string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/projects/%s/envs/%s", c.baseURL, projectID, key), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return &APIError{StatusCode: resp.StatusCode, Code: errResp.Error.Code, Message: errResp.Error.Message}
	}

	return nil
}
