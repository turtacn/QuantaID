## PHASE 6: 性能优化与生产环境部署

> **(Phase 6: Performance Optimization & Production Deployment)**

* **Phase ID:** `P6`
* **Branch:** `feat/round1-phase6-optimization`
* **Dependencies:** `P1`, `P2`, `P3`, `P4`, `P5`（需要完整的系统功能）

### 🎯 目标 (Objectives)

* 优化数据库查询性能（索引优化、查询计划分析）
* 实现 Redis 缓存策略（用户会话、OAuth Token）
* 配置生产环境部署（Kubernetes Helm Chart）
* 实现监控和告警（Prometheus + Grafana）
* 实现日志聚合（ELK Stack 或 Loki）
* 实现备份和恢复策略

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `ADD`: `internal/cache/redis_cache.go` - Redis 缓存抽象层
  * `ADD`: `internal/metrics/prometheus.go` - Prometheus 指标导出
  * `ADD`: `deployments/kubernetes/helm-chart/` - Helm Chart 配置
  * `ADD`: `deployments/docker-compose.prod.yml` - 生产环境 Docker Compose
  * `ADD`: `scripts/backup-database.sh` - 数据库备份脚本
  * `MODIFY`: `internal/storage/postgres/*.go` - 添加数据库索引

* **[Dependency Change]** (依赖变更):

  * `ADD`: `github.com/prometheus/client_golang` - Prometheus 客户端
  * `ADD`: `github.com/redis/go-redis/v9` - Redis 客户端
  * `ADD`: `github.com/uber-go/zap` - 结构化日志库（替换标准库）

* **[Infrastructure Change]** (基础设施变更):

  * `ADD`: Kubernetes 集群配置（3 个 Worker 节点）
  * `ADD`: PostgreSQL 主从复制配置
  * `ADD`: Redis Sentinel 高可用配置
  * `ADD`: Prometheus + Grafana 监控栈

### 📝 关键任务 (Key Tasks)

* [ ] `P6-T1`: **[Optimization]** 数据库性能优化

  * 分析慢查询日志（使用 `pg_stat_statements`）
  * 添加索引：

    ```sql
    -- 用户表索引
    CREATE INDEX idx_users_username ON users(username);
    CREATE INDEX idx_users_email ON users(email);
    CREATE INDEX idx_users_created_at ON users(created_at DESC);

    -- 审计日志索引
    CREATE INDEX idx_audit_logs_user_action ON audit_logs(user_id, action, created_at DESC);
    CREATE INDEX idx_audit_logs_ip ON audit_logs(ip_address, created_at DESC);

    -- OAuth Token 索引
    CREATE INDEX idx_oauth_tokens_access_token ON oauth_tokens(access_token);
    CREATE INDEX idx_oauth_tokens_refresh_token ON oauth_tokens(refresh_token);
    CREATE INDEX idx_oauth_tokens_expires_at ON oauth_tokens(expires_at) WHERE revoked = false;
    ```
  * 优化复杂查询（使用 `EXPLAIN ANALYZE` 分析）
  * 配置连接池：最小连接数 10，最大连接数 100

* [ ] `P6-T2`: **[Implement]** 实现 Redis 缓存策略 (`internal/cache/redis_cache.go`)

  * 缓存用户会话（Key: `session:{session_id}`, TTL: 30 分钟）
  * 缓存 OAuth Access Token（Key: `token:{access_token}`, TTL: Token 过期时间）
  * 缓存用户信息（Key: `user:{user_id}`, TTL: 5 分钟）
  * 实现缓存穿透保护（布隆过滤器）
  * 实现缓存雪崩保护（随机 TTL：基础 TTL ± 10%）
  * 示例代码：

    ```go
    type RedisCache struct {
        client *redis.Client
    }

    func (rc *RedisCache) GetUser(ctx context.Context, userID string) (*types.User, error) {
        // 1. 尝试从缓存获取
        cached, err := rc.client.Get(ctx, "user:"+userID).Result()
        if err == nil {
            var user types.User
            json.Unmarshal([]byte(cached), &user)
            return &user, nil
        }
        
        // 2. 缓存未命中，从数据库查询
        user, err := rc.userRepo.GetByID(ctx, userID)
        if err != nil {
            return nil, err
        }
        
        // 3. 写入缓存
        data, _ := json.Marshal(user)
        rc.client.Set(ctx, "user:"+userID, data, 5*time.Minute)
        
        return user, nil
    }
    ```

