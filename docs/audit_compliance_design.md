# Audit & Compliance System Design

This document outlines the architecture and design of the QuantaID Audit & Compliance system.

## 1. Objectives

- **Comprehensive Auditing:** To create a detailed, immutable record of all sensitive actions within the system.
- **Regulatory Compliance:** To provide a framework for meeting compliance standards like GDPR, SOC2, and ISO 27001.
- **Data Lifecycle Management:** To automatically manage audit data through its entire lifecycle, from hot storage to archival and deletion.
- **Actionable Insights:** To enable security and operations teams to query, analyze, and generate reports from audit data.

## 2. Architecture

The audit system is designed as a high-throughput, asynchronous pipeline that decouples the event creation from the event processing and storage.

### Core Components:

1.  **`AuditEvent` Struct (`event_types.go`):** A standardized, structured format for all audit events. It includes detailed information about the actor, action, target, and result.

2.  **`AuditLogger` (`logger.go`):** The entry point for the audit pipeline. It provides a non-blocking `Record()` method that places events into an in-memory buffer.

3.  **Asynchronous Engine:** A background goroutine in the `AuditLogger` manages the flushing of the buffer. It batches events and writes them to the persistence layer based on two triggers:
    *   **Batch Size:** When the number of buffered events reaches a configured threshold (e.g., 1000 events).
    *   **Flush Interval:** Periodically, based on a time interval (e.g., every 5 seconds).

4.  **`AuditRepository` Interface:** An abstraction for the persistence layer. This allows the core audit logic to remain independent of the database implementation.

5.  **PostgreSQL Implementation (`audit_repo.go`):** The concrete implementation of the `AuditRepository` that writes audit events to a partitioned PostgreSQL table for efficient time-series data management.

## 3. Data Model & Schema

The `audit_logs` table in PostgreSQL is the source of truth for all audit data.

-   **Schema:** The schema is designed for efficient querying, with indexes on common filter fields like `timestamp`, `actor_id`, `event_type`, and `target_id`.
-   **Partitioning:** The table is partitioned by a time range (e.g., monthly) to ensure that queries are fast and database maintenance (like archiving or deleting old partitions) is efficient.
-   **Data Integrity:** The system relies on the database for data integrity. Future enhancements could include periodic hashing or signing of log batches to ensure immutability.

## 4. Compliance & Data Lifecycle

-   **`ComplianceChecker` (`compliance_checker.go`):** A framework for running automated compliance checks. Rules are defined in `configs/compliance_rules.yaml` and can be extended to support various standards.
-   **`RetentionPolicyManager` (`retention_policy.go`):** A background service that enforces data retention policies defined in `configs/audit_policies.yaml`. It handles the automatic archiving and deletion of logs.

## 5. Reporting & Exporting

-   **`ReportGenerator` (`report_generator.go`):** A service that queries the audit repository and formats the data into various formats (JSON, CSV, PDF).
-   **`audit-exporter` CLI:** A command-line tool that allows administrators to perform on-demand data exports and generate compliance reports.
