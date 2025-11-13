package auth

import "time"

type RiskFactor string

const (
	RiskFactorNewDevice    RiskFactor = "new_device"
	RiskFactorGeoVelocity  RiskFactor = "geo_velocity"
	RiskFactorUnusualTime  RiskFactor = "unusual_time"
	RiskFactorIPReputation RiskFactor = "ip_reputation"
)

type RiskScore float64

type RiskDecision string

const (
	RiskDecisionAllow      RiskDecision = "allow"
	RiskDecisionRequireMFA RiskDecision = "require_mfa"
	RiskDecisionDeny       RiskDecision = "deny"
)

type RiskAssessment struct {
	Score    RiskScore
	Factors  []RiskFactor
	Decision RiskDecision
}

type LoginContext struct {
	UserID           string
	CurrentIP        string
	CurrentCountry   string
	UserAgent        string
	Now              time.Time
	LastLoginIP      string
	LastLoginCountry string
	LastLoginAt      time.Time
}
