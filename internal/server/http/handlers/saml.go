package handlers

import (
	"github.com/turtacn/QuantaID/internal/protocols/saml"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
)

// SAMLHandlers provides HTTP handlers for SAML protocol endpoints.
type SAMLHandlers struct {
	samlService *saml.Service
	logger      utils.Logger
}

// NewSAMLHandlers creates a new set of SAML handlers.
//
// Parameters:
//   - samlService: The service that contains the core SAML IdP logic.
//   - logger: The logger for handler-specific messages.
//
// Returns:
//   A new SAMLHandlers instance.
func NewSAMLHandlers(samlService *saml.Service, logger utils.Logger) *SAMLHandlers {
	return &SAMLHandlers{
		samlService: samlService,
		logger:      logger,
	}
}

// HandleSSO is the main handler for the SAML Single Sign-On (SSO) endpoint.
// It receives the SAMLRequest from the Service Provider (SP), initiates the
// authentication flow, and returns a SAMLResponse. It delegates the core
// SAML processing to the SAML protocol service.
func (h *SAMLHandlers) HandleSSO(w http.ResponseWriter, r *http.Request) {
	h.samlService.HandleSSO(w, r)
}

// HandleMetadata is the handler for the IdP metadata endpoint.
// It returns the IdP's SAML metadata XML, which Service Providers use
// to configure their side of the trust relationship. It delegates the core
// metadata generation to the SAML protocol service.
func (h *SAMLHandlers) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	h.samlService.HandleMetadata(w, r)
}