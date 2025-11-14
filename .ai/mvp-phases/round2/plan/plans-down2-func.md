好，我们直接把 **P1–P5 全部“下钻到函数级”**，顺带给出关键 demo 代码骨架（只片段，不是完整文件）。
尽量沿用你之前那套目录命名：`internal/...`, `configs/...`, `docs/...`，并显式列出**每个文件里要新增/修改哪些函数**。

---

## PHASE 1: 语义对齐 & 核心分层收敛（P1）

> **(Phase 1: Semantic Alignment & Core Layer Hardening)**
> 目标：
>
> * 建立 `CapabilityMapping`，把架构能力映射到代码 package；
> * 为 Jules 环境提供 memory backend，`go test` 无外部 DB 依赖；
> * 至少跑通一个 Login + Protected API 的 E2E 流程。

---

### 🧩 函数级 TODO 列表

#### 1. `internal/architecture/map.go`（新增文件）

**目标：** 定义 Layer/Capability 枚举 + 默认映射表。

**新增类型与函数：**

```go
// Layer 和 Capability 定义
type Layer string
type Capability string

const (
    LayerPresentation Layer = "presentation"
    LayerGateway      Layer = "gateway"
    LayerAppService   Layer = "app_service"
    LayerDomain       Layer = "domain"
    LayerInfra        Layer = "infra"
)

const (
    CapabilityAuthMultiProtocol Capability = "auth.multi_protocol"
    CapabilityAuthMFA           Capability = "auth.mfa.basic"
    CapabilityIdentityLifecycle Capability = "identity.lifecycle.basic"
    CapabilityConnectorLDAP     Capability = "connector.ldap.basic"
    CapabilityAuditLog          Capability = "audit.log.basic"
    CapabilityMetricsPrometheus Capability = "metrics.prometheus.basic"
)

type CapabilityMapping struct {
    Capability Capability
    Layer      Layer
    Packages   []string
    Status     string // "planned" / "partial" / "done"
}

// TODO: 手动维护的默认映射
var DefaultMappings = []CapabilityMapping{
    // TODO: 填写具体包路径，如：
    {
        Capability: CapabilityAuthMultiProtocol,
        Layer:      LayerAppService,
        Packages: []string{
            "internal/services/auth",
            "internal/server/http/handlers/auth.go",
        },
        Status: "partial",
    },
    // ...
}

// TODO: 提供查询工具函数
func FindMappingsByCapability(c Capability) []CapabilityMapping {
    var res []CapabilityMapping
    for _, m := range DefaultMappings {
        if m.Capability == c {
            res = append(res, m)
        }
    }
    return res
}

func FindMappingsByPackage(pkg string) []CapabilityMapping {
    var res []CapabilityMapping
    for _, m := range DefaultMappings {
        for _, p := range m.Packages {
            if p == pkg {
                res = append(res, m)
                break
            }
        }
    }
    return res
}
```

> TODO 要点：
>
> * 把架构文档中列出的关键能力逐一填入 `DefaultMappings`；
> * 为后续 Phase 文档/工具提供查询接口。

---

#### 2. `internal/storage/memory/identity_memory_repository.go`（新增）

**目标：** 提供内存版 Identity Repository，实现现有 interface。

**需要实现的函数（示例）：**

```go
type identityMemoryRepository struct {
    mu     sync.RWMutex
    users  map[string]*identity.User
    groups map[string]*identity.Group
}

// NewIdentityMemoryRepository 创建实例
func NewIdentityMemoryRepository() *identityMemoryRepository {
    return &identityMemoryRepository{
        users:  make(map[string]*identity.User),
        groups: make(map[string]*identity.Group),
    }
}

// TODO: 实现 interface IdentityRepository 所需方法
func (r *identityMemoryRepository) CreateUser(ctx context.Context, u *identity.User) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // TODO: 生成 ID（如 UUID），防重复校验，写入 map
    return nil
}

func (r *identityMemoryRepository) GetUserByID(ctx context.Context, id string) (*identity.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // TODO: 从 map 查找，找不到返回 domain 定义的 NotFound 错误
    return nil, nil
}

func (r *identityMemoryRepository) UpdateUser(ctx context.Context, u *identity.User) error {
    // TODO
    return nil
}

func (r *identityMemoryRepository) DeleteUser(ctx context.Context, id string) error {
    // TODO
    return nil
}

// 其他：ListUsers / Group 相关 CRUD ...
```

