package memory

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
)

// MFAFactorMemoryRepository is an in-memory implementation of the MFAFactorRepository interface.
type MFAFactorMemoryRepository struct {
	sync.RWMutex
	factors map[string]*types.MFAFactor
}

// NewMFAFactorMemoryRepository creates a new MFAFactorMemoryRepository.
func NewMFAFactorMemoryRepository() *MFAFactorMemoryRepository {
	return &MFAFactorMemoryRepository{
		factors: make(map[string]*types.MFAFactor),
	}
}

// CreateMFAFactor stores a new MFA factor in memory.
func (r *MFAFactorMemoryRepository) CreateMFAFactor(ctx context.Context, factor *types.MFAFactor) error {
	r.Lock()
	defer r.Unlock()

	if factor.ID == uuid.Nil {
		factor.ID = uuid.New()
	}

	r.factors[factor.ID.String()] = factor
	return nil
}

// UpdateMFAFactor updates an existing MFA factor in memory.
func (r *MFAFactorMemoryRepository) UpdateMFAFactor(ctx context.Context, factor *types.MFAFactor) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.factors[factor.ID.String()]; !ok {
		return fmt.Errorf("factor not found")
	}

	r.factors[factor.ID.String()] = factor
	return nil
}

// GetMFAFactorsByUserID retrieves all MFA factors for a given user from memory.
func (r *MFAFactorMemoryRepository) GetMFAFactorsByUserID(ctx context.Context, userID string) ([]*types.MFAFactor, error) {
	r.RLock()
	defer r.RUnlock()

	var userFactors []*types.MFAFactor
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	for _, factor := range r.factors {
		if factor.UserID == parsedUserID {
			userFactors = append(userFactors, factor)
		}
	}

	return userFactors, nil
}
