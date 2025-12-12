# Network Device Integration Guide

## Overview

This guide details how to integrate various network appliances with QuantaID via RADIUS.

## Common Parameters

* **Authentication Port**: 1812 (UDP)
* **Accounting Port**: 1813 (UDP)
* **Shared Secret**: configured per client IP in QuantaID

## VPN Integration

### OpenVPN (via PAM radius)
Install `pam_radius_auth` and configure `/etc/pam_radius_auth.conf`:
```
127.0.0.1       secret      1
10.0.0.5        sharedsecret  3
```

### Fortigate SSL VPN
1. Go to **User & Device > RADIUS Servers**.
2. Create New.
   * **Name**: QuantaID
   * **IP/Name**: <QuantaID IP>
   * **Secret**: <Shared Secret>
3. Create a User Group mapping to the RADIUS server.
4. Assign Group to SSL-VPN settings.

## Wi-Fi Enterprise (WPA2/WPA3-Enterprise)

QuantaID supports PEAP-MSCHAPv2 which is standard for Enterprise Wi-Fi.

### Ubiquiti UniFi
1. Settings > Profiles > RADIUS.
2. Create New RADIUS Profile.
   * **IP Address**: <QuantaID IP>
   * **Port**: 1812
   * **Password**: <Shared Secret>
3. Apply to Wireless Network (WPA Enterprise).

## Troubleshooting

* Check `radius_accounting` table for connection attempts.
* Ensure firewall permits UDP 1812/1813.
* Verify Shared Secret matches exactly.
* Use `radtest` tool for verification:
  ```bash
  radtest user password localhost 1812 secret
  ```