> 约束：
>
> * 不依赖任何外部库（UUID 可用 `crypto/rand` / 简易自增 ID）；
> * 线程安全（使用 `sync.RWMutex`）。

---

#### 3. `internal/storage/memory/auth_memory_repository.go`（新增）

**目标：** 为会话 / token 等提供内存存储。

**核心函数 TODO：**

```go
type authMemoryRepository struct {
    mu      sync.RWMutex
    sessions map[string]*auth.Session
    tokens   map[string]*auth.Token
}

func NewAuthMemoryRepository() *authMemoryRepository {
    // TODO 初始化 map
    return nil
}

func (r *authMemoryRepository) CreateSession(ctx context.Context, s *auth.Session) error {
    // TODO
    return nil
}

func (r *authMemoryRepository) GetSession(ctx context.Context, id string) (*auth.Session, error) {
    // TODO
    return nil, nil
}

func (r *authMemoryRepository) DeleteSession(ctx context.Context, id string) error {
    // TODO
    return nil
}

// Token 相关 CRUD ...
```

---

#### 4. `internal/storage/memory/policy_memory_repository.go`（新增）

**目标：** 存储 Policy。

**函数级 TODO：**

```go
type policyMemoryRepository struct {
    mu      sync.RWMutex
    policies map[string]*policy.Policy
}

func NewPolicyMemoryRepository() *policyMemoryRepository {
    // TODO
    return nil
}

func (r *policyMemoryRepository) CreatePolicy(ctx context.Context, p *policy.Policy) error {
    // TODO
    return nil
}

func (r *policyMemoryRepository) GetPolicyByID(ctx context.Context, id string) (*policy.Policy, error) {
    // TODO
    return nil, nil
}

func (r *policyMemoryRepository) ListPolicies(ctx context.Context, filter policy.Filter) ([]*policy.Policy, error) {
    // TODO: 简单过滤实现
    return nil, nil
}
```

---

#### 5. `internal/server/http/server.go`（修改）

**目标：** 支持 `STORAGE_MODE=memory` / `cfg.Storage.Mode=memory` 时，注入 memory repos。

**关键 TODO 函数示例：**

```go
// Config 中增加字段
type StorageConfig struct {
    Mode string `yaml:"mode"` // "postgres" / "memory"
    // ...
}

// TODO: 新工厂函数
func NewServerWithConfig(cfg *Config) (*Server, error) {
    // TODO: 根据 cfg.Storage.Mode 决定构造哪种 repository
    // if cfg.Storage.Mode == "memory" { use memory repos }
    // else { use postgres repos }
    return nil, nil
}

// TODO: 在 main / cmd 中调用 NewServerWithConfig
```

---

#### 6. `configs/server.jules.yaml`（新增）

**内容要点：**

```yaml
storage:
  mode: "memory"

http:
  listen: ":8080"
  # TODO: 简化 TLS / Auth 配置，适配 Jules 环境
```

---

#### 7. 测试文件函数 TODO

* `tests/unit/identity_memory_repository_test.go`

  ```go
  func TestIdentityMemoryRepository_CreateAndGetUser(t *testing.T) { /* TODO */ }
  func TestIdentityMemoryRepository_UpdateUser(t *testing.T)      { /* TODO */ }
  func TestIdentityMemoryRepository_DeleteUser(t *testing.T)      { /* TODO */ }
  ```
* `tests/integration/server_jules_memory_test.go`

  ```go
  func TestServerWithMemoryBackend_Healthz(t *testing.T)   { /* TODO */ }
  func TestServerWithMemoryBackend_LoginFlow(t *testing.T) { /* TODO: 注册用户 + 登录 + 调用受保护 API */ }
  ```

---

## PHASE 2: 零信任授权 & 策略引擎基础（P2）

> **(Phase 2: Zero-Trust Authorization & Policy Engine Foundation)**

---

### 🧩 函数级 TODO 列表

#### 1. `internal/domain/policy/model.go`（扩展）

