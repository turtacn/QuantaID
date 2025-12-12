# RADIUS Server Configuration

QuantaID includes a built-in RADIUS server to support network device authentication (VPN, Wi-Fi, Switches).

## Configuration

Enable RADIUS in `server.yaml` (or environment variables):

```yaml
radius:
  enabled: true
  auth_port: 1812
  acct_port: 1813
  read_timeout: 5s
  write_timeout: 5s
  worker_count: 10
  proxy:
    enabled: false
```

## Adding NAS Clients

NAS Clients (Network Access Servers) must be registered in the database before they can authenticate users.

Currently, clients can be added via database insertion into `radius_clients` table.

```sql
INSERT INTO radius_clients (id, name, ip_address, secret, tenant_id, enabled)
VALUES ('uuid-1', 'VPN-Gateway', '192.168.1.10', 'sharedsecret', 'tenant-1', true);
```

## Supported Authentication Methods

1. **PAP**: Standard password authentication.
2. **CHAP**: Challenge-Handshake Authentication Protocol. Requires cleartext password availability (or reversible encryption).
3. **MS-CHAPv2**: Microsoft CHAP version 2. Commonly used for VPNs (L2TP/IPsec) and Wi-Fi (PEAP).

## Accounting

The server listens on port 1813 for accounting packets and logs them to the `radius_accounting` table.

## Integration Examples

### Cisco IOS
```
aaa new-model
radius-server host 10.0.0.5 auth-port 1812 acct-port 1813 key sharedsecret
aaa authentication login default group radius local
```

### Mikrotik
```
/radius add address=10.0.0.5 secret=sharedsecret service=ppp,wireless
```
