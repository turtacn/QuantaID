package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

func RegisterMFAHandlers(router *http.ServeMux, mfaPolicy *auth.MFAPolicy, logger utils.Logger, crypto *utils.CryptoManager) {
	router.HandleFunc("/api/v1/users/me/mfa/totp/setup", handleTOTPSetup(mfaPolicy, logger))
	router.HandleFunc("/api/v1/users/me/mfa/totp/verify", handleTOTPVerify(mfaPolicy, logger))
	router.HandleFunc("/api/v1/auth/mfa/challenge", handleMFAChallenge(mfaPolicy, logger))
	router.HandleFunc("/api/v1/users/me/mfa/recovery-codes", handleGenerateRecoveryCodes(mfaPolicy, logger, crypto))
}

func handleTOTPSetup(mfaPolicy *auth.MFAPolicy, logger utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is a simplified example. In a real application, you would get the user ID from the session.
		// userID := uuid.New()

		// In a real application, you would get the issuer and account name from the user's data.
		issuer := "QuantaID"
		accountName := "user@example.com"

		// The TOTP provider would be a dependency of the MFA policy.
		key, err := mfaPolicy.TotpProvider.GenerateSecret(issuer, accountName)
		if err != nil {
			http.Error(w, "Failed to generate secret", http.StatusInternalServerError)
			return
		}

		// In a real application, you would save the secret to the database.

		qrCodeURL := mfaPolicy.TotpProvider.GenerateQRCodeURL(key)

		// Respond with the QR code URL.
		json.NewEncoder(w).Encode(map[string]string{"qr_code_url": qrCodeURL})
	}
}

func handleTOTPVerify(mfaPolicy *auth.MFAPolicy, logger utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is a simplified example. In a real application, you would get the user ID from the session.
		userID := uuid.New()

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// In a real application, you would get the user's TOTP secret from the database.
		valid, err := mfaPolicy.VerifyTOTP(r.Context(), userID, req.Code)
		if err != nil {
			http.Error(w, "Failed to verify TOTP code", http.StatusInternalServerError)
			return
		}

		if !valid {
			http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
			return
		}

		// In a real application, you would mark the user's TOTP as verified in the database.

		w.WriteHeader(http.StatusOK)
	}
}

func handleGenerateRecoveryCodes(mfaPolicy *auth.MFAPolicy, logger utils.Logger, crypto *utils.CryptoManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is a simplified example. In a real application, you would get the user ID from the session.
		// userID := uuid.New()

		codes, err := crypto.GenerateRecoveryCodes()
		if err != nil {
			http.Error(w, "Failed to generate recovery codes", http.StatusInternalServerError)
			return
		}

		// In a real application, you would hash and save the recovery codes to the database.

		json.NewEncoder(w).Encode(map[string][]string{"recovery_codes": codes})
	}
}

func handleMFAChallenge(mfaPolicy *auth.MFAPolicy, logger utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ChallengeID string `json:"challenge_id"`
			Method      string `json:"method"`
			Code        string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := mfaPolicy.VerifyMFAChallenge(r.Context(), req.ChallengeID, req.Method, req.Code); err != nil {
			if err, ok := err.(*types.Error); ok {
				http.Error(w, err.Message, err.HttpStatus)
				return
			}
			http.Error(w, "Failed to verify MFA challenge", http.StatusInternalServerError)
			return
		}

		// In a real application, you would issue the final access token here.

		w.WriteHeader(http.StatusOK)
	}
}