**新增结构 & 类型：**

```go
type Subject struct {
    UserID     string
    Groups     []string
    Attributes map[string]string
}

type Resource struct {
    Type       string
    ID         string
    Attributes map[string]string
}

type Action string

type Environment struct {
    IP          string
    Time        time.Time
    DeviceTrust string
}

type EvaluationContext struct {
    Subject     Subject
    Resource    Resource
    Action      Action
    Environment Environment
}

type Decision string

const (
    DecisionAllow Decision = "allow"
    DecisionDeny  Decision = "deny"
)
```

> TODO：
>
> * 若已有 Policy 类型，需要在不破坏现有逻辑的前提下扩展或新增；
> * 增加必要的 JSON/YAML tag 以支持配置加载。

---

#### 2. `internal/services/authorization/evaluator.go`（新增）

**核心接口 & 默认实现：**

```go
// Evaluator 接口
type Evaluator interface {
    Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error)
}

// DefaultEvaluator 从配置加载策略并在内存中评估
type DefaultEvaluator struct {
    rules []Rule // 自定义规则结构
}

type Rule struct {
    Name        string
    Effect      policy.Decision // allow / deny
    Actions     []policy.Action
    Subjects    []string        // 用户ID或组名
    IPWhitelist []string
    TimeRanges  []TimeRange
}

type TimeRange struct {
    Start string // "09:00"
    End   string // "18:00"
}

// TODO: 构造函数
func NewDefaultEvaluatorFromConfig(cfg *Config) (*DefaultEvaluator, error) {
    // TODO: 读取 policy/basic.yaml，解析到 []Rule
    return nil, nil
}

func (e *DefaultEvaluator) Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error) {
    // TODO:
    // 1. 遍历规则，匹配 Action、Subject、IP、Time
    // 2. 命中第一条规则即返回 Effect
    // 3. 未命中则默认 Deny
    return policy.DecisionDeny, nil
}
```

---

#### 3. `internal/services/authorization/service.go`（修改）

**目标：** 所有授权判断统一走 `Evaluator`。

```go
type Service struct {
    evaluator Evaluator
    // TODO: 现有依赖 ...
}

func NewService(e Evaluator /* other deps... */) *Service {
    return &Service{evaluator: e}
}

func (s *Service) Authorize(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error) {
    // TODO: 包薄封装，调用 evaluator
    return s.evaluator.Evaluate(ctx, evalCtx)
}
```

> TODO：
>
> * 替换原本在 service 内硬编码的角色判断逻辑；
> * 确认不再有 handler 直接操作用户角色的判断代码。

---

#### 4. `internal/server/middleware/authz.go`（新增或修改）

**目标：** 将 HTTP 请求转换成 `EvaluationContext`，调用授权服务。

```go
func AuthorizationMiddleware(authzSvc *authorization.Service, action policy.Action, resourceType string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // TODO: 从 context 中获取用户信息（JWT Claims）→ Subject
        // TODO: 从 URL / Path / Params 构建 Resource
        // TODO: 从 request 中获取 IP / User-Agent → Environment

        evalCtx := policy.EvaluationContext{
            Subject: policy.Subject{
                UserID:  userID,
                Groups:  groups,
                // Attributes: 可填充部门、租户等
            },
            Resource: policy.Resource{
                Type: resourceType,
                ID:   resourceID,
            },
            Action: policy.Action(action),
            Environment: policy.Environment{
                IP:          clientIP,
                Time:        time.Now().UTC(),
                DeviceTrust: deviceTrust,
            },
        }

        decision, err := authzSvc.Authorize(c.Request.Context(), evalCtx)
        if err != nil || decision != policy.DecisionAllow {
            c.AbortWithStatus(http.StatusForbidden)
            return
        }

        c.Next()
    }
}
```

> TODO：
>
> * 把管理员接口、敏感 API 都挂上此 middleware；
> * IP 获取优先读取 `X-Forwarded-For` 或 `X-Real-IP`，再 fallback 到 `RemoteAddr`。

---

#### 5. `configs/policy/basic.yaml`（新增）

**结构示例（对应 Rule 结构）：**

