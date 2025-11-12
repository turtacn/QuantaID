## PHASE 4: 多因素认证 (MFA) 插件实现

> **(Phase 4: Multi-Factor Authentication (MFA) Plugin Implementation)**

* **Phase ID:** `P4`
* **Branch:** `feat/round1-phase4-mfa`
* **Dependencies:** `P1`, `P2`（需要 OAuth 流程支持）

### 🎯 目标 (Objectives)

* 实现 TOTP（Time-based One-Time Password）二次认证
* 实现 SMS OTP（短信验证码）二次认证
* 实现邮箱 OTP 二次认证
* 支持 MFA 策略配置（强制启用、可选启用）
* 提供 MFA 备用恢复码功能

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `ADD`: `pkg/plugins/mfa/totp/totp_provider.go` - TOTP 认证提供者
  * `ADD`: `pkg/plugins/mfa/sms/sms_provider.go` - SMS OTP 提供者（集成阿里云短信）
  * `ADD`: `pkg/plugins/mfa/email/email_provider.go` - 邮箱 OTP 提供者
  * `ADD`: `internal/domain/auth/mfa_policy.go` - MFA 策略引擎
  * `ADD`: `internal/storage/postgres/mfa_repository.go` - MFA 配置存储
  * `ADD`: `tests/integration/mfa_flow_test.go` - MFA 完整流程测试

* **[Dependency Change]** (依赖变更)

  * `ADD`: `github.com/pquerna/otp` - TOTP 算法实现
  * `ADD`: `github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi` - 阿里云短信 SDK

* **[Database Change]** (数据库变更)

  * `ADD`: 表 `user_mfa_configs` - 用户 MFA 配置（TOTP 密钥、备用码等）
  * `ADD`: 表 `mfa_verification_logs` - MFA 验证日志

* **[API Change]** (API 变更)

  * `ADD`: `POST /api/v1/users/me/mfa/totp/setup` - 初始化 TOTP 设置
  * `ADD`: `POST /api/v1/users/me/mfa/totp/verify` - 验证 TOTP 代码
  * `ADD`: `POST /api/v1/auth/mfa/challenge` - MFA 挑战端点
  * `ADD`: `POST /api/v1/users/me/mfa/recovery-codes` - 生成备用恢复码

### 📝 关键任务 (Key Tasks)

* [ ] `P4-T1`: **[Implement]** 创建 `pkg/plugins/mfa/totp/totp_provider.go`

  * 实现 `GenerateSecret()` - 生成 32 字节密钥
  * 实现 `GenerateQRCode(secret, issuer, account)` - 生成 QR 码（otpauth:// URL）
  * 实现 `VerifyCode(secret, code)` - 验证 6 位数字代码（容错 ±1 时间窗口）
  * 使用 RFC 6238 标准（时间窗口 30 秒）

* [ ] `P4-T2`: **[Implement]** 创建 `pkg/plugins/mfa/sms/sms_provider.go`

  * 实现 `SendCode(phoneNumber, code)` - 调用阿里云短信 API
  * 生成 6 位数字验证码（有效期 5 分钟）
  * 限流策略：同一手机号每分钟最多 1 条、每小时最多 5 条
  * 验证码存储到 Redis（key: `sms:otp:{phone}`, value: `{code}`, TTL: 5 分钟）

* [ ] `P4-T3`: **[Implement]** 创建 `pkg/plugins/mfa/email/email_provider.go`

  * 实现 `SendCode(email, code)` - 发送邮件验证码
  * 使用 SMTP 或 SendGrid API
  * 邮件模板：包含 6 位验证码 + 过期时间提示

* [ ] `P4-T4`: **[Implement]** 创建 `internal/domain/auth/mfa_policy.go`

  * 实现 `ShouldEnforceMFA(user)` - 判断用户是否需要强制 MFA

    * 规则示例：管理员角色强制启用、普通用户可选
  * 实现 `GetAvailableMFAMethods(user)` - 返回用户可用的 MFA 方法列表
  * 实现 `VerifyMFAChallenge(sessionID, method, code)` - 验证 MFA 挑战

* [ ] `P4-T5`: **[Database Design]** 创建数据库迁移

  * 表 `user_mfa_configs`:

    ```sql
    CREATE TABLE user_mfa_configs (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        method VARCHAR(20) NOT NULL, -- 'totp', 'sms', 'email'
        config JSONB NOT NULL, -- TOTP: {secret, verified}, SMS: {phone}, Email: {email}
        backup_codes TEXT[], -- 备用恢复码（加密存储）
        enabled BOOLEAN DEFAULT true,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW(),
        UNIQUE(user_id, method)
    );
    ```
  * 表 `mfa_verification_logs`:

    ```sql
    CREATE TABLE mfa_verification_logs (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id UUID NOT NULL REFERENCES users(id),
        method VARCHAR(20) NOT NULL,
        success BOOLEAN NOT NULL,
        ip_address INET,
        user_agent TEXT,
        created_at TIMESTAMPTZ DEFAULT NOW()
    );
    CREATE INDEX idx_mfa_logs_user ON mfa_verification_logs(user_id, created_at DESC);
    ```

* [ ] `P4-T6`: **[Implement]** 实现 MFA 挑战流程

  * 登录成功后，如果用户启用了 MFA，返回 `mfa_required: true` + `challenge_id`
  * 客户端使用 `challenge_id` 调用 `/api/v1/auth/mfa/challenge` 提交验证码
  * 验证成功后，更新会话状态为 `mfa_verified`，签发最终的 Access Token

