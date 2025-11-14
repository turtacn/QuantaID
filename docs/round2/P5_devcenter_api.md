# P5: DevCenter API

This document provides a summary of the DevCenter API endpoints, along with examples using `curl`.

## Authentication

All DevCenter API endpoints require administrator privileges. You must include a valid JWT in the `Authorization` header of your requests.

```bash
export TOKEN="your-admin-jwt"
```

## API Endpoints

### Applications

#### List Applications

* **Endpoint:** `GET /api/devcenter/apps`
* **Description:** Retrieves a list of all applications.
* **Example:**

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/devcenter/apps
```

#### Create Application

* **Endpoint:** `POST /api/devcenter/apps`
* **Description:** Creates a new application.
* **Example:**

```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My New App",
    "protocol": "oidc",
    "redirect_uri": "http://localhost:3000/callback"
  }' \
  http://localhost:8080/api/v1/devcenter/apps
```

### Connectors

#### List Connectors

* **Endpoint:** `GET /api/devcenter/connectors`
* **Description:** Retrieves a list of all connectors.
* **Example:**

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/devcenter/connectors
```

#### Enable Connector

* **Endpoint:** `POST /api/devcenter/connectors/{id}/enable`
* **Description:** Enables a connector.
* **Example:**

```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/devcenter/connectors/ldap-1/enable
```

### Diagnostics

#### Get Diagnostics

* **Endpoint:** `GET /api/devcenter/diagnostics`
* **Description:** Retrieves diagnostic information.
* **Example:**

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/devcenter/diagnostics
```
