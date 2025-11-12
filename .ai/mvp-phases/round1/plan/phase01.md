## PHASE 1: 核心基础设施完善与 Redis 客户端实现

> **(Phase 1: Core Infrastructure Enhancement & Redis Client Implementation)**

* **Phase ID:** `P1`
* **Branch:** `feat/round1-phase1-infrastructure`
* **Dependencies:** 无（基于当前 main 分支）

### 🎯 目标 (Objectives)

* 完善数据持久化层，实现真正的 Redis 客户端（当前为内存模拟）
* 建立完整的配置管理系统，支持多环境配置
* 实现结构化日志和基础可观测性
* 建立单元测试和集成测试框架

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `ADD`: `internal/storage/redis/client.go` - 真正的 Redis 客户端实现
  * `MODIFY`: `internal/storage/redis/session.go` - 替换内存实现为 Redis 调用
  * `MODIFY`: `internal/storage/redis/cache.go` - 替换内存实现为 Redis 调用
  * `ADD`: `pkg/observability/metrics.go` - Prometheus 指标定义
  * `ADD`: `pkg/observability/tracing.go` - OpenTelemetry 追踪初始化
  * `MODIFY`: `pkg/utils/logger.go` - 增强日志功能（增加 trace_id、结构化字段）
  * `ADD`: `configs/server.yaml.example` - 完整配置示例文件
  * `ADD`: `tests/integration/redis_test.go` - Redis 集成测试

* **[Dependency Change]** (依赖变更)

  * `ADD`: `github.com/redis/go-redis/v9` - Redis 官方客户端
  * `ADD`: `github.com/prometheus/client_golang` - Prometheus 客户端
  * `ADD`: `go.opentelemetry.io/otel` - OpenTelemetry SDK
  * `ADD`: `github.com/testcontainers/testcontainers-go` - 集成测试容器

* **[Config Change]** (配置变更)

  * `ADD`: `configs/server.yaml.example` - 包含 Redis 连接池、日志级别、监控端口等配置
  * `ADD`: `configs/testing.yaml` - 测试环境专用配置

### 📝 关键任务 (Key Tasks)

* [ ] `P1-T1`: **[Implement]** 创建 `internal/storage/redis/client.go`，实现 Redis 连接池管理器

  * 使用 `go-redis/v9`，支持哨兵模式和集群模式
  * 实现健康检查方法 `HealthCheck()`
  * 实现优雅关闭逻辑
  * 配置项：`host`, `port`, `password`, `db`, `pool_size`, `timeout`

* [ ] `P1-T2`: **[Refactor]** 重构 `internal/storage/redis/session.go`

  * 将 `InMemorySessionRepository` 重命名为 `RedisSessionRepository`
  * 替换 `map[string]sessionWithValue` 为 Redis 的 `SETEX` 和 `GET` 操作
  * 实现 `GetUserSessions` 使用 Redis `KEYS` 或 `SCAN` 命令（考虑性能）
  * 关键方法：`CreateSession()`, `GetSession()`, `DeleteSession()`, `GetUserSessions()`

* [ ] `P1-T3`: **[Refactor]** 重构 `internal/storage/redis/cache.go`

  * 将 `InMemoryTokenRepository` 重命名为 `RedisTokenRepository`
  * Refresh Token 存储使用 Redis `SETEX`，JTI 黑名单使用 `SETEX`
  * 实现 TTL 自动过期逻辑

* [ ] `P1-T4`: **[Implement]** 创建 `pkg/observability/metrics.go`

  * 定义 Prometheus 指标：

    * `quantaid_auth_requests_total` (Counter) - 认证请求总数
    * `quantaid_auth_duration_seconds` (Histogram) - 认证耗时
    * `quantaid_active_sessions` (Gauge) - 当前活跃会话数
    * `quantaid_redis_operations_total` (Counter) - Redis 操作计数
  * 暴露 `/metrics` HTTP 端点

* [ ] `P1-T5`: **[Implement]** 创建 `pkg/observability/tracing.go`

  * 初始化 OpenTelemetry Tracer Provider
  * 配置 OTLP exporter（支持 stdout 和 Jaeger）
  * 为关键路径（认证流程）添加 Span

* [ ] `P1-T6`: **[Test Design]** 创建 `tests/integration/redis_test.go`

  * 使用 `testcontainers-go` 启动 Redis 容器
  * 测试用例：

    * `TestRedisSessionCRUD`: 会话的创建、读取、过期
    * `TestRedisTokenDenyList`: Token 黑名单功能
    * `TestRedisConnectionPoolExhaustion`: 连接池耗尽场景