```yaml
rules:
  - name: "admin-dashboard-access"
    effect: "allow"
    actions: ["dashboard.read"]
    subjects:
      - "group:admins"
    ip_whitelist: ["10.0.0.0/8", "192.168.0.0/16"]
    time_ranges:
      - start: "08:00"
        end: "20:00"
  - name: "default-deny"
    effect: "deny"
    actions: ["*"]
```

---

#### 6. 测试函数 TODO

* `tests/unit/policy_evaluator_test.go`

  ```go
  func TestDefaultEvaluator_AdminAllowDashboard(t *testing.T)   { /* TODO */ }
  func TestDefaultEvaluator_UserDenyDashboard(t *testing.T)     { /* TODO */ }
  func TestDefaultEvaluator_IpNotInWhitelist_Deny(t *testing.T) { /* TODO */ }
  ```
* `tests/integration/authz_middleware_test.go`

  ```go
  func TestAuthorizationMiddleware_AdminAccessGranted(t *testing.T) { /* TODO */ }
  func TestAuthorizationMiddleware_UserAccessDenied(t *testing.T)   { /* TODO */ }
  ```

---

## PHASE 3: 自适应多因素 & 风险引擎雏形（P3）

> **(Phase 3: Adaptive MFA & Risk Engine Bootstrap)**

---

### 🧩 函数级 TODO 列表

#### 1. `internal/domain/auth/risk_model.go`（新增）

```go
type RiskFactor string

const (
    RiskFactorNewDevice     RiskFactor = "new_device"
    RiskFactorGeoVelocity   RiskFactor = "geo_velocity"
    RiskFactorUnusualTime   RiskFactor = "unusual_time"
    RiskFactorIPReputation  RiskFactor = "ip_reputation"
)

type RiskScore float64

type RiskAssessment struct {
    Score   RiskScore
    Factors []RiskFactor
}

type LoginContext struct {
    UserID         string
    CurrentIP      string
    CurrentCountry string
    UserAgent      string
    Now            time.Time

    LastLoginIP      string
    LastLoginCountry string
    LastLoginAt      time.Time
}
```

---

#### 2. `internal/services/auth/risk_engine.go`（新增）

```go
// RiskEngine 接口
type RiskEngine interface {
    Assess(ctx context.Context, loginCtx auth.LoginContext) (*auth.RiskAssessment, error)
}

// SimpleRiskEngine 规则驱动实现
type SimpleRiskEngine struct {
    cfg SimpleRiskConfig
}

type SimpleRiskConfig struct {
    NewDeviceScore   float64 `yaml:"new_device_score"`
    GeoVelocityScore float64 `yaml:"geo_velocity_score"`
    UnusualTimeScore float64 `yaml:"unusual_time_score"`
    BlockThreshold   float64 `yaml:"block_threshold"`
    MfaThreshold     float64 `yaml:"mfa_threshold"`
}

func NewSimpleRiskEngine(cfg SimpleRiskConfig) *SimpleRiskEngine {
    return &SimpleRiskEngine{cfg: cfg}
}

func (e *SimpleRiskEngine) Assess(ctx context.Context, loginCtx auth.LoginContext) (*auth.RiskAssessment, error) {
    score := 0.0
    var factors []auth.RiskFactor

    // TODO: 规则 1 - 新设备
    if loginCtx.LastLoginIP != "" && loginCtx.LastLoginIP != loginCtx.CurrentIP {
        score += e.cfg.NewDeviceScore
        factors = append(factors, auth.RiskFactorNewDevice)
    }

    // TODO: 规则 2 - 跨国
    if loginCtx.LastLoginCountry != "" && loginCtx.LastLoginCountry != loginCtx.CurrentCountry {
        score += e.cfg.GeoVelocityScore
        factors = append(factors, auth.RiskFactorGeoVelocity)
    }

    // TODO: 规则 3 - 非工作时间
    hour := loginCtx.Now.Hour()
    if hour < 7 || hour > 22 {
        score += e.cfg.UnusualTimeScore
        factors = append(factors, auth.RiskFactorUnusualTime)
    }

    return &auth.RiskAssessment{
        Score:   auth.RiskScore(score),
        Factors: factors,
    }, nil
}
```

---

