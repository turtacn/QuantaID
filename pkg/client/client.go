package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/types"
	"io"
	"net/http"
	"time"
)

// Client provides a client for interacting with the QuantaID API.
// It handles making HTTP requests, authentication, and error handling.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string

	// Services for different API categories
	Auth     *AuthService
	Identity *IdentityService
}

// NewClient creates a new QuantaID API client.
//
// Parameters:
//   - baseURL: The base URL of the QuantaID API server.
//   - timeout: The timeout for HTTP requests.
//
// Returns:
//   A new QuantaID API client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
	c.Auth = &AuthService{client: c}
	c.Identity = &IdentityService{client: c}
	return c
}

// SetAuthToken sets the authentication token to be used for subsequent requests.
// The token will be included in the Authorization header of all API requests.
//
// Parameters:
//   - token: The authentication token.
func (c *Client) SetAuthToken(token string) {
	c.token = token
}

// newRequest is a helper function to create a new HTTP request.
// It sets up the URL, method, and body, and adds the necessary headers
// like Content-Type, Accept, and the Authorization token if available.
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

// do is a helper function that sends an HTTP request and handles the response.
// It executes the request, checks for non-successful status codes, and decodes
// the JSON response body into the provided interface `v`. It also handles
// standardized API error responses.
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError struct {
			Error *types.Error `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err == nil && apiError.Error != nil {
			return apiError.Error
		}
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}

