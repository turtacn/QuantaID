package application

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// Repository defines the persistence interface for application-related data.
// It outlines the CRUD and query operations for Application entities.
type Repository interface {
	// CreateApplication saves a new application to the database.
	CreateApplication(ctx context.Context, app *types.Application) error
	// GetApplicationByID retrieves an application by its unique ID.
	GetApplicationByID(ctx context.Context, id string) (*types.Application, error)
	// GetApplicationByName retrieves an application by its unique name.
	GetApplicationByName(ctx context.Context, name string) (*types.Application, error)
	// UpdateApplication modifies an existing application's data.
	UpdateApplication(ctx context.Context, app *types.Application) error
	// DeleteApplication removes an application from the database by its ID.
	DeleteApplication(ctx context.Context, id string) error
	// ListApplications retrieves a paginated list of all applications.
	ListApplications(ctx context.Context, pq types.PaginationQuery) ([]*types.Application, error)
}