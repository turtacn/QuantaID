# QuantaID API 参考文档

## 概述

QuantaID 提供完整的 RESTful API 和 GraphQL 接口，支持企业级身份认证与访问控制的所有核心功能。本文档详细描述了 API 的结构、端点、参数和响应格式。

## API 设计原则

### RESTful 设计规范

QuantaID API 遵循 REST 架构风格和 OpenAPI 3.0 规范：

```mermaid
graph TB
    subgraph AR[API 资源（API Resources）]
        U[用户资源<br/>Users]
        G[用户组资源<br/>Groups]  
        A[应用资源<br/>Applications]
        P[策略资源<br/>Policies]
        S[会话资源<br/>Sessions]
        T[令牌资源<br/>Tokens]
    end
    
    subgraph AM[API 方法（API Methods）]
        GET[GET - 查询资源]
        POST[POST - 创建资源]
        PUT[PUT - 更新资源]
        PATCH[PATCH - 部分更新]
        DELETE[DELETE - 删除资源]
    end
    
    AR --> AM
````

### 版本策略

| 版本策略     | 说明                | 示例                       |
| -------- | ----------------- | ------------------------ |
| URL 版本控制 | API 版本通过 URL 路径指定 | `/api/v1/users`          |
| 向后兼容     | 新版本保持向后兼容性        | v1.1 兼容 v1.0             |
| 废弃通知     | 通过响应头提供废弃通知       | `X-API-Deprecated: true` |
| 迁移指南     | 提供详细的版本迁移文档       | 版本升级最佳实践                 |

## 认证与授权

### API 认证机制

```mermaid
sequenceDiagram
    participant C as 客户端（Client）
    participant API as API 网关（Gateway）
    participant AUTH as 认证服务（Auth Service）
    participant RES as 资源服务（Resource Service）
    
    C->>API: 1. 请求访问令牌<br/>POST /api/v1/auth/token
    API->>AUTH: 2. 验证客户端凭据
    AUTH-->>API: 3. 返回访问令牌
    API-->>C: 4. 返回令牌响应
    
    C->>API: 5. 携带令牌访问资源<br/>Authorization: Bearer <token>
    API->>AUTH: 6. 验证访问令牌
    AUTH-->>API: 7. 返回令牌信息
    API->>RES: 8. 转发已认证请求
    RES-->>API: 9. 返回资源数据
    API-->>C: 10. 返回 API 响应
```

### 支持的认证方式

| 认证方式         | 用途       | 示例                                              |
| ------------ | -------- | ----------------------------------------------- |
| Bearer Token | API 访问认证 | `Authorization: Bearer eyJhbGciOiJSUzI1NiIs...` |
| Basic Auth   | 客户端凭据认证  | `Authorization: Basic Y2xpZW50OnNlY3JldA==`     |
| API Key      | 服务间认证    | `X-API-Key: qid_ak_1234567890abcdef`            |
| mTLS         | 高安全场景    | 双向 TLS 证书认证                                     |

## 核心 API 端点

### 认证管理 API

#### 获取访问令牌

```http
POST /api/v1/auth/token
Content-Type: application/json

{
  "grant_type": "client_credentials",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "scope": "read:users write:users"
}
```

**响应示例：**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read:users write:users",
  "issued_at": "2024-01-15T10:00:00Z"
}
```

#### 用户认证流程

```http
POST /api/v1/auth/authenticate
Content-Type: application/json

{
  "username": "john.doe@example.com",
  "password": "SecurePassword123!",
  "mfa_token": "123456",
  "device_id": "device-12345",
  "remember_me": true
}
```

**响应示例：**

```json
{
  "authentication_result": "SUCCESS",
  "session_id": "sess_1234567890abcdef",
  "user": {
    "id": "user_12345",
    "username": "john.doe@example.com",
    "display_name": "John Doe",
    "roles": ["user", "admin"],
    "permissions": ["read:dashboard", "write:users"]
  },
  "expires_at": "2024-01-15T18:00:00Z"
}
```

### 身份管理 API

#### 用户管理

**创建用户：**

```http
POST /api/v1/users
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "username": "jane.smith@example.com",
  "email": "jane.smith@example.com",
  "display_name": "Jane Smith",
  "first_name": "Jane",
  "last_name": "Smith",
  "phone": "+1-555-0123",
  "attributes": {
    "department": "Engineering",
    "employee_id": "EMP001"
  },
  "groups": ["developers", "employees"]
}
```

