package types

// TokenRequest represents the request to the token endpoint.
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CodeVerifier string `json:"code_verifier"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Scope        string `json:"scope"`
}

// TokenResponse represents the successful response from the token endpoint.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// UserInfo represents the user information returned by the UserInfo endpoint.
type UserInfo struct {
	Subject         string `json:"sub"`
	Name            string `json:"name,omitempty"`
	GivenName       string `json:"given_name,omitempty"`
	FamilyName      string `json:"family_name,omitempty"`
	MiddleName      string `json:"middle_name,omitempty"`
	Nickname        string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile         string `json:"profile,omitempty"`
	Picture         string `json:"picture,omitempty"`
	Website         string `json:"website,omitempty"`
	Email           string `json:"email,omitempty"`
	EmailVerified   bool   `json:"email_verified,omitempty"`
	Gender          string `json:"gender,omitempty"`
	Birthdate       string `json:"birthdate,omitempty"`
	Zoneinfo        string `json:"zoneinfo,omitempty"`
	Locale          string `json:"locale,omitempty"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool `json:"phone_number_verified,omitempty"`
	Address         string `json:"address,omitempty"`
	UpdatedAt       int64  `json:"updated_at,omitempty"`
}

// Client represents an OAuth 2.0 client.
type Client struct {
	ID                      string   `json:"id"`
	Secret                  string   `json:"secret"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	ClientName              string   `json:"client_name"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

// ClientMetadata represents the metadata for a new client registration.
type ClientMetadata struct {
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	ClientName              string   `json:"client_name"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}
