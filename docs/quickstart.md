# QuantaID Quickstart

This guide will help you get QuantaID up and running quickly for development or testing.

## Prerequisites

*   Go 1.18+
*   Docker & Docker Compose
*   PostgreSQL 13+ (if not using Docker)
*   Redis 6+ (if not using Docker)

## Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/turtacn/QuantaID.git
    cd QuantaID
    ```

2.  **Configuration:**

    Copy the example configuration:

    ```bash
    cp configs/server.yaml.example configs/server.yaml
    ```

    Edit `configs/server.yaml` to match your environment (DB credentials, etc.).

3.  **Start Services (Docker):**

    Use Docker Compose to start PostgreSQL and Redis:

    ```bash
    docker-compose -f deployments/docker-compose/infrastructure.yaml up -d
    ```

4.  **Run the Server:**

    ```bash
    go run ./cmd/qid-server
    ```

    The server will start on port 8080 (default).

## API Key Management (New!)

QuantaID now supports API Keys for Machine-to-Machine (M2M) authentication and rate limiting.

### 1. Generate an API Key

(Currently via programmatic service injection or future CLI/UI. Below is a conceptual usage if CLI is updated)

### 2. Use the API Key

Include the `X-API-Key` header in your requests to protected endpoints (e.g., `/api/v1/m2m/*`).

```bash
curl -H "X-API-Key: qid_live_<key_id><secret>" http://localhost:8080/api/v1/m2m/ping
```

### 3. Rate Limiting

Rate limits are enforced based on the Application ID associated with the API Key, or falling back to IP address. Default limits are configured in `server.yaml`.

```yaml
security:
  rate_limit:
    enabled: true
    default_limit: 1000
    default_window: 60 # seconds
```

## Running Tests

To run unit tests:

```bash
go test ./...
```

To run integration tests (requires Docker):

```bash
go test -tags=integration ./...
```