**查询用户：**

```http
GET /api/v1/users?page=1&limit=20&search=john&filter[status]=active
Authorization: Bearer <access_token>
```

**响应示例：**

```json
{
  "data": [
    {
      "id": "user_12345",
      "username": "john.doe@example.com",
      "email": "john.doe@example.com",
      "display_name": "John Doe",
      "status": "active",
      "created_at": "2024-01-01T09:00:00Z",
      "last_login": "2024-01-15T08:30:00Z",
      "groups": ["employees", "admins"]
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_count": 98,
    "page_size": 20
  },
  "metadata": {
    "query_time": "45ms",
    "cache_hit": false
  }
}
```

#### 用户组管理

```http
POST /api/v1/groups
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "DevOps Team",
  "description": "DevOps and Infrastructure Team",
  "type": "security",
  "parent_group_id": "group_engineering",
  "attributes": {
    "cost_center": "CC-001",
    "manager": "user_54321"
  },
  "permissions": [
    "infrastructure:read",
    "infrastructure:write",
    "monitoring:admin"
  ]
}
```

### 应用管理 API

#### OIDC 应用注册

```http
POST /api/v1/applications
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "Corporate Dashboard",
  "type": "oidc",
  "client_id": "corp-dashboard-prod",
  "redirect_uris": [
    "https://dashboard.corp.com/auth/callback",
    "https://dashboard.corp.com/silent-renew"
  ],
  "post_logout_redirect_uris": [
    "https://dashboard.corp.com/logout"
  ],
  "response_types": ["code"],
  "grant_types": ["authorization_code", "refresh_token"],
  "token_endpoint_auth_method": "client_secret_post",
  "scope": "openid profile email groups",
  "settings": {
    "access_token_lifetime": 3600,
    "refresh_token_lifetime": 86400,
    "require_pkce": true
  }
}
```

#### SAML 应用配置

```http
POST /api/v1/applications
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "AWS SSO",
  "type": "saml",
  "entity_id": "https://signin.aws.amazon.com/saml",
  "acs_url": "https://signin.aws.amazon.com/saml",
  "slo_url": "https://signin.aws.amazon.com/saml/logout",
  "name_id_format": "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent",
  "attribute_mapping": {
    "email": "https://aws.amazon.com/SAML/Attributes/RoleSessionName",
    "groups": "https://aws.amazon.com/SAML/Attributes/Role"
  },
  "settings": {
    "assertion_lifetime": 300,
    "sign_assertions": true,
    "encrypt_assertions": false
  }
}
```

### 策略管理 API

#### 访问策略配置

```http
POST /api/v1/policies
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "Admin Dashboard Access",
  "description": "Controls access to admin dashboard",
  "type": "authorization",
  "rules": [
    {
      "effect": "allow",
      "subjects": ["group:admins", "user:john.doe@example.com"],
      "resources": ["dashboard:admin"],
      "actions": ["read", "write"],
      "conditions": {
        "time_range": {
          "start": "08:00",
          "end": "18:00"
        },
        "ip_whitelist": ["192.168.1.0/24", "10.0.0.0/8"]
      }
    }
  ],
  "priority": 100,
  "enabled": true
}
```

#### MFA 策略配置

```http
POST /api/v1/policies
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "name": "Adaptive MFA Policy",
  "type": "authentication",
  "rules": [
    {
      "name": "High Risk MFA",
      "conditions": {
        "risk_score": {"gte": 0.7},
        "or": [
          {"location.country": {"neq": "trusted_countries"}},
          {"device.trusted": {"eq": false}},
          {"time_of_day": {"not_between": ["08:00", "18:00"]}}
        ]
      },
      "actions": [
        {
          "type": "require_mfa",
          "methods": ["webauthn", "totp"],
          "timeout": 300
        }
      ]
    }
  ]
}
```

### 审计和监控 API

#### 审计日志查询

```http
GET /api/v1/audit/logs?start_time=2024-01-15T00:00:00Z&end_time=2024-01-15T23:59:59Z&event_type=authentication&user_id=user_12345
Authorization: Bearer <access_token>
```

**响应示例：**

```json
{
  "logs": [
    {
      "id": "log_abcdef123456",
      "timestamp": "2024-01-15T10:30:15.123Z",
      "event_type": "authentication",
      "event_name": "user_login_success",
      "user_id": "user_12345",
      "username": "john.doe@example.com",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...",
      "details": {
        "authentication_method": "password+totp",
        "application": "corp-dashboard-prod",
        "session_id": "sess_xyz789",
        "risk_score": 0.2
      }
    }
  ],
  "total_count": 1245,
  "query_time": "120ms"
}
```

