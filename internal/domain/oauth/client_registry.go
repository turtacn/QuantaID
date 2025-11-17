package oauth

import (
	"context"
	"fmt"

	"github.com/turtacn/QuantaID/pkg/types"
)

// ClientRegistry handles dynamic client registration.
type ClientRegistry struct {
	repo   ClientRepository
	config ClientRegistryConfig
}

// ClientRegistryConfig holds the configuration for client registration.
type ClientRegistryConfig struct {
	// DefaultGrantTypes is a list of grant types assigned to new clients if not specified.
	DefaultGrantTypes []string `yaml:"defaultGrantTypes"`
	// DefaultResponseTypes is a list of response types assigned to new clients if not specified.
	DefaultResponseTypes []string `yaml:"defaultResponseTypes"`
	// DefaultScopes is a list of scopes assigned to new clients if not specified.
	DefaultScopes []string `yaml:"defaultScopes"`
}

// ClientRepository defines the interface for storing and retrieving OAuth clients.
type ClientRepository interface {
	CreateClient(ctx context.Context, client *types.Client) error
	GetClient(ctx context.Context, clientID string) (*types.Client, error)
}

// NewClientRegistry creates a new ClientRegistry.
func NewClientRegistry(repo ClientRepository, config ClientRegistryConfig) *ClientRegistry {
	return &ClientRegistry{repo: repo, config: config}
}

// RegisterClient registers a new OAuth client.
func (r *ClientRegistry) RegisterClient(ctx context.Context, req *types.ClientMetadata) (*types.Client, error) {
	// TODO: Add validation for client metadata (e.g., redirect URIs, grant types).

	client := &types.Client{
		ID:           generateClientID(),
		Secret:       generateClientSecret(),
		RedirectURIs: req.RedirectURIs,
		GrantTypes:   req.GrantTypes,
		ResponseTypes: req.ResponseTypes,
		ClientName:   req.ClientName,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
	}

	if len(client.GrantTypes) == 0 {
		client.GrantTypes = r.config.DefaultGrantTypes
	}
	if len(client.ResponseTypes) == 0 {
		client.ResponseTypes = r.config.DefaultResponseTypes
	}

	if err := r.repo.CreateClient(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// TODO: Implement these functions with proper random string generation.
func generateClientID() string {
	return "temp_client_id"
}

func generateClientSecret() string {
	return "temp_client_secret"
}
