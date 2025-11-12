## PHASE 7: 高级功能扩展与 Round 1 收尾

> **(Phase 7: Advanced Features & Round 1 Finalization)**

* **Phase ID:** `P7`
* **Branch:** `feat/round1-phase7-advanced`
* **Dependencies:** `P1`, `P2`, `P3`, `P4`, `P5`, `P6`（需要稳定的生产环境）

### 🎯 目标 (Objectives)

* 实现 OAuth 2.0 高级流程（PKCE、Device Flow）
* 实现 OpenID Connect (OIDC) 协议支持
* 实现 SSO 单点登录（企业集成）
* 实现用户自助服务（密码重置、账号恢复）
* 实现 API 限流和防滥用机制
* 完成 Round 1 的文档、测试和发布

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `ADD`: `internal/oauth/pkce.go` - PKCE 流程实现
  * `ADD`: `internal/oauth/device_flow.go` - Device Flow 实现
  * `ADD`: `internal/oidc/provider.go` - OIDC Provider 实现
  * `ADD`: `internal/sso/saml_handler.go` - SAML SSO 集成
  * `ADD`: `internal/ratelimit/token_bucket.go` - Token Bucket 限流算法
  * `ADD`: `web/self-service/` - 用户自助服务前端
  * `MODIFY`: `internal/oauth/authorization_handler.go` - 支持 PKCE 和 Device Flow

* **[API Change]** (API 变更):

  * `ADD`: `POST /oauth/device/authorize` - Device Flow 授权端点
  * `ADD`: `POST /oauth/device/token` - Device Flow Token 端点
  * `ADD`: `GET /.well-known/openid-configuration` - OIDC Discovery 端点
  * `ADD`: `GET /oidc/userinfo` - OIDC UserInfo 端点
  * `ADD`: `POST /self-service/password-reset` - 密码重置请求
  * `ADD`: `POST /self-service/account-recovery` - 账号恢复

* **[Documentation]** (文档):

  * `ADD`: `docs/oauth/pkce-guide.md` - PKCE 使用指南
  * `ADD`: `docs/oidc/integration-guide.md` - OIDC 集成指南
  * `ADD`: `docs/sso/enterprise-sso.md` - 企业 SSO 集成文档
  * `ADD`: `docs/api/rate-limiting.md` - API 限流策略说明
  * `ADD`: `CHANGELOG.md` - Round 1 完整变更日志
  * `ADD`: `README.md` - 项目概述、快速开始、架构图

### 📝 关键任务 (Key Tasks)

* [ ] `P7-T1`: **[Implement]** 实现 PKCE 支持 (`internal/oauth/pkce.go`)

  * 支持 Authorization Code Flow with PKCE（RFC 7636）
  * 客户端在授权请求时提供 `code_challenge` 和 `code_challenge_method`
  * Token 请求时验证 `code_verifier` 是否匹配
  * 强制公共客户端（移动应用、SPA）使用 PKCE
  * 示例流程：

    ```
    1. 客户端生成随机 code_verifier（43-128 字符）
    2. 计算 code_challenge = BASE64URL(SHA256(code_verifier))
    3. 授权请求：GET /oauth/authorize?code_challenge=xxx&code_challenge_method=S256
    4. Token 请求：POST /oauth/token (body: code_verifier=xxx)
    5. 服务端验证：SHA256(code_verifier) == code_challenge
    ```

* [ ] `P7-T2`: **[Implement]** 实现 Device Flow (`internal/oauth/device_flow.go`)

  * 支持设备授权流程（RFC 8628）- 用于智能电视、IoT 设备
  * 实现端点：

    * `POST /oauth/device/authorize` - 返回 `device_code` 和 `user_code`
    * `POST /oauth/device/token` - 轮询 Token（使用 `device_code`）
  * 流程：

    ```
    1. 设备请求授权：POST /oauth/device/authorize
       响应：{ "device_code": "xxx", "user_code": "ABCD-1234", "verification_uri": "https://quantaid.com/activate" }
    2. 用户在浏览器中访问 verification_uri，输入 user_code
    3. 设备轮询 Token：POST /oauth/device/token (interval: 5 秒)
       - 用户未授权：返回 "authorization_pending"
       - 用户已授权：返回 Access Token
    ```
  * 防止暴力破解：`user_code` 长度至少 8 位，支持大小写字母和数字

* [ ] `P7-T3`: **[Implement]** 实现 OIDC Provider (`internal/oidc/provider.go`)

  * 实现 OIDC Discovery 端点（`/.well-known/openid-configuration`）：

    ```json
    {
      "issuer": "https://auth.quantaid.com",
      "authorization_endpoint": "https://auth.quantaid.com/oauth/authorize",
      "token_endpoint": "https://auth.quantaid.com/oauth/token",
      "userinfo_endpoint": "https://auth.quantaid.com/oidc/userinfo",
      "jwks_uri": "https://auth.quantaid.com/oidc/jwks",
      "response_types_supported": ["code", "token", "id_token"],
      "scopes_supported": ["openid", "profile", "email"],
      "claims_supported": ["sub", "name", "email", "email_verified"]
    }
    ```
  * 实现 UserInfo 端点（`GET /oidc/userinfo`）：

    * 验证 Access Token
    * 返回用户信息（根据 scope）
  * 签发 ID Token（JWT 格式）：

    ```json
    {
      "iss": "https://auth.quantaid.com",
      "sub": "user-123",
      "aud": "client-456",
      "exp": 1700000000,
      "iat": 1699996400,
      "name": "John Doe",
      "email": "john@example.com"
    }
    ```

* [ ] `P7-T4`: **[Implement]** 实现 SAML SSO 支持 (`internal/sso/saml_handler.go`)

  * 支持 SAML 2.0 Service Provider (SP) 角色
  * 实现端点：

    * `POST /sso/saml/acs` - Assertion Consumer Service（接收 Identity Provider 的断言）
    * `GET /sso/saml/metadata` - 导出 SP Metadata XML
  * 集成企业 Identity Provider（如 Okta、Azure AD、Google Workspace）
  * 支持属性映射（SAML Attribute → 本地用户字段）
  * 示例配置：

    ```yaml
    saml:
      entity_id: "https://auth.quantaid.com/sso/saml/metadata"
      acs_url: "https://auth.quantaid.com/sso/saml/acs"
      idp_metadata_url: "https://idp.example.com/metadata"
      attribute_mapping:
        email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
        name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
    ```

* [ ] `P7-T5`: **[Implement]** 实现用户自助服务 - 密码重置 (`internal/handlers/password_reset.go`)

  * 流程：

    1. 用户请求密码重置：`POST /self-service/password-reset` (body: `{ "email": "user@example.com" }`)
    2. 系统发送重置邮件（包含 Token，有效期 1 小时）
    3. 用户点击邮件中的链接：`GET /self-service/reset-password?token=xxx`
    4. 用户设置新密码：`POST /self-service/reset-password` (body: `{ "token": "xxx", "new_password": "xxx" }`)
  * 安全措施：

    * Token 一次性使用（使用后立即失效）
    * Token 绑定 IP 地址（可选，防止 Token 泄露）
    * 限制重置频率（同一邮箱 10 分钟内只能请求一次）

* [ ] `P7-T6`: **[Implement]** 实现账号恢复机制 (`internal/handlers/account_recovery.go`)

  * 支持多种恢复方式：

    * 邮箱验证码（发送 6 位数字验证码）
    * 备用邮箱（设置时要求用户提供备用邮箱）
    * 安全问题（设置 3 个安全问题，恢复时需回答至少 2 个）
  * 恢复流程：

    1. 用户请求恢复：`POST /self-service/account-recovery` (body: `{ "email": "user@example.com", "method": "email" }`)
    2. 系统验证用户身份（发送验证码或显示安全问题）
    3. 用户提交验证信息：`POST /self-service/verify-recovery` (body: `{ "token": "xxx", "code": "123456" }`)
    4. 验证通过后，允许用户重置密码或恢复账号

* [ ] `P7-T7`: **[Implement]** 实现 API 限流 (`internal/ratelimit/token_bucket.go`)

  * 使用 Token Bucket 算法（基于 Redis 实现）
  * 限流策略：

    * 登录端点：5 次/分钟（同一 IP）
    * Token 端点：10 次/分钟（同一 Client ID）
    * UserInfo 端点：100 次/分钟（同一 Access Token）
    * 管理 API：1000 次/小时（同一管理员）
  * 超过限制时返回 `429 Too Many Requests`，并在响应头中包含：

    ```
    X-RateLimit-Limit: 5
    X-RateLimit-Remaining: 0
    X-RateLimit-Reset: 1699996400
    Retry-After: 60
    ```
  * 示例代码：

    ```go
    type TokenBucket struct {
        redis *redis.Client
    }

    func (tb *TokenBucket) Allow(key string, limit int, window time.Duration) bool {
        now := time.Now().Unix()
        bucketKey := fmt.Sprintf("ratelimit:%s", key)
        
        // 使用 Redis Lua 脚本实现原子操作
        script := `
            local key = KEYS[1]
            local limit = tonumber(ARGV[1])
            local window = tonumber(ARGV[2])
            local now = tonumber(ARGV[3])
            
            local current = redis.call('GET', key)
            if current == false then
                redis.call('SET', key, 1, 'EX', window)
                return 1
            end
            
            current = tonumber(current)
            if current < limit then
                redis.call('INCR', key)
                return 1
            end
            
            return 0
        `
        
        result, err := tb.redis.Eval(context.Background(), script, []string{bucketKey}, limit, window.Seconds(), now).Int()
        return err == nil && result == 1
    }
    ```

* [ ] `P7-T8`: **[Documentation]** 完善项目文档

  * 创建 `README.md`（项目首页）：

    * 项目简介（一句话描述）
    * 核心功能列表
    * 架构图（使用 Mermaid 或 PlantUML）
    * 快速开始（Docker Compose 一键部署）
    * 链接到详细文档
  * 创建 `CHANGELOG.md`（变更日志）：

    * 按版本组织（v0.1.0、v0.2.0...）
    * 每个版本包含：新功能、Bug 修复、破坏性变更
  * 创建 `CONTRIBUTING.md`（贡献指南）：

    * 代码风格规范
    * Commit 消息规范（Conventional Commits）
    * Pull Request 流程
  * 创建 `docs/architecture/system-design.md`（系统设计文档）：

    * 整体架构图
    * 数据流图
    * 技术选型说明
    * 安全设计考虑

* [ ] `P7-T9`: **[Test]** 完成 Round 1 的端到端测试

  * 创建完整的 E2E 测试套件（覆盖所有 Phase 的功能）
  * 测试场景：

    * `E2E::TestCompleteAuthFlow` - 完整的认证流程（注册 → 登录 → MFA → Token → UserInfo）
    * `E2E::TestOAuthCodeFlowWithPKCE` - PKCE 流程
    * `E2E::TestDeviceFlow` - Device Flow
    * `E2E::TestSAMLSSO` - SAML SSO 集成
    * `E2E::TestPasswordReset` - 密码重置
    * `E2E::TestRateLimiting` - 限流机制
    * `E2E::TestAdminConsoleAllFeatures` - 管理控制台所有功能
  * 使用 Playwright 录制测试（生成测试脚本）

* [ ] `P7-T10`: **[Release]** 发布 Round 1 版本

  * 版本号：`v1.0.0-round1`
  * 创建 GitHub Release：

    * Release Notes（总结所有 Phase 的功能）
    * 二进制文件（Linux、macOS、Windows）
    * Docker 镜像（推送到 Docker Hub）
  * 更新文档网站（使用 MkDocs 或 Docusaurus）
  * 发布公告（技术博客、社交媒体）

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[OIDC Compliance Test]** (OIDC 合规性测试):

  * `Test Case 1`: `oidc::TestDiscoveryEndpoint` - 验证 Discovery 端点返回正确的配置
  * `Test Case 2`: `oidc::TestIDTokenSignature` - 验证 ID Token 签名有效（使用 JWKS 公钥）
  * `Test Case 3`: `oidc::TestUserInfoEndpoint` - 验证 UserInfo 端点返回正确的用户信息

* **[Device Flow Test]** (设备流程测试):

  * `Test Case 4`: `deviceflow::TestUserCodeGeneration` - 验证 User Code 格式正确（大写字母+数字，易读）
  * `Test Case 5`: `deviceflow::TestPollingInterval` - 验证设备轮询间隔正确（5 秒）
  * `Test Case 6`: `deviceflow::TestUserCodeExpiration` - 验证 User Code 过期后无法使用（15 分钟）

* **[Rate Limiting Test]** (限流测试):

  * `Test Case 7`: `ratelimit::TestLoginEndpointLimit` - 同一 IP 连续请求 10 次登录，第 6 次应返回 429
  * `Test Case 8`: `ratelimit::TestResetAfterWindow` - 等待限流窗口结束后，验证可以继续请求

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (功能完整性) 所有 Phase 1-7 的功能均已实现并通过测试
* `AC-2`: (OIDC 合规性) 通过 OpenID Foundation 的 Conformance Test（如 OP-Basic）
* `AC-3`: (安全性) 所有敏感端点启用限流，防止暴力破解
* `AC-4`: (文档完整性) 所有功能都有对应的文档和示例代码
* `AC-5`: (测试覆盖率) 单元测试覆盖率 > 80%，E2E 测试覆盖所有核心流程
* `AC-6`: (生产就绪) 系统在生产环境稳定运行 30 天，无 P0/P1 级别故障

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P7` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] OIDC 合规性测试通过
* [ ] 完整的 E2E 测试套件已创建并通过
* [ ] 所有文档已完成并发布到文档网站
* [ ] GitHub Release `v1.0.0-round1` 已发布
* [ ] Docker 镜像已推送到 Docker Hub
* [ ] 代码已合并到 `main` 分支，Tag `v1.0.0-round1`

---