#### 系统指标查询

```http
GET /api/v1/metrics?metric=authentication_rate&timerange=1h&aggregation=avg
Authorization: Bearer <access_token>
```

## GraphQL API

### GraphQL 端点

```
POST /api/v1/graphql
Content-Type: application/json
Authorization: Bearer <access_token>
```

### 查询示例

**用户信息查询：**

```graphql
query GetUserDetails($userId: ID!) {
  user(id: $userId) {
    id
    username
    email
    displayName
    status
    groups {
      id
      name
      permissions {
        resource
        actions
      }
    }
    sessions {
      id
      createdAt
      lastActivity
      deviceInfo
    }
    auditLogs(limit: 10) {
      timestamp
      eventType
      details
    }
  }
}
```

**批量用户操作：**

```graphql
mutation BulkUpdateUsers($input: BulkUserUpdateInput!) {
  bulkUpdateUsers(input: $input) {
    successCount
    failureCount
    errors {
      userId
      error
    }
    updatedUsers {
      id
      username
      status
    }
  }
}
```

## 错误处理

### 标准错误响应格式

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "The provided credentials are invalid",
    "details": {
      "field": "password",
      "reason": "Password does not meet complexity requirements"
    },
    "trace_id": "req_1234567890abcdef",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### 错误码表

| HTTP 状态码 | 错误码              | 描述      | 处理建议       |
| -------- | ---------------- | ------- | ---------- |
| 400      | INVALID_REQUEST | 请求参数错误  | 检查请求格式和参数  |
| 401      | UNAUTHORIZED     | 未认证访问   | 提供有效的认证凭据  |
| 403      | FORBIDDEN        | 权限不足    | 联系管理员获取权限  |
| 404      | NOT_FOUND       | 资源不存在   | 确认资源 ID 正确 |
| 409      | CONFLICT         | 资源冲突    | 处理资源冲突后重试  |
| 429      | RATE_LIMITED    | 超出速率限制  | 降低请求频率     |
| 500      | INTERNAL_ERROR  | 服务器内部错误 | 联系技术支持     |

## SDK 和示例代码

### Go SDK 示例

```go
package main

import (
    "context"
    "log"
    "github.com/turtacn/QuantaID/pkg/client"
)

func main() {
    // 初始化客户端
    qid, err := client.New(client.Config{
        Endpoint: "https://your-quantaid-instance.com",
        ClientID: "your-client-id",
        ClientSecret: "your-client-secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 创建用户
    user := &client.User{
        Username:    "test.user@example.com",
        Email:       "test.user@example.com", 
        DisplayName: "Test User",
        Groups:      []string{"employees"},
    }

    ctx := context.Background()
    createdUser, err := qid.Users.Create(ctx, user)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Created user: %s", createdUser.ID)
}
```

### JavaScript/Node.js 示例

```javascript
const { QuantaIDClient } = require('@quantaid/client');

const qid = new QuantaIDClient({
  endpoint: 'https://your-quantaid-instance.com',
  clientId: 'your-client-id',
  clientSecret: 'your-client-secret'
});

// 用户认证
async function authenticateUser(username, password) {
  try {
    const result = await qid.auth.authenticate({
      username,
      password
    });
    
    console.log('Authentication successful:', result.user.displayName);
    return result.sessionId;
  } catch (error) {
    console.error('Authentication failed:', error.message);
    throw error;
  }
}

// 获取用户信息
async function getUserInfo(userId) {
  const user = await qid.users.get(userId);
  return {
    id: user.id,
    name: user.displayName,
    groups: user.groups.map(g => g.name)
  };
}
```

### Python 示例