* [ ] `P6-T3`: **[Implement]** 实现 Prometheus 指标导出 (`internal/metrics/prometheus.go`)

  * 导出指标：

    * `quantaid_http_requests_total` - HTTP 请求总数（按状态码、路径分组）
    * `quantaid_http_request_duration_seconds` - 请求延迟（直方图）
    * `quantaid_db_queries_total` - 数据库查询总数
    * `quantaid_cache_hits_total` / `quantaid_cache_misses_total` - 缓存命中/未命中
    * `quantaid_oauth_tokens_issued_total` - 签发的 Token 数量
    * `quantaid_mfa_verifications_total` - MFA 验证次数（成功/失败）
  * 使用 Gin 中间件自动记录 HTTP 请求指标
  * 暴露 `/metrics` 端点供 Prometheus 抓取

* [ ] `P6-T4`: **[Config]** 创建 Helm Chart (`deployments/kubernetes/helm-chart/`)

  * Chart 结构：

    ```
    quantaid/
    ├── Chart.yaml
    ├── values.yaml
    ├── templates/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   ├── ingress.yaml
    │   ├── configmap.yaml
    │   ├── secret.yaml
    │   └── hpa.yaml  # Horizontal Pod Autoscaler
    ```
  * `values.yaml` 配置项：

    ```yaml
    replicaCount: 3

    image:
      repository: quantaid/quantaid
      tag: "v0.7.0"
      pullPolicy: IfNotPresent

    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "500m"

    autoscaling:
      enabled: true
      minReplicas: 3
      maxReplicas: 10
      targetCPUUtilizationPercentage: 70

    postgresql:
      enabled: true
      auth:
        username: quantaid
        password: <generated>
        database: quantaid
      primary:
        persistence:
          size: 20Gi

    redis:
      enabled: true
      auth:
        enabled: true
        password: <generated>
      master:
        persistence:
          size: 8Gi
    ```

* [ ] `P6-T5`: **[Infrastructure]** 配置生产环境部署

  * 部署 Kubernetes 集群（使用 kubeadm 或托管 Kubernetes 服务）
  * 配置 Ingress Controller（Nginx Ingress）
  * 配置 TLS 证书（Let's Encrypt + cert-manager）
  * 配置持久化存储（使用 StorageClass）
  * 部署 Prometheus + Grafana：

    ```bash
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm install prometheus prometheus-community/kube-prometheus-stack 
      --namespace monitoring --create-namespace
    ```

* [ ] `P6-T6`: **[Monitoring]** 创建 Grafana Dashboard

  * Dashboard 面板：

    * **系统概览**：请求速率、错误率、P95 延迟
    * **数据库**：连接数、查询 QPS、慢查询数量
    * **缓存**：命中率、内存使用率
    * **业务指标**：活跃用户数、登录成功率、MFA 验证成功率
  * 配置告警规则（Prometheus AlertManager）：

    * `HighErrorRate`: HTTP 5xx 错误率 > 5% 持续 5 分钟
    * `SlowDatabaseQueries`: P95 查询延迟 > 1 秒
    * `LowCacheHitRate`: 缓存命中率 < 70%
    * `HighMemoryUsage`: Pod 内存使用率 > 80%

