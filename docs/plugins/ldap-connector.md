# LDAP Connector Plugin

This document provides a guide for configuring and using the LDAP Connector plugin for QuantaID.

## Overview

The LDAP Connector plugin allows QuantaID to connect to an external LDAP or Active Directory server to authenticate users and synchronize user information.

## Configuration

The LDAP Connector is configured in the `configs/plugins/ldap.yaml` file. The following is an example configuration with explanations of each option.

```yaml
ldap:
  host: "ldap.example.com"
  port: 389
  use_tls: true
  bind_dn: "cn=admin,dc=example,dc=com"
  bind_password: "secret"
  base_dn: "ou=users,dc=example,dc=com"
  user_filter: "(objectClass=inetOrgPerson)"
  attribute_mapping:
    username: "uid"
    email: "mail"
    display_name: "displayName"
    phone: "telephoneNumber"
  sync:
    enabled: true
    interval: "1h"
    full_sync_cron: "0 2 * * *"  # Daily at 2 AM
```

### Configuration Options

- `host`: The hostname or IP address of the LDAP server.
- `port`: The port number of the LDAP server.
- `use_tls`: Whether to use TLS to connect to the LDAP server.
- `bind_dn`: The Distinguished Name (DN) of the user to bind to the LDAP server with. This user should have permission to search for users.
- `bind_password`: The password for the `bind_dn` user.
- `base_dn`: The base DN to search for users in.
- `user_filter`: The LDAP filter to use to find users.
- `attribute_mapping`: A map of QuantaID user attributes to LDAP attributes.
  - `username`: The LDAP attribute to map to the QuantaID username.
  - `email`: The LDAP attribute to map to the QuantaID email.
  - `display_name`: The LDAP attribute to map to the QuantaID display name.
  - `phone`: The LDAP attribute to map to the QuantaID phone number.
- `sync`: Configuration for user synchronization.
  - `enabled`: Whether to enable user synchronization.
  - `interval`: The interval at which to perform incremental user synchronization.
  - `full_sync_cron`: A cron expression for when to perform a full user synchronization.

## Troubleshooting

### Connection Errors

If you are seeing connection errors, please check the following:

- The `host` and `port` are correct.
- The LDAP server is running and accessible from the QuantaID server.
- The `use_tls` setting is correct for your LDAP server.
- The `bind_dn` and `bind_password` are correct.

### Authentication Errors

If you are seeing authentication errors, please check the following:

- The `base_dn` is correct.
- The `user_filter` is correct.
- The `attribute_mapping` is correct.
- The user you are trying to authenticate as exists in the LDAP server and matches the `user_filter`.
- The password you are using is correct.
