package sync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

type SyncStats struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    string    `json:"duration"`
	Created     int       `json:"created"`
	Updated     int       `json:"updated"`
	Disabled    int       `json:"disabled"`
	Unchanged   int       `json:"unchanged"`
	Errors      int       `json:"errors"`
	TotalRemote int       `json:"total_remote"`
	TotalLocal  int       `json:"total_local"`
}

type LDAPSyncService struct {
	ldapConnector LDAPConnector
	userRepo      identity.UserRepository
	config        *LDAPSyncConfig
	auditService  AuditService
	logger        *zap.Logger
	mu            sync.Mutex
	lastSyncStats *SyncStats
}

func NewLDAPSyncService(
	ldapConnector LDAPConnector,
	userRepo identity.UserRepository,
	config *LDAPSyncConfig,
	auditService AuditService,
	logger *zap.Logger,
) *LDAPSyncService {
	return &LDAPSyncService{
		ldapConnector: ldapConnector,
		userRepo:      userRepo,
		config:        config,
		auditService:  auditService,
		logger:        logger.Named("LDAPSyncService"),
	}
}

func (s *LDAPSyncService) GetLastSyncStatus() *SyncStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastSyncStats == nil {
		return &SyncStats{}
	}
	return s.lastSyncStats
}

func (s *LDAPSyncService) FullSync(ctx context.Context) (*SyncStats, error) {
	stats := &SyncStats{StartTime: time.Now()}
	s.logger.Info("Starting full LDAP sync")

	ldapUsers, err := s.ldapConnector.SyncUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch users from LDAP", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch users from LDAP: %w", err)
	}
	stats.TotalRemote = len(ldapUsers)
	ldapUserMap := make(map[string]*types.User, len(ldapUsers))
	for _, user := range ldapUsers {
		ldapUserMap[user.Username] = user
	}

	localUsers, err := s.userRepo.ListUsers(ctx, identity.PaginationQuery{PageSize: 10000, Offset: 0}) // Note: In production, we need a better pagination strategy
	if err != nil {
		s.logger.Error("Failed to list local users", zap.Error(err))
		return nil, fmt.Errorf("failed to list local users: %w", err)
	}
	stats.TotalLocal = len(localUsers)
	localUserMap := make(map[string]*types.User, len(localUsers))
	for _, user := range localUsers {
		localUserMap[user.Username] = user
	}

	// Process updates for existing users and create new ones
	for _, ldapUser := range ldapUsers {
		if localUser, ok := localUserMap[ldapUser.Username]; ok {
			s.updateUser(ctx, localUser, ldapUser, stats)
		} else {
			s.createUser(ctx, ldapUser, stats)
		}
	}

	// Process local users that no longer exist in LDAP
	for _, localUser := range localUsers {
		if _, ok := ldapUserMap[localUser.Username]; !ok {
			if localUser.Status != types.UserStatusInactive {
				s.logger.Info("Disabling user not found in LDAP", zap.String("username", localUser.Username))
				localUser.Status = types.UserStatusInactive
				if err := s.userRepo.UpdateUser(ctx, localUser); err != nil {
					s.logger.Error("Failed to disable user", zap.String("username", localUser.Username), zap.Error(err))
					stats.Errors++
				} else {
					stats.Disabled++
				}
			}
		}
	}

	s.finalizeSync(ctx, "full", stats)
	return stats, nil
}

func (s *LDAPSyncService) IncrementalSync(ctx context.Context, since time.Time) (*SyncStats, error) {
	stats := &SyncStats{StartTime: time.Now()}
	s.logger.Info("Starting incremental LDAP sync", zap.Time("since", since))

	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(modifyTimestamp>=%s))", since.Format("20060102150405Z"))
	changedUsers, err := s.ldapConnector.SearchUsers(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to fetch incremental changes from LDAP", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch incremental changes from LDAP: %w", err)
	}
	stats.TotalRemote = len(changedUsers)

	for _, ldapUser := range changedUsers {
		localUser, err := s.userRepo.GetUserByUsername(ctx, ldapUser.Username)
		if err != nil {
			var e *types.Error
			if errors.As(err, &e) && e.Code == types.ErrUserNotFound.Code {
				s.createUser(ctx, ldapUser, stats)
			} else {
				s.logger.Error("Failed to get local user during incremental sync", zap.String("username", ldapUser.Username), zap.Error(err))
				stats.Errors++
			}
		} else {
			s.updateUser(ctx, localUser, ldapUser, stats)
		}
	}

	s.finalizeSync(ctx, "incremental", stats)
	return stats, nil
}

func (s *LDAPSyncService) createUser(ctx context.Context, ldapUser *types.User, stats *SyncStats) {
	s.logger.Info("Creating new user from LDAP", zap.String("username", ldapUser.Username))
	s.applyLifecycleRules(ldapUser) // Apply rules before creation
	if err := s.userRepo.CreateUser(ctx, ldapUser); err != nil {
		s.logger.Error("Failed to create user", zap.String("username", ldapUser.Username), zap.Error(err))
		stats.Errors++
	} else {
		stats.Created++
	}
}

func (s *LDAPSyncService) updateUser(ctx context.Context, localUser, ldapUser *types.User, stats *SyncStats) {
	if s.config.ConflictStrategy == ConflictPreferLocal {
		stats.Unchanged++
		return // Do nothing if local version is preferred
	}

	// PreferRemote: Update attributes if they differ
	updated := false
	if localUser.Email != ldapUser.Email {
		localUser.Email = ldapUser.Email
		updated = true
	}
	if localUser.Phone != ldapUser.Phone {
		localUser.Phone = ldapUser.Phone
		updated = true
	}
	// Note: a more robust implementation would compare attributes map
	localUser.Attributes = ldapUser.Attributes

	s.applyLifecycleRules(localUser)

	if updated {
		if err := s.userRepo.UpdateUser(ctx, localUser); err != nil {
			s.logger.Error("Failed to update user", zap.String("username", localUser.Username), zap.Error(err))
			stats.Errors++
		} else {
			stats.Updated++
		}
	} else {
		stats.Unchanged++
	}
}

func (s *LDAPSyncService) applyLifecycleRules(user *types.User) {
	for _, rule := range s.config.LifecycleRules {
		if val, ok := user.Attributes[rule.SourceAttr]; ok && val == rule.MatchValue {
			s.logger.Info("Applying lifecycle rule",
				zap.String("username", user.Username),
				zap.String("rule_attr", rule.SourceAttr),
				zap.String("target_status", rule.TargetStatus))
			user.Status = types.UserStatus(rule.TargetStatus)
			break // First matching rule wins
		}
	}
}

func (s *LDAPSyncService) finalizeSync(ctx context.Context, syncType string, stats *SyncStats) {
	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime).String()
	s.logger.Info("LDAP sync finished",
		zap.String("type", syncType),
		zap.Int("created", stats.Created),
		zap.Int("updated", stats.Updated),
		zap.Int("disabled", stats.Disabled),
		zap.Int("errors", stats.Errors),
		zap.String("duration", stats.Duration),
	)

	s.mu.Lock()
	s.lastSyncStats = stats
	s.mu.Unlock()

	// Audit the sync event
	details := map[string]any{
		"sync_type":    syncType,
		"stats":        stats,
		"strategy":     s.config.ConflictStrategy,
		"num_rules":    len(s.config.LifecycleRules),
	}
	s.auditService.RecordAdminAction(ctx, "system", "internal", "ldap_sync", "LDAP Sync Completed", "", details)
}
