# QuantaID Deployment Guide

This guide provides instructions on how to deploy the QuantaID application to a production environment.

## Prerequisites

- Docker
- Kubernetes cluster (e.g., Minikube, Kind, or a cloud provider's managed Kubernetes service)
- Helm 3

## Docker Deployment

1.  **Build the Docker image:**

    ```bash
    docker build -t turtacn/quantaid:latest .
    ```

2.  **Run the Docker container:**

    ```bash
    docker run -p 8080:8080 turtacn/quantaid:latest
    ```

## Helm Deployment

1.  **Package the Helm chart:**

    ```bash
    helm package deploy/helm/quantaid
    ```

2.  **Install the Helm chart:**

    ```bash
    helm install quantaid quantaid-0.1.0.tgz
    ```

    You can customize the deployment by creating a `values.yaml` file and passing it to the `helm install` command:

    ```bash
    helm install quantaid quantaid-0.1.0.tgz -f my-values.yaml
    ```
