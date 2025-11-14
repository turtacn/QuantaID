# Identity Synchronization with LDAP

This document describes how to configure and use the LDAP identity synchronization feature in QuantaID.

## Overview

The LDAP sync service allows you to synchronize user identities from an LDAP directory into the QuantaID database. This enables you to manage your users in a central location and have the changes automatically reflected in QuantaID.

The service supports both full and incremental synchronization, conflict resolution strategies, and lifecycle management rules.

## Configuration

The LDAP sync service is configured in the `configs/sync/ldap_sync.yaml` file. Here is an example configuration:

```yaml
# Configuration for the LDAP Synchronization Service

# Sync intervals
full_sync_interval: "24h"
incremental_interval: "1h"

# Conflict resolution strategy when a user exists in both LDAP and the local database.
# Can be 'prefer_local' or 'prefer_remote'.
conflict_strategy: "prefer_remote"

# Lifecycle rules to automatically change a user's status based on LDAP attributes.
lifecycle_rules:
  - source_attr: "hr_status"
    match_value: "terminated"
    target_status: "disabled"
  - source_attr: "account_status"
    match_value: "inactive"
    target_status: "disabled"
```

### Configuration Options

- `full_sync_interval`: The interval at which a full synchronization is performed.
- `incremental_interval`: The interval at which an incremental synchronization is performed.
- `conflict_strategy`: The strategy to use when a user exists in both LDAP and the local database.
  - `prefer_local`: Keep the local user's attributes.
  - `prefer_remote`: Update the local user's attributes with the ones from LDAP.
- `lifecycle_rules`: A list of rules to automatically change a user's status based on LDAP attributes.
  - `source_attr`: The name of the LDAP attribute to check.
  - `match_value`: The value of the attribute to match.
  - `target_status`: The status to set for the user if the attribute matches.

## Usage

You can trigger a synchronization manually using the following REST API endpoints:

- `POST /admin/sync/ldap/full`: Triggers a full synchronization.
- `POST /admin/sync/ldap/incremental`: Triggers an incremental synchronization.
- `GET /admin/sync/ldap/status`: Retrieves the status of the last synchronization.

### Triggering a Full Sync

To trigger a full synchronization, send a POST request to `/admin/sync/ldap/full`.

```bash
curl -X POST http://localhost:8080/admin/sync/ldap/full
```

### Triggering an Incremental Sync

To trigger an incremental synchronization, send a POST request to `/admin/sync/ldap/incremental`.

You can also specify a `since` query parameter to synchronize changes since a specific time. The `since` parameter should be in RFC3339 format.

```bash
curl -X POST http://localhost:8080/admin/sync/ldap/incremental?since=2023-10-27T10:00:00Z
```

If you don't specify a `since` parameter, the service will synchronize changes since the last 24 hours.

### Checking the Sync Status

To check the status of the last synchronization, send a GET request to `/admin/sync/ldap/status`.

```bash
curl http://localhost:8080/admin/sync/ldap/status
```

The response will be a JSON object containing the statistics of the last synchronization.

## Auditing

Every synchronization event is audited. You can view the audit logs to see the details of each synchronization, including the number of users created, updated, and disabled.
