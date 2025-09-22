package client

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/pkg/types"
)

// CreateUser creates a new user in QuantaID.
func (c *Client) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	createReq := identity.CreateUserRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	req, err := c.newRequest(ctx, "POST", "/api/v1/users", createReq)
	if err != nil {
		return nil, err
	}

	var userResp types.User
	if err := c.do(req, &userResp); err != nil {
		return nil, err
	}

	return &userResp, nil
}

// GetUserByID retrieves a single user by their ID.
func (c *Client) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	path := fmt.Sprintf("/api/v1/users/%s", userID)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var userResp types.User
	if err := c.do(req, &userResp); err != nil {
		return nil, err
	}

	return &userResp, nil
}

//Personal.AI order the ending