```python
from quantaid_client import QuantaIDClient
from quantaid_client.exceptions import AuthenticationError

client = QuantaIDClient(
    endpoint='https://your-quantaid-instance.com',
    client_id='your-client-id',
    client_secret='your-client-secret'
)

# 批量用户操作
def bulk_update_user_groups(user_ids, group_id):
    """将多个用户添加到指定用户组"""
    results = []
    
    for user_id in user_ids:
        try:
            user = client.users.get(user_id)
            user.groups.append(group_id)
            updated_user = client.users.update(user_id, user)
            results.append({
                'user_id': user_id,
                'status': 'success',
                'groups': updated_user.groups
            })
        except Exception as e:
            results.append({
                'user_id': user_id,
                'status': 'error',
                'error': str(e)
            })
    
    return results

# 策略验证
def check_access_permission(user_id, resource, action):
    """检查用户是否有权限执行特定操作"""
    try:
        result = client.authorization.check({
            'user_id': user_id,
            'resource': resource,
            'action': action
        })
        return result.allowed
    except Exception as e:
        print(f"Permission check failed: {e}")
        return False
```

## Webhook 集成

### Webhook 事件

QuantaID 支持通过 Webhook 推送关键事件：

```mermaid
graph LR
    subgraph QID[QuantaID 事件源]
        E1[用户登录<br/>User Login]
        E2[用户创建<br/>User Created]
        E3[权限变更<br/>Permission Changed]
        E4[安全事件<br/>Security Event]
    end
    
    subgraph WH[Webhook 处理]
        W1[事件队列<br/>Event Queue]
        W2[重试机制<br/>Retry Logic]
        W3[签名验证<br/>Signature Verification]
    end
    
    subgraph EP[外部端点]
        S1[SIEM 系统<br/>Security System]
        S2[监控系统<br/>Monitoring System]
        S3[业务系统<br/>Business System]
    end
    
    E1 --> W1
    E2 --> W1
    E3 --> W1
    E4 --> W1
    
    W1 --> W2
    W2 --> W3
    
    W3 --> S1
    W3 --> S2
    W3 --> S3
```

### Webhook 配置

```http
POST /api/v1/webhooks
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "url": "https://your-system.com/webhooks/quantaid",
  "events": [
    "user.created",
    "user.updated", 
    "user.deleted",
    "authentication.success",
    "authentication.failed",
    "permission.granted",
    "permission.revoked"
  ],
  "secret": "webhook-signing-secret",
  "active": true,
  "retry_config": {
    "max_attempts": 3,
    "backoff_strategy": "exponential"
  }
}
```

### Webhook 有效载荷示例

```json
{
  "id": "event_123456789",
  "type": "user.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "user": {
      "id": "user_12345",
      "username": "new.user@example.com",
      "email": "new.user@example.com",
      "display_name": "New User",
      "groups": ["employees"],
      "created_by": "admin@example.com"
    }
  },
  "metadata": {
    "source": "quantaid",
    "version": "1.0",
    "trace_id": "req_abcdef123456"
  }
}
```

## 性能和限制

### API 限制

| 限制类型 | 默认值    | 企业版     | 说明             |
| ---- | ------ | ------- | -------------- |
| 请求频率 | 100/分钟 | 1000/分钟 | 每个 API 密钥的请求限制 |
| 并发连接 | 50     | 500     | 同时活跃的连接数       |
| 响应大小 | 10MB   | 100MB   | 单次响应的最大大小      |
| 批量操作 | 100 记录 | 1000 记录 | 批量操作的记录数限制     |

### 性能优化建议

1. **使用分页查询**：对于大数据集查询，始终使用分页参数
2. **实现客户端缓存**：缓存不经常变更的数据（如用户组信息）
3. **批量操作**：使用批量 API 减少网络往返次数
4. **异步处理**：对于耗时操作，使用异步 API 和 Webhook 通知

## 参考资料

[1] OpenAPI Specification 3.0 - [https://swagger.io/specification/](https://swagger.io/specification/)

[2] GraphQL Specification - [https://spec.graphql.org/](https://spec.graphql.org/)

[3] OAuth 2.1 Authorization Framework - [https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1)

[4] OpenID Connect 1.0 - [https://openid.net/specs/openid-connect-core-1_0.html](https://openid.net/specs/openid-connect-core-1_0.html)

[5] SAML 2.0 Web SSO Profile - [http://docs.oasis-open.org/security/saml/v2.0/saml-profiles-2.0-os.pdf](http://docs.oasis-open.org/security/saml/v2.0/saml-profiles-2.0-os.pdf)

[6] JSON Web Token (JWT) - [https://datatracker.ietf.org/doc/html/rfc7519](https://datatracker.ietf.org/doc/html/rfc7519)

[7] RESTful API Design Best Practices - [https://restfulapi.net/](https://restfulapi.net/)

[8] Webhook Security Best Practices - [https://webhooks.fyi/security/](https://webhooks.fyi/security/)