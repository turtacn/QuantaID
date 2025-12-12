package radius

import (
	"context"
	"crypto/md5"
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/password"
	"github.com/turtacn/QuantaID/pkg/types"
)

type AuthenticatorConfig struct {
	RequireMFA            bool
	MFAInPassword         bool   // MFA code appended to password
	MFASeparator          string // Separator, default ","
	DefaultSessionTimeout int
}

type Authenticator struct {
	userService     identity.IService
	passwordService password.IService
	// mfaService      mfa.IService // Not available in context yet, assuming simplified for now or add if needed
	codec           *AttributeCodec
	config          AuthenticatorConfig
	mschap          *MSCHAPHandler
}

func NewAuthenticator(userService identity.IService, passwordService password.IService, codec *AttributeCodec, config AuthenticatorConfig) *Authenticator {
	auth := &Authenticator{
		userService:     userService,
		passwordService: passwordService,
		codec:           codec,
		config:          config,
	}
	auth.mschap = NewMSCHAPHandler(userService, passwordService, codec)
	return auth
}

func (a *Authenticator) Authenticate(ctx context.Context, request *Packet, client *RADIUSClient) (*Packet, error) {
	username := request.GetString(AttrUserName)
	if username == "" {
		return a.createReject(request, "Missing username"), nil
	}

	// 1. Check for CHAP
	if chapPassword := request.GetAttribute(AttrCHAPPassword); chapPassword != nil {
		return a.authenticateCHAP(ctx, request, client, username, chapPassword)
	}

	// 2. Check for MS-CHAPv2
	// MS-CHAP response is in Vendor-Specific attribute
	if a.isMSCHAPRequest(request) {
		return a.mschap.Authenticate(ctx, request, client, username)
	}

	// 3. Default to PAP
	return a.authenticatePAP(ctx, request, client, username)
}

func (a *Authenticator) authenticatePAP(ctx context.Context, request *Packet, client *RADIUSClient, username string) (*Packet, error) {
	encPassword := request.GetAttribute(AttrUserPassword)
	if encPassword == nil {
		return a.createReject(request, "Missing password"), nil
	}

	// Decode password
	passwordStr := a.codec.DecodePassword(encPassword.Value, request.Authenticator, []byte(client.Secret))

	// Split MFA if configured
	// mfaCode := ""
	if a.config.MFAInPassword {
		parts := strings.SplitN(passwordStr, a.config.MFASeparator, 2)
		passwordStr = parts[0]
		if len(parts) > 1 {
			// mfaCode = parts[1]
		}
	}

	// Find user
	u, err := a.userService.GetUserByUsername(ctx, username)
	if err != nil || u == nil {
		return a.createReject(request, "Invalid credentials"), nil
	}

	// Verify password
	// Assuming password service has Verify method
	valid, err := a.passwordService.Verify(ctx, u.ID, passwordStr)
	if err != nil || !valid {
		return a.createReject(request, "Invalid credentials"), nil
	}

	// MFA Verification (Placeholder)
	// if a.config.RequireMFA { ... }

	return a.createAccept(request, u), nil
}

func (a *Authenticator) authenticateCHAP(ctx context.Context, request *Packet, client *RADIUSClient, username string, chapPassword *Attribute) (*Packet, error) {
	if len(chapPassword.Value) < 17 {
		return a.createReject(request, "Invalid CHAP response"), nil
	}

	chapID := chapPassword.Value[0]
	chapResponse := chapPassword.Value[1:17]

	challenge := request.GetAttribute(AttrCHAPChallenge)
	var challengeVal []byte
	if challenge == nil {
		challengeVal = request.Authenticator[:]
	} else {
		challengeVal = challenge.Value
	}

	u, err := a.userService.GetUserByUsername(ctx, username)
	if err != nil || u == nil {
		return a.createReject(request, "Invalid credentials"), nil
	}

	// We need plain text password for CHAP
	// Assuming a method exists on passwordService or we fail.
	// As discussed in MSCHAP, likely this is not readily available in typical systems storing hashes.
	// For P6, we assume it's possible or we fail.

	// Mock implementation:
	plainPassword, err := a.getPlainPassword(ctx, u.ID)
	if err != nil {
		return a.createReject(request, "CHAP not supported for this user"), nil
	}

	expected := a.calculateCHAPResponse(chapID, plainPassword, challengeVal)
	if !bytes.Equal(chapResponse, expected) {
		return a.createReject(request, "Invalid credentials"), nil
	}

	return a.createAccept(request, u), nil
}

func (a *Authenticator) isMSCHAPRequest(request *Packet) bool {
	// Check for Microsoft Vendor Attributes
	for _, attr := range request.Attributes {
		if attr.Type == AttrVendorSpecific {
			vid := binary.BigEndian.Uint32(attr.Value[0:4])
			if vid == VendorMicrosoft {
				return true
			}
		}
	}
	return false
}

func (a *Authenticator) calculateCHAPResponse(id byte, password string, challenge []byte) []byte {
	h := md5.New()
	h.Write([]byte{id})
	h.Write([]byte(password))
	h.Write(challenge)
	return h.Sum(nil)
}

func (a *Authenticator) createAccept(request *Packet, user *types.User) *Packet {
	response := request.CreateResponse(CodeAccessAccept)

	// Add Standard Attributes
	// Service-Type = Framed-User (2)
	response.AddAttribute(AttrServiceType, encodeInteger(2))

	if a.config.DefaultSessionTimeout > 0 {
		response.AddAttribute(AttrSessionTimeout, encodeInteger(a.config.DefaultSessionTimeout))
	}

	// Class attribute for accounting correlation
	// User struct in pkg/types/user.go doesn't have TenantID field, using hardcoded default or extracting from attributes?
	// For now, removing TenantID or using empty/default. Multitenancy usually involves context or specific field.
	// Since User struct doesn't have it, we assume single tenant or it's in metadata.
	// We'll just use UserID for now.
	classValue := fmt.Sprintf("user=%s", user.ID)
	response.AddAttribute(AttrClass, []byte(classValue))

	return response
}

func (a *Authenticator) createReject(request *Packet, message string) *Packet {
	response := request.CreateResponse(CodeAccessReject)
	response.AddAttribute(AttrReplyMessage, []byte(message))
	return response
}

func (a *Authenticator) getPlainPassword(ctx context.Context, userID string) (string, error) {
	// Placeholder.
	// In real world, we might fetch from a reversible encryption store if available.
	// Or this method simply fails if we only store one-way hashes.
	return "", fmt.Errorf("cleartext password not available")
}

func encodeInteger(val int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(val))
	return b
}