* [ ] `P4-T7`: **[Implement]** 实现备用恢复码功能

  * 生成 10 个 8 位字母数字恢复码（示例：`A3F7-B2G9`）
  * 恢复码使用 bcrypt 哈希后存储到数据库
  * 每个恢复码仅能使用一次
  * 用户可在 MFA 验证时使用恢复码代替 TOTP/SMS 码

* [ ] `P4-T8`: **[Test Design]** 创建 `tests/integration/mfa_flow_test.go`

  * 测试用例：

    * `TestTOTPSetupAndVerify` - 设置 TOTP + 验证正确/错误代码
    * `TestSMSOTPSendAndVerify` - 发送短信验证码 + 验证
    * `TestMFALoginFlow` - 完整登录流程：用户名密码 → MFA 挑战 → 获取 Token
    * `TestRecoveryCodeUsage` - 使用备用恢复码绕过 MFA
    * `TestMFARateLimiting` - 验证短信发送限流

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[Unit Test]** (单元测试):

  * `Test Case 1`: `pkg/plugins/mfa/totp/totp_test.go::TestTOTPVerifyCode` - 验证 TOTP 算法正确性
  * `Test Case 2`: `internal/domain/auth/mfa_policy_test.go::TestMFAPolicyEnforcement` - 验证强制 MFA 策略

* **[Integration Test]** (集成测试):

  * `Test Case 3`: `tests/integration/mfa_flow_test.go::TestTOTPSetupAndVerify` - 完整 TOTP 设置和验证流程
  * `Test Case 4`: `tests/integration/mfa_flow_test.go::TestSMSOTPWithMockProvider` - 使用 Mock SMS 提供者测试

* **[Security Test]** (安全测试):

  * `Test Case 5`: 尝试暴力破解 TOTP 代码（1000 次尝试），验证账号锁定机制
  * `Test Case 6`: 使用已用过的恢复码尝试二次验证，验证被拒绝
  * `Test Case 7`: TOTP 密钥泄露场景，验证重新生成密钥能够使旧密钥失效

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (功能完整性) `Test Case 3`, `Test Case 4` 全部通过
* `AC-2`: (安全性) `Test Case 5`, `Test Case 6`, `Test Case 7` 全部通过
* `AC-3`: (用户体验) TOTP QR 码能够被 Google Authenticator 和 Authy 正确识别
* `AC-4`: (性能) SMS OTP 发送延迟 < 3 秒
* `AC-5`: (文档) 新增 `docs/features/mfa-setup-guide.md`，包含用户操作步骤截图

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P4` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] 数据库迁移脚本已提交并在测试环境验证
* [ ] 用户 MFA 设置指南已发布到文档站点
* [ ] 代码已合并到 `main` 分支，Tag `v0.5.0-phase4`

### 🔧 开发指南与约束 (Development Guidelines & Constraints)

**关键实现思路（Demo Code）：**

**示例 1：TOTP 设置** (`pkg/plugins/mfa/totp/totp_provider.go`)

```go
package totp

import (
    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
)

type TOTPProvider struct{}

func (p *TOTPProvider) GenerateSecret(issuer, accountName string) (*otp.Key, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      issuer,      // "QuantaID"
        AccountName: accountName, // user email
        SecretSize:  32,
    })
    if err != nil {
        return nil, err
    }
    return key, nil
}

func (p *TOTPProvider) VerifyCode(secret, code string) bool {
    return totp.Validate(code, secret)
}

func (p *TOTPProvider) GenerateQRCodeURL(key *otp.Key) string {
    return key.URL() // otpauth://totp/QuantaID:user@example.com?secret=...&issuer=QuantaID
}
```

**示例 2：MFA 挑战验证** (`internal/domain/auth/mfa_policy.go`)

```go
func (mp *MFAPolicy) VerifyMFAChallenge(ctx context.Context, challengeID, method, code string) error {
    // 1. 从 Redis 获取挑战信息
    challenge, err := mp.redis.Get(ctx, "mfa:challenge:"+challengeID).Result()
    if err != nil {
        return types.NewError(types.ErrCodeInvalidRequest, "invalid challenge")
    }
    
    var challengeData struct {
        UserID string
        Method string
    }
    json.Unmarshal([]byte(challenge), &challengeData)
    
    // 2. 根据方法验证代码
    switch method {
    case "totp":
        mfaConfig, _ := mp.mfaRepo.GetUserMFAConfig(ctx, challengeData.UserID, "totp")
        if !mp.totpProvider.VerifyCode(mfaConfig.Secret, code) {
            return types.NewError(types.ErrCodeUnauthorized, "invalid TOTP code")
        }
    case "sms":
        storedCode, _ := mp.redis.Get(ctx, "sms:otp:"+challengeData.UserID).Result()
        if storedCode != code {
            return types.NewError(types.ErrCodeUnauthorized, "invalid SMS code")
        }
    }
    
    // 3. 验证成功，删除挑战
    mp.redis.Del(ctx, "mfa:challenge:"+challengeID)
    
    return nil
}
```

**测试约束：**

* TOTP 测试必须模拟时间偏移（±30 秒），验证容错机制
* SMS 测试必须使用 Mock 提供者（不真实发送短信）
* 备用恢复码必须使用 `crypto/rand` 生成（不使用 `math/rand`）

---

