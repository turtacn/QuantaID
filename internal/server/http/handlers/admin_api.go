package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
)

type AdminHandler struct {
    userRepo types.UserRepository
    auditRepo auth.AuditLogRepository
}

func NewAdminHandler(userRepo types.UserRepository, auditRepo auth.AuditLogRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, auditRepo: auditRepo}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    // 1. 验证管理员权限
    user, ok := r.Context().Value("user").(*types.User)
    if !ok || !user.HasRole("admin") {
        http.Error(w, "insufficient permissions", http.StatusForbidden)
        return
    }

    // 2. 解析查询参数
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    size, _ := strconv.Atoi(r.URL.Query().Get("size"))
    search := r.URL.Query().Get("search")

    if size == 0 {
        size = 10
    }

    // 3. 查询用户
    users, total, err := h.userRepo.ListUsers(r.Context(), types.UserFilter{
        Search: search,
        Offset: page * size,
        Limit:  size,
    })
    if err != nil {
        http.Error(w, "query failed", http.StatusInternalServerError)
        return
    }

    // 4. 返回结果
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data": users,
        "total": total,
        "page": page,
        "size": size,
    })
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string   `json:"username"`
        Email    string   `json:"email"`
        RoleIDs  []string `json:"role_ids"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 创建用户逻辑...
    user := &types.User{
        Username: req.Username,
        Email:    req.Email,
    }

    if err := h.userRepo.CreateUser(r.Context(), user); err != nil {
        http.Error(w, "create failed", http.StatusInternalServerError)
        return
    }

    // 记录审计日志
    currentUser, _ := r.Context().Value("user").(*types.User)
    auditLog := &types.AuditLog{
        ActorID:     currentUser.ID,
        Action:     "user.create",
        Resource:   user.ID,
    }
    h.auditRepo.CreateLogEntry(r.Context(), auditLog)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