* [ ] `P1-T7`: **[Config]** 完善 `configs/server.yaml.example`

  * 包含完整的配置注释
  * 示例值：Redis 连接字符串、日志级别、监控端口

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[Unit Test]** (单元测试):

  * `Test Case 1`: `pkg/utils/logger_test.go::TestLoggerWithTraceID` - 验证日志是否包含 trace_id 字段
  * `Test Case 2`: `internal/storage/redis/client_test.go::TestRedisHealthCheck` - 验证健康检查在 Redis 不可用时返回错误

* **[Integration Test]** (集成测试 - 对应 `P1-T6`):

  * `Test Case 3`: `tests/integration/redis_test.go::TestRedisSessionCRUD` - 启动 Redis 容器，验证会话存储和检索
  * `Test Case 4`: `tests/integration/redis_test.go::TestRedisTokenExpiry` - 验证 Refresh Token 的 TTL 自动过期

* **[Manual Test]** (手动测试):

  * `Test Case 5`: 启动服务，访问 `/metrics` 端点，验证 Prometheus 指标是否正常暴露
  * `Test Case 6`: 连接外部 Redis 实例，进行 100 次并发认证请求，观察 Redis 连接池使用情况

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (代码质量) 所有新增代码通过 `golangci-lint` 检查，无 critical 错误
* `AC-2`: (测试覆盖率) `internal/storage/redis/` 包的单元测试覆盖率 ≥ 80%
* `AC-3`: (集成测试) `Test Case 3` 和 `Test Case 4` 在 CI 环境中 100% 通过
* `AC-4`: (功能验证) `Test Case 5` 手动验证通过，`/metrics` 端点返回至少 4 个自定义指标
* `AC-5`: (性能) `Test Case 6` 中 Redis 连接池无泄漏，所有连接最终正确归还
* `AC-6`: (文档) `configs/server.yaml.example` 包含不少于 20 个配置项的详细注释

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P1` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] CI 流水线中集成测试全部通过
* [ ] 代码已通过至少 1 名 Reviewer 的 Code Review
* [ ] 更新 `CHANGELOG.md`，记录本 Phase 的主要变更
* [ ] 合并到 `main` 分支，并打上 Tag `v0.2.0-phase1`

### 🔧 开发指南与约束 (Development Guidelines & Constraints)

**开发环境要求：**

* Go 1.21+
* Docker 24.0+（用于 testcontainers）
* Redis 7.0+（本地测试或使用 Docker Compose）

**关键实现思路（Demo Code）：**

**示例 1：Redis 客户端初始化** (`internal/storage/redis/client.go`)

```go
package redis

import (
    "context"
    "github.com/redis/go-redis/v9"
    "time"
)

type RedisClient struct {
    client *redis.Client
    cfg    *RedisConfig
}

type RedisConfig struct {
    Host        string
    Port        int
    Password    string
    DB          int
    PoolSize    int
    DialTimeout time.Duration
}

func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password:     cfg.Password,
        DB:           cfg.DB,
        PoolSize:     cfg.PoolSize,
        DialTimeout:  cfg.DialTimeout,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
    
    // 健康检查
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis ping failed: %w", err)
    }
    
    return &RedisClient{client: rdb, cfg: cfg}, nil
}

func (rc *RedisClient) HealthCheck(ctx context.Context) error {
    return rc.client.Ping(ctx).Err()
}

func (rc *RedisClient) Close() error {
    return rc.client.Close()
}
```

**示例 2：Session 存储重构** (`internal/storage/redis/session.go` 部分)

```go
func (r *RedisSessionRepository) CreateSession(ctx context.Context, session *types.UserSession) error {
    key := fmt.Sprintf("session:%s", session.SessionID)
    data, err := json.Marshal(session)
    if err != nil {
        return fmt.Errorf("marshal session: %w", err)
    }
    
    ttl := time.Until(session.ExpiresAt)
    if ttl <= 0 {
        return fmt.Errorf("session already expired")
    }
    
    return r.client.SetEx(ctx, key, data, ttl).Err()
}

func (r *RedisSessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    data, err := r.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, types.NewError(types.ErrCodeNotFound, "session not found")
    }
    if err != nil {
        return nil, fmt.Errorf("redis get: %w", err)
    }
    
    var session types.UserSession
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("unmarshal session: %w", err)
    }
    
    return &session, nil
}
```

**测试约束：**

* 所有集成测试必须使用 `testcontainers`，不依赖外部 Redis 实例
* 测试完成后必须清理测试数据（会话、Token 等）
* 禁止在测试中使用 `time.Sleep()` 进行同步，使用 `context.WithTimeout()` 或 `Eventually()` 模式

---