* [ ] `P6-T7`: **[Logging]** 实现日志聚合

  * 使用 Grafana Loki 收集日志
  * 配置日志格式（JSON 格式，包含 trace_id）：

    ```json
    {
      "timestamp": "2025-11-11T12:34:56Z",
      "level": "info",
      "message": "user logged in",
      "trace_id": "abc123",
      "user_id": "user-456",
      "ip_address": "192.168.1.1"
    }
    ```
  * 部署 Promtail（Loki 日志采集器）
  * 在 Grafana 中配置 Loki 数据源

* [ ] `P6-T8`: **[Backup]** 实现备份和恢复策略

  * 数据库备份脚本 (`scripts/backup-database.sh`):

    ```bash
    #!/bin/bash
    BACKUP_DIR="/backups/postgres"
    DATE=$(date +%Y%m%d_%H%M%S)
    pg_dump -h $DB_HOST -U $DB_USER -d quantaid > "$BACKUP_DIR/quantaid_$DATE.sql"
    # 保留最近 7 天的备份
    find $BACKUP_DIR -name "quantaid_*.sql" -mtime +7 -delete
    ```
  * 配置 Cron Job 每天凌晨 3 点自动备份
  * 测试恢复流程：

    ```bash
    psql -h $DB_HOST -U $DB_USER -d quantaid < quantaid_backup.sql
    ```
  * 备份到远程存储（AWS S3 或 MinIO）

* [ ] `P6-T9`: **[Load Testing]** 进行压力测试

  * 使用 k6 进行负载测试：

    ```javascript
    import http from 'k6/http';
    import { check } from 'k6';

    export let options = {
      stages: [
        { duration: '2m', target: 100 }, // 2 分钟内增加到 100 VU
        { duration: '5m', target: 100 }, // 保持 100 VU 5 分钟
        { duration: '2m', target: 0 },   // 2 分钟内降到 0 VU
      ],
    };

    export default function () {
      let res = http.post('https://api.quantaid.com/v1/auth/login', JSON.stringify({
        username: 'testuser',
        password: 'password123',
      }), {
        headers: { 'Content-Type': 'application/json' },
      });
      
      check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
      });
    }
    ```
  * 测试目标：

    * 支持 1000 并发用户
    * P95 响应时间 < 500ms
    * 错误率 < 0.1%

* [ ] `P6-T10`: **[Documentation]** 完善生产部署文档

  * 创建 `docs/deployment/production-guide.md`，包含：

    * 硬件要求（CPU、内存、磁盘）
    * 网络架构图（负载均衡器 → Ingress → Service → Pods）
    * 数据库高可用配置（主从复制、故障转移）
    * 备份和恢复流程（包含演练步骤）
    * 滚动更新策略（Blue-Green 部署）
    * 故障排查手册（常见问题和解决方案）
  * 创建 `docs/operations/runbook.md`（运维手册）：

    * 日常巡检清单（检查日志、监控指标）
    * 扩容/缩容操作步骤
    * 证书续期流程（Let's Encrypt 90 天有效期）
    * 安全事件响应流程（如发现可疑登录）

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[Performance Test]** (性能测试):

  * `Test Case 1`: `k6::TestLoginEndpointUnder1000Users` - 验证 1000 并发用户登录，P95 延迟 < 500ms
  * `Test Case 2`: `k6::TestOAuthTokenIssuanceUnder500QPS` - 验证 Token 签发接口在 500 QPS 下稳定运行 10 分钟
  * `Test Case 3`: `k6::TestDatabaseConnectionPoolExhaustion` - 模拟连接池耗尽场景，验证系统是否返回友好错误

* **[Chaos Engineering Test]** (混沌工程测试):

  * `Test Case 4`: `chaos::TestPodFailure` - 随机杀掉 1 个 Pod，验证系统自动恢复（健康检查通过）
  * `Test Case 5`: `chaos::TestDatabaseLatencyInjection` - 注入 500ms 数据库延迟，验证系统是否超时并降级
  * `Test Case 6`: `chaos::TestRedisConnectionFailure` - 断开 Redis 连接，验证系统回退到数据库查询

