# Identity Profile & Data Quality (Phase 4)

## Overview
Phase 4 introduces a comprehensive Identity Profile system to the QuantaID platform. This module aggregates user behavior, risk indicators, and data quality metrics into a unified `UserProfile` model. It enables adaptive security policies, automated user segmentation (tagging), and identity data governance.

## Key Components

### 1. User Profile Model (`internal/identity/profile/profile_model.go`)
The `UserProfile` struct acts as the central aggregation point:
- **BehaviorMetrics**: Statistics on logins, devices, locations, and session habits.
- **RiskIndicators**: Tracks anomalies like geo-jumps, suspicious IPs, and failed MFA attempts.
- **Tags**: A mix of automated system tags (e.g., `frequent_traveler`) and manual admin tags.
- **QualityDetails**: Assessment of the user's data completeness and verification status.

### 2. Risk Scorer (`risk_scorer.go`)
Calculates a risk score (0-100) based on weighted risk indicators.
- **Decay**: Risk scores decay over time if no new anomalies occur.
- **Levels**: low (<25), medium (<50), high (<75), critical (>=75).

### 3. Tag Manager (`tag_manager.go`)
Manages user segmentation:
- **Auto Tags**: Rule-based tagging (e.g., `high_value` if login frequency > 10/week).
- **Manual Tags**: Admin-assigned labels for custom workflows.

### 4. Quality Scorer (`quality_scorer.go`)
Evaluates the "trustworthiness" of the identity data itself.
- Factors: Verified email/phone, MFA enrollment, profile completeness.
- Outputs: A score (0-100) and improvement suggestions.

### 5. Profile Builder (`profile_builder.go`)
A service that aggregates data from:
- `AccessLogRepository` (Behavior)
- `DeviceRepository` (Device counts)
- `IdentityService` (User attributes)
- `MFAService` (Security posture)

It supports both full rebuilds and incremental updates via event processing.

## Configuration
The system is configured via `config.yaml` (or `pkg/utils/config.go` defaults):

```yaml
profile:
  enabled: true
  risk_scorer:
    anomaly_weight: 15.0
    geo_jump_weight: 20.0
    decay_days: 30
  quality_weights:
    email_verified: 10
    mfa: 20
```

## Usage

### Retrieving a Profile
```go
service := di.GetProfileService()
profile, err := service.GetProfile(ctx, userID)
```

### Handling Events
The `ProfileEventHandler` listens for system events:
- `user.login`: Updates login counts and last activity.
- `device.anomaly`: Updates risk indicators and triggers re-scoring.

### Architecture
The module follows Clean Architecture principles:
- **Domain**: `UserProfile` entities and interfaces.
- **Repository**: `PostgresProfileRepository` for persistence.
- **Service**: `ProfileService` for business logic orchestration.
