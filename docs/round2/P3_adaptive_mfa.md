# P3: Adaptive MFA

This document describes the adaptive MFA feature, which introduces a risk-based approach to multi-factor authentication.

## Flow

The adaptive MFA feature introduces a risk assessment step into the authentication workflow. The workflow is as follows:

1.  **Password Verification**: The user's password is verified.
2.  **Risk Assessment**: The risk of the login attempt is assessed based on a set of rules.
3.  **MFA Decision**: Based on the risk score, a decision is made whether to:
    *   Allow the login without MFA.
    *   Require MFA.
    *   Block the login.

## Risk Engine

The risk engine is responsible for assessing the risk of a login attempt. The `SimpleRiskEngine` is a rule-based engine that calculates a risk score based on the following factors:

*   **New Device**: A new device is detected.
*   **Geo-Velocity**: The user is logging in from a location that is geographically distant from their last login.
*   **Unusual Time**: The user is logging in at an unusual time.

The risk scores and thresholds are configurable in `configs/auth/risk_rules.yaml`.

## Testing

To simulate different risk scenarios in a testing environment, you can provide the following parameters in the initial state of the `standard_auth_flow` workflow:

*   `last_login_ip`: The IP address of the last login.
*   `last_login_country`: The country of the last login.
*   `now`: The current time.

By providing different values for these parameters, you can trigger different risk factors and test the adaptive MFA flow.
