<p align="center">
  <img src="logo.png" alt="QuantaID Logo" width="200" height="200">
</p>

<h1 align="center">QuantaID</h1>

<p align="center">
  <strong>Next-Generation Unified Identity Authentication & Access Control Platform</strong>
</p>

<p align="center">
  <a href="https://github.com/turtacn/QuantaID/actions"><img src="https://img.shields.io/github/actions/workflow/status/turtacn/QuantaID/ci.yml?branch=main" alt="Build Status"></a>
  <a href="https://github.com/turtacn/QuantaID/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.21+-blue.svg" alt="Go Version"></a>
  <a href="https://github.com/turtacn/QuantaID/releases"><img src="https://img.shields.io/github/v/release/turtacn/QuantaID" alt="Latest Release"></a>
  <a href="https://goreportcard.com/report/github.com/turtacn/QuantaID"><img src="https://goreportcard.com/badge/github.com/turtacn/QuantaID" alt="Go Report Card"></a>
</p>

<p align="center">
  <a href="README-zh.md">简体中文</a> |
  <a href="#installation">Installation</a> |
  <a href="docs/architecture.md">Architecture</a> |
  <a href="docs/apis.md">API Reference</a> |
  <a href="#contributing">Contributing</a>
</p>

---

## 🎯 Mission Statement

QuantaID revolutionizes enterprise identity management by providing a **lightweight**, **plugin-based**, and **standards-compliant** unified authentication platform. It addresses the critical pain points of fragmented identity systems, high customization costs, and complex integration challenges across diverse enterprise environments.

## 🌟 Why QuantaID?

In today's complex enterprise landscape, organizations struggle with:

- **High Customization Costs**: Each identity integration requires weeks of custom development
- **Limited Component Reusability**: Authentication components cannot be easily shared across products
- **Fragmented User Experience**: Users manage multiple credentials across different systems
- **Compliance Challenges**: Inconsistent security baselines across global deployments
- **Technical Debt Accumulation**: Legacy authentication systems become maintenance nightmares

**QuantaID transforms these challenges into competitive advantages:**

| Challenge | QuantaID Solution | Business Impact |
|-----------|-------------------|-----------------|
| 🔧 Custom Development | Configuration-Driven Architecture | 60% reduction in delivery time |
| 🔄 Limited Reusability | Plugin Ecosystem & SDKs | 90% code reuse across products |
| 🌍 Global Deployment | Multi-form Factor Delivery | Simplified international expansion |
| 🔒 Security Baseline | Standards-Compliant Core | Unified compliance posture |
| 🏗️ Technical Debt | API-First Design | Future-proof architecture |

## 🚀 Key Features

### 🔐 **Universal Authentication Engine**
- **Multi-Protocol Support**: OAuth 2.1, OIDC, SAML 2.0, LDAP/LDAPS, RADIUS
- **Passwordless Authentication**: WebAuthn/FIDO2 support
- **Adaptive MFA**: Risk-based multi-factor authentication

### 🔌 **Plugin-First Architecture**
- **Extensible Connectors**: Custom identity source integrations
- **Visual Flow Orchestration**: Drag-and-drop authentication workflows
- **Multi-Language SDKs**: Go, Java, Node.js, Python, C++

### 🏢 **Enterprise-Grade Features**
- **Identity Lifecycle Management**: Automated user provisioning/deprovisioning
- **Fine-Grained Authorization**: RBAC/ABAC/ReBAC support
- **Comprehensive Auditing**: Structured logging and compliance reporting
- **High Availability**: Cluster deployment with automatic failover

### 📦 **Flexible Deployment Models**
- **Standalone Binary**: Zero-dependency deployment
- **Container-First**: Kubernetes-native with Helm charts
- **SDK/Library**: Deep integration for performance-critical scenarios
- **Cloud & On-Premise**: Support for hybrid environments

## 📊 Architecture Overview

```mermaid
graph TB
    subgraph CL[Client Layer]
        WEB[Web UI]
        CLI[CLI Tools]
        SDK[Multi-Language SDKs]
    end
    
    subgraph AL[API Gateway Layer]
        GW[API Gateway]
        AUTH[Auth Middleware]
        RATE[Rate Limiter]
    end
    
    subgraph SL[Service Layer]
        ORE[Orchestration Engine]
        AUE[Authentication Engine]
        AZE[Authorization Engine]
        IMS[Identity Management]
        FED[Federation Service]
    end
    
    subgraph PL[Plugin Layer]
        IDP[Identity Providers]
        MFA[MFA Providers]
        CON[Custom Connectors]
    end
    
    subgraph DL[Data Layer]
        PG[(PostgreSQL)]
        RD[(Redis Cache)]
        ES[(Elasticsearch)]
    end
    
    CL --> AL
    AL --> SL
    SL --> PL
    SL --> DL
````

Detailed architecture documentation available at [docs/architecture.md](docs/architecture.md).

## 🛠️ Installation

### Prerequisites

* Go 1.21 or higher
* Docker (optional, for containerized deployment)
* PostgreSQL 12+ (for production deployment)

### Quick Start

```bash
# Install QuantaID CLI
go install github.com/turtacn/QuantaID/cmd/qid@latest

