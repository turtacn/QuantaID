package storage

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// requireDocker checks if the Docker environment is available and responsive.
// If Docker is not available or the daemon cannot be reached, it skips the test.
func requireDocker(t *testing.T) {
	t.Helper()

	// Using the internal core.NewClient function from testcontainers-go to ping the Docker daemon.
	// This is a reliable way to check for a working Docker environment.
	if _, err := core.NewClient(context.Background()); err != nil {
		t.Skipf("skipping test: Docker not available or permission denied: %v", err)
	}
}
