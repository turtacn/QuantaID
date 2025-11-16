package main

import (
	"net/http"

	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/auth/protocols"
	"github.com/turtacn/QuantaID/pkg/utils"
)

func RegisterOAuthHandlers(router *http.ServeMux, logger utils.Logger) {
	// Initialize adapters
	oauthAdapter := protocols.NewOAuthAdapter().(*protocols.OAuthAdapter)
	oidcAdapter := protocols.NewOIDCAdapter().(*protocols.OIDCAdapter)
	engine := orchestrator.NewEngine(logger)

	// Initialize handlers
	oauthHandler := handlers.NewOAuthHandler(oauthAdapter, engine, logger)
	oidcHandler := handlers.NewOIDCHandler(oidcAdapter, logger, "")

	// Register handlers
	router.HandleFunc("/oauth/authorize", oauthHandler.Authorize)
	router.HandleFunc("/oauth/token", oauthHandler.Token)
	router.HandleFunc("/oauth/userinfo", oidcHandler.UserInfo)
	router.HandleFunc("/.well-known/openid-configuration", oidcHandler.Discovery)
	router.HandleFunc("/.well-known/jwks.json", oidcHandler.JWKS)
}
