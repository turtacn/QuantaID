## PHASE 5: Web 管理控制台 - 用户与应用管理

> **(Phase 5: Web Admin Console - User & Application Management)**

* **Phase ID:** `P5`
* **Branch:** `feat/round1-phase5-admin-console`
* **Dependencies:** `P1`, `P2`, `P4`（需要完整的后端 API）

### 🎯 目标 (Objectives)

* 实现基于 React + TypeScript 的管理控制台前端
* 提供用户管理界面（CRUD、角色分配、MFA 状态查看）
* 提供应用管理界面（OAuth 客户端注册、Redirect URI 配置）
* 实现权限管理界面（角色和权限关系可视化）
* 提供审计日志查询界面

### 📦 交付物与变更 (Deliverables & Changes)

* **[Code Change]** (代码变更)

  * `ADD`: `web/admin-console/` - React 前端项目
  * `ADD`: `web/admin-console/src/pages/UserManagement.tsx` - 用户管理页面
  * `ADD`: `web/admin-console/src/pages/ApplicationManagement.tsx` - 应用管理页面
  * `ADD`: `web/admin-console/src/pages/RoleManagement.tsx` - 角色管理页面
  * `ADD`: `web/admin-console/src/pages/AuditLogs.tsx` - 审计日志页面
  * `ADD`: `web/admin-console/src/services/api.ts` - API 客户端封装
  * `ADD`: `internal/server/http/handlers/admin_api.go` - 后端管理 API 端点

* **[Dependency Change]** (依赖变更 - 前端):

  * `ADD`: `react@18`, `react-dom@18`
  * `ADD`: `@tanstack/react-query` - 数据获取和缓存
  * `ADD`: `react-router-dom@6` - 路由管理
  * `ADD`: `@mui/material` - UI 组件库
  * `ADD`: `axios` - HTTP 客户端

* **[API Change]** (API 变更 - 后端):

  * `ADD`: `GET /api/v1/admin/users` - 获取用户列表（分页、搜索）
  * `ADD`: `POST /api/v1/admin/users` - 创建用户
  * `ADD`: `PATCH /api/v1/admin/users/:id` - 更新用户信息
  * `ADD`: `DELETE /api/v1/admin/users/:id` - 删除用户
  * `ADD`: `GET /api/v1/admin/applications` - 获取应用列表
  * `ADD`: `POST /api/v1/admin/applications` - 注册新应用
  * `ADD`: `GET /api/v1/admin/audit-logs` - 查询审计日志

### 📝 关键任务 (Key Tasks)

* [ ] `P5-T1`: **[Setup]** 初始化 React 项目

  * 使用 Vite 脚手架创建项目（`npm create vite@latest admin-console -- --template react-ts`）
  * 配置 ESLint + Prettier
  * 配置代理（开发环境代理到后端 `http://localhost:8080`）

* [ ] `P5-T2`: **[Implement]** 创建用户管理页面 (`web/admin-console/src/pages/UserManagement.tsx`)

  * 功能：

    * 用户列表展示（表格，支持排序、搜索）
    * 添加用户对话框（表单：用户名、邮箱、角色）
    * 编辑用户对话框（更新邮箱、角色、启用/禁用状态）
    * 批量删除用户（多选 + 确认对话框）
    * 查看用户 MFA 状态（TOTP 是否启用）
  * 使用 MUI DataGrid 组件展示数据
  * 使用 React Query 管理数据获取和缓存

* [ ] `P5-T3`: **[Implement]** 创建应用管理页面 (`web/admin-console/src/pages/ApplicationManagement.tsx`) - 续

  * 功能：

    * 应用列表展示（卡片布局）
    * 注册新应用表单（应用名称、描述、Redirect URIs）
    * 查看应用详情（Client ID、Client Secret、授权类型）
    * 编辑 Redirect URIs（动态添加/删除）
    * 重新生成 Client Secret（带确认提示）
    * 删除应用（软删除，保留审计记录）
  * 使用 MUI Card 组件展示应用信息
  * Client Secret 显示时使用 "点击显示" 按钮（默认隐藏为 `********`）

