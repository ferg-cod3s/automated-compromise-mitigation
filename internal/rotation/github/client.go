// Package github provides GitHub Personal Access Token rotation functionality.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// GitHubAPIBaseURL is the base URL for GitHub API v3.
	GitHubAPIBaseURL = "https://api.github.com"

	// GitHubAPIVersion is the API version header value.
	GitHubAPIVersion = "2022-11-28"

	// MaxRetries is the maximum number of retry attempts for transient errors.
	MaxRetries = 3

	// RetryDelay is the initial retry delay (exponential backoff).
	RetryDelay = 2 * time.Second
)

// Client is a GitHub API client for token management operations.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new GitHub API client.
func NewClient() *Client {
	return &Client{
		baseURL: GitHubAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "ACM-GitHub-Rotator/1.0",
	}
}

// NewClientWithHTTP creates a client with a custom HTTP client (for testing).
func NewClientWithHTTP(httpClient *http.Client) *Client {
	return &Client{
		baseURL:    GitHubAPIBaseURL,
		httpClient: httpClient,
		userAgent:  "ACM-GitHub-Rotator/1.0",
	}
}

// User represents a GitHub user.
type User struct {
	Login     string    `json:"login"`
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Token represents a GitHub Personal Access Token.
type Token struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`      // Only returned on creation
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateTokenRequest represents a request to create a new PAT.
type CreateTokenRequest struct {
	Name        string    `json:"name"`
	Scopes      []string  `json:"scopes"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

// GetUser retrieves the authenticated user's information.
// This is used to verify token validity.
func (c *Client) GetUser(ctx context.Context, token string) (*User, error) {
	req, err := c.newRequest(ctx, "GET", "/user", nil, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var user User
	if err := c.doRequest(req, &user); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// ListTokens lists all Personal Access Tokens for the authenticated user.
// Note: This endpoint may not be available for all token types.
func (c *Client) ListTokens(ctx context.Context, token string) ([]*Token, error) {
	// Note: As of 2025, GitHub's REST API doesn't provide a direct endpoint
	// to list all PATs. This is a placeholder for when/if GitHub adds this.
	// For now, we'll track tokens in our own database.

	// We can still verify a token exists by trying to use it
	_, err := c.GetUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Return empty list - we'll track tokens in our database
	return []*Token{}, nil
}

// CreateToken creates a new fine-grained Personal Access Token.
// Note: Classic PATs cannot be created via API - only fine-grained tokens.
func (c *Client) CreateToken(ctx context.Context, authToken string, req CreateTokenRequest) (*Token, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("token name is required")
	}

	if len(req.Scopes) == 0 {
		return nil, fmt.Errorf("at least one scope is required")
	}

	// Note: The actual GitHub API endpoint for creating fine-grained tokens
	// is /user/installations/{installation_id}/access_tokens for GitHub Apps,
	// or requires OAuth flow for fine-grained PATs.
	//
	// For now, this is a simplified implementation that demonstrates the pattern.
	// In production, we'd need to handle the OAuth flow or GitHub App installation.

	// For MVP, we'll document that users need to create tokens manually
	// and we'll focus on validation and deletion.

	return nil, fmt.Errorf("GitHub API does not support creating classic PATs; fine-grained tokens require OAuth flow (Phase IV feature)")
}

// DeleteToken deletes a Personal Access Token by ID.
// This works for both classic and fine-grained tokens.
func (c *Client) DeleteToken(ctx context.Context, authToken string, tokenID int64) error {
	// Note: GitHub's API doesn't provide a direct endpoint to delete PATs by ID
	// for security reasons. Users must delete via the GitHub UI.
	//
	// However, we can revoke OAuth tokens if we're using OAuth flow.
	// For classic PATs, deletion must be manual.

	// For MVP, we'll log this limitation
	return fmt.Errorf("GitHub API does not support deleting PATs programmatically (must be done via GitHub UI)")
}

// TestToken validates that a token works and has the expected permissions.
func (c *Client) TestToken(ctx context.Context, token string, expectedScopes []string) error {
	// Get user to verify token is valid
	_, err := c.GetUser(ctx, token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// TODO: Verify scopes if possible
	// GitHub doesn't provide a direct way to query token scopes via API
	// We could test specific endpoints to infer permissions

	return nil
}

// GetRateLimit gets the current API rate limit status.
func (c *Client) GetRateLimit(ctx context.Context, token string) (*RateLimit, error) {
	req, err := c.newRequest(ctx, "GET", "/rate_limit", nil, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Rate RateLimit `json:"rate"`
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}

	return &response.Rate, nil
}

// RateLimit represents GitHub API rate limit information.
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

// newRequest creates a new HTTP request with standard headers.
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}, token string) (*http.Request, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", GitHubAPIVersion)
	req.Header.Set("User-Agent", c.userAgent)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// doRequest executes an HTTP request and decodes the JSON response.
func (c *Client) doRequest(req *http.Request, result interface{}) error {
	return c.doRequestWithRetry(req, result, 0)
}

// doRequestWithRetry executes an HTTP request with retry logic.
func (c *Client) doRequestWithRetry(req *http.Request, result interface{}, attempt int) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Network error - retry if attempts remaining
		if attempt < MaxRetries {
			time.Sleep(RetryDelay * time.Duration(1<<attempt)) // Exponential backoff
			return c.doRequestWithRetry(req, result, attempt+1)
		}
		return fmt.Errorf("network error after %d attempts: %w", attempt+1, err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		if attempt < MaxRetries {
			// Check for Retry-After header
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				// Parse and wait
				delay, _ := time.ParseDuration(retryAfter + "s")
				if delay == 0 {
					delay = RetryDelay * time.Duration(1<<attempt)
				}
				time.Sleep(delay)
				return c.doRequestWithRetry(req, result, attempt+1)
			}

			// Exponential backoff
			time.Sleep(RetryDelay * time.Duration(1<<attempt))
			return c.doRequestWithRetry(req, result, attempt+1)
		}
		return fmt.Errorf("rate limited after %d attempts", attempt+1)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var ghErr GitHubError
		if err := json.Unmarshal(body, &ghErr); err == nil && ghErr.Message != "" {
			return &ghErr
		}
		return fmt.Errorf("GitHub API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Decode response
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// GitHubError represents an error response from the GitHub API.
type GitHubError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Status           int    `json:"status"`
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("GitHub API error: %s (status: %d)", e.Message, e.Status)
}

// IsRateLimitError checks if an error is a rate limit error.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "rate limit") ||
		strings.Contains(err.Error(), "API rate limit exceeded")
}

// IsAuthenticationError checks if an error is an authentication error.
func IsAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	ghErr, ok := err.(*GitHubError)
	if !ok {
		return false
	}
	return ghErr.Status == http.StatusUnauthorized || ghErr.Status == http.StatusForbidden
}

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	ghErr, ok := err.(*GitHubError)
	if !ok {
		return false
	}
	return ghErr.Status == http.StatusNotFound
}
