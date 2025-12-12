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
  <a href="docs/quickstart.md">Getting Started</a> |
  <a href="#-development-setup">Development Setup</a> |
  <a href="#-architecture-overview">Architecture</a> |
  <a href="#-contributing">Contributing</a>
</p>

---

## 🎯 Mission Statement

QuantaID revolutionizes enterprise identity management by providing a **lightweight**, **plugin-based**, and **standards-compliant** unified authentication platform. It addresses the critical pain points of fragmented identity systems, high customization costs, and complex integration challenges across diverse enterprise environments.

## ✨ Getting Started

For a fast and easy setup, please follow our **[Quickstart Guide](docs/quickstart.md)**.

This guide will walk you through cloning the repository, building the binary, and running the server with its dependencies in under 5 minutes.

## 🛠️ Development Setup

QuantaID is designed to be easy to set up for development.

### Prerequisites
* Go 1.21 or higher
* Docker (optional, for containerized deployment)
* PostgreSQL 13+ (optional, for production-like deployment)
* Redis 6+ (optional, for distributed rate limiting and sessions)

### Running for Development
The server supports both in-memory (quick start) and persistent (PostgreSQL + Redis) modes.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/turtacn/QuantaID.git
    cd QuantaID
    ```
2.  **Install dependencies:**
    ```bash
    go mod download
    ```
3.  **Run the server:**
    ```bash
    go run ./cmd/qid-server/
    ```
    The server will start on `http://localhost:8080`.

4.  **Run tests:**
    ```bash
    go test ./...
    ```

## 🏗️ Project Structure

The project follows the standard Go project layout. All custom source code is in the `cmd`, `internal`, and `pkg` directories.

```
QuantaID/
├── cmd/               # Command-line applications
│   ├── qid/           # Main CLI tool for managing the server.
│   └── qid-server/    # The server daemon itself.
├── pkg/               # Public Go packages, intended for use by external applications.
│   ├── client/        # A Go client SDK for interacting with the QuantaID API.
│   ├── types/         # Core type definitions (structs, constants) used across the project.
│   ├── auth/          # The core authentication engine logic.
│   └── plugins/       # The plugin framework, including interfaces and base implementations.
├── internal/          # Private application code, not intended for external use.
│   ├── domain/        # Core business logic and entities, decoupled from frameworks.
│   ├── orchestrator/  # A workflow engine for multi-step processes like authentication flows.
│   ├── server/        # HTTP server setup, handlers, and middleware.
│   ├── services/      # Application services that act as a facade over the domain layer.
│   └── storage/       # Data persistence implementations (e.g., PostgreSQL, Redis, in-memory).
├── deployments/       # Deployment configurations (e.g., Docker, Kubernetes).
└── docs/              # Project documentation.
```

## 📊 Architecture Overview

QuantaID is built on a clean, layered architecture that separates concerns and promotes modularity.

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

## 🛡️ Security Features

### Continuous Session Evaluation (Phase 5)
QuantaID implements a **Continuous Session Evaluation (CSE)** mechanism to dynamically assess and mitigate session risks in real-time.
*   **Real-time Risk Monitoring**: Detects IP changes, geo-jumps, device changes, and inactivity.
*   **Dynamic Policy Enforcement**: Automatically downgrades permissions, requires MFA (Step-Up), or terminates sessions based on risk levels.
*   **Adaptive Scheduling**: Evaluates high-risk sessions more frequently.

For more details, see [Session Evaluation Documentation](docs/session-evaluation.md).

### Identity Profiling & Risk (Phase 4)
*   **Behavioral Profiling**: Aggregates user login habits and device usage.
*   **Risk Scoring**: Real-time risk assessment (0-100) based on anomalies.
*   **Data Quality**: Scores identity data completeness and verification status.
*   **Tagging**: Automated and manual user segmentation.

Learn more in the [Identity Profile Documentation](docs/identity-profile.md).


## 🤝 Contributing

We welcome contributions from the community! Please read our [Contributing Guide](CONTRIBUTING.md) to get started.

### Contribution Areas

* 🔌 **Plugin Development**: Create connectors for new identity providers.
* 🛡️ **Device Trust**: Enhance device fingerprinting and anomaly detection.
* 🌐 **Internationalization**: Add support for new languages.
* 📚 **Documentation**: Improve guides and API documentation.
* 🐛 **Bug Reports**: Help us identify and fix issues.
* ✨ **Feature Requests**: Propose new capabilities.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
