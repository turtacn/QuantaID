package policy

import (
	"context"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/policy/engine"
	"go.uber.org/zap"
)

type service struct {
	repo        policy.RBACRepository
	opaProvider *engine.OPAProvider
	watcher     *fsnotify.Watcher
	logger      *zap.Logger
}

func NewService(repo policy.RBACRepository, opaProvider *engine.OPAProvider, logger *zap.Logger) PolicyService {
	s := &service{
		repo:        repo,
		opaProvider: opaProvider,
		logger:      logger,
	}

	if opaProvider != nil && opaProvider.Config().Enabled && opaProvider.Config().Mode == "sdk" {
		if err := s.startFileWatcher(opaProvider.Config().PolicyFile); err != nil {
			logger.Error("Failed to start policy file watcher", zap.Error(err))
		}
	}

	return s
}

func (s *service) startFileWatcher(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = watcher

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					s.logger.Info("Policy file modified, reloading...", zap.String("file", event.Name))
					if err := s.opaProvider.Reload(context.Background()); err != nil {
						s.logger.Error("Failed to reload policy", zap.Error(err))
					} else {
						s.logger.Info("Policy reloaded successfully")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				s.logger.Error("Policy watcher error", zap.Error(err))
			}
		}
	}()

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {
		return err
	}

	s.logger.Info("Started watching policy directory", zap.String("directory", dir))
	return nil
}

func (s *service) CreateRole(ctx context.Context, role *policy.Role) error {
	// Add any validation or business logic here
	return s.repo.CreateRole(ctx, role)
}

func (s *service) ListRoles(ctx context.Context) ([]*policy.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *service) UpdateRole(ctx context.Context, role *policy.Role) error {
	return s.repo.UpdateRole(ctx, role)
}

func (s *service) DeleteRole(ctx context.Context, roleID uint) error {
	return s.repo.DeleteRole(ctx, roleID)
}

func (s *service) CreatePermission(ctx context.Context, permission *policy.Permission) error {
	return s.repo.CreatePermission(ctx, permission)
}

func (s *service) ListPermissions(ctx context.Context) ([]*policy.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *service) AddPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	return s.repo.AddPermissionToRole(ctx, roleID, permissionID)
}

func (s *service) AssignRoleToUser(ctx context.Context, userID string, roleID uint) error {
	return s.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (s *service) UnassignRoleFromUser(ctx context.Context, userID string, roleID uint) error {
	return s.repo.UnassignRoleFromUser(ctx, userID, roleID)
}
