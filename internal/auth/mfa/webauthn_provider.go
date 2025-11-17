package mfa

import (
	"context"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
)

type WebAuthnProvider struct {
	webauthn *webauthn.WebAuthn
	repo     *postgresql.PostgresMFARepository
	config   WebAuthnConfig
}

type WebAuthnConfig struct {
	RPDisplayName string `yaml:"rp_display_name"`
	RPID          string `yaml:"rp_id"`
	RPOrigin      string `yaml:"rp_origin"`
	Timeout       int    `yaml:"timeout"`
}

type WebAuthnUser struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return u.ID
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.Name
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (wp *WebAuthnProvider) Enroll(ctx context.Context, userID string, params EnrollParams) (*EnrollResult, error) {
	// user, _ := wp.repo.GetUser(ctx, userID) // TODO: Get user from identity service
	user := &WebAuthnUser{
		ID:          []byte(userID),
		Name:        params.Email,
		DisplayName: params.Email,
	}

	_, sessionData, err := wp.webauthn.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyNotRequired(),
			UserVerification:   protocol.VerificationPreferred,
		}),
	)
	if err != nil {
		return nil, err
	}

	// TODO: store registration session data
	_ = sessionData

	return &EnrollResult{
		// RegistrationOptions: options,
	}, nil
}

func (wp *WebAuthnProvider) FinishEnrollment(ctx context.Context, userID string, response *protocol.CredentialCreationResponse) error {
	// TODO: get registration session data
	return nil
}

func (wp *WebAuthnProvider) Verify(ctx context.Context, userID string, credential string) (bool, error) {
	// TODO:
	return false, nil
}

func (wp *WebAuthnProvider) Revoke(ctx context.Context, userID string, credentialID string) error {
	// TODO:
	return nil
}
