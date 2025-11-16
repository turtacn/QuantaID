package ldap

import (
	"context"
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
)

type LDAPClientInterface interface {
	Connect() error
	Search(ctx context.Context, baseDN string, filter string, attributes []string) ([]*ldap.Entry, error)
	SearchPaged(ctx context.Context, baseDN string, filter string, pageSize uint32, cookie string) ([]*ldap.Entry, string, error)
	PersistentSearch(ctx context.Context, baseDN, filter string, attributes []string) (*ldap.SearchResult, error)
	Close()
}

type LDAPClient struct {
	conn *ldap.Conn
	url  string
	user string
	pass string
}

func NewLDAPClient(url, user, pass string) *LDAPClient {
	return &LDAPClient{
		url:  url,
		user: user,
		pass: pass,
	}
}

func (c *LDAPClient) Connect() error {
	l, err := ldap.DialURL(c.url, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	if err != nil {
		return err
	}
	c.conn = l
	return c.conn.Bind(c.user, c.pass)
}

func (c *LDAPClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *LDAPClient) Search(ctx context.Context, baseDN string, filter string, attributes []string) ([]*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil,
	)
	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	return sr.Entries, nil
}

func (c *LDAPClient) SearchPaged(ctx context.Context, baseDN string, filter string, pageSize uint32, cookie string) ([]*ldap.Entry, string, error) {
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"*"},
		[]ldap.Control{ldap.NewControlPaging(pageSize)},
	)
	if cookie != "" {
		searchRequest.Controls[0].(*ldap.ControlPaging).SetCookie([]byte(cookie))
	}

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, "", err
	}

	var newCookie string
	for _, ctrl := range sr.Controls {
		if pagingCtrl, ok := ctrl.(*ldap.ControlPaging); ok {
			newCookie = string(pagingCtrl.Cookie)
			break
		}
	}

	return sr.Entries, newCookie, nil
}

func (c *LDAPClient) PersistentSearch(ctx context.Context, baseDN, filter string, attributes []string) (*ldap.SearchResult, error) {
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil,
	)
	return c.conn.Search(searchRequest)
}