* **[Backup & Recovery Test]** (备份恢复测试):

  * `Test Case 7`: `backup::TestFullDatabaseRestore` - 删除生产数据库，从备份恢复，验证数据完整性（对比用户数、应用数）
  * `Test Case 8`: `backup::TestPointInTimeRecovery` - 恢复到 2 小时前的时间点（使用 PostgreSQL PITR）

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (性能) 系统支持 1000 并发用户，P95 响应时间 < 500ms，P99 < 1 秒
* `AC-2`: (可用性) 系统 SLA 达到 99.9%（月度停机时间 < 43 分钟）
* `AC-3`: (扩展性) 使用 HPA 自动扩容，CPU 使用率超过 70% 时自动增加 Pod
* `AC-4`: (监控) Grafana Dashboard 实时显示所有核心指标，告警规则触发时发送邮件/Slack 通知
* `AC-5`: (备份) 数据库每天自动备份，保留最近 30 天的备份文件
* `AC-6`: (安全) 所有生产环境 Secret 使用 Kubernetes Secrets 或外部密钥管理（如 AWS Secrets Manager）
* `AC-7`: (文档) 运维手册包含所有关键操作步骤，新团队成员可在 1 天内完成生产部署

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P6` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] 压力测试报告已生成（包含 TPS、延迟分布、错误率）
* [ ] 混沌工程测试通过（系统能够自动恢复）
* [ ] 生产环境已部署，并运行 7 天无重大故障
* [ ] Grafana Dashboard 已导出为 JSON 文件（`deployments/monitoring/grafana-dashboard.json`）
* [ ] 代码已合并到 `main` 分支，Tag `v0.7.0-phase6`

### 🔧 开发指南与约束 (Development Guidelines & Constraints)

**关键实现思路（Demo Code）：**

**示例 1：Redis 缓存中间件** (`internal/middleware/cache_middleware.go`)

```go
package middleware

import (
    "context"
    "encoding/json"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

type CacheMiddleware struct {
    redis *redis.Client
}

// CacheUserInfo 缓存用户信息
func (cm *CacheMiddleware) CacheUserInfo() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.Next()
            return
        }
        
        // 1. 尝试从缓存获取
        cacheKey := "user:" + userID
        cached, err := cm.redis.Get(c.Request.Context(), cacheKey).Result()
        if err == nil {
            var user types.User
            json.Unmarshal([]byte(cached), &user)
            c.Set("user", &user)
            c.Next()
            return
        }
        
        // 2. 缓存未命中，继续执行后续逻辑
        c.Next()
        
        // 3. 在响应后将用户信息写入缓存
        if user, exists := c.Get("user"); exists {
            data, _ := json.Marshal(user)
            // 随机 TTL 防止缓存雪崩
            ttl := 5*time.Minute + time.Duration(rand.Intn(60))*time.Second
            cm.redis.Set(c.Request.Context(), cacheKey, data, ttl)
        }
    }
}
```

**示例 2：Prometheus 指标中间件** (`internal/middleware/metrics_middleware.go`)

```go
package middleware

import (
    "strconv"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "quantaid_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "quantaid_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func PrometheusMetrics() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        path := c.FullPath() // 使用路由模板而非实际路径（避免高基数）
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
    }
}
```

**示例 3：Kubernetes Deployment** (`deployments/kubernetes/helm-chart/templates/deployment.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "quantaid.fullname" . }}
  labels:
    {{- include "quantaid.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0  # 确保零停机更新
  selector:
    matchLabels:
      {{- include "quantaid.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "quantaid.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: quantaid
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: {{ include "quantaid.fullname" . }}-config
              key: db_host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: {{ include "quantaid.fullname" . }}-secret
              key: db_password
        - name: REDIS_HOST
          value: "{{ .Release.Name }}-redis-master"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: {{ .Release.Name }}-redis
              key: redis-password
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
```

**测试约束：**

* 压力测试必须在独立的测试环境进行（避免影响生产）
* 混沌工程测试必须在非工作时间进行（降低风险）
* 备份恢复测试每季度至少演练一次

---

