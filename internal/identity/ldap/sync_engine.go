package ldap

import (
	"context"
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

	var pageCookie string
	for {
		entries, newCookie, err := se.ldapClient.SearchPaged(ctx, baseDN, filter, uint32(se.config.BatchSize), pageCookie)
		if err != nil {
			se.metrics.RecordSyncError("search_paged")
			return err
		}

		var users []*types.User
		for _, entry := range entries {
			user, err := se.schemaMapper.MapEntry(entry)
			if err != nil {
				se.logger.Warn("failed to map entry", zap.Error(err))
				continue
			}
			users = append(users, user)
		}

		dedupedUsers, err := se.deduplicator.Process(ctx, users)
		if err != nil {
			se.metrics.RecordSyncError("deduplication")
			se.logger.Error("failed to process users", zap.Error(err))
			continue
		}

		if err := se.identityRepo.UpsertBatch(ctx, dedupedUsers); err != nil {
			se.metrics.RecordSyncError("upsert_batch")
			se.logger.Error("failed to upsert batch", zap.Error(err))
			continue
		}

		if err := se.syncStateRepo.UpdateProgress(ctx, sourceID, len(users)); err != nil {
			se.logger.Error("failed to update progress", zap.Error(err))
		}

		pageCookie = newCookie
		if pageCookie == "" {
			break
		}
	}

	if err := se.syncStateRepo.MarkCompleted(ctx, sourceID, time.Now()); err != nil {
		se.logger.Error("failed to mark sync as completed", zap.Error(err))
	}

	se.metrics.RecordFullSyncDuration(time.Since(startTime))
	se.logger.Info("full sync completed", zap.String("sourceID", sourceID), zap.Duration("duration", time.Since(startTime)))
	return nil
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