* [ ] `P5-T4`: **[Implement]** 创建角色管理页面 (`web/admin-console/src/pages/RoleManagement.tsx`)

  * 功能：

    * 角色列表展示（树形结构，支持父子角色）
    * 创建角色对话框（角色名称、描述、父角色选择）
    * 编辑角色权限（多选框列表，支持按模块分组）
    * 查看角色成员（显示拥有该角色的用户列表）
    * 删除角色（检查是否有用户关联）
  * 使用 MUI TreeView 展示角色层级关系
  * 权限选择使用分组多选框（例如：用户管理、应用管理、审计日志）

* [ ] `P5-T5`: **[Implement]** 创建审计日志页面 (`web/admin-console/src/pages/AuditLogs.tsx`)

  * 功能：

    * 日志列表展示（时间倒序，分页加载）
    * 高级搜索（用户名、操作类型、时间范围、IP 地址）
    * 日志详情抽屉（显示完整的请求/响应数据）
    * 导出功能（导出为 CSV 或 JSON）
    * 实时刷新（可选，WebSocket 推送新日志）
  * 使用 MUI Table 展示日志
  * 操作类型使用不同颜色标签（成功=绿色、失败=红色、警告=黄色）

* [ ] `P5-T6`: **[Implement]** 创建 API 客户端封装 (`web/admin-console/src/services/api.ts`)

  * 封装所有后端 API 调用
  * 统一错误处理（401 自动跳转登录、500 显示错误提示）
  * 请求拦截器（自动添加 Authorization Header）
  * 响应拦截器（处理分页元数据）
  * 示例代码：

    ```typescript
    import axios from 'axios';

    const apiClient = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
      timeout: 10000,
    });

    apiClient.interceptors.request.use((config) => {
      const token = localStorage.getItem('access_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    apiClient.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );

    export const userAPI = {
      list: (params: { page: number; size: number; search?: string }) =>
        apiClient.get('/admin/users', { params }),
      create: (data: { username: string; email: string; role_ids: string[] }) =>
        apiClient.post('/admin/users', data),
      update: (id: string, data: Partial<User>) =>
        apiClient.patch(`/admin/users/${id}`, data),
      delete: (id: string) => apiClient.delete(`/admin/users/${id}`),
    };
    ```

* [ ] `P5-T7`: **[Backend API]** 创建后端管理 API 端点 (`internal/server/http/handlers/admin_api.go`)

  * 实现用户管理 API：

    * `GET /api/v1/admin/users` - 分页查询（支持 `search` 参数）
    * `POST /api/v1/admin/users` - 创建用户（验证邮箱格式、用户名唯一性）
    * `PATCH /api/v1/admin/users/:id` - 更新用户（支持部分更新）
    * `DELETE /api/v1/admin/users/:id` - 软删除用户（标记为 `deleted_at`）
  * 实现应用管理 API：

    * `GET /api/v1/admin/applications` - 查询应用列表
    * `POST /api/v1/admin/applications` - 注册应用（生成 Client ID/Secret）
    * `PATCH /api/v1/admin/applications/:id` - 更新 Redirect URIs
    * `POST /api/v1/admin/applications/:id/rotate-secret` - 重新生成 Secret
  * 实现审计日志 API：

    * `GET /api/v1/admin/audit-logs` - 查询日志（支持多维度过滤）

* [ ] `P5-T8`: **[Security]** 实现管理端权限控制

  * 所有 `/api/v1/admin/*` 端点必须验证用户拥有 `admin` 角色
  * 使用中间件 `RequireRole("admin")` 进行拦截
  * 敏感操作（删除用户、重置密码）记录到审计日志

* [ ] `P5-T9`: **[Test Design]** 创建前端单元测试

  * 使用 Vitest + React Testing Library
  * 测试用例：

    * `UserManagement.test.tsx::TestUserListRendering` - 验证用户列表正确渲染
    * `ApplicationManagement.test.tsx::TestCreateApplication` - 验证创建应用表单提交
    * `api.test.ts::TestAPIErrorHandling` - 验证 401 错误自动跳转登录

