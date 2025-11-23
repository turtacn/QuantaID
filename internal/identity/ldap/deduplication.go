package ldap

import (
	"context"
	"fmt"
	pkg_types "github.com/turtacn/QuantaID/pkg/types"
	"strings"
	"time"
)

type Deduplicator struct {
	rules       []DeduplicationRule
	conflictMgr *ConflictManager
}

type DeduplicationRule struct {
	MatchFields []string
	Priority    int
}

type ConflictResolution string

const (
	ResolvePriority  ConflictResolution = "priority"
	ResolveTimestamp ConflictResolution = "timestamp"
	ResolveManual    ConflictResolution = "manual"
)

type Conflict struct {
	Existing *pkg_types.User
	New      *pkg_types.User
}

func NewDeduplicator(rules []DeduplicationRule, conflictMgr *ConflictManager) *Deduplicator {
	return &Deduplicator{
		rules:       rules,
		conflictMgr: conflictMgr,
	}
}

type ConflictManager struct {
	// For now, this is a placeholder.
	// In a real implementation, this would interact with the database.
}

func (cm *ConflictManager) ResolveConflict(existing, newUser *pkg_types.User) (resolution struct{ Action string; MergeStrategy string }) {
	// A real implementation would use the configured conflict resolution strategy.
	// For now, we'll just keep the existing user.
	return struct {
		Action        string
		MergeStrategy string
	}{Action: "keep_existing"}
}

func (cm *ConflictManager) SaveConflicts(ctx context.Context, conflicts []*Conflict) error {
	// A real implementation would save the conflicts to the database.
	return nil
}

func (d *Deduplicator) Process(ctx context.Context, users []*pkg_types.User) ([]*pkg_types.User, error) {
	dedupMap := make(map[string]*pkg_types.User)
	conflicts := []*Conflict{}

	for _, user := range users {
		key := d.generateDeduplicationKey(user)
		if key == "" {
			continue
		}

		if existing, found := dedupMap[key]; found {
			resolution := d.conflictMgr.ResolveConflict(existing, user)
			switch resolution.Action {
			case "keep_existing":
				continue
			case "replace":
				dedupMap[key] = user
			case "merge":
				dedupMap[key] = d.mergeIdentities(existing, user, resolution.MergeStrategy)
			case "defer":
				conflicts = append(conflicts, &Conflict{Existing: existing, New: user})
			}
		} else {
			dedupMap[key] = user
		}
	}

	if len(conflicts) > 0 {
		d.conflictMgr.SaveConflicts(ctx, conflicts)
	}

	result := make([]*pkg_types.User, 0, len(dedupMap))
	for _, identity := range dedupMap {
		result = append(result, identity)
	}

	return result, nil
}

func (d *Deduplicator) generateDeduplicationKey(user *pkg_types.User) string {
	for _, rule := range d.rules {
		var keyParts []string
		for _, field := range rule.MatchFields {
			switch field {
			case "email":
				if user.Email != "" {
					keyParts = append(keyParts, fmt.Sprintf("email:%s", user.Email))
				}
			}
		}
		if len(keyParts) > 0 {
			return strings.Join(keyParts, ";")
		}
	}
	return ""
}

func (d *Deduplicator) mergeIdentities(a, b *pkg_types.User, strategy string) *pkg_types.User {
	merged := &pkg_types.User{}

	merged.Username = coalesce(a.Username, b.Username)
	merged.Email = pkg_types.EncryptedString(coalesce(string(a.Email), string(b.Email)))

	merged.Attributes = mergeJSON(a.Attributes, b.Attributes, strategy)

	merged.MergeHistory = append(a.MergeHistory, pkg_types.MergeRecord{
		SourceIDs: []string{a.ID, b.ID},
		MergedAt:  time.Now(),
		Strategy:  strategy,
	})

	return merged
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func mergeJSON(a, b map[string]interface{}, strategy string) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}
	return merged
}
