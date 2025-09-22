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

// Client is the main client for interacting with the QuantaID API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new QuantaID API client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetAuthToken sets the authentication token to be used for subsequent requests.
func (c *Client) SetAuthToken(token string) {
	c.token = token
}

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

//Personal.AI order the ending
