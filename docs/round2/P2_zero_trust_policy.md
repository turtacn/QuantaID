# P2: Zero-Trust Authorization & Policy Engine

This document outlines the architecture and implementation of the centralized authorization engine introduced in Phase 2. The new engine provides a unified, policy-driven approach to access control, moving away from scattered, hardcoded authorization logic.

## Policy Model

The policy engine is built around a set of core data models that represent the context of an authorization request.

*   **Subject**: Represents the entity performing an action. It includes the user's ID, their group memberships, and any other relevant attributes.
*   **Resource**: Represents the object being accessed. It includes the resource's type, its unique ID, and any other relevant attributes.
*   **Action**: Represents the operation being performed (e.g., `dashboard.read`, `users.create`).
*   **Environment**: Represents the context of the request, including the user's IP address, the time of the request, and the device's trust level.
*   **EvaluationContext**: A container that bundles all of the above models into a single object that can be passed to the policy engine for evaluation.

## Policy Evaluator

The `Evaluator` is the core of the policy engine. It's an interface that defines a single method, `Evaluate`, which takes an `EvaluationContext` and returns a `Decision` (either `Allow` or `Deny`).

The `DefaultEvaluator` is the default implementation of the `Evaluator` interface. It's a rule-based engine that loads a set of rules from a YAML file and evaluates them in order. The first rule that matches the `EvaluationContext` determines the outcome. If no rules match, the default decision is to deny access.

## Configuration

The `DefaultEvaluator` is configured using a YAML file that defines a list of rules. Each rule has the following properties:

*   `name`: A descriptive name for the rule.
*   `effect`: The effect of the rule, either `allow` or `deny`.
*   `actions`: A list of actions that the rule applies to.
*   `subjects`: A list of subjects that the rule applies to. Subjects can be specified as `user:<user-id>` or `group:<group-id>`.
*   `ip_whitelist`: A list of IP addresses or CIDR ranges that are allowed to perform the specified actions.
*   `time_ranges`: A list of time ranges during which the specified actions are allowed.

### Example Configuration

```yaml
rules:
  - name: "admin-dashboard-access"
    effect: "allow"
    actions: ["dashboard.read"]
    subjects:
      - "group:admins"
    ip_whitelist: ["10.0.0.0/8", "192.168.0.0/16"]
    time_ranges:
      - start: "08:00"
        end: "20:00"
  - name: "default-deny"
    effect: "deny"
    actions: ["*"]
    subjects: ["*"]
```

## Zero-Trust Alignment

This new policy engine aligns with the principles of zero-trust by:

*   **Explicitly verifying** every access request against a set of policies.
*   **Enforcing least privilege** by default-denying access and only allowing access to specific resources for specific subjects.
*   **Using rich context** (e.g., user, group, IP address, time of day) to make authorization decisions.

The current implementation provides a solid foundation for a zero-trust architecture. Future phases will build on this foundation by adding support for more complex policies, device trust, and integration with external policy engines like OPA.