#### 3. `internal/services/auth/service.go` / `internal/orchestrator/workflows/auth_flow.go`（修改）

**核心 TODO：** 在密码验证后调用 RiskEngine，并决策是否要求 MFA。

伪代码示例：

```go
type AuthService struct {
    // ...
    riskEngine RiskEngine
    mfaService MFAService
}

// 登录主流程
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
    // Step 1: 验证用户名/密码
    user, err := s.verifyPassword(ctx, req.Username, req.Password)
    if err != nil {
        return nil, err
    }

    // Step 2: 构造 LoginContext
    loginCtx := auth.LoginContext{
        UserID:         user.ID,
        CurrentIP:      req.ClientIP,
        CurrentCountry: req.ClientCountry,
        UserAgent:      req.UserAgent,
        Now:            time.Now().UTC(),
        // TODO: 从 audit / session repo 拉用户最近一次登录记录
    }

    assessment, err := s.riskEngine.Assess(ctx, loginCtx)
    if err != nil {
        return nil, err
    }

    // Step 3: 根据风险分数决定 MFA 流程
    if float64(assessment.Score) >= s.riskConfig.BlockThreshold {
        return nil, ErrHighRiskBlocked
    }

    if float64(assessment.Score) >= s.riskConfig.MfaThreshold {
        // TODO: 返回需要 MFA 的状态（如 pending_mfa），不直接签发最终 token
        return s.startMFAFlow(ctx, user, assessment)
    }

    // 低风险：正常签发 token
    return s.issueTokens(ctx, user)
}
```

> TODO：
>
> * 明确 `LoginResponse` 中需要增加的字段（如 `RequiresMFA bool` / `MFAChallengeID string`）；
> * 兼容当前 MFA 流程，避免 breaking change。

---

#### 4. `configs/auth/risk_rules.yaml`（新增）

示例：

```yaml
new_device_score: 0.3
geo_velocity_score: 0.3
unusual_time_score: 0.2

block_threshold: 0.8
mfa_threshold: 0.3
```

---

#### 5. 测试函数 TODO

* `tests/unit/risk_engine_test.go`

  ```go
  func TestSimpleRiskEngine_LowRisk(t *testing.T)      { /* TODO */ }
  func TestSimpleRiskEngine_MediumRisk(t *testing.T)   { /* TODO */ }
  func TestSimpleRiskEngine_HighRisk(t *testing.T)     { /* TODO */ }
  ```
* `tests/integration/adaptive_mfa_flow_test.go`

  ```go
  func TestLoginFlow_NoMFA_WhenLowRisk(t *testing.T)   { /* TODO */ }
  func TestLoginFlow_RequireMFA_WhenMediumRisk(t *testing.T) { /* TODO */ }
  func TestLoginFlow_Block_WhenHighRisk(t *testing.T)  { /* TODO */ }
  ```

---

## PHASE 4: 审计 & 可观测性 & 安全运营基础（P4）

> **(Phase 4: Audit, Observability & Security Operations Foundation)**

---

### 🧩 函数级 TODO 列表

#### 1. `internal/audit/event.go`（新增）

```go
type AuditEvent struct {
    ID        string                 `json:"id"`
    Timestamp time.Time              `json:"ts"`
    Category  string                 `json:"category"` // auth/policy/admin/mfa/risk
    Action    string                 `json:"action"`
    UserID    string                 `json:"user_id,omitempty"`
    IP        string                 `json:"ip,omitempty"`
    Resource  string                 `json:"resource,omitempty"`
    Result    string                 `json:"result"` // success/fail/deny
    TraceID   string                 `json:"trace_id,omitempty"`
    Details   map[string]any         `json:"details,omitempty"`
}
```

---

#### 2. `internal/audit/pipeline.go`（新增）

