package client

import (
	"context"
	"github.com/turtacn/QuantaID/internal/services/auth"
)

// AuthService handles authentication-related API calls.
type AuthService struct {
	client *Client
}

// Login authenticates with the QuantaID server and stores the token for future requests.
//
// Parameters:
//   - ctx: The context for the API request.
//   - username: The user's username.
//   - password: The user's password.
//
// Returns:
//   A LoginResponse containing the access token, or an error if authentication fails.
func (s *AuthService) Login(ctx context.Context, username, password string) (*auth.LoginResponse, error) {
	loginReq := auth.LoginRequest{
		Username: username,
		Password: password,
	}

	req, err := s.client.newRequest(ctx, "POST", "/api/v1/auth/login", loginReq)
	if err != nil {
		return nil, err
	}

	var loginResp auth.LoginResponse
	if err := s.client.do(req, &loginResp); err != nil {
		return nil, err
	}

	if loginResp.AccessToken != "" {
		s.client.SetAuthToken(loginResp.AccessToken)
	}

	return &loginResp, nil
}

// Logout clears the local authentication token.
func (s *AuthService) Logout() {
	s.client.SetAuthToken("")
}
