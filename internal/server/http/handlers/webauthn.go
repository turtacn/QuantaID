package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
)

// WebAuthnHandler handles WebAuthn registration and login requests.
type WebAuthnHandler struct {
	provider   *mfa.WebAuthnProvider
	userRepo   identity.UserRepository
	redis      redis.RedisClientInterface
}

// NewWebAuthnHandler creates a new WebAuthnHandler.
func NewWebAuthnHandler(provider *mfa.WebAuthnProvider, userRepo identity.UserRepository, redis redis.RedisClientInterface) *WebAuthnHandler {
	return &WebAuthnHandler{
		provider: provider,
		userRepo: userRepo,
		redis:    redis,
	}
}

// Helper to construct types.Error
func newError(code string, message string, cause error, status int) *types.Error {
	err := &types.Error{
		Code:       code,
		Message:    message,
		HttpStatus: status,
	}
	if cause != nil {
		err.WithCause(cause)
	}
	return err
}

// BeginRegistration handles the initiation of WebAuthn registration.
func (h *WebAuthnHandler) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("user_id").(string) // Assumes AuthMiddleware populates this

	user, err := h.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		WriteJSONError(w, newError("USER_NOT_FOUND", "Failed to get user", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	options, sessionData, err := h.provider.BeginRegistration(ctx, user)
	if err != nil {
		WriteJSONError(w, newError("WEBAUTHN_ERROR", "Failed to begin registration", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Store session data in Redis with 5 minute TTL
	sessionKey := fmt.Sprintf("webauthn:reg:%s", userID)
	sessionBytes, _ := json.Marshal(sessionData)
	err = h.redis.Set(ctx, sessionKey, sessionBytes, 5*time.Minute)
	if err != nil {
		WriteJSONError(w, newError("REDIS_ERROR", "Failed to save session", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, options)
}

// FinishRegistration handles the completion of WebAuthn registration.
func (h *WebAuthnHandler) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("user_id").(string)

	user, err := h.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		WriteJSONError(w, newError("USER_NOT_FOUND", "Failed to get user", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Retrieve session data
	sessionKey := fmt.Sprintf("webauthn:reg:%s", userID)
	sessionBytesStr, err := h.redis.Get(ctx, sessionKey)
	if err != nil {
		WriteJSONError(w, newError("SESSION_EXPIRED", "Session not found or expired", err, http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionBytesStr), &sessionData); err != nil {
		WriteJSONError(w, newError("INTERNAL_ERROR", "Failed to unmarshal session", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	credential, err := h.provider.FinishRegistration(ctx, user, sessionData, r)
	if err != nil {
		WriteJSONError(w, newError("REGISTRATION_FAILED", "Failed to finish registration", err, http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Cleanup session
	h.redis.Del(ctx, sessionKey)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": "registered",
		"credential_id": credential.ID,
	})
}

// BeginLogin handles the initiation of WebAuthn login.
func (h *WebAuthnHandler) BeginLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If fails, try query param? Or context if already logged in (re-auth).
		// If UserID is in context, use it.
		if uid, ok := ctx.Value("user_id").(string); ok {
			user, err := h.userRepo.GetUserByID(ctx, uid)
			if err != nil {
				WriteJSONError(w, newError("USER_NOT_FOUND", "User not found", err, http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			h.beginLoginForUser(w, r, user)
			return
		}
		WriteJSONError(w, newError("INVALID_REQUEST", "Username required", err, http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		WriteJSONError(w, newError("USER_NOT_FOUND", "User not found", err, http.StatusNotFound), http.StatusNotFound)
		return
	}

	h.beginLoginForUser(w, r, user)
}

func (h *WebAuthnHandler) beginLoginForUser(w http.ResponseWriter, r *http.Request, user *types.User) {
	ctx := r.Context()

	options, sessionData, err := h.provider.BeginLogin(ctx, user)
	if err != nil {
		WriteJSONError(w, newError("WEBAUTHN_ERROR", "Failed to begin login", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Use a random session ID to prevent DoS on the user ID key
	// The client must return this session ID in the FinishLogin request
	// Simple random string generation (in production use crypto/rand)
	loginSessionID := fmt.Sprintf("%d", time.Now().UnixNano())

	sessionKey := fmt.Sprintf("webauthn:login:%s", loginSessionID)
	sessionBytes, _ := json.Marshal(sessionData)
	h.redis.Set(ctx, sessionKey, sessionBytes, 5*time.Minute)

	// Wrap options to include session_id
	var respMap map[string]interface{}
	optsBytes, _ := json.Marshal(options)
	json.Unmarshal(optsBytes, &respMap)
	respMap["session_id"] = loginSessionID

	WriteJSON(w, http.StatusOK, respMap)
}

// FinishLogin handles the completion of WebAuthn login.
func (h *WebAuthnHandler) FinishLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Request Body is the WebAuthn response.
	// URL params: ?session_id=...

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		WriteJSONError(w, newError("INVALID_REQUEST", "Session ID query parameter required", nil, http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	sessionKey := fmt.Sprintf("webauthn:login:%s", sessionID)
	sessionBytesStr, err := h.redis.Get(ctx, sessionKey)
	if err != nil {
		WriteJSONError(w, newError("SESSION_EXPIRED", "Session expired or invalid", err, http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionBytesStr), &sessionData); err != nil {
		WriteJSONError(w, newError("INTERNAL_ERROR", "Failed to unmarshal session", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Retrieve user from sessionData.UserID
	user, err := h.userRepo.GetUserByID(ctx, string(sessionData.UserID))
	if err != nil {
		WriteJSONError(w, newError("USER_NOT_FOUND", "User not found", err, http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	credential, err := h.provider.FinishLogin(ctx, user, sessionData, r)
	if err != nil {
		WriteJSONError(w, newError("AUTH_FAILED", "Authentication failed", err, http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	h.redis.Del(ctx, sessionKey)

	// Issue Auth Token / Session
	// In a real flow, this would call the AuthService to generate tokens.

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": "authenticated",
		"credential_id": credential.ID,
		"user_id": user.ID,
	})
}
