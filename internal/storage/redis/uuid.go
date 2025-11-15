package redis

import "github.com/google/uuid"

// GoogleUUIDGenerator is a UUID generator that uses the google/uuid package.
type GoogleUUIDGenerator struct{}

// New creates a new UUID string.
func (g *GoogleUUIDGenerator) New() string {
	return uuid.New().String()
}
