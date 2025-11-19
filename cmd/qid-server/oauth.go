package main

import (
	"net/http"

	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/auth/protocols"
	"github.com/turtacn/QuantaID/pkg/utils"
)

func RegisterOAuthHandlers(router *http.ServeMux, logger utils.Logger) {
	// Initialize adapters
	oauthAdapter := protocols.NewOAuthAdapter()
	oidcAdapter := protocols.NewOIDCAdapter()

	// Initialize handlers
	oauthHandler := handlers.NewOAuthHandler(oauthAdapter.(*protocols.OAuthAdapter), logger)
	oidcHandler := handlers.NewOIDCHandler(oidcAdapter.(*protocols.OIDCAdapter), logger, "")

	// Register handlers
	router.HandleFunc("/oauth/authorize", oauthHandler.Authorize)
	router.HandleFunc("/oauth/token", oauthHandler.Token)
	router.HandleFunc("/oauth/userinfo", oidcHandler.UserInfo)
	router.HandleFunc("/.well-known/openid-configuration", oidcHandler.Discovery)
	router.HandleFunc("/.well-known/jwks.json", oidcHandler.JWKS)
}
