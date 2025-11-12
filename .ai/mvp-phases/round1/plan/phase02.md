## PHASE 2: OAuth 2.1 与 OIDC 协议完整实现

> **(Phase 2: OAuth 2.1 & OIDC Protocol Implementation)**

* **Phase ID:** `P2`
* **Branch:** `feat/round1-phase2-oauth-oidc`
* **Dependencies:** `P1`（需要 Redis 会话管理）

### 🎯 目标 (Objectives)

* 实现完整的 OAuth 2.1 授权码流程（Authorization Code Flow with PKCE）
* 实现 OpenID Connect 1.0 核心协议
* 提供 Token 端点、UserInfo 端点、JWKS 端点
* 实现 Client Credentials Grant（机器对机器认证）

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `MODIFY`: `pkg/auth/protocols/oauth.go` - 从空实现改为完整实现
  * `MODIFY`: `pkg/auth/protocols/oidc.go` - 从空实现改为完整实现
  * `ADD`: `internal/server/http/handlers/oauth.go` - OAuth 端点处理器
  * `ADD`: `internal/server/http/handlers/oidc.go` - OIDC 端点处理器
  * `ADD`: `internal/domain/auth/pkce.go` - PKCE 验证逻辑
  * `ADD`: `pkg/types/oauth.go` - OAuth 请求/响应类型
  * `ADD`: `tests/e2e/oauth_flow_test.go` - OAuth 完整流程 E2E 测试

* **[Dependency Change]** (依赖变更)

  * `ADD`: `github.com/golang-jwt/jwt/v5` - 升级 JWT 库版本
  * `ADD`: `gopkg.in/square/go-jose.v2` - JWKS 生成

* **[API Change]** (API 变更)

  * `ADD`: `POST /oauth/authorize` - OAuth 授权端点
  * `ADD`: `POST /oauth/token` - Token 交换端点
  * `ADD`: `GET /oauth/userinfo` - OIDC UserInfo 端点
  * `ADD`: `GET /.well-known/openid-configuration` - OIDC Discovery 端点
  * `ADD`: `GET /.well-known/jwks.json` - JWKS 公钥端点

### 📝 关键任务 (Key Tasks)

* [ ] `P2-T1`: **[Implement]** 实现 `pkg/auth/protocols/oauth.go::HandleAuthRequest()`

  * 支持 `response_type=code`（授权码模式）
  * 验证 `client_id`、`redirect_uri`、`scope`
  * 验证 PKCE 参数（`code_challenge`、`code_challenge_method`）
  * 生成并存储授权码（存入 Redis，TTL 10 分钟）

* [ ] `P2-T2`: **[Implement]** 实现 `pkg/auth/protocols/oauth.go::handleAuthorizationCode()`

  * Token 端点处理授权码交换
  * 验证 `code_verifier` 与 `code_challenge` 匹配（PKCE 验证）
  * 生成 Access Token（JWT，有效期 1 小时）和 Refresh Token（随机字符串，有效期 7 天）
  * 撤销已使用的授权码

* [ ] `P2-T3`: **[Implement]** 实现 `pkg/auth/protocols/oauth.go::handleClientCredentials()`

  * 验证 `client_id` 和 `client_secret`
  * 生成 Access Token（不含 Refresh Token）

* [ ] `P2-T4`: **[Implement]** 实现 `pkg/auth/protocols/oidc.go::generateIDToken()`

  * 生成符合 OIDC 规范的 ID Token（包含 `sub`, `aud`, `iss`, `exp`, `iat`, `nonce`）
  * 使用 RS256 算法签名（私钥来自配置）

* [ ] `P2-T5`: **[Implement]** 实现 `pkg/auth/protocols/oidc.go::GetUserInfo()`

  * 从 Access Token 中提取 `sub`
  * 从身份服务查询用户信息
  * 返回符合 OIDC 标准的 UserInfo JSON（`sub`, `name`, `email`, `email_verified`）

* [ ] `P2-T6`: **[Implement]** 创建 `internal/server/http/handlers/oidc.go`

  * 实现 `GET /.well-known/openid-configuration` - 返回 OIDC Discovery 元数据
  * 实现 `GET /.well-known/jwks.json` - 返回 RSA 公钥的 JWK Set

