## PHASE 5: 平台服务 & 开发者中心最小版（P5）

> **(Phase 5: Minimal Platform Services & Developer Center)**

---

### 🧩 函数级 TODO 列表

#### 1. `internal/services/platform/devcenter_service.go`（新增）

```go
type DevCenterService struct {
    appSvc       *application.Service      // 应用管理
    connectorSvc *connector.Service       // Connector 管理（如 LDAP）
    policySvc    *authorization.Service   // 策略视图
    mfaSvc       *mfa.Service             // MFA 配置
}

func NewDevCenterService(
    appSvc *application.Service,
    connectorSvc *connector.Service,
    policySvc *authorization.Service,
    mfaSvc *mfa.Service,
) *DevCenterService {
    return &DevCenterService{
        appSvc:       appSvc,
        connectorSvc: connectorSvc,
        policySvc:    policySvc,
        mfaSvc:       mfaSvc,
    }
}

// TODO: 应用相关
func (s *DevCenterService) ListApps(ctx context.Context) ([]*DevCenterAppDTO, error) {
    // 调用 appSvc，转换为 DTO
    return nil, nil
}

func (s *DevCenterService) CreateApp(ctx context.Context, req CreateAppRequest) (*DevCenterAppDTO, error) {
    // TODO: 调用 appSvc.CreateOIDCClient / CreateSAMLApp
    return nil, nil
}

// TODO: Connector 相关
func (s *DevCenterService) ListConnectors(ctx context.Context) ([]*DevCenterConnectorDTO, error) {
    return nil, nil
}

func (s *DevCenterService) EnableConnector(ctx context.Context, id string) error {
    // TODO: 调用 connectorSvc.Enable
    return nil
}

// TODO: 诊断
func (s *DevCenterService) Diagnostics(ctx context.Context) (*DiagnosticsDTO, error) {
    // TODO: 汇总版本信息、配置片段、健康状态等
    return nil, nil
}
```

---

#### 2. `internal/server/http/handlers/devcenter.go`（新增）

```go
type DevCenterHandler struct {
    svc *platform.DevCenterService
}

func NewDevCenterHandler(svc *platform.DevCenterService) *DevCenterHandler {
    return &DevCenterHandler{svc: svc}
}

func (h *DevCenterHandler) RegisterRoutes(r *gin.RouterGroup, authz middleware.AuthorizationMiddlewareProvider) {
    // 仅管理员可访问
    adminGroup := r.Group("/devcenter")
    adminGroup.Use(authz.RequireAction("devcenter.admin", "devcenter"))
    {
        adminGroup.GET("/apps", h.ListApps)
        adminGroup.POST("/apps", h.CreateApp)
        adminGroup.GET("/connectors", h.ListConnectors)
        adminGroup.POST("/connectors/:id/enable", h.EnableConnector)
        adminGroup.GET("/diagnostics", h.Diagnostics)
    }
}

func (h *DevCenterHandler) ListApps(c *gin.Context) {
    ctx := c.Request.Context()
    apps, err := h.svc.ListApps(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, apps)
}

func (h *DevCenterHandler) CreateApp(c *gin.Context) {
    var req platform.CreateAppRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    ctx := c.Request.Context()
    app, err := h.svc.CreateApp(ctx, req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, app)
}

// ListConnectors / EnableConnector / Diagnostics 同理
```

> TODO：
>
> * 在 `admin_routes.go` 中调用 `NewDevCenterHandler` 并注册到 `/api` 下；
> * 授权使用 P2 的策略引擎（action: `devcenter.admin`，subject: admin 组）。

---

#### 3. `pkg/types/devcenter.go`（新增）

```go
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
```

---

#### 4. `configs/devcenter/jules.yaml`（新增）

示例：

```yaml
devcenter:
  enabled: true
  allow_write: true   # 是否允许在 Jules 环境真正创建 App / Connector
```

> TODO：
>
> * 在 DevCenterService 或 Handler 层读取此配置，若 `allow_write=false` 则对写操作返回 403/feature disabled。

---

#### 5. 测试函数 TODO

* `tests/unit/devcenter_service_test.go`

  ```go
  func TestDevCenterService_ListApps(t *testing.T)        { /* TODO */ }
  func TestDevCenterService_CreateApp(t *testing.T)       { /* TODO */ }
  func TestDevCenterService_EnableConnector(t *testing.T) { /* TODO */ }
  ```
* `tests/integration/devcenter_api_test.go`

  ```go
  func TestDevCenterAPI_AdminCanManageApps(t *testing.T)      { /* TODO */ }
  func TestDevCenterAPI_NonAdminForbidden(t *testing.T)       { /* TODO */ }
  func TestDevCenterAPI_Diagnostics(t *testing.T)             { /* TODO */ }
  ```

---