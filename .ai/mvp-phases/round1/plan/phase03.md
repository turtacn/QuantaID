## PHASE 3: LDAP/AD 身份源连接器插件实现

> **(Phase 3: LDAP/AD Identity Connector Plugin Implementation)**

* **Phase ID:** `P3`
* **Branch:** `feat/round1-phase3-ldap-connector`
* **Dependencies:** `P1`（需要插件管理器）

### 🎯 目标 (Objectives)

* 实现 LDAP/Active Directory 身份源连接器插件
* 支持 LDAP 用户认证、属性查询、组成员关系查询
* 实现用户同步功能（从 LDAP 导入用户到本地数据库）
* 提供连接池和重试机制

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `ADD`: `pkg/plugins/connectors/ldap/ldap_connector.go` - LDAP 连接器主实现
  * `ADD`: `pkg/plugins/connectors/ldap/config.go` - LDAP 配置结构
  * `ADD`: `pkg/plugins/connectors/ldap/mapper.go` - 属性映射器
  * `ADD`: `internal/services/sync/ldap_sync_service.go` - LDAP 用户同步服务
  * `ADD`: `tests/integration/ldap_connector_test.go` - LDAP 集成测试

* **[Dependency Change]** (依赖变更)

  * `ADD`: `github.com/go-ldap/ldap/v3` - LDAP 客户端库

* **[Config Change]** (配置变更)

  * `ADD`: `configs/plugins/ldap.yaml.example` - LDAP 插件配置示例

### 📝 关键任务 (Key Tasks)

* [ ] `P3-T1`: **[Implement]** 创建 `pkg/plugins/connectors/ldap/ldap_connector.go`

  * 实现 `IIdentityConnector` 接口
  * 方法：`Authenticate(username, password)` - LDAP Bind 认证
  * 方法：`GetUser(username)` - 查询用户属性（DN、CN、mail、memberOf）
  * 方法：`SearchUsers(filter)` - 搜索用户列表
  * 连接池管理：支持 TLS/STARTTLS 加密连接

* [ ] `P3-T2`: **[Implement]** 创建 `pkg/plugins/connectors/ldap/mapper.go`

  * 定义 LDAP 属性到 `types.User` 的映射规则
  * 支持自定义属性映射（配置文件定义）
  * 示例：`sAMAccountName` → `username`, `mail` → `email`, `memberOf` → `groups`

* [ ] `P3-T3`: **[Implement]** 创建 `internal/services/sync/ldap_sync_service.go`

  * 实现增量同步逻辑（基于 `modifyTimestamp` 或 `uSNChanged`）
  * 支持全量同步（首次导入）
  * 同步策略：

    * 新用户：自动创建到本地数据库
    * 已有用户：更新属性（邮箱、手机号等）
    * 已删除用户：标记

* [ ] `P3-T4`: **[Implement]** 实现 LDAP 连接池与健康检查

  * 连接池大小：10-50 个连接（可配置）
  * 连接超时：5 秒，查询超时：10 秒
  * 健康检查：每 30 秒执行一次 `whoami` 扩展操作
  * 实现指数退避重试（最多 3 次）

* [ ] `P3-T5`: **[Implement]** 实现 LDAP 分页查询

  * 支持大量用户场景（>1000 用户）
  * 使用 LDAP Paged Results Control（RFC 2696）
  * 每页查询 100 条记录

* [ ] `P3-T6`: **[Test Design]** 创建 `tests/integration/ldap_connector_test.go`

  * 使用 `testcontainers` 启动 OpenLDAP 容器
  * 预置测试数据（10 个用户、3 个组）
  * 测试用例：

    * `TestLDAPAuthenticate` - 验证正确/错误密码
    * `TestLDAPGetUser` - 查询用户属性
    * `TestLDAPSearchUsers` - 搜索过滤器（`objectClass=person`）
    * `TestLDAPUserSync` - 全量同步 + 增量同步

* [ ] `P3-T7`: **[Config]** 创建 `configs/plugins/ldap.yaml.example`

  * 配置项：

    ```yaml
    ldap:
      host: "ldap.example.com"
      port: 389
      use_tls: true
      bind_dn: "cn=admin,dc=example,dc=com"
      bind_password: "secret"
      base_dn: "ou=users,dc=example,dc=com"
      user_filter: "(objectClass=inetOrgPerson)"
      attribute_mapping:
        username: "uid"
        email: "mail"
        display_name: "displayName"
        phone: "telephoneNumber"
      sync:
        enabled: true
        interval: "1h"
        full_sync_cron: "0 2 * * *"  # 每天凌晨 2 点全量同步
    ```

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[Integration Test]** (集成测试):

  * `Test Case 1`: `TestLDAPAuthenticateSuccess` - 使用正确凭证绑定成功
  * `Test Case 2`: `TestLDAPAuthenticateFailure` - 使用错误密码返回认证失败
  * `Test Case 3`: `TestLDAPGetUserAttributes` - 查询用户返回正确的邮箱、电话号码
  * `Test Case 4`: `TestLDAPSearchWithFilter` - 使用复杂过滤器 `(&(objectClass=person)(mail=*@example.com))` 搜索
  * `Test Case 5`: `TestLDAPPaginatedSearch` - 模拟 1500 个用户，验证分页查询返回所有记录

* **[Sync Test]** (同步测试):

  * `Test Case 6`: `TestLDAPFullSync` - 全量同步 10 个用户到本地数据库
  * `Test Case 7`: `TestLDAPIncrementalSync` - 修改 LDAP 用户属性后，增量同步更新本地记录
  * `Test Case 8`: `TestLDAPSyncDeletedUser` - LDAP 中删除用户后，本地用户状态标记为 `disabled`

* **[Performance Test]** (性能测试):

  * `Test Case 9`: 并发 50 个 LDAP 认证请求，验证连接池不耗尽
  * `Test Case 10`: 同步 10000 个用户，验证耗时 < 5 分钟

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (功能完整性) `Test Case 1-8` 全部通过
* `AC-2`: (性能) `Test Case 9` 中所有请求在 3 秒内完成
* `AC-3`: (可靠性) LDAP 服务宕机时，系统能够优雅降级（使用本地缓存认证）
* `AC-4`: (安全性) LDAP 密码使用 TLS 加密传输，配置文件中的 `bind_password` 支持从环境变量读取
* `AC-5`: (文档) 新增 `docs/plugins/ldap-connector.md`，包含配置指南和故障排查步骤

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P3` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] 插件文档已提交并包含至少 3 个实际配置示例
* [ ] 集成测试在 CI 中自动运行（使用 OpenLDAP 容器）
* [ ] 代码已合并到 `main` 分支，Tag `v0.4.0-phase3`

### 🔧 开发指南与约束 (Development Guidelines & Constraints)

**关键实现思路（Demo Code）：**

**示例 1：LDAP 连接器实现** (`pkg/plugins/connectors/ldap/ldap_connector.go`)

```go
package ldap

import (
    "fmt"
    "github.com/go-ldap/ldap/v3"
    "quantaid/pkg/types"
)

type LDAPConnector struct {
    config *LDAPConfig
    conn   *ldap.Conn
}

type LDAPConfig struct {
    Host         string
    Port         int
    UseTLS       bool
    BindDN       string
    BindPassword string
    BaseDN       string
    UserFilter   string
    AttrMapping  map[string]string
}

func NewLDAPConnector(cfg *LDAPConfig) (*LDAPConnector, error) {
    conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port))
    if err != nil {
        return nil, fmt.Errorf("ldap dial: %w", err)
    }
    
    // TLS 升级
    if cfg.UseTLS {
        if err := conn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
            return nil, fmt.Errorf("ldap starttls: %w", err)
        }
    }
    
    // 绑定管理员账号
    if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
        return nil, fmt.Errorf("ldap bind: %w", err)
    }
    
    return &LDAPConnector{config: cfg, conn: conn}, nil
}

func (lc *LDAPConnector) Authenticate(username, password string) (*types.User, error) {
    // 1. 搜索用户 DN
    searchRequest := ldap.NewSearchRequest(
        lc.config.BaseDN,
        ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
        fmt.Sprintf("(&%s(%s=%s))", lc.config.UserFilter, lc.config.AttrMapping["username"], username),
        []string{"dn", "mail", "displayName"},
        nil,
    )
    
    sr, err := lc.conn.Search(searchRequest)
    if err != nil || len(sr.Entries) == 0 {
        return nil, types.NewError(types.ErrCodeUnauthorized, "user not found")
    }
    
    userDN := sr.Entries[0].DN
    
    // 2. 使用用户凭证绑定验证密码
    if err := lc.conn.Bind(userDN, password); err != nil {
        return nil, types.NewError(types.ErrCodeUnauthorized, "invalid password")
    }
    
    // 3. 映射用户属性
    user := &types.User{
        Username: username,
        Email:    sr.Entries[0].GetAttributeValue("mail"),
        FullName: sr.Entries[0].GetAttributeValue("displayName"),
    }
    
    return user, nil
}

func (lc *LDAPConnector) GetUser(username string) (*types.User, error) {
    // 实现用户查询（类似 Authenticate，但不验证密码）
    // ...
}

func (lc *LDAPConnector) SearchUsers(filter string) ([]*types.User, error) {
    // 实现分页搜索
    // ...
}
```

**示例 2：用户同步服务** (`internal/services/sync/ldap_sync_service.go`)

```go
package sync

import (
    "context"
    "quantaid/pkg/plugins/connectors/ldap"
    "quantaid/internal/storage/postgres"
)

type LDAPSyncService struct {
    ldapConnector *ldap.LDAPConnector
    userRepo      *postgres.UserRepository
}

func (s *LDAPSyncService) FullSync(ctx context.Context) error {
    // 1. 从 LDAP 查询所有用户
    users, err := s.ldapConnector.SearchUsers("(objectClass=person)")
    if err != nil {
        return err
    }
    
    // 2. 批量插入/更新到数据库
    for _, ldapUser := range users {
        existingUser, _ := s.userRepo.GetByUsername(ctx, ldapUser.Username)
        
        if existingUser == nil {
            // 新用户：创建
            if err := s.userRepo.Create(ctx, ldapUser); err != nil {
                return err
            }
        } else {
            // 已有用户：更新属性
            existingUser.Email = ldapUser.Email
            existingUser.FullName = ldapUser.FullName
            if err := s.userRepo.Update(ctx, existingUser); err != nil {
                return err
            }
        }
    }
    
    return nil
}

func (s *LDAPSyncService) IncrementalSync(ctx context.Context, lastSyncTime time.Time) error {
    // 使用 modifyTimestamp >= lastSyncTime 过滤增量变更
    filter := fmt.Sprintf("(&(objectClass=person)(modifyTimestamp>=%s))", 
        lastSyncTime.Format("20060102150405Z"))
    
    changedUsers, err := s.ldapConnector.SearchUsers(filter)
    if err != nil {
        return err
    }
    
    // 更新变更的用户
    for _, user := range changedUsers {
        if err := s.userRepo.Update(ctx, user); err != nil {
            return err
        }
    }
    
    return nil
}
```

**测试约束：**

* 集成测试必须使用 `rroemhild/test-openldap` Docker 镜像（预置测试数据）
* LDAP 连接超时必须设置为 5 秒，避免测试挂起
* 测试完成后必须关闭 LDAP 连接（`defer conn.Close()`）

---