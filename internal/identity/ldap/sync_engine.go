package ldap

import (
	"context"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
	"time"
)

type SyncEngine struct {
	ldapClient    LDAPClientInterface
	identityRepo  identity.IdentityRepository
	schemaMapper  *SchemaMapper
	deduplicator  *Deduplicator
	syncStateRepo identity.SyncStateRepository
	metrics       *metrics.SyncMetrics
	config        SyncConfig
	logger        *zap.Logger
}

type SyncConfig struct {
	Mode              SyncMode
	FullSyncCron      string
	IncrementalEnable bool
	BatchSize         int
	ConcurrencyLimit  int
}

type SyncMode string

const (
	SyncModeFull        SyncMode = "full"
	SyncModeIncremental SyncMode = "incremental"
)

func NewSyncEngine(
	ldapClient LDAPClientInterface,
	identityRepo identity.IdentityRepository,
	schemaMapper *SchemaMapper,
	deduplicator *Deduplicator,
	syncStateRepo identity.SyncStateRepository,
	metrics *metrics.SyncMetrics,
	config SyncConfig,
	logger *zap.Logger,
) *SyncEngine {
	return &SyncEngine{
		ldapClient:    ldapClient,
		identityRepo:  identityRepo,
		schemaMapper:  schemaMapper,
		deduplicator:  deduplicator,
		syncStateRepo: syncStateRepo,
		metrics:       metrics,
		config:        config,
		logger:        logger,
	}
}

func (se *SyncEngine) StartFullSync(ctx context.Context, sourceID, baseDN, filter string) error {
	startTime := time.Now()
	se.logger.Info("starting full sync", zap.String("sourceID", sourceID))

	// 1. Fetch all LDAP users
	ldapUsers, err := se.fetchAllLdapUsers(ctx, baseDN, filter)
	if err != nil {
		return fmt.Errorf("failed to fetch all LDAP users: %w", err)
	}

	// 2. Fetch all local users linked to this source
	localUsers, err := se.identityRepo.FindUsersBySource(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to fetch local users: %w", err)
	}

	// 3. Create maps for efficient lookup
	ldapMap := make(map[string]*types.User)
	for _, u := range ldapUsers {
		ldapMap[u.Username] = u
	}
	localMap := make(map[string]*types.User)
	for _, u := range localUsers {
		localMap[u.Username] = u
	}

	var toCreate, toUpdate []*types.User
	var toDelete []string

	// 4. Iterate LDAP users -> find users to create or update
	for _, ldapUser := range ldapUsers {
		if localUser, exists := localMap[ldapUser.Username]; exists {
			// User exists, check for changes
			if se.hasChanged(localUser, ldapUser) {
				ldapUser.ID = localUser.ID // Preserve the original ID
				toUpdate = append(toUpdate, ldapUser)
			}
		} else {
			// User doesn't exist, create them
			toCreate = append(toCreate, ldapUser)
		}
	}

	// 5. Iterate local users -> find users to delete
	for _, localUser := range localUsers {
		if _, exists := ldapMap[localUser.Username]; !exists {
			toDelete = append(toDelete, localUser.ID)
		}
	}

	// 6. Perform batch operations
	if len(toCreate) > 0 {
		if err := se.identityRepo.CreateBatch(ctx, toCreate); err != nil {
			se.logger.Error("failed to create users", zap.Error(err))
		}
	}
	if len(toUpdate) > 0 {
		if err := se.identityRepo.UpdateBatch(ctx, toUpdate); err != nil {
			se.logger.Error("failed to update users", zap.Error(err))
		}
	}
	if len(toDelete) > 0 {
		if err := se.identityRepo.DeleteBatch(ctx, toDelete); err != nil {
			se.logger.Error("failed to delete users", zap.Error(err))
		}
	}

	if err := se.syncStateRepo.MarkCompleted(ctx, sourceID, time.Now()); err != nil {
		se.logger.Error("failed to mark sync as completed", zap.Error(err))
	}

	se.metrics.RecordFullSyncDuration(time.Since(startTime))
	se.logger.Info("full sync completed",
		zap.String("sourceID", sourceID),
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("created", len(toCreate)),
		zap.Int("updated", len(toUpdate)),
		zap.Int("deleted", len(toDelete)),
	)
	return nil
}

func (se *SyncEngine) fetchAllLdapUsers(ctx context.Context, baseDN, filter string) ([]*types.User, error) {
	var allEntries []*ldap.Entry
	var pageCookie string
	for {
		entries, newCookie, err := se.ldapClient.SearchPaged(ctx, baseDN, filter, uint32(se.config.BatchSize), pageCookie)
		if err != nil {
			se.metrics.RecordSyncError("search_paged")
			return nil, err
		}
		allEntries = append(allEntries, entries...)
		pageCookie = newCookie
		if pageCookie == "" {
			break
		}
	}

	var users []*types.User
	for _, entry := range allEntries {
		user, err := se.schemaMapper.MapEntry(entry)
		if err != nil {
			se.logger.Warn("failed to map entry", zap.Error(err))
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (se *SyncEngine) hasChanged(local, ldap *types.User) bool {
	if local.Email != ldap.Email {
		return true
	}
	if local.Status != ldap.Status {
		return true
	}

	// A more robust implementation would compare all mapped attributes
	return false
}

func (se *SyncEngine) StartIncrementalSync(ctx context.Context, sourceID, baseDN, filter string) error {
	se.logger.Info("starting incremental sync", zap.String("sourceID", sourceID))

	sr, err := se.ldapClient.PersistentSearch(ctx, baseDN, filter, []string{"*", "+"})
	if err != nil {
		return err
	}

	for _, entry := range sr.Entries {
		se.processChangeNotification(ctx, entry)
	}

	se.logger.Info("LDAP persistent search channel closed")
	return nil
}

func (se *SyncEngine) processChangeNotification(ctx context.Context, entry *ldap.Entry) {
	// For simplicity, we'll just upsert the user.
	// A real implementation would need to handle deletes and renames.
	user, err := se.schemaMapper.MapEntry(entry)
	if err != nil {
		se.logger.Warn("failed to map entry", zap.Error(err))
		return
	}

	if err := se.identityRepo.UpsertBatch(ctx, []*types.User{user}); err != nil {
		se.metrics.RecordSyncError("upsert_batch")
		se.logger.Error("failed to upsert user", zap.Error(err))
	}
}