* [ ] `P2-T7`: **[Test Design]** 创建 `tests/e2e/oauth_flow_test.go`

  * 测试完整的授权码流程：

    1. 客户端发起授权请求（带 PKCE）
    2. 用户登录并授权
    3. 获取授权码
    4. 交换 Access Token 和 ID Token
    5. 使用 Access Token 访问 UserInfo 端点
    6. 使用 Refresh Token 刷新 Access Token

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[Unit Test]** (单元测试):

  * `Test Case 1`: `pkg/auth/protocols/oauth_test.go::TestPKCEValidation` - 验证 PKCE 挑战验证逻辑
  * `Test Case 2`: `pkg/auth/protocols/oidc_test.go::TestIDTokenGeneration` - 验证 ID Token 包含正确的 Claims

* **[E2E Test]** (端到端测试 - 对应 `P2-T7`):

  * `Test Case 3`: `tests/e2e/oauth_flow_test.go::TestOAuthAuthorizationCodeFlow` - 完整授权码流程
  * `Test Case 4`: `tests/e2e/oauth_flow_test.go::TestOAuthClientCredentialsFlow` - 客户端凭证流程
  * `Test Case 5`: `tests/e2e/oauth_flow_test.go::TestOAuthTokenRefresh` - Refresh Token 刷新流程

* **[Security Test]** (安全测试):

  * `Test Case 6`: 尝试重放已使用的授权码，验证是否被拒绝
  * `Test Case 7`: 使用错误的 `code_verifier`，验证 PKCE 验证是否失败
  * `Test Case 8`: 使用过期的 Access Token 访问 UserInfo，验证返回 401

* **[Compliance Test]** (合规性测试):

  * `Test Case 9`: 使用 `oidc-client-ts` 库进行集成测试，验证符合 OIDC 规范
  * `Test Case 10`: 验证 OIDC Discovery 文档包含所有必需字段（`issuer`, `authorization_endpoint`, `token_endpoint`, `jwks_uri`）

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (功能完整性) `Test Case 3`, `Test Case 4`, `Test Case 5` 100% 通过
* `AC-2`: (安全性) `Test Case 6`, `Test Case 7`, `Test Case 8` 全部通过，攻击被正确阻断
* `AC-3`: (合规性) `Test Case 9` 与标准 OIDC 客户端库兼容
* `AC-4`: (性能) 单次 Token 交换耗时 < 100ms（P95）
* `AC-5`: (文档) 新增 API 文档 `docs/apis/oauth2-oidc.md`，包含完整的请求/响应示例

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P2` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] 新增 API 文档已提交并通过评审
* [ ] E2E 测试在 CI 中每次提交自动运行
* [ ] 代码已合并到 `main` 分支，Tag `v0.3.0-phase2`

### 🔧 开发指南与约束 (Development Guidelines & Constraints)

**关键实现思路（Demo Code）：**

**示例 1：PKCE 验证** (`internal/domain/auth/pkce.go`)

```go
package auth

import (
    "crypto/sha256"
    "encoding/base64"
)

func VerifyPKCE(codeVerifier, codeChallenge, method string) bool {
    if method != "S256" {
        return false // 仅支持 SHA256
    }
    
    hash := sha256.Sum256([]byte(codeVerifier))
    computed := base64.RawURLEncoding.EncodeToString(hash[:])
    
    return computed == codeChallenge
}
```

**示例 2：授权码生成** (`pkg/auth/protocols/oauth.go` 部分)

```go
func (o *OAuthAdapter) HandleAuthRequest(ctx context.Context, req *types.AuthRequest) (*types.AuthResponse, error) {
    // 1. 验证 client_id 和 redirect_uri
    app, err := o.appRepo.GetApplicationByClientID(ctx, req.ClientID)
    if err != nil || !contains(app.RedirectURIs, req.RedirectURI) {
        return nil, types.NewError(types.ErrCodeInvalidRequest, "invalid client or redirect_uri")
    }
    
    // 2. 生成授权码
    code := generateRandomCode(32) // 实现随机字符串生成
    
    // 3. 存储授权码到 Redis（关联 PKCE challenge）
    authCodeData := map[string]interface{}{
        "user_id":         req.UserID,
        "client_id":       req.ClientID,
        "redirect_uri":    req.RedirectURI,
        "code_challenge":  req.Params["code_challenge"],
        "challenge_method": req.Params["code_challenge_method"],
        "scope":           req.Scope,
    }
    
    if err := o.redis.SetWithExpiry(ctx, "authcode:"+code, authCodeData, 10*time.Minute); err != nil {
        return nil, err
    }
    
    return &types.AuthResponse{
        Code:        code,
        RedirectURI: req.RedirectURI,
    }, nil
}
```

**测试约束：**

* E2E 测试必须模拟完整的浏览器重定向流程（可使用 `httptest`）
* 所有 JWT Token 必须在测试中验证签名有效性
* 测试密钥对使用固定的测试密钥（不使用生产密钥）

---