package identity

import (
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

// ConflictStrategy defines the strategy for resolving conflicts between local and remote users.
type ConflictStrategy string

const (
	// StrategyRemoteWins updates the local user with remote attributes, overwriting local changes.
	StrategyRemoteWins ConflictStrategy = "RemoteWins"
	// StrategyLocalWins ignores remote changes and keeps local data.
	StrategyLocalWins ConflictStrategy = "LocalWins"
	// StrategyMerge updates only fields that are empty locally or specific allowed fields.
	StrategyMerge ConflictStrategy = "Merge"
)

// ConflictManager handles conflict resolution during synchronization.
type ConflictManager struct{}

// NewConflictManager creates a new instance of ConflictManager.
func NewConflictManager() *ConflictManager {
	return &ConflictManager{}
}

// Resolve resolves the conflict between a local user and a remote user based on the strategy.
// It modifies the localUser in place or returns a merged copy.
// Here we modify localUser to reflect the desired state.
func (cm *ConflictManager) Resolve(localUser, remoteUser *types.User, strategy ConflictStrategy) *types.User {
	if localUser == nil {
		return remoteUser
	}
	if remoteUser == nil {
		return localUser // Should not happen in sync context, but safe fallback
	}

	switch strategy {
	case StrategyRemoteWins:
		// Overwrite mutable fields from remote to local
		localUser.Username = remoteUser.Username
		localUser.Email = remoteUser.Email
		localUser.Phone = remoteUser.Phone
		localUser.Attributes = remoteUser.Attributes // Deep merge could be better, but RemoteWins implies overwrite
		localUser.SourceType = remoteUser.SourceType
		localUser.ExternalID = remoteUser.ExternalID
		// Note: We typically preserve ID, CreatedAt, Password (unless synced), etc.

	case StrategyLocalWins:
		// Do nothing, keep local state.
		// Only update metadata like sync timestamp if needed outside this function.
		return localUser

	case StrategyMerge:
		// Update only if local is empty or specifically allowed
		if localUser.Username == "" {
			localUser.Username = remoteUser.Username
		}
		if localUser.Email == "" {
			localUser.Email = remoteUser.Email
		}
		if localUser.Phone == "" {
			localUser.Phone = remoteUser.Phone
		}
		// Merge attributes
		if localUser.Attributes == nil {
			localUser.Attributes = make(map[string]interface{})
		}
		for k, v := range remoteUser.Attributes {
			if _, exists := localUser.Attributes[k]; !exists {
				localUser.Attributes[k] = v
			}
		}
	}

	// Always record merge history
	record := types.MergeRecord{
		MergedAt: time.Now(),
		Strategy: string(strategy),
	}
	// We might want to store SourceID if available, currently just appending record
	localUser.MergeHistory = append(localUser.MergeHistory, record)

	return localUser
}
