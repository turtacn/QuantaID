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