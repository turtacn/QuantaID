先直接给一个**偏差分析总览**，然后按要求一段一段给出 Phase 计划（P1/P2/P3/...）。
（注意：下面只用 architecture.md 的架构作为“设计真相源”，和 semantic_digest.yaml 里的 Go 代码结构作为“当前实现语义”。）

---

# 一、设计 vs 实现：GAP 总体偏差分析

> 粗略结论：当前 Go 代码已经初步实现了 **多协议认证、身份管理、LDAP Connector、MFA、HTTP Server、Prometheus 指标** 等核心能力，但离架构里设想的“统一认证中台 + 零信任 + 安全运营 + 平台化”还有一大段路；尤其是：
>
> * **策略引擎/零信任授权** 还停留在基础 Policy 模型，缺乏 ABAC/OPA / 持续验证；
> * **自适应风控、威胁检测、SOC 级能力** 几乎尚未落地；
> * **用户/管理员门户、开发者中心、可视化编排、合规/隐私能力** 只体现在文档/目录结构，代码层面基本缺席或偏骨架；
> * **观测性与运维治理** 有 Prometheus metrics 和审计表，但缺少统一审计流水线与 SIEM 集成。

下面按架构层次来拆：

---

## 1. 分层架构对比（L1–L5）

### 1.1 展现层 / 平台服务 (L1 + L3)

**设计里：**

* Web 管理控制台、登录页面、多语言 SDK、CLI 工具；
* 用户体验门户（自助服务 + 管理控制台）、开发者中心（API 网关 + SDK + 插件市场）、集成适配层（连接器生态）。

**当前实现 (semantic_digest)：**

