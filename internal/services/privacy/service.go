package privacy

import (
	"context"

	"gorm.io/gorm"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"github.com/turtacn/QuantaID/internal/domain/privacy"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"encoding/json"
	"github.com/turtacn/QuantaID/pkg/utils"
)

type GrantConsentRequest struct {
	UserID        string `json:"userId"`
	PolicyID      string `json:"policyId"`
	PolicyVersion string `json:"policyVersion"`
	UserAgent     string `json:"userAgent"`
	IPAddress     string `json:"ipAddress"`
}

type ExportData struct {
	User           *types.User               `json:"user"`
	AuditHistory   []*events.AuditEvent      `json:"auditHistory"`
	ConsentHistory []*privacy.ConsentRecord `json:"consentHistory"`
}

// SessionRevoker defines the behavior required from the session management layer.
// This interface allows us to swap out Redis for mocks or other backends.
type SessionRevoker interface {
	RevokeAllUserSessions(ctx context.Context, userID string) error
}

// AuditRecorder defines the behavior required from the audit logging layer.
// This interface decouples privacy logic from the specific audit service implementation.
type AuditRecorder interface {
	// RecordAdminAction is used here to log the critical action of account erasure.
	// Ensure your concrete audit.Service implements this method signature.
	RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]any)
}

// Service handles privacy-related operations such as account erasure (Right to be Forgotten).
type Service struct {
	db             *gorm.DB
	sessionRevoker SessionRevoker
	auditRecorder  AuditRecorder
	privacyRepo    privacy.Repository
	identityRepo   identity.UserRepository
	auditRepo      auth.AuditLogRepository
	config         *utils.Config
}

// NewService creates a new privacy service with dependencies injected via interfaces.
func NewService(db *gorm.DB, sessionRevoker SessionRevoker, auditRecorder AuditRecorder, privacyRepo privacy.Repository, identityRepo identity.UserRepository, auditRepo auth.AuditLogRepository, config *utils.Config) *Service {
	return &Service{
		db:             db,
		sessionRevoker: sessionRevoker,
		auditRecorder:  auditRecorder,
		privacyRepo:    privacyRepo,
		identityRepo:   identityRepo,
		auditRepo:      auditRepo,
		config:         config,
	}
}

func (s *Service) HasConsentedLatest(ctx context.Context, userID string, policyID string) (bool, error) {
	// 1. 获取系统配置的 LatestVersion
	latestVersion := s.config.Privacy.PolicyVersions[policyID]
	if latestVersion == "" {
		return false, types.ErrNotFound.WithDetails(map[string]string{"reason": "Policy not found."})
	}

	// 2. 查询用户最后一条 GRANTED 记录
	record, err := s.privacyRepo.GetLastConsentRecord(ctx, userID, policyID)
	if err != nil {
		if err == types.ErrNotFound {
			return false, nil
		}
		return false, types.ErrInternal.WithCause(err)
	}

	// 3. 比较 Version >= LatestVersion
	// This is a simplified comparison. A proper semver comparison might be needed.
	return record.PolicyVersion >= latestVersion, nil
}

func (s *Service) GrantConsent(ctx context.Context, req GrantConsentRequest) error {
	// 1. 验证 PolicyVersion 是否有效 (simple check for now)
	if req.PolicyVersion == "" {
		return types.ErrInvalidRequest.WithDetails(map[string]string{"reason": "PolicyVersion is required."})
	}

	// 2. 创建 ConsentRecord
	record := &privacy.ConsentRecord{
		ID:            utils.GenerateUUID(),
		UserID:        req.UserID,
		PolicyID:      req.PolicyID,
		PolicyVersion: req.PolicyVersion,
		Action:        privacy.ConsentActionGranted,
		UserAgent:     req.UserAgent,
		IPAddress:     req.IPAddress,
	}

	// 3. 保存至 DB
	if err := s.privacyRepo.CreateConsentRecord(ctx, record); err != nil {
		return types.ErrInternal.WithCause(err)
	}

	// 4. 更新 User.PrivacySettings
	user, err := s.identityRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return types.ErrInternal.WithCause(err)
	}
	if user.PrivacySettings == nil {
		user.PrivacySettings = make(map[string]interface{})
	}
	user.PrivacySettings[req.PolicyID] = req.PolicyVersion
	if err := s.identityRepo.UpdateUser(ctx, user); err != nil {
		return types.ErrInternal.WithCause(err)
	}

	// Audit log can be added here if needed

	return nil
}

func (s *Service) CollectUserData(ctx context.Context, userID string) (*ExportData, error) {
	data := &ExportData{}
	var err error

	// 1. Identity: Get user data
	data.User, err = s.identityRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}
	// Redact password
	data.User.Password = ""

	// 2. Audit: Get audit history
	auditLogs, err := s.auditRepo.GetLogsForUser(ctx, userID, types.PaginationQuery{PageSize: 1000, Offset: 0})
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}
	for _, log := range auditLogs {
		data.AuditHistory = append(data.AuditHistory, &events.AuditEvent{
			ID:        log.ID,
			Timestamp: log.Timestamp,
			EventType: events.EventType(log.Action),
			Actor:     events.Actor{ID: log.ActorID},
			Target:    events.Target{ID: log.Resource},
			Result:    events.Result(log.Status),
			IPAddress: log.IPAddress,
			UserAgent: log.UserAgent,
		})
	}

	// 3. Consent: Get consent history
	data.ConsentHistory, err = s.privacyRepo.GetConsentHistory(ctx, userID)
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}

	return data, nil
}

func (s *Service) ExportToJSON(data *ExportData) ([]byte, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}
	return bytes, nil
}

// EraseAccount performs a complete removal of a user's data, ensuring sessions are revoked
// and the action is audited.
func (s *Service) EraseAccount(ctx context.Context, userID string) error {
	// 1. Validate Input (Basic check)
	if userID == "" {
		return gorm.ErrRecordNotFound
	}

	// 2. Perform Erasure Transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// A. Invalidate all active sessions for the user immediately
		// We do this inside the transaction (or immediately before/after) to ensure security.
		if err := s.sessionRevoker.RevokeAllUserSessions(ctx, userID); err != nil {
			return err
		}

		// B. Log the erasure event for compliance
		// We log this before the actual delete to ensure a trail exists even if the DB commit fails,
		// or we can log it after. Logging inside the transaction is usually safe if the audit sink is async/robust.
		// Using "system" as actor if triggered automatically, or userID if self-service.
		auditDetails := map[string]interface{}{
			"reason": "gdpr_right_to_be_forgotten",
		}
		s.auditRecorder.RecordAdminAction(ctx, userID, "", "user", "account.erased", "", auditDetails)

		// C. Soft Delete or Hard Delete the user record
		// Assuming a generic User model here; adjust table name as necessary.
		user, err := s.identityRepo.GetUserByID(ctx, userID)
		if err != nil {
			return err
		}
		anonymizedUUID := utils.GenerateUUID()
		user.Username = "deleted_" + anonymizedUUID
		user.Email = types.EncryptedString("deleted_" + anonymizedUUID + "@anonymized.local")
		user.Phone = ""
		user.Password = ""
		user.Status = "deleted"
		if err := s.identityRepo.UpdateUser(ctx, user); err != nil {
			return err
		}

		return nil
	})
}