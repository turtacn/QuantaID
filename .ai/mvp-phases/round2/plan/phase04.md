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