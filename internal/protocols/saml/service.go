package saml

import (
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/turtacn/QuantaID/internal/domain/application"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// Service now acts as a wrapper around the crewjam/saml IdentityProvider
// and implements the necessary provider interfaces.
type Service struct {
	logger         utils.Logger
	appRepo        application.Repository
	identityDomain identity.IService
	crypto         *utils.CryptoManager
	IDP            *saml.IdentityProvider
}

// NewService creates a new SAML protocol service and configures the underlying IdP.
func NewService(
	logger utils.Logger,
	appRepo application.Repository,
	identityDomain identity.IService,
	crypto *utils.CryptoManager,
	idpKey *rsa.PrivateKey,
	idpCert *x509.Certificate,
	idpIssuerURL string,
) (*Service, error) {
	ssoURL, err := url.Parse(idpIssuerURL + "/sso")
	if err != nil {
		return nil, err
	}
	metadataURL, err := url.Parse(idpIssuerURL)
	if err != nil {
		return nil, err
	}

	s := &Service{
		logger:         logger,
		appRepo:        appRepo,
		identityDomain: identityDomain,
		crypto:         crypto,
	}

	s.IDP = &saml.IdentityProvider{
		SSOURL:                  *ssoURL,
		MetadataURL:             *metadataURL,
		Key:                     idpKey,
		Certificate:             idpCert,
		ServiceProviderProvider: s,
		SessionProvider:         s,
	}
	return s, nil
}

// GetServiceProvider looks up a service provider by its metadata URL (entityID).
// This method implements the saml.ServiceProviderProvider interface.
func (s *Service) GetServiceProvider(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	// In a real implementation, we would look up the application by its entity ID
	// from s.appRepo and construct the EntityDescriptor from its configuration.
	// For now, we return a hardcoded dummy SP.
	acsURL, _ := url.Parse("http://localhost:3000/v1/auth/saml/acs")

	return &saml.EntityDescriptor{
		EntityID: serviceProviderID,
		SPSSODescriptors: []saml.SPSSODescriptor{
			{
				AssertionConsumerServices: []saml.IndexedEndpoint{
					{
						Binding:  saml.HTTPPostBinding,
						Location: acsURL.String(),
						Index:    1,
					},
				},
			},
		},
	}, nil
}

// GetSession retrieves the currently authenticated user session.
// This method implements the saml.SessionProvider interface.
func (s *Service) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	// This is where you would integrate with your actual session management.
	// If the user is not logged in, you must redirect them to a login page and return nil.
	// Returning nil signals to the IdP that a new authentication flow should be initiated.
	// For this example, we create a dummy session for a hardcoded user.

	return &saml.Session{
		ID:           s.crypto.GenerateUUID(),
		NameID:       "test.user@example.com",
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
		UserEmail:    "test.user@example.com",
		UserName:     "Test User",
		Groups:       []string{"Admins", "Developers"},
	}
}

// HandleSSO now delegates directly to the underlying IdP's ServeSSO method.
func (s *Service) HandleSSO(w http.ResponseWriter, r *http.Request) {
	s.IDP.ServeSSO(w, r)
}

// HandleMetadata now delegates directly to the underlying IdP's ServeMetadata method.
func (s *Service) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	s.IDP.ServeMetadata(w, r)
}