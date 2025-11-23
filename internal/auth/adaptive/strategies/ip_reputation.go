package strategies

import (
	"context"
)

// IPReputationStrategy checks IP address against lists and frequency.
type IPReputationStrategy struct {
	// In a real implementation, this would have access to a threat intelligence feed or repository.
}

// NewIPReputationStrategy creates a new IPReputationStrategy.
func NewIPReputationStrategy() *IPReputationStrategy {
	return &IPReputationStrategy{}
}

// CalculateRisk assesses the risk of the given IP address.
func (s *IPReputationStrategy) CalculateRisk(ctx context.Context, ip string) (float64, error) {
	// Placeholder logic as per requirements
	// 1. Check blacklist
	// 2. Check frequency (maybe via Redis if passed in)

	if ip == "1.2.3.4" { // Example bad IP
		return 0.9, nil
	}
	if ip == "8.8.8.8" { // Example good IP
		return 0.1, nil
	}

	return 0.4, nil // Neutral
}
