package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"github.com/turtacn/QuantaID/tests/testutils"
)

func TestMFAFlow(t *testing.T) {
	// Initialize logger
	logger := utils.NewNoopLogger()

	// Initialize repositories
	mfaRepo := &testutils.MockMFARepository{}

	// Initialize MFA policy
	mfaPolicy := auth.NewMFAPolicy(mfaRepo, nil, nil)

	// Initialize CryptoManager
	cryptoManager := utils.NewCryptoManager("test-secret-key")

	// Initialize router
	router := http.NewServeMux()
	RegisterMFAHandlers(router, mfaPolicy, logger, cryptoManager)

	// --- Test TOTP Setup ---
	req, _ := http.NewRequest("POST", "/api/v1/users/me/mfa/totp/setup", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var setupResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &setupResp)
	assert.NotEmpty(t, setupResp["qr_code_url"])

	// --- Test TOTP Verify ---
	// In a real test, you would extract the secret from the QR code URL and generate a valid TOTP code.
	secret := "JBSWY3DPEHPK3PXP"
	code, err := totp.GenerateCode(secret, time.Now())
	assert.NoError(t, err)

	mfaRepo.On("GetUserFactors", mock.Anything, mock.Anything).Return([]*types.MFAFactor{
		{
			Type:   "totp",
			Secret: secret,
		},
	}, nil)

	verifyReqBody := `{"code": "` + code + `"}`
	req, _ = http.NewRequest("POST", "/api/v1/users/me/mfa/totp/verify", strings.NewReader(verifyReqBody))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