* [ ] `P5-T10`: **[Deployment]** 配置前端构建和部署

  * 添加 `web/admin-console/Dockerfile`：

    ```dockerfile
    FROM node:20-alpine AS builder
    WORKDIR /app
    COPY package*.json ./
    RUN npm ci
    COPY . .
    RUN npm run build

    FROM nginx:alpine
    COPY --from=builder /app/dist /usr/share/nginx/html
    COPY nginx.conf /etc/nginx/conf.d/default.conf
    EXPOSE 80
    ```
  * 配置 Nginx 反向代理（API 请求转发到后端）

### 🧪 测试设计与验收 (Test Design & Acceptance)

**1. 测试设计 (Test Design):**

* **[Frontend Unit Test]** (前端单元测试):

  * `Test Case 1`: `UserManagement.test.tsx::TestUserSearch` - 输入搜索关键词后，验证 API 请求参数正确
  * `Test Case 2`: `ApplicationManagement.test.tsx::TestSecretVisibilityToggle` - 点击 "显示 Secret" 按钮，验证 Secret 显示/隐藏

* **[Frontend Integration Test]** (前端集成测试):

  * `Test Case 3`: `E2E::TestUserCRUDFlow` - 使用 Playwright 测试完整的用户创建、编辑、删除流程
  * `Test Case 4`: `E2E::TestApplicationRegistration` - 测试应用注册表单提交后，能够在列表中看到新应用

* **[Backend API Test]** (后端 API 测试):

  * `Test Case 5`: `handlers/admin_api_test.go::TestAdminUsersListWithSearch` - 验证搜索功能返回正确的用户
  * `Test Case 6`: `handlers/admin_api_test.go::TestNonAdminUserAccessDenied` - 验证普通用户访问管理 API 返回 403

**2. 效果验收 (Acceptance Criteria):**

* `AC-1`: (功能完整性) 所有 CRUD 操作在前端界面可正常执行
* `AC-2`: (性能) 用户列表加载时间 < 1 秒（1000 个用户）
* `AC-3`: (安全性) 非管理员用户无法访问管理控制台（后端返回 403）
* `AC-4`: (用户体验) 所有表单验证错误有明确的提示信息（例如："邮箱格式不正确"）
* `AC-5`: (响应式设计) 管理控制台在移动端（宽度 < 768px）能够正常使用
* `AC-6`: (文档) 新增 `docs/admin-console/user-guide.md`，包含操作截图

### ✅ 完成标准 (Definition of Done - DoD)

