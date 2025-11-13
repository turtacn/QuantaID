package sync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/plugins/connectors/ldap"
	"github.com/turtacn/QuantaID/pkg/types"
)

type LDAPSyncService struct {
	ldapConnector *ldap.LDAPConnector
	userRepo      identity.UserRepository
}

func NewLDAPSyncService(ldapConnector *ldap.LDAPConnector, userRepo identity.UserRepository) *LDAPSyncService {
	return &LDAPSyncService{
		ldapConnector: ldapConnector,
		userRepo:      userRepo,
	}
}

func (s *LDAPSyncService) FullSync(ctx context.Context) error {
	ldapUsers, err := s.ldapConnector.SyncUsers(ctx)
	if err != nil {
		return err
	}

	ldapUserMap := make(map[string]*types.User)
	for _, user := range ldapUsers {
		ldapUserMap[user.Username] = user
	}

	// This is not efficient for large user bases, but it's a start.
	// A better implementation would use a streaming or paginated approach.
	localUsers, err := s.userRepo.ListUsers(ctx, identity.PaginationQuery{PageSize: 10000, Offset: 0})
	if err != nil {
		return err
	}

	for _, localUser := range localUsers {
		if _, ok := ldapUserMap[localUser.Username]; !ok {
			// User exists locally but not in LDAP, so mark as inactive
			localUser.Status = types.UserStatusInactive
			if err := s.userRepo.UpdateUser(ctx, localUser); err != nil {
				return err
			}
		}
	}

	for _, ldapUser := range ldapUsers {
		existingUser, err := s.userRepo.GetUserByUsername(ctx, ldapUser.Username)
		if err != nil {
			var e *types.Error
			if errors.As(err, &e) && e.Code == types.ErrUserNotFound.Code {
				// User not found, so create them
				if err := s.userRepo.CreateUser(ctx, ldapUser); err != nil {
					return err
				}
			} else {
				// A different error occurred, so we should abort the sync
				return err
			}
		} else {
			// User exists, so update their attributes
			existingUser.Email = ldapUser.Email
			existingUser.Phone = ldapUser.Phone
			existingUser.Attributes = ldapUser.Attributes
			existingUser.Status = types.UserStatusActive // Ensure user is active
			if err := s.userRepo.UpdateUser(ctx, existingUser); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *LDAPSyncService) IncrementalSync(ctx context.Context, lastSyncTime time.Time) error {
	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(modifyTimestamp>=%s))",
		lastSyncTime.Format("20060102150405Z"))

	changedUsers, err := s.ldapConnector.SearchUsers(ctx, filter)
	if err != nil {
		return err
	}

	for _, user := range changedUsers {
		existingUser, err := s.userRepo.GetUserByUsername(ctx, user.Username)
		if err != nil {
			var e *types.Error
			if errors.As(err, &e) && e.Code == types.ErrUserNotFound.Code {
				// User not found, so create them
				if err := s.userRepo.CreateUser(ctx, user); err != nil {
					return err
				}
			} else {
				// A different error occurred, so we should abort the sync
				return err
			}
		} else {
			// User exists, so update their attributes
			existingUser.Email = user.Email
			existingUser.Phone = user.Phone
			existingUser.Attributes = user.Attributes
			if err := s.userRepo.UpdateUser(ctx, existingUser); err != nil {
				return err
			}
		}
	}

	return nil
}