```go
// Sink 接口
type Sink interface {
    Write(ctx context.Context, event *AuditEvent) error
}

// Pipeline
type Pipeline struct {
    sinks []Sink
}

func NewPipeline(sinks ...Sink) *Pipeline {
    return &Pipeline{sinks: sinks}
}

func (p *Pipeline) Emit(ctx context.Context, event *AuditEvent) {
    for _, s := range p.sinks {
        // TODO: 逐个写入，单个 sink 出错仅打日志，不影响其他 sink
        if err := s.Write(ctx, event); err != nil {
            // TODO: log error / metrics
        }
    }
}

// FileSink 示例
type FileSink struct {
    mu   sync.Mutex
    file *os.File
}

func NewFileSink(path string) (*FileSink, error) {
    // TODO: 打开/创建文件，按行写 JSON
    return nil, nil
}

func (s *FileSink) Write(ctx context.Context, event *AuditEvent) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    // TODO: marshal JSON, append newline, write to file
    return nil
}
```

---

#### 3. `internal/services/audit/service.go`（修改）

**目标：** 审计服务使用 Pipeline，而非仅写 DB。

```go
type Service struct {
    pipeline *audit.Pipeline
    // 保留 DB repository 可选：dbRepo AuditRepository
}

func NewService(p *audit.Pipeline /* dbRepo ... */) *Service {
    return &Service{pipeline: p}
}

func (s *Service) RecordLoginSuccess(ctx context.Context, userID, ip, traceID string, details map[string]any) {
    event := &audit.AuditEvent{
        ID:        generateAuditID(),
        Timestamp: time.Now().UTC(),
        Category:  "auth",
        Action:    "login_success",
        UserID:    userID,
        IP:        ip,
        Result:    "success",
        TraceID:   traceID,
        Details:   details,
    }
    s.pipeline.Emit(ctx, event)
}

// TODO: RecordLoginFailed, RecordPolicyDecision, RecordAdminAction, RecordHighRiskLogin ...
```

---

#### 4. `internal/metrics/http_middleware.go`（新增）

```go
func HTTPMetricsMiddleware(reg *prometheus.Registry) gin.HandlerFunc {
    // TODO: 定义 Histogram / Counter，例如：
    // requestDuration := prometheus.NewHistogramVec(...)
    // requestTotal := prometheus.NewCounterVec(...)
    // reg.MustRegister(requestDuration, requestTotal)

    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath()
        method := c.Request.Method

        c.Next()

        status := c.Writer.Status()
        duration := time.Since(start).Seconds()

        // TODO: 更新 metrics
        _ = path
        _ = method
        _ = status
        _ = duration
    }
}
```

> TODO：
>
> * 在 HTTP server 初始化时加载此 middleware；
> * 确保 `/metrics` 路径本身也暴露相关指标。

---

#### 5. `configs/audit/pipeline.jules.yaml`（新增）

示例：

```yaml
sinks:
  - type: "file"
    path: "./logs/audit_jules.log"
  - type: "stdout"
```

> TODO：
>
> * 配置解析逻辑可放在 `internal/config/audit_config.go` 中，提供 `NewPipelineFromConfig(cfg)`。

---

#### 6. 测试函数 TODO

* `tests/unit/audit_pipeline_test.go`

  ```go
  func TestPipeline_EmitFanout(t *testing.T)       { /* TODO */ }
  func TestFileSink_WriteJSONLine(t *testing.T)    { /* TODO */ }
  ```
* `tests/integration/audit_http_flow_test.go`

  ```go
  func TestLogin_AuditEventsEmitted(t *testing.T)  { /* TODO：检查文件中是否有 login_success/failed */ }
  ```

---

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

## 小结 & 使用建议

* 上面已经把 **P1–P5 每个 Phase** 细化到了：

  * 要新增/修改的 **文件路径**；
  * 每个文件中需要新增/修改的 **函数级 TODO（函数签名 + 粗实现思路）**；
  * 关键的 **demo code 片段**（足以让 jules 自动补全 & 你自己扩展实现）。

* 建议执行方式：

  1. **从 P1 开始**，优先完成 memory backend + `NewServerWithConfig`，确保 Jules 环境一键 `go test`；
  2. 然后 P2 (策略引擎) + P3 (RiskEngine + 自适应 MFA)，把“零信任认证闭环”打通；
  3. 再做 P4 (审计+观测) & P5 (DevCenter API)，把平台化能力补齐。

如果你愿意，我还可以针对“某一个具体文件”（比如 `risk_engine.go` 或 `devcenter_service.go`）写出更接近可直接粘贴运行的完整实现版本。
