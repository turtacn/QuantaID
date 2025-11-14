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
