package client

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IdentityService handles identity-related API calls.
type IdentityService struct {
	client *Client
}

// CreateUser creates a new user in QuantaID.
//
// Parameters:
//   - ctx: The context for the API request.
//   - username: The username for the new user.
//   - email: The email address for the new user.
//   - password: The password for the new user.
//
// Returns:
//   The created user, or an error if the creation fails.
func (s *IdentityService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	createReq := identity.CreateUserRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	req, err := s.client.newRequest(ctx, "POST", "/api/v1/users", createReq)
	if err != nil {
		return nil, err
	}

	var userResp types.User
	if err := s.client.do(req, &userResp); err != nil {
		return nil, err
	}

	return &userResp, nil
}

// GetUserByID retrieves a single user by their ID.
//
// Parameters:
//   - ctx: The context for the API request.
//   - userID: The ID of the user to retrieve.
//
// Returns:
//   The user, or an error if the user is not found.
func (s *IdentityService) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	path := fmt.Sprintf("/api/v1/users/%s", userID)
	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var userResp types.User
	if err := s.client.do(req, &userResp); err != nil {
		return nil, err
	}

	return &userResp, nil
}
