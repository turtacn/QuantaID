package types

type DevCenterAppDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Protocol    string `json:"protocol"` // oidc/saml
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	Enabled     bool   `json:"enabled"`
}

type CreateAppRequest struct {
	Name        string `json:"name" binding:"required"`
	Protocol    string `json:"protocol" binding:"required"`
	RedirectURI string `json:"redirect_uri"`
	// TODO: 其他配置项：回调地址、签名算法等
}

type DevCenterConnectorDTO struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // ldap/...
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	TenantID string `json:"tenant_id,omitempty"`
}

type DiagnosticsDTO struct {
	Version    string            `json:"version"`
	BuildTime  string            `json:"build_time"`
	GoVersion  string            `json:"go_version"`
	ConfigInfo map[string]string `json:"config_info"`
	// TODO: 可包含当前启用的 connectors/app 数量等
}