* 有 `cmd/qid` CLI、`cmd/qid-server` 服务入口；
* 有 HTTP Server：`internal/server/http/server.go` + 一组 handlers（`auth.go / oidc.go / saml.go / oauth.go / admin_api.go / identity.go` 等）；
* metrics：`internal/metrics/prometheus.go`；
* plugins 目录下已存在 MFA / Connector 插件（如 LDAP）相关代码；
* 语义摘要只覆盖 Go 代码，**不包括 web/* 前端**，因此从当前语义视角看不到真正 UI 门户。

**主要 GAP：**

1. **用户体验门户 & 管理控制台**

   * 设计要求自助密码重置、MFA 管理、登录历史、管理员仪表盘等；
   * 实现侧没有“门户层”抽象，HTTP handler 更像“API 管理面”，而非用户自助门户。

2. **开发者中心 / API 网关功能**

   * 设计中提到统一 API 网关、限流熔断、版本管理；
   * 实现中 HTTP Server 更像直接暴露业务接口，没有独立的 Gateway 层（限流、熔断、Routing Policy）。

3. **集成适配层可视化/配置化能力不足**

   * LDAP Connector 已有实现 (`pkg/plugins/connectors/ldap/*`, `internal/services/sync/ldap_sync_service.go`)，但：

     * Connector 注册/生命周期管理多是代码级，而非“平台可配置 + 开发者自助管理”；
     * 无“可视化流程编排”“低代码集成”的影子。

---

### 1.2 应用服务层 & 领域层 (L3/L4)

**设计里：**

* 清晰的 L3/L4 分层：

  * L3：身份管理服务、认证编排服务、授权策略服务、审计日志服务；
  * L4：认证引擎、授权引擎、身份联邦、策略引擎，面向 use case。

**当前实现：**

* `internal/services/identity/*`、`internal/services/auth/*`、`internal/services/authorization/*`、`internal/services/audit/*`：应用服务层基本齐全；
* `internal/domain/identity/*`、`internal/domain/auth/*`、`internal/domain/policy/*`：领域模型 + Repository 接口存在；
* `internal/orchestrator/engine.go` + `internal/orchestrator/workflows/auth_flow.go`：已经有认证编排引擎的雏形。

**GAP：**

1. **认证编排“脚本化/配置化”不足**

   * 文档里期望的是“可视化/低代码编排”，支持基于上下文动态更改流程；
   * 当前 orchestrator 更偏“写死的 Go 流程”，外部不可配置/不可动态调整。

2. **策略引擎缺少 ABAC / ReBAC / OPA 集成**

   * 有 policy 域模型和 repository，但看不到：

     * 属性驱动（用户属性、资源属性、环境上下文）的策略表达；
     * 与 OPA 或同类 Policy Engine 的交互；
     * 连续/会话内持续评估。

3. **智能多因素 / 风险评估缺失**

   * 已有 MFA manager (`pkg/auth/mfa/manager.go`) 和 `postres_mfa_repository.go`，TOTP provider 也在；
   * 但没有 risk engine、无 “基于风险的动态 MFA”，认证策略基本静态。

---

### 1.3 基础设施 & 运维治理 (L5 + L4)

**设计里：**

* PostgreSQL + Redis + Elasticsearch；
* 高可用部署、日志标准化、SIEM/SOC 集成、隐私/合规中心（GDPR、国密、脱敏、数据本地化）、可观测性三件套 (metrics/logs/traces)。

**当前实现：**

* `internal/storage/postgresql/*` + `internal/storage/redis/*`：Postgres & Redis 完整 repository 层；
* `internal/metrics/prometheus.go`：Prometheus 指标暴露；
* `internal/storage/postgresql/postgres_auditlog_repository.go` + `internal/services/audit/service.go`：有审计日志表与服务。

**GAP：**

1. **日志管线 / SIEM 集成缺失**

   * 目前审计日志写 DB；看不到标准化格式输出到 Syslog/Kafka 或对接 SIEM 的 pipeline。

2. **隐私/合规中心缺失**

   * 没有数据脱敏、Consent、“被遗忘权”、地区/租户隔离等逻辑的实现痕迹。

3. **Trace & 全链路观测缺**

   * 只见 metrics，无 OpenTelemetry / trace / log correlation。

4. **配置治理 & 多形态交付的 runtime hook**

   * 文档里写了 Docker/K8s/Helm/Standalone 等多形态交付，但服务内部对部署形态差异的配置抽象有限（大多在部署脚本里解决）。

---

## 2. 能力级 GAP 小结

从功能点来看，当前实现已覆盖：

* 多协议认证（OIDC / OAuth / SAML，RADIUS/LDAP Bind 等传统协议待确认）；
* 用户/身份/应用/策略 基础 CRUD；
* LDAP Connector + 基础同步服务；
* TOTP MFA 能力；
* Prometheus metrics；
* PostgreSQL + Redis 存储抽象。

而与架构目标相比，主要缺如下几个“能力簇”：

1. **“零信任授权服务”** 的全套（ABAC、持续验证、路径分析、OPA 集成）。
2. **“智能多因素认证”** 真正的“自适应 & 风险决策”。
3. **“安全运营中心 & 威胁检测与响应”**：异常行为检测 + 自动响应 + SIEM 集成。
4. **“数据隐私与合规中心”**：脱敏、合规 API、国密、数据本地化策略。
5. **“开发者中心 & 平台服务”**：统一 API 网关特性、插件/Connector 管理 API、可视化编排。
6. **“用户体验门户”**：自服务中心、品牌化登录页、多端体验。

下面的多 Phase 计划，就是围绕这些 GAP，用 **jules.google.com 环境可行** 的方式分阶段补齐。

> 假定 Jules 环境的主要约束：
>
> * 无 Docker/K8s、本地只跑 `go test ./...` / 单进程 `go run`；
> * 对外网络受限，不能依赖外部 Redis/Postgres 实例（测试中用 embedded / in-memory / sqlite / fake）；
> * 需要“一键” test & run（脚本和 Makefile 友好）；
> * 无长生命周期 daemon，只能通过 jules workflow/cron 触发定期任务。
>   后文每个 Phase 都会基于这些假设设计约束与测试方案。

---

# PHASE 1：语义对齐 & 核心分层收敛（P1）

> **(Phase 1: Semantic Alignment & Core Layer Hardening)**

* **Phase ID:** `P1`
* **Branch:** `feat/round2-phase1-semantic-layering`
* **Dependencies:** `round1 输出的基础 PoC（若无，视为无依赖）`

---

### 🎯 目标 (Objectives)

1. 建立 **“设计架构 ↔ 当前代码” 的显式映射层**（结构化语义），让后续所有 Phase 能直接使用；
2. 收敛 & 固化 **L3 应用服务层 / L4 领域层 / L5 存储层** 的接口边界，避免后续功能堆砌再造技术债；
3. 在 **Jules 环境** 下完成一套可以快速跑通的：

   * 单元测试覆盖 skeleton；
   * 轻量 In-Memory / sqlite 存储替身；
   * 基础 E2E “健康验证用例”（Login 成功 + LDAP 同步一次）。

---

### 📦 交付物与变更 (Deliverables & Changes)

**[Code Change]**

* **ADD**: `internal/architecture/map.go`

  * 提供 `type Capability string`、`type Layer string` 等枚举，和一组常量：`LayerPresentation`, `LayerGateway`, `LayerAppService`, `LayerDomain`, `LayerInfra`；
  * 定义 `type CapabilityMapping struct { Capability Capability; Layer Layer; Packages []string; Status string /* planned/partial/done */ }`；
  * 由人工维护一份 `var DefaultMappings []CapabilityMapping`，把架构中的关键能力（如 `Auth.MultiProtocol`, `Identity.Lifecycle`, `MFA.Adaptive` 等）映射到当前 Go package 名称。
* **ADD**: `internal/storage/memory/`

  * `identity_memory_repository.go`
  * `auth_memory_repository.go`
  * `policy_memory_repository.go`
  * 这些实现遵守现有 domain 层的 Repository 接口，基于 Go map/in-memory 实现，用于 Jules 环境测试（无需外部 DB）。
* **MODIFY**:

  * `internal/server/http/server.go`：

    * 增加对“memory backend” 模式的配置判断（例如 env `QID_STORAGE_MODE=memory` 时，wire memory repositories）。
  * `internal/domain/*/repository.go`：如有直接依赖 Postgres struct 的地方，抽象成 interface，确保 memory 实现可以 drop-in。

**[Config Change]**

* **ADD**: `configs/server.jules.yaml`

  * 仅使用 in-memory 存储；
  * 关闭对真正 Redis/Postgres 的依赖；
  * 简化 TLS、日志配置（stdout 即可）。
* **MODIFY**: `Makefile` / `scripts/test.sh`

  * 增加 `test-jules` 目标：`STORAGE_MODE=memory go test ./...`。

**[Doc Change]**

* **ADD**: `docs/round2/P1_semantic_alignment.md`

  * 说明架构能力列表、`CapabilityMapping` 的使用方式；
  * 列出当前 “设计 vs 实现 vs 状态(planned/partial/done)” 的表格；
* **MODIFY**: `docs/architecture.md`

  * 在末尾增加一节 “Code Mapping Overview”，指向上面的 map 文件和 P1 文档。

---

### 📝 关键任务 (Key Tasks)

* [ ] `P1-T1`: **[Capability Map]**

  * 新建 `internal/architecture/map.go`，定义 Layer/Capability 枚举 & `DefaultMappings`；
  * 至少覆盖以下 capability：

    * `Auth.MultiProtocol`（OIDC/OAuth/SAML）
    * `Auth.MFA.Basic`（TOTP）
    * `Identity.Lifecycle.Basic`（用户 CRUD）
    * `Connector.LDAP.Basic`
    * `Audit.Log.Basic`
    * `Metrics.Prometheus.Basic`
* [ ] `P1-T2`: **[Memory Storage]**

  * 实现 identity/auth/policy 的 memory repositories；
  * 确保所有现有 services 的依赖可以切到 memory，实现不改业务逻辑。
* [ ] `P1-T3`: **[Server Wiring for Jules]**

  * 在 `internal/server/http/server.go` 中注入一个 `InitWithConfig(cfg Config)`，根据 `cfg.Storage.Mode` 选择 memory 或 postgres；
  * Jules 环境使用 `configs/server.jules.yaml`。
* [ ] `P1-T4`: **[Tests]**

  * 为每个 memory repo 加基础单测；
  * 编写一个 E2E 测试（见下节），使用 memory backend 和假 LDAP 服务（fake interface）。
* [ ] `P1-T5`: **[Docs]**

  * 写 P1 doc，更新 architecture.md，说明“从 P1 起，所有能力评估依赖 `CapabilityMapping`”。

---

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计**

* **[Unit Tests]**

  * `tests/unit/identity_memory_repository_test.go`

    * 用例：Create / Get / Update / Delete / List 用户；验证线程安全性（可用 `t.Parallel` + RWMutex）。
  * `tests/unit/auth_memory_repository_test.go`

    * 用例：存取凭据/会话/refresh token 的基础 CRUD。
* **[Integration Tests]**

  * `tests/integration/server_jules_memory_test.go`

    * 使用 Go test 启动 HTTP server（使用 `configs/server.jules.yaml`），发起真实 HTTP 请求；
    * 替换 LDAP Connector 为内置 Fake（实现同样接口但数据在内存）。
* **[E2E Tests]**

  * `tests/e2e/jules_login_flow_test.go`：

    * Step 1：通过 API 注册一个用户；
    * Step 2：使用用户名/密码登录，获得 token；
    * Step 3：调用带授权保护的 API，验证 token 生效；
    * 所有存储后端使用 memory，无 DB 依赖。

**2. 效果验收 (Acceptance Criteria)**

* `AC-P1-1`: 所有 memory repos 单测覆盖率 ≥ 80%。
* `AC-P1-2`: `go test ./tests/integration -run TestServerJulesMemory` 在 Jules 环境可一键通过，无外部依赖。
* `AC-P1-3`: `docs/round2/P1_semantic_alignment.md` 中列出的 capability 状态与实际代码一致，并通过一次 peer review。
* `AC-P1-4`: 新的 `CapabilityMapping` 被至少一个后续 Phase 文档引用为依据（即已经成为“事实源”）。

**✅ 完成标准 (Definition of Done - DoD)**

* [ ] `P1` 关键任务全部勾选完成；
* [ ] 所有 `AC-P1-*` 通过；
* [ ] 分支 `feat/round2-phase1-semantic-layering` 已合并；
* [ ] Jules 环境下 `make test-jules` 成为标准“健康检查”。

---

# PHASE 2：零信任授权 & 策略引擎基础（P2）

> **(Phase 2: Zero-Trust Authorization & Policy Engine Foundation)**

* **Phase ID:** `P2`
* **Branch:** `feat/round2-phase2-zero-trust-policy`
* **Dependencies:** `P1`

---

### 🎯 目标 (Objectives)

1. 将当前零散的授权逻辑提升为 **统一策略引擎**，支持 RBAC + 初级 ABAC；
2. 为未来接入 OPA / 更复杂策略留出接口与数据模型；
3. 在 Jules 环境中通过 memory backend，就能完整跑通“登录 + 授权判定”的 E2E 流程。

---

### 📦 交付物与变更 (Deliverables & Changes)

**[Code Change]**

* **ADD**: `internal/domain/policy/model.go`（如未存在则扩展）

  * 增加：

    * `type Subject struct { UserID string; Groups []string; Attributes map[string]string }`
    * `type Resource struct { Type string; ID string; Attributes map[string]string }`
    * `type Action string`
    * `type Environment struct { IP string; Time time.Time; DeviceTrust string }`
    * `type EvaluationContext struct { Subject; Resource; Action; Environment }`
* **ADD**: `internal/services/authorization/evaluator.go`

  * 定义 `type Evaluator interface { Evaluate(ctx context.Context, evalCtx EvaluationContext) (Decision, error) }`；
  * 实现 `DefaultEvaluator`，支持：

    * 基于角色/用户的 allow/deny；
    * 简单属性条件（如 IP 白名单、工作时间段）。
* **MODIFY**:

  * `internal/services/authorization/service.go`

    * 将原有散落的权限判断统一委托给 `Evaluator`；
  * `internal/server/middleware/auth.go`

    * 在 JWT 验证后构建 `EvaluationContext` 并调用授权服务。
* **ADD (可选 demo)**: `internal/services/authorization/demo_opa_adapter.go`

  * 仅提供接口和假实现，用注释说明未来如何接入 OPA（Jules 环境不实际调用 OPA）。

**[Config Change]**

* **ADD**: `configs/policy/basic.yaml`

  * 支持配置：

    * 基于用户组/角色的策略；
    * 基于 `ip_whitelist`、`business_hours` 等简单环境条件；
  * 由 `DefaultEvaluator` 加载。

**[Doc Change]**

* **ADD**: `docs/round2/P2_zero_trust_policy.md`

  * 描述策略模型、EvaluationContext、配置样例；
  * 对应架构中“零信任授权服务”的对齐情况。

---

### 📝 关键任务 (Key Tasks)

* [ ] `P2-T1`: **[模型扩展]**

  * 扩展 policy domain 模型，加入 Subject/Resource/Environment 抽象；
* [ ] `P2-T2`: **[Evaluator 实现]**

  * 实现 `DefaultEvaluator`，支持：

    * RBAC：用户 / 组 / 角色 + action；
    * 简单 ABAC 条件：时间/IP/设备可信度（通过 context 中 attribute 填充）。
* [ ] `P2-T3`: **[Middleware 集成]**

  * 在 auth middleware 中，解析 JWT → 填充 Subject/Env；
  * 对受保护路由统一调用授权服务，不再在 handler 里写 `if user.Role != "admin"`.
* [ ] `P2-T4`: **[Jules 测试适配]**

  * 使用 P1 的 memory backend + policy.yaml，提供一组 E2E 授权测试。
* [ ] `P2-T5`: **[文档 & 示例策略]**

  * 在 P2 doc 里给出几套示例策略（Admin Dashboard、只允许公司网段访问等）。

---

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计**

* **[Unit Tests]**

  * `tests/unit/policy_evaluator_test.go`

    * 用例：

      * 用户在 admins 组，对 `dashboard:admin` 的 `read` → allow；
      * 用户在 employees 组，对 `dashboard:admin` → deny；
      * 用户满足 `ip in whitelist` + 业务时间段 → allow；否则 → deny。
* **[Integration Tests]**

  * `tests/integration/authz_middleware_test.go`

    * 使用 memory backend + policy.yaml，启动 HTTP server；
    * 场景：

      * 带有 admin JWT 调用 `/admin/...`，返回 200；
      * 带普通用户 JWT 调用同样 API，返回 403；
      * 模拟 IP header 改变，触发 deny。
* **[E2E Tests]**

  * `tests/e2e/jules_zero_trust_flow_test.go`：

    * Step 1：创建 admin 用户 + 普通用户；
    * Step 2：两者分别登录，获得 token；
    * Step 3：调用 admin-only API；
    * Step 4：验证返回 200 / 403；
    * 所有测试不依赖外部 DB，通过 memory backend 运行于 Jules。

**2. 效果验收**

* `AC-P2-1`: 所有授权决策都通过统一 Evaluator，不存在 handler 层“绕过策略引擎”的权限判断。
* `AC-P2-2`: policy.yaml 中修改策略后，无需改代码即可改变授权行为（单测验证）。
* `AC-P2-3`: Jules 环境下 E2E 授权测试一次性通过。
* `AC-P2-4`: P2 文档中把“零信任授权服务”需求逐条对照当前能力，说明哪些已实现、哪些留到后续 Phase。

**✅ 完成标准 (DoD)**

* [ ] 所有 `P2` 关键任务完成；
* [ ] 所有 `AC-P2-*` 满足；
* [ ] 分支 `feat/round2-phase2-zero-trust-policy` 合并；
* [ ] 架构文档中“零信任授权服务”对应 section 标记为 “基础版已落地”。

---

# PHASE 3：自适应多因素 & 风险引擎雏形（P3）

> **(Phase 3: Adaptive MFA & Risk Engine Bootstrap)**

* **Phase ID:** `P3`
* **Branch:** `feat/round2-phase3-adaptive-mfa`
* **Dependencies:** `P1`, `P2`

---

### 🎯 目标 (Objectives)

1. 在现有 TOTP MFA 能力之上，引入 **风险评分 + 自适应策略** 的基本实现；
2. 不追求 ML/大数据，先以规则引擎实现“风险分层 → 不同 MFA 策略”；
3. Jules 环境中可通过纯单测 / 集成测试完整验证风险决策逻辑。

---

### 📦 交付物与变更 (Deliverables & Changes)

**[Code Change]**

* **ADD**: `internal/domain/auth/risk_model.go`

  * `type RiskFactor string` （如 `GeoVelocity`, `NewDevice`, `UnusualTime`, `IPReputation`）；
  * `type RiskScore float64` + `type RiskAssessment struct { Score RiskScore; Factors []RiskFactor }`.
* **ADD**: `internal/services/auth/risk_engine.go`

  * `type RiskEngine interface { Assess(ctx context.Context, loginCtx LoginContext) (RiskAssessment, error) }`
  * `LoginContext` 包含：用户 ID、历史登录记录（通过 audit service 或 repository 提供）、当前 IP/UA/时间等；
  * `SimpleRiskEngine` 使用规则：

    * 首次登录新设备 → +0.3；
    * 异常时间（如本地深夜）→ +0.2；
    * 与上一次登录地理位置差距过大（如国家变化）→ +0.3。
* **MODIFY**:

  * `internal/services/auth/service.go` / `internal/orchestrator/workflows/auth_flow.go`：

    * 在密码验证后调用 `RiskEngine`，依据 Score 选择：

      * `< 0.3`：可选 MFA（如用户已绑定则触发，否则允许单因子）；
      * `0.3–0.7`：必须 TOTP MFA；
      * `>= 0.7`：拒绝/通知 + 必须强 MFA（当前仅有 TOTP，可统一走 TOTP）。
* **ADD**: `pkg/types/risk.go`（如需要在 handler/API 层传递信息）。

**[Config Change]**

* **ADD**: `configs/auth/risk_rules.yaml`

  * 规则参数化，如 “new_device_score: 0.3” 等，方便运营调整。

**[Doc Change]**

* **ADD**: `docs/round2/P3_adaptive_mfa.md`

  * 用流程图解释“密码 → 风险评估 → MFA 决策”；
  * 描述如何在 Jules 环境用测试数据模拟地理位移、时间段等。

---

### 📝 关键任务 (Key Tasks)

* [ ] `P3-T1`: **[Risk Model 定义]**

  * 实现 RiskAssessment / LoginContext 结构；
  * 确定至少 3–4 个规则因子。
* [ ] `P3-T2`: **[RiskEngine 实现 & 配置化]**

  * `SimpleRiskEngine` 从 `risk_rules.yaml` 中读取权重；
* [ ] `P3-T3`: **[认证流程集成]**

  * 在 login 流程中插入风险评估 & MFA 决策；
  * 确保所有 MFA 逻辑集中在 orchestrator/service 层，不散落在 handler。
* [ ] `P3-T4`: **[Jules 测试支持]**

  * 构造 fake 登录历史（通过 memory backend），在单测/集成测试中验证 risk 行为。
* [ ] `P3-T5`: **[文档 & 示例]**

  * 给出几个典型场景故事：内网白天 vs 外网凌晨；新设备 vs 老设备。

---

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计**

* **[Unit Tests]**

  * `tests/unit/risk_engine_test.go`：

    * TestCase 1：同一 IP/设备/白天 → Score < 0.3；
    * TestCase 2：新设备 + 夜间 → Score 介于 0.3–0.7；
    * TestCase 3：跨国访问 + 新设备 → Score > 0.7。
* **[Integration Tests]**

  * `tests/integration/adaptive_mfa_flow_test.go`：

    * 场景 1：低风险登录 → 允许单因子；
    * 场景 2：中风险登录 → 要求 TOTP；
    * 场景 3：高风险登录 → 返回错误码或要求管理员干预。
  * 实现方式：通过自定义 header 或 test-only 参数注入 “IP/UA/Time”等上下文。
* **[E2E Tests]**

  * `tests/e2e/jules_adaptive_mfa_test.go`：

    * 在 Jules 环境下使用 memory backend + 风险规则 config；
    * 全链路跑通 login API + MFA 验证。

**2. 效果验收**

* `AC-P3-1`: 风险评分逻辑完全被单测覆盖，且可通过修改 YAML 配置改变行为。
* `AC-P3-2`: 认证流程中所有 MFA 决策都可在日志中看到清晰“Score + Factors”记录。
* `AC-P3-3`: Jules 环境中 `go test ./tests/e2e -run AdaptiveMFA` 成功执行。
* `AC-P3-4`: 架构文档中“智能多因素认证”段落可以标记为“基础版已实现（规则驱动）”。

**✅ 完成标准 (DoD)**

* [ ] 所有 `P3` 关键任务完成；
* [ ] 所有 `AC-P3-*` 通过；
* [ ] 分支 `feat/round2-phase3-adaptive-mfa` 合并；
* [ ] RiskEngine 成为后续威胁检测与响应的输入基础。

---

# PHASE 4：审计 & 可观测性 & 安全运营基础（P4）

> **(Phase 4: Audit, Observability & Security Operations Foundation)**

* **Phase ID:** `P4`
* **Branch:** `feat/round2-phase4-audit-observability`
* **Dependencies:** `P1`, `P2`, `P3`

---

### 🎯 目标 (Objectives)

1. 将当前散落的日志和 audit 表升级为 **统一审计流水线**；
2. 提供最小可用的“安全运营视角”：

   * 标准化事件格式；
   * 可导出到外部 SIEM（在 Jules 中通过 file sink 模拟）；
3. 补齐基础可观测性：metrics + structured logs + trace id。

---

### 📦 交付物与变更 (Deliverables & Changes)

**[Code Change]**

* **ADD**: `internal/audit/event.go`

  * 定义通用事件模型 `AuditEvent`：

    ```go
    type AuditEvent struct {
        ID        string
        Timestamp time.Time
        Category  string // auth, policy, admin, mfa, risk
        Action    string // login_success, login_failed, policy_evaluated, ...
        UserID    string
        IP        string
        Resource  string
        Result    string // success/fail/deny
        TraceID   string
        Details   map[string]any
    }
    ```
* **ADD**: `internal/audit/pipeline.go`

  * 抽象 `Sink` 接口：DBSink、FileSink、StdoutSink（Jules 环境主要用 File/Stdout）；
  * `Pipeline` 实现 fan-out 到多个 sink。
* **MODIFY**:

  * `internal/services/audit/service.go` 改为使用 Pipeline；
  * 登录/授权/MFA/RiskEngine 等调用 audit service 记录标准化事件。
* **ADD**: `internal/metrics/http_middleware.go`

  * 提供 HTTP-level metrics（latency, status code, route）；
  * 结合已有 `internal/metrics/prometheus.go`。

**[Config Change]**

* **ADD**: `configs/audit/pipeline.jules.yaml`

  * 只启用 FileSink + StdoutSink；
* **ADD**: `configs/audit/pipeline.prod.example.yaml`

  * 示意如何启用 KafkaSink / SyslogSink（仅配置示例，不要求在 Jules 环境运行）。

**[Doc Change]**

* **ADD**: `docs/round2/P4_audit_observability.md`

  * 定义事件分类、字段含义；
  * 示例：如何把 FileSink 输出对接到 SIEM（在 Jules 中用 `tail -f` + 简单脚本模拟）。

---

### 📝 关键任务 (Key Tasks)

* [ ] `P4-T1`: **[事件模型 & Pipeline]**

  * 实现 `AuditEvent` + `Sink` + `Pipeline`；
  * 确保新增 sink 容易。
* [ ] `P4-T2`: **[业务埋点]**

  * 在 login 成功/失败、策略评估结果、高风险登录、管理员操作等场景调用 audit；
  * 风险评估中的 Score / Factors 要写入 Details。
* [ ] `P4-T3`: **[Metrics 中间件]**

  * 增加 HTTP metrics（请求量、延迟、status code 分布）；
  * 结合 Prometheus endpoint。
* [ ] `P4-T4`: **[Jules 环境适配]**

  * FileSink 输出路径放在工作目录下，如 `./logs/audit_jules.log`；
  * 确保在 Jules 上不需要任何外部 agent。
* [ ] `P4-T5`: **[文档 & 使用手册]**

  * 指导安全同学如何基于 audit.log 做简单的事件搜索与导出。

---

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计**

* **[Unit Tests]**

  * `tests/unit/audit_pipeline_test.go`：

    * 验证 fan-out 正确，错误 sink 不影响其他 sink（可记录 error metric）。
* **[Integration Tests]**

  * `tests/integration/audit_http_flow_test.go`：

    * 模拟 login 成功/失败、admin 操作；
    * 检查 FileSink 输出的 JSON 行是否符合 schema。
* **[E2E Tests]**

  * `tests/e2e/jules_audit_observability_test.go`：

    * 在 Jules 环境中跑 server + 触发一系列操作；
    * 确认：

      * `/metrics` 暴露 HTTP 指标；
      * audit 日志文件存在且含有 expected 事件。

**2. 效果验收**

* `AC-P4-1`: 所有安全相关动作（认证/授权/MFA/风险）都有对应 `AuditEvent`。
* `AC-P4-2`: Jules 环境中可通过简单脚本（例如 `jq`）对 audit.log 做条件查询。
* `AC-P4-3`: Prometheus 指标中能看到 HTTP 请求与授权决策相关 metrics。
* `AC-P4-4`: 架构文档中“安全运营中心”部分可标记为“基础审计与观测已具备”。

**✅ 完成标准 (DoD)**

* [ ] `P4` 关键任务完成；
* [ ] 所有 `AC-P4-*` 满足；
* [ ] 分支 `feat/round2-phase4-audit-observability` 合并；
* [ ] 安全部门可基于 audit.log 做最小可用的事件排查。

---

# PHASE 5：平台服务 & 开发者中心最小版（P5）

> **(Phase 5: Minimal Platform Services & Developer Center)**

* **Phase ID:** `P5`
* **Branch:** `feat/round2-phase5-devcenter-platform`
* **Dependencies:** `P1`, `P2`, `P3`, `P4`

---

### 🎯 目标 (Objectives)

1. 为未来的平台化扩展打基础：

   * 提供一组 **平台级 API** 用于管理应用、Connector、MFA Provider 等；
   * 提供“开发者中心”的最小 REST 接口；
2. 仍以 Jules 环境可运行为硬约束，前端 UI 可暂时缺位，用 API + 文档替代。

---

### 📦 交付物与变更 (Deliverables & Changes)

**[Code Change]**

* **ADD**: `internal/services/platform/devcenter_service.go`

  * 提供：

    * 列出 / 注册 / 更新 应用（OIDC/SAML）；
    * 列出 / 启用 / 禁用 Connector（如 LDAP）；
    * 查看策略 / MFA 配置的汇总视图。
* **ADD**: `internal/server/http/handlers/devcenter.go`

  * REST API：

    * `GET /api/devcenter/apps`
    * `POST /api/devcenter/apps`
    * `GET /api/devcenter/connectors`
    * `POST /api/devcenter/connectors/{id}/enable`
    * `GET /api/devcenter/diagnostics`（聚合部分 metrics / version / 配置信息）
* **ADD**: `pkg/types/devcenter.go`

  * DTO 定义，用于统一 API 输入/输出。

**[Config Change]**

* **ADD**: `configs/devcenter/jules.yaml`

  * 控制哪些操作在 Jules 环境下可用（例如禁止真正写入生产 Connector 配置，只允许 mock）。

**[Doc Change]**

* **ADD**: `docs/round2/P5_devcenter_api.md`

  * Swagger/OpenAPI 摘要或手写 API 文档；
  * 使用 curl / httpie 示例演示如何管理一个 OIDC Client 与 LDAP Connector。

---

### 📝 关键任务 (Key Tasks)

* [ ] `P5-T1`: **[DevCenter Service]**

  * 实现 devcenter_service，内部复用已有 identity/auth/application/policy 服务；
* [ ] `P5-T2`: **[HTTP Handler & Routing]**

  * 在 `admin_routes.go` 中增加 `/api/devcenter/*` 路由并挂接 handler；
* [ ] `P5-T3`: **[权限控制]**

  * 使用 P2 的策略引擎，**只允许管理员** 调用 devcenter API；
* [ ] `P5-T4`: **[Jules 环境兼容]**

  * 通过配置使得 devcenter API 在 Jules 下仍然可用（使用 memory backend）；
* [ ] `P5-T5`: **[API 文档 & 示例]**

  * 补充文档、写一份“5 分钟体验 devcenter API”的教程。

---

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计**

* **[Unit Tests]**

  * `tests/unit/devcenter_service_test.go`：

    * 使用 fake repositories，测试 app/connector 的增删改查逻辑。
* **[Integration Tests]**

  * `tests/integration/devcenter_api_test.go`：

    * 以管理员身份调用 devcenter API，验证：

      * 能列出应用、创建新 OIDC 客户端；
      * 能列出/启用/禁用 LDAP Connector（在 Jules 下只是标记字段，但不真正连外部 LDAP）。
* **[E2E Tests]**

  * `tests/e2e/jules_devcenter_smoke_test.go`：

    * 在 Jules 环境中启动 server，顺序调用一套 devcenter API 流程，最后验证配置生效（如新创建 app 可用于登录）。

**2. 效果验收**

* `AC-P5-1`: DevCenter API 完整覆盖“应用 + Connector + 基础诊断”视角。
* `AC-P5-2`: Jules 环境下可以仅通过 API & Docs 完成一个“集成新 OIDC 应用”的 end-to-end 流程。
* `AC-P5-3`: 所有 devcenter API 都经过 P2 策略引擎授权控制，普通用户无法访问。
* `AC-P5-4`: 架构文档中“开发者中心（Developer Center）”字段可标记为“API 版已落地”。

**✅ 完成标准 (DoD)**

* [ ] 所有 `P5` 关键任务完成；
* [ ] 所有 `AC-P5-*` 满足；
* [ ] 分支 `feat/round2-phase5-devcenter-platform` 合并；
* [ ] 平台能力从“仅核心认证引擎”升级为“带平台服务入口”的形态。

---

## 总结

* 以上五个 Phase 已经覆盖了 **设计中的关键能力簇** 和 **当前实现的主要 GAP**，并且每个 Phase 都能在 Jules 环境中用 memory backend + Go test 跑通；
* 你可以从 `P1` 开始逐个 Phase 执行，每个 Phase 都是信息全景化的任务包，不需要来回查其它文档就能上手开发；
* 如果你希望，我可以在后续轮次里针对某个 Phase（比如 P2 策略引擎或 P3 RiskEngine）进一步细化到「函数级 TODO 列表」或给出关键 demo 代码片段。
