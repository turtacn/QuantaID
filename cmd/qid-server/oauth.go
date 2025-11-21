package main

import (
	"net/http"

	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/utils"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
)

func RegisterOAuthHandlers(router *http.ServeMux, logger utils.Logger, authService *auth.ApplicationService, identityService identity.IService, cryptoManager *utils.CryptoManager) {
	// Initialize handlers
	oauthHandler := handlers.NewOAuthHandlers(authService, identityService, logger)

	// Register handlers
	router.HandleFunc("/oauth/authorize", oauthHandler.Authorize)
	router.HandleFunc("/oauth/token", oauthHandler.Token)
}
