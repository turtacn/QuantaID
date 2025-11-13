# P4: Audit & Observability

This document details the unified audit and observability pipeline implemented in Phase 4.

## Unified Audit Pipeline

The audit pipeline provides a standardized way to record security-sensitive events throughout the system. It is designed to be extensible, allowing events to be fanned out to multiple destinations (sinks) for monitoring, alerting, and long-term storage.

### Audit Event Schema

All audit events follow a common JSON schema, defined in `internal/audit/event.go`.

| Field       | Type           | Description                                                                                             | Example                               |
|-------------|----------------|---------------------------------------------------------------------------------------------------------|---------------------------------------|
| `id`        | string (UUID)  | A unique identifier for the event.                                                                      | `a1b2c3d4-e5f6-7890-1234-567890abcdef` |
| `ts`        | string (RFC3339) | The UTC timestamp of when the event occurred.                                                           | `2023-10-27T10:00:00Z`                  |
| `category`  | string         | Broad classification of the event. One of: `auth`, `policy`, `admin`, `mfa`, `risk`.                        | `auth`                                |
| `action`    | string         | The specific action that occurred.                                                                      | `login_success`                       |
| `user_id`   | string         | The ID of the user associated with the event. Omitted for system events.                                | `usr_12345`                           |
| `ip`        | string         | The source IP address of the request that triggered the event.                                          | `192.168.1.100`                       |
| `resource`  | string         | The resource that was targeted by the action.                                                           | `user:usr_67890`                      |
| `result`    | string         | The outcome of the action. One of: `success`, `fail`, `deny`.                                           | `success`                             |
| `trace_id`  | string         | A correlation ID for tracing the request through the system.                                            | `trace_xyz`                           |
| `details`   | object         | A flexible map for storing additional, action-specific context.                                         | `{"reason": "invalid_credentials"}`   |

### Pipeline Configuration

The pipeline is configured in `configs/audit/pipeline.jules.yaml`. In the Jules environment, it is configured to use two sinks:

1.  **`stdout`**: Prints JSON-formatted events to the console.
2.  **`file`**: Appends JSON-formatted events to a log file at `./logs/audit_jules.log`.

### Example: Searching Audit Logs

Security operators can use standard command-line tools like `grep` and `jq` to perform simple queries on the audit log file.

**Find all failed login attempts:**

```bash
cat ./logs/audit_jules.log | jq 'select(.action == "login_failed")'
```

**Find all high-risk events for a specific user:**

```bash
cat ./logs/audit_jules.log | jq 'select(.category == "risk" and .user_id == "usr_12345")'
```

## Observability

### Metrics

The system exposes key performance and security metrics in a Prometheus-compatible format at the `/metrics` endpoint.

**Key HTTP Metrics:**

*   `quantaid_http_requests_total`: A counter for the total number of HTTP requests, labeled by `method`, `path`, and `status`.
*   `quantaid_http_request_duration_seconds`: A histogram of request latency, labeled by `method`, `path`, and `status`.

These metrics provide a high-level overview of the system's health and can be used to create dashboards and alerts.
