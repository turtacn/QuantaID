package client

import (
	"context"
	"github.com/turtacn/QuantaID/internal/services/auth"
)

// Login authenticates with the QuantaID server and stores the token for future requests.
func (c *Client) Login(ctx context.Context, username, password string) (*auth.LoginResponse, error) {
	loginReq := auth.LoginRequest{
		Username: username,
		Password: password,
	}

	req, err := c.newRequest(ctx, "POST", "/api/v1/auth/login", loginReq)
	if err != nil {
		return nil, err
	}

	var loginResp auth.LoginResponse
	if err := c.do(req, &loginResp); err != nil {
		return nil, err
	}

	if loginResp.AccessToken != "" {
		c.SetAuthToken(loginResp.AccessToken)
	}

	return &loginResp, nil
}

// Logout clears the local authentication token.
func (c *Client) Logout() {
	c.SetAuthToken("")
}

//Personal.AI order the ending
