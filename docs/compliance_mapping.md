# Compliance Standard Mapping

This document maps the features of the QuantaID Audit & Compliance system to the requirements of various regulatory standards. This is a living document and should be updated as new compliance features are added.

## GDPR (General Data Protection Regulation)

| GDPR Article | Requirement | QuantaID Feature Mapping | Status |
| :--- | :--- | :--- | :--- |
| **Art. 5(1)(e)** | **Storage Limitation:** Personal data kept no longer than necessary. | `RetentionPolicyManager`: Automatically deletes user data and associated audit logs after a configurable period (e.g., 7 years). | ✅ Implemented |
| **Art. 15** | **Right of Access:** Data subjects have the right to access their personal data. | `audit-exporter` CLI: Can be used to export all audit logs related to a specific `actor_id` (user). | ✅ Implemented |
| **Art. 17** | **Right to Erasure ('Right to be Forgotten'):** Obligation to erase personal data without undue delay. | User deletion API triggers the removal of user data. Retention policy ensures associated logs are eventually purged. | ✅ Implemented |
| **Art. 30** | **Records of Processing Activities:** Maintain a record of processing activities under its responsibility. | The entire `audit_logs` table serves as a comprehensive record of data processing activities. | ✅ Implemented |
| **Art. 32** | **Security of Processing:** Implement appropriate technical and organizational measures. | - **Access Controls:** QuantaID's own authorization model restricts access to audit data.<br>- **Immutability:** The append-only nature of the log provides a degree of tamper evidence. | ✅ Implemented |

## SOC 2 (Service Organization Control 2)

| Trust Services Criteria | Requirement | QuantaID Feature Mapping | Status |
| :--- | :--- | :--- | :--- |
| **CC6.1** | **Logical Access Control:** Restrict logical access to authorized users. | Audit logs for `auth.login.success`, `auth.login.failure`, `authz.permission.granted`, and `authz.permission.revoked` provide a complete trail of access events. | ✅ Implemented |
| **CC7.1** | **Monitoring Controls:** Monitor the system to detect changes and anomalies. | `ComplianceChecker`: Can be configured with rules to ensure monitoring is active (e.g., critical services are logging). | ✅ Implemented |
| **CC7.2** | **System Monitoring for Malicious Activity:** Monitor for security incidents and anomalies. | The `audit_logs` can be streamed to a SIEM for real-time analysis of malicious patterns (e.g., high rate of `auth.login.failure`). | ✅ Implemented |
| **CC3.2** | **Change Management:** A process for authorizing, testing, and approving changes. | Audit logs for `system.config.changed` provide a record of all significant system changes. | ✅ Implemented |

## ISO 27001

| ISO 27001 Control | Requirement | QuantaID Feature Mapping | Status |
| :--- | :--- | :--- | :--- |
| **A.12.4.1** | **Event Logging:** Produce, maintain, and review logs of user activities, exceptions, and security events. | The `AuditLogger` and `audit_logs` table are the core implementation of this control. | ✅ Implemented |
| **A.12.4.3** | **Administrator and Operator Logs:** Log privileged user activities. | `AuditEvent`'s `Actor` field can distinguish between regular users and administrators, allowing for specific reports on privileged actions. | ✅ Implemented |
| **A.16.1** | **Information Security Incident Management:** Log and manage security incidents. | Security-related events (e.g., `auth.login.failure`, MFA failures) are specifically typed for easy filtering and alerting. | ✅ Implemented |
