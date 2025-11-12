package ldap

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-ldap/ldap/v3"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"time"
)

// TODO: Implement a connection pool for the LDAP connector.
// The current implementation uses a single, long-lived connection, which is not ideal for production use.
// A connection pool would provide better performance and reliability.
type LDAPConnector struct {
	config *LDAPConfig
	conn   *ldap.Conn
	logger utils.Logger
	mapper *Mapper
}

func NewLDAPConnector(cfg *LDAPConfig) (*LDAPConnector, error) {
	var conn *ldap.Conn
	operation := func() error {
		var err error
		conn, err = ldap.DialURL(fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port))
		return err
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 30 * time.Second

	err := backoff.Retry(operation, expBackoff)
	if err != nil {
		return nil, fmt.Errorf("ldap dial: %w", err)
	}

	if cfg.UseTLS {
		if err := conn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			return nil, fmt.Errorf("ldap starttls: %w", err)
		}
	}

	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		return nil, fmt.Errorf("ldap bind: %w", err)
	}

	return &LDAPConnector{
		config: cfg,
		conn:   conn,
		mapper: NewMapper(cfg.AttrMapping),
		logger: utils.NewNoopLogger(),
	}, nil
}

func (lc *LDAPConnector) Name() string {
	return "ldap-connector"
}

func (lc *LDAPConnector) Type() types.PluginType {
	return types.PluginTypeIdentityConnector
}

func (lc *LDAPConnector) Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error {
	lc.logger = logger
	return nil
}

func (lc *LDAPConnector) Start(ctx context.Context) error {
	return nil
}

func (lc *LDAPConnector) Stop(ctx context.Context) error {
	lc.conn.Close()
	return nil
}

func (lc *LDAPConnector) HealthCheck(ctx context.Context) error {
	_, err := lc.conn.Search(ldap.NewSearchRequest(
		"",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		[]string{"dn"},
		nil,
	))
	return err
}

func (lc *LDAPConnector) Authenticate(ctx context.Context, credentials map[string]string) (*types.AuthResponse, error) {
	username, ok := credentials["username"]
	if !ok {
		return nil, types.ErrBadRequest.WithDetails(map[string]string{"field": "username"})
	}
	password, ok := credentials["password"]
	if !ok {
		return nil, types.ErrBadRequest.WithDetails(map[string]string{"field": "password"})
	}

	searchRequest := ldap.NewSearchRequest(
		lc.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&%s(%s=%s))", lc.config.UserFilter, lc.config.AttrMapping["username"], username),
		lc.mapper.GetAttributeList(),
		nil,
	)

	sr, err := lc.conn.Search(searchRequest)
	if err != nil {
		lc.logger.Error(ctx, "LDAP search failed", zap.Error(err))
		return nil, types.ErrInternal
	}
	if len(sr.Entries) == 0 {
		return &types.AuthResponse{Success: false, Error: types.ErrUserNotFound}, nil
	}

	userDN := sr.Entries[0].DN

	// Attempt to bind with the user's credentials using the existing connection
	userBindErr := lc.conn.Bind(userDN, password)

	// Immediately re-bind as the admin to restore the connection for the next operation
	if rebindErr := lc.conn.Bind(lc.config.BindDN, lc.config.BindPassword); rebindErr != nil {
		lc.logger.Error(ctx, "Failed to re-bind as admin after user authentication attempt", zap.Error(rebindErr))
		// This is a critical failure, the connection is now in a bad state
		return nil, types.ErrInternal
	}

	// Now, check the result of the user's bind attempt
	if userBindErr != nil {
		return &types.AuthResponse{Success: false, Error: types.ErrInvalidCredentials}, nil
	}

	// If the user bind was successful, map the user attributes and return
	user, err := lc.mapper.MapEntryToUser(sr.Entries[0])
	if err != nil {
		lc.logger.Error(ctx, "Failed to map user attributes", zap.Error(err))
		return nil, types.ErrInternal
	}

	return &types.AuthResponse{
		Success: true,
		User:    user,
	}, nil
}

func (lc *LDAPConnector) GetUser(ctx context.Context, identifier string) (*types.User, error) {
	searchRequest := ldap.NewSearchRequest(
		lc.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&%s(%s=%s))", lc.config.UserFilter, lc.config.AttrMapping["username"], identifier),
		lc.mapper.GetAttributeList(),
		nil,
	)

	sr, err := lc.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) == 0 {
		return nil, types.ErrUserNotFound
	}

	return lc.mapper.MapEntryToUser(sr.Entries[0])
}

func (lc *LDAPConnector) GetGroup(ctx context.Context, identifier string) (*types.UserGroup, error) {
	return nil, fmt.Errorf("not implemented")
}

func (lc *LDAPConnector) SearchUsers(ctx context.Context, filter string) ([]*types.User, error) {
	searchRequest := ldap.NewSearchRequest(
		lc.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		lc.mapper.GetAttributeList(),
		nil,
	)

	sr, err := lc.conn.SearchWithPaging(searchRequest, 100)
	if err != nil {
		return nil, err
	}

	var users []*types.User
	for _, entry := range sr.Entries {
		user, err := lc.mapper.MapEntryToUser(entry)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (lc *LDAPConnector) SyncUsers(ctx context.Context) ([]*types.User, error) {
	return lc.SearchUsers(ctx, lc.config.UserFilter)
}

func (lc *LDAPConnector) SyncGroups(ctx context.Context) ([]*types.UserGroup, error) {
	return nil, fmt.Errorf("not implemented")
}
