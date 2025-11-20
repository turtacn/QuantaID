# Policy Engine Design Guide

This document outlines the design and usage of the hybrid policy engine in QuantaID.

## Overview

The policy engine is a hybrid model that combines Role-Based Access Control (RBAC) and Attribute-Based Access Control (ABAC) to provide a flexible and powerful authorization system.

- **RBAC**: Provides a fast and simple way to manage permissions based on user roles.
- **ABAC**: Allows for more granular control by evaluating policies based on attributes of the user, resource, and environment.

## Architecture

The policy engine is composed of the following components:

- **HybridEvaluator**: The main entry point for policy decisions. It first checks for a definitive "allow" from the RBAC provider, and if so, it then checks for any "deny" rules from the ABAC provider.
- **RBACProvider**: Fetches user roles and permissions from the database and caches them for fast lookups.
- **ABACProvider**: Evaluates conditional policies based on attributes of the user, resource, and environment.

## Defining Roles and Permissions

Roles and permissions are managed through the admin API.

- **Roles**: A role is a collection of permissions. Users can be assigned one or more roles.
- **Permissions**: A permission is the ability to perform an action on a resource.

## Protecting Routes

Routes can be protected using the `RequirePermission` middleware. This middleware takes a required permission string in the format `resource:action` and checks if the user has the required permission before allowing access to the route.

Example:

```go
router.Handle("/api/users", middleware.RequirePermission(evaluator, "users:read")(myHandler)).Methods("GET")
```