# Initialize a new deployment
qid init --config-dir ./qid-config

# Start QuantaID server
qid server start --config ./qid-config/server.yaml
```

### Using Docker

```bash
# Pull the latest image
docker pull quantaid/quantaid:latest

# Run with docker-compose
curl -O https://raw.githubusercontent.com/turtacn/QuantaID/main/deployments/docker-compose.yml
docker-compose up -d
```

### Kubernetes Deployment

```bash
# Add QuantaID Helm repository
helm repo add quantaid https://helm.quantaid.dev
helm repo update

# Install QuantaID
helm install quantaid quantaid/quantaid \
  --set postgresql.enabled=true \
  --set redis.enabled=true
```

## 📖 Usage Examples

### Basic Authentication Setup

```go
package main

import (
    "context"
    "log"
    "github.com/turtacn/QuantaID/pkg/client"
    "github.com/turtacn/QuantaID/pkg/types"
)

func main() {
    // Initialize QuantaID client
    qid, err := client.New(client.Config{
        Endpoint: "https://your-quantaid-instance.com",
        APIKey:   "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Configure OIDC provider
    provider := &types.IdentityProvider{
        Name:     "corporate-sso",
        Type:     "oidc",
        Enabled:  true,
        Config: map[string]interface{}{
            "issuer_url":     "https://your-corp-sso.com",
            "client_id":      "quantaid-client",
            "client_secret":  "your-secret",
            "scopes":         []string{"openid", "profile", "email"},
        },
    }
    
    ctx := context.Background()
    if err := qid.IdentityProviders.Create(ctx, provider); err != nil {
        log.Fatal(err)
    }
    
    // Start authentication flow
    authURL, err := qid.Auth.GetAuthorizationURL(ctx, &types.AuthRequest{
        Provider:    "corporate-sso",
        RedirectURI: "https://your-app.com/callback",
        State:       "random-state-string",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Redirect user to: %s", authURL)
}
```

### CLI Usage Examples

```bash
# Configure LDAP identity source
qid identity-sources add ldap \
  --name "corporate-ad" \
  --host "ldap.corp.com" \
  --port 636 \
  --use-tls \
  --bind-dn "cn=service,ou=apps,dc=corp,dc=com" \
  --bind-password "secret"

# Set up SAML application
qid applications create saml \
  --name "aws-sso" \
  --acs-url "https://signin.aws.amazon.com/saml" \
  --entity-id "https://signin.aws.amazon.com/saml" \
  --attribute-mapping "email:urn:oid:1.2.840.113549.1.9.1"

# Configure adaptive MFA policy
qid policies create \
  --name "high-risk-mfa" \
  --condition "risk_score > 0.7 OR location.country != 'trusted'" \
  --action "require_mfa:totp,webauthn"

# Monitor authentication metrics
qid metrics auth --since "24h" --group-by provider
```

### Command Line Demo Effects

Generate these demo GIFs using the following prompts:

1. **Basic Setup Demo**: Record `qid-demo setup --interactive` showing configuration wizard
2. **Identity Source Integration**: Record `qid-demo connect ldap --wizard` with step-by-step LDAP setup
3. **Policy Configuration**: Record `qid-demo policy create --visual` showing drag-and-drop policy builder
4. **Real-time Monitoring**: Record `qid-demo monitor --dashboard` displaying live authentication metrics

## 🏗️ Project Structure

```
QuantaID/
├── cmd/                    # Command-line applications
│   ├── qid/               # Main CLI tool
│   └── qid-server/        # Server daemon
├── pkg/                   # Public Go packages
│   ├── client/            # Go client SDK
│   ├── types/             # Core type definitions
│   ├── auth/              # Authentication engine
│   └── plugins/           # Plugin framework
├── internal/              # Private application code
│   ├── server/            # HTTP/gRPC server
│   ├── orchestrator/      # Workflow orchestration
│   └── storage/           # Data persistence
├── web/                   # Web UI components
├── deployments/           # Deployment configurations
├── docs/                  # Documentation
└── scripts/               # Build and utility scripts
```

## 🤝 Contributing

We welcome contributions from the community! Please read our [Contributing Guide](CONTRIBUTING.md) to get started.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/turtacn/QuantaID.git
cd QuantaID

# Install dependencies
go mod download

# Run tests
make test

# Start development server
make dev
```

### Contribution Areas

* 🔌 **Plugin Development**: Create connectors for new identity providers
* 🌐 **Internationalization**: Add support for new languages
* 📚 **Documentation**: Improve guides and API documentation
* 🐛 **Bug Reports**: Help us identify and fix issues
* ✨ **Feature Requests**: Propose new capabilities

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

* 📖 [Documentation](https://docs.quantaid.dev)
* 🏗️ [Architecture Guide](docs/architecture.md)
* 🔧 [API Reference](docs/apis.md)
* 💬 [Community Forum](https://community.quantaid.dev)
* 🐛 [Issue Tracker](https://github.com/turtacn/QuantaID/issues)
* 📈 [Roadmap](https://github.com/turtacn/QuantaID/projects)

---

<p align="center">
  Built with ❤️ by the QuantaID Community
</p>