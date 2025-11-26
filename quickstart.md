# QuantaID Quickstart Guide

This guide provides the essential steps to get the QuantaID server up and running on your local machine for development and testing purposes.

## Prerequisites

Before you begin, ensure you have the following installed:

*   **Go:** Version 1.18 or higher.
*   **Docker:** The latest version of Docker and Docker Compose.
*   **Git:** For cloning the repository.

## 1. Clone the Repository

First, clone the QuantaID repository to your local machine:

```bash
git clone https://github.com/turtacn/QuantaID.git
cd QuantaID
```

## 2. Build the Server Binary

Compile the main application binary from the `cmd/qid-server` directory. This command will create a `qid-server` executable in the root of the project.

```bash
go build -o qid-server ./cmd/qid-server
```

## 3. Start Dependent Services

QuantaID requires PostgreSQL and Redis to run. A Docker Compose file is provided to easily start these services.

```bash
sudo docker compose -f deployments/docker-compose/infrastructure.yaml up -d
```

This command will download the required Docker images and start the containers in the background.

## 4. Run the Application

Once the database and cache services are running, you can start the QuantaID server:

```bash
./qid-server
```

The server will start, and you should see log output indicating that it is listening for requests. By default, the server runs on port 8080.

## 5. (Optional) Shut Down Services

When you are finished, you can stop the dependent services with the following command:

```bash
sudo docker compose -f deployments/docker-compose/infrastructure.yaml down
```

You are now ready to start developing and testing QuantaID!