* [ ] 所有 `P5` 关键任务均已勾选完成
* [ ] 所有 `AC` 均已满足
* [ ] 前端单元测试覆盖率 > 70%
* [ ] E2E 测试在 CI 中自动运行（使用 Playwright）
* [ ] 前端已部署到测试环境（[https://admin.quantaid-test.com）](https://admin.quantaid-test.com）)
* [ ] 代码已合并到 `main` 分支，Tag `v0.6.0-phase5`

### 🔧 开发指南与约束 (Development Guidelines & Constraints)

**关键实现思路（Demo Code）：**

**示例 1：用户管理页面** (`web/admin-console/src/pages/UserManagement.tsx`)

```typescript
import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { DataGrid, GridColDef } from '@mui/x-data-grid';
import { Button, Dialog, DialogTitle, DialogContent, TextField } from '@mui/material';
import { userAPI } from '../services/api';

export const UserManagement: React.FC = () => {
  const [page, setPage] = React.useState(0);
  const [search, setSearch] = React.useState('');
  const [openDialog, setOpenDialog] = React.useState(false);
  const queryClient = useQueryClient();
  
  // 查询用户列表
  const { data: users, isLoading } = useQuery({
    queryKey: ['users', page, search],
    queryFn: () => userAPI.list({ page, size: 20, search }),
  });
  
  // 创建用户 Mutation
  const createUserMutation = useMutation({
    mutationFn: userAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setOpenDialog(false);
    },
  });
  
  const columns: GridColDef[] = [
    { field: 'username', headerName: '用户名', width: 150 },
    { field: 'email', headerName: '邮箱', width: 200 },
    { field: 'role', headerName: '角色', width: 120 },
    {
      field: 'mfa_enabled',
      headerName: 'MFA 状态',
      width: 100,
      renderCell: (params) => (
        <span style={{ color: params.value ? 'green' : 'gray' }}>
          {params.value ? '已启用' : '未启用'}
        </span>
      ),
    },
    {
      field: 'actions',
      headerName: '操作',
      width: 150,
      renderCell: (params) => (
        <>
          <Button size="small" onClick={() => handleEdit(params.row.id)}>
            编辑
          </Button>
          <Button size="small" color="error" onClick={() => handleDelete(params.row.id)}>
            删除
          </Button>
        </>
      ),
    },
  ];
  
  const handleEdit = (id: string) => {
    // TODO: 打开编辑对话框
  };
  
  const handleDelete = async (id: string) => {
    if (confirm('确认删除该用户？')) {
      await userAPI.delete(id);
      queryClient.invalidateQueries({ queryKey: ['users'] });
    }
  };
  
  return (
    <div>
      <h1>用户管理</h1>
      <div style={{ marginBottom: 16 }}>
        <TextField
          label="搜索用户"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          size="small"
        />
        <Button variant="contained" onClick={() => setOpenDialog(true)} style={{ marginLeft: 16 }}>
          添加用户
        </Button>
      </div>
      
      <DataGrid
        rows={users?.data || []}
        columns={columns}
        loading={isLoading}
        pagination
        paginationMode="server"
        rowCount={users?.total || 0}
        page={page}
        onPageChange={setPage}
        pageSize={20}
      />
      
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)}>
        <DialogTitle>添加用户</DialogTitle>
        <DialogContent>
          {/* TODO: 添加用户表单 */}
        </DialogContent>
      </Dialog>
    </div>
  );
};
```

**示例 2：后端管理 API** (`internal/server/http/handlers/admin_api.go`)

```go
package handlers

import (
    "net/http"
    "quantaid/internal/storage/postgres"
    "quantaid/pkg/types"
    "github.com/gin-gonic/gin"
)

type AdminHandler struct {
    userRepo *postgres.UserRepository
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
    // 1. 验证管理员权限
    user := c.MustGet("user").(*types.User)
    if !user.HasRole("admin") {
        c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
        return
    }
    
    // 2. 解析查询参数
    var params struct {
        Page   int    `form:"page" binding:"min=0"`
        Size   int    `form:"size" binding:"min=1,max=100"`
        Search string `form:"search"`
    }
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 3. 查询用户
    users, total, err := h.userRepo.ListWithPagination(c.Request.Context(), postgres.UserFilter{
        Search: params.Search,
        Offset: params.Page * params.Size,
        Limit:  params.Size,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
        return
    }
    
    // 4. 返回结果
    c.JSON(http.StatusOK, gin.H{
        "data": users,
        "total": total,
        "page": params.Page,
        "size": params.Size,
    })
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
    var req struct {
        Username string   `json:"username" binding:"required,min=3,max=32"`
        Email    string   `json:"email" binding:"required,email"`
        RoleIDs  []string `json:"role_ids" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 创建用户逻辑...
    user := &types.User{
        Username: req.Username,
        Email:    req.Email,
    }
    
    if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
        return
    }
    
    // 记录审计日志
    auditLog := &types.AuditLog{
        UserID:     c.MustGet("user").(*types.User).ID,
        Action:     "user.create",
        ResourceID: user.ID,
        IPAddress:  c.ClientIP(),
    }
    // 保存审计日志...
    
    c.JSON(http.StatusCreated, user)
}
```

**测试约束：**

* E2E 测试必须在无头浏览器模式下运行（Playwright headless: true）
* API 请求超时必须设置为 10 秒（避免测试挂起）
* 前端测试必须 Mock 所有 API 请求（使用 MSW - Mock Service Worker）

---

