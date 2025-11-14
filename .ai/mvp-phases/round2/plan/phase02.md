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
