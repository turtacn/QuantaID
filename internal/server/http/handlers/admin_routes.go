package handlers

import (
	"net/http"

	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
)

func RegisterAdminHandlers(router *http.ServeMux, userRepo types.UserRepository, auditRepo auth.AuditLogRepository) {
	adminHandler := NewAdminHandler(userRepo, auditRepo)

	router.HandleFunc("/api/v1/admin/users", adminHandler.ListUsers)
	router.HandleFunc("/api/v1/admin/users/create", adminHandler.CreateUser)
}
