package auth

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IApplicationService defines the interface for the auth application service.
type IApplicationService interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, *types.Error)
	Logout(ctx context.Context, req LogoutRequest) *types.Error
}

//Personal.AI order the ending
