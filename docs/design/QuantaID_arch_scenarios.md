# QuantaID Architecture Scenarios

This document outlines key architectural scenarios for QuantaID.

## Scenario One: User Authentication

This scenario describes the process of a user authenticating with QuantaID.

...

## Scenario Two: API Authorization

This scenario describes the process of a client application accessing a protected API.

...

## Scenario Three: Adaptive Multi-Factor Authentication

This scenario describes the process of a user authenticating with adaptive MFA.

### Status: Implemented

The adaptive MFA flow has been implemented as described in [P1_auth_mfa_flow.md](./P1_auth_mfa_flow.md).

### Differences

*   The current implementation uses a simplified policy engine that makes decisions based on the risk level. A more advanced policy engine that supports complex rules is planned for a future release.
*   The current implementation includes placeholder logic for risk factor calculation and MFA enrollment. These will be replaced with real implementations in a future release.
