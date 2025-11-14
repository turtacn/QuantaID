package handlers

import (
	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/sync"
	"go.uber.org/zap"
)

func RegisterSyncHandlers(router *mux.Router, syncService *sync.LDAPSyncService, logger *zap.Logger) {
	syncHandler := NewSyncHandler(syncService, logger)
	syncHandler.RegisterRoutes(router)
}
