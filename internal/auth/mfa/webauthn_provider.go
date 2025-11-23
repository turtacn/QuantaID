package mfa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/datatypes"
)

// MFARepository defines the interface needed by WebAuthnProvider to access MFA factors.
// We reproduce it here to avoid circular imports if it was imported from domain/auth.
// Ideally, this should be in a shared interface package.
type MFARepository interface {
	GetUserFactorsByType(ctx context.Context, userID uuid.UUID, factorType string) ([]*types.MFAFactor, error)
	CreateFactor(ctx context.Context, factor *types.MFAFactor) error
	UpdateFactor(ctx context.Context, factor *types.MFAFactor) error
}

// WebAuthnConfig holds the configuration for WebAuthn.
type WebAuthnConfig struct {
	RPID          string
	RPDisplayName string
	RPOrigin      string
}

// WebAuthnProvider implements the MFAProvider interface for WebAuthn.
type WebAuthnProvider struct {
	w       *webauthn.WebAuthn
	mfaRepo MFARepository
}

// NewWebAuthnProvider creates a new WebAuthnProvider.
func NewWebAuthnProvider(cfg WebAuthnConfig, mfaRepo MFARepository) (*WebAuthnProvider, error) {
	w, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webauthn: %w", err)
	}

	return &WebAuthnProvider{
		w:       w,
		mfaRepo: mfaRepo,
	}, nil
}

// WebAuthnUserAdapter adapts types.User to webauthn.User interface.
type WebAuthnUserAdapter struct {
	user        *types.User
	credentials []webauthn.Credential
}

func (u WebAuthnUserAdapter) WebAuthnID() []byte {
	return []byte(u.user.ID)
}

func (u WebAuthnUserAdapter) WebAuthnName() string {
	return u.user.Username
}

func (u WebAuthnUserAdapter) WebAuthnDisplayName() string {
	return u.user.Username
}

func (u WebAuthnUserAdapter) WebAuthnIcon() string {
	return ""
}

func (u WebAuthnUserAdapter) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// loadCredentials fetches WebAuthn credentials for the user from the repository.
func (p *WebAuthnProvider) loadCredentials(ctx context.Context, user *types.User) ([]webauthn.Credential, error) {
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	factors, err := p.mfaRepo.GetUserFactorsByType(ctx, userID, "webauthn")
	if err != nil {
		return nil, err
	}

	var credentials []webauthn.Credential
	for _, f := range factors {
		var cred webauthn.Credential
		// Basic fields
		// Decode CredentialID from Base64URL string
		credID, err := base64.RawURLEncoding.DecodeString(f.CredentialID)
		if err != nil {
			// Log error but continue with other credentials?
			// Or return error. Since DB is source of truth, failure here is bad.
			return nil, fmt.Errorf("failed to decode credential ID for factor %s: %w", f.ID, err)
		}
		cred.ID = credID
		cred.PublicKey = f.PublicKey

		// Load metadata
		if len(f.Metadata) > 0 {
			type CredMetadata struct {
				AttestationType string
				Transport       []protocol.AuthenticatorTransport
				Flags           struct {
					UserPresent    bool `json:"userPresent"`
					UserVerified   bool `json:"userVerified"`
					BackupEligible bool `json:"backupEligible"`
					BackupState    bool `json:"backupState"`
				}
				Authenticator webauthn.Authenticator
			}
			var meta CredMetadata
			if err := json.Unmarshal(f.Metadata, &meta); err == nil {
				cred.AttestationType = meta.AttestationType
				cred.Transport = meta.Transport
				// cred.Flags = meta.Flags // flags are complicated to map back directly if structure changed, skipping for now as not critical for exclude list
				cred.Authenticator = meta.Authenticator
			}
		}
		credentials = append(credentials, cred)
	}
	return credentials, nil
}

// BeginRegistration initiates the WebAuthn registration process.
func (p *WebAuthnProvider) BeginRegistration(ctx context.Context, user *types.User) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	// We don't strictly need existing credentials for registration, but it's good practice to prevent duplicates.
	creds, err := p.loadCredentials(ctx, user)
	if err != nil {
		// Log error but maybe proceed?
		// For now fail.
		return nil, nil, err
	}

	adapter := WebAuthnUserAdapter{
		user:        user,
		credentials: creds,
	}

	// Exclude existing credentials
	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.CredentialExcludeList = make([]protocol.CredentialDescriptor, len(creds))
		for i, cred := range creds {
			credCreationOpts.CredentialExcludeList[i] = protocol.CredentialDescriptor{
				Type: protocol.PublicKeyCredentialType,
				CredentialID: cred.ID,
			}
		}
	}

	return p.w.BeginRegistration(adapter, registerOptions)
}

