# Storage Integration Tests

This directory contains integration tests for the PostgreSQL and Redis repositories. These tests use `testcontainers-go` to spin up real database instances in Docker, ensuring that the repository logic is tested against a live environment.

## Prerequisites

- **Docker:** You must have Docker installed and running.
- **Docker Permissions:** Your user must have permission to interact with the Docker daemon. If you encounter `permission denied` errors, you may need to add your user to the `docker` group or run your IDE/terminal with elevated privileges.

## Running the Tests

To run all integration tests in this directory, use the following command from the root of the repository:

```bash
go test ./tests/integration/storage/...
```

The tests will automatically be skipped if a working Docker environment is not detected.