// FinishRegistration completes the WebAuthn registration process.
func (p *WebAuthnProvider) FinishRegistration(ctx context.Context, user *types.User, sessionData webauthn.SessionData, r *http.Request) (*webauthn.Credential, error) {
	adapter := WebAuthnUserAdapter{
		user: user,
	}

	credential, err := p.w.FinishRegistration(adapter, sessionData, r)
	if err != nil {
		return nil, err
	}

	// Persist the credential
	userID, _ := uuid.Parse(user.ID)

	// Serialize metadata
	meta := struct {
		AttestationType string
		Transport       []protocol.AuthenticatorTransport
		Flags           struct {
			UserPresent    bool `json:"userPresent"`
			UserVerified   bool `json:"userVerified"`
			BackupEligible bool `json:"backupEligible"`
			BackupState    bool `json:"backupState"`
		}
		Authenticator webauthn.Authenticator
	}{
		AttestationType: credential.AttestationType,
		Transport:       credential.Transport,
		Flags: struct {
			UserPresent    bool `json:"userPresent"`
			UserVerified   bool `json:"userVerified"`
			BackupEligible bool `json:"backupEligible"`
			BackupState    bool `json:"backupState"`
		}{
			UserPresent:    credential.Flags.UserPresent,
			UserVerified:   credential.Flags.UserVerified,
			BackupEligible: credential.Flags.BackupEligible,
			BackupState:    credential.Flags.BackupState,
		},
		Authenticator: credential.Authenticator,
	}
	metaBytes, _ := json.Marshal(meta)

	factor := &types.MFAFactor{
		UserID:       userID,
		Type:         "webauthn",
		Status:       "active",
		CredentialID: base64.RawURLEncoding.EncodeToString(credential.ID),
		PublicKey:    credential.PublicKey,
		Metadata:     datatypes.JSON(metaBytes),
	}

	err = p.mfaRepo.CreateFactor(ctx, factor)
	if err != nil {
		return nil, err
	}

	return credential, nil
}

// BeginLogin initiates the WebAuthn login process.
func (p *WebAuthnProvider) BeginLogin(ctx context.Context, user *types.User) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	creds, err := p.loadCredentials(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	if len(creds) == 0 {
		return nil, nil, fmt.Errorf("no webauthn credentials found")
	}

	adapter := WebAuthnUserAdapter{
		user:        user,
		credentials: creds,
	}

	return p.w.BeginLogin(adapter)
}

// FinishLogin completes the WebAuthn login process.
func (p *WebAuthnProvider) FinishLogin(ctx context.Context, user *types.User, sessionData webauthn.SessionData, r *http.Request) (*webauthn.Credential, error) {
	creds, err := p.loadCredentials(ctx, user)
	if err != nil {
		return nil, err
	}

	adapter := WebAuthnUserAdapter{
		user:        user,
		credentials: creds,
	}

	credential, err := p.w.FinishLogin(adapter, sessionData, r)
	if err != nil {
		return nil, err
	}

	return credential, nil
}

// Implement MFAProvider Interface methods (wrappers/placeholders as WebAuthn flow is different)

// Challenge generates a new WebAuthn challenge.
// Note: This generic method doesn't fit the Begin/Finish flow perfectly.
func (p *WebAuthnProvider) Challenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error) {
	// For generic MFA flow, we might return a challenge that prompts the client to call the specific WebAuthn endpoints.
	return &types.MFAChallenge{
		MFAProvider: "webauthn",
		ChallengeID: "use_webauthn_endpoints",
	}, nil
}

// Verify validates a WebAuthn response.
func (p *WebAuthnProvider) Verify(ctx context.Context, user *types.User, response string) (bool, error) {
	return false, fmt.Errorf("use FinishLogin endpoint for webauthn verification")
}

// ListMethods returns the available WebAuthn methods for the user.
func (p *WebAuthnProvider) ListMethods(ctx context.Context, user *types.User) ([]*types.MFAMethod, error) {
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, err
	}
	factors, err := p.mfaRepo.GetUserFactorsByType(ctx, userID, "webauthn")
	if err != nil {
		return nil, err
	}

	var methods []*types.MFAMethod
	for _, f := range factors {
		methods = append(methods, &types.MFAMethod{
			ID:   f.ID.String(),
			Type: "webauthn",
		})
	}
	return methods, nil
}

// GetStrength returns the strength of the WebAuthn provider.
func (p *WebAuthnProvider) GetStrength() StrengthLevel {
	return StrengthLevelStrong
}

// Enroll is a placeholder for the WebAuthn enrollment process.
func (p *WebAuthnProvider) Enroll(ctx context.Context, user *types.User) (*types.MFAEnrollment, error) {
	return nil, fmt.Errorf("use BeginRegistration endpoint for webauthn enrollment")
}
