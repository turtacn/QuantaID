package ldap

import (
	"context"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"
)

func (s *Server) HandleBind(ctx context.Context, messageID int64, req *ber.Packet) *ber.Packet {
	// BindRequest ::= [APPLICATION 0] SEQUENCE {
	//     version                 INTEGER (1 .. 127),
	//     name                    LDAPDN,
	//     authentication          AuthenticationChoice }
	// AuthenticationChoice ::= CHOICE {
	//     simple                  [0] OCTET STRING,
	//     ... }

	if len(req.Children) < 3 {
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultProtocolError, "", "Invalid BindRequest")
	}

	version := req.Children[0].Value.(int64)
	dn := req.Children[1].Value.(string)
	authChoice := req.Children[2]

	s.logger.Debug("Bind Request", zap.Int64("version", version), zap.String("dn", dn))

	if version != 3 {
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultProtocolError, "", "Only LDAPv3 supported")
	}

	if authChoice.Tag == 0 { // Simple Bind
		password := ""
		if authChoice.Value != nil {
			if p, ok := authChoice.Value.(string); ok {
				password = p
			} else if p, ok := authChoice.Value.([]byte); ok {
				password = string(p)
			}
		} else if len(authChoice.Data.Bytes()) > 0 {
			password = string(authChoice.Data.Bytes())
		}
		return s.doSimpleBind(ctx, messageID, dn, password)
	}

	// SASL not fully implemented yet, or other types
	return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultAuthMethodNotSupported, "", "Unsupported auth method")
}

func (s *Server) doSimpleBind(ctx context.Context, messageID int64, dn string, passwordStr string) *ber.Packet {
	// Parse DN to find username
	// Assuming DN format: uid=username,ou=users,dc=example,dc=com
	// or cn=username,ou=users...
	// For simplicity, we extract the first component value if it is uid or cn.

	// If DN is empty, it's anonymous bind.
	if dn == "" {
		// Anonymous bind allowed? Let's say yes for now, or configurable.
		// If password is also empty.
		if passwordStr == "" {
			return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultSuccess, "", "")
		}
	}

	parsedDN, err := ldap.ParseDN(dn)
	if err != nil || len(parsedDN.RDNs) == 0 {
		s.logger.Warn("Bind failed: invalid DN", zap.String("dn", dn), zap.Error(err))
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultInvalidCredentials, "", "Invalid DN")
	}

	// Extract the first RDN. Assuming uid=username or cn=username
	// For AD/LDAP, it's usually the first component of the DN.
	firstRDN := parsedDN.RDNs[0]
	if len(firstRDN.Attributes) == 0 {
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultInvalidCredentials, "", "Invalid RDN")
	}

	val := firstRDN.Attributes[0].Value

	// Lookup user
	user, err := s.userService.GetUserByUsername(ctx, val)
	if err != nil {
		// Differentiate not found vs error? To be safe, just invalid credentials
		s.logger.Warn("Bind failed: user not found", zap.String("username", val), zap.Error(err))
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultInvalidCredentials, "", "Invalid Credentials")
	}

	// Verify password
	valid, err := s.pwdService.Verify(ctx, user.ID, passwordStr)
	if err != nil {
		s.logger.Warn("Bind failed: verify error", zap.String("username", val), zap.Error(err))
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultInvalidCredentials, "", "Invalid Credentials")
	}
	if !valid {
		s.logger.Warn("Bind failed: invalid password", zap.String("username", val), zap.String("password_received", passwordStr))
		return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultInvalidCredentials, "", "Invalid Credentials")
	}

	return encodeLDAPResult(messageID, ApplicationBindResponse, LDAPResultSuccess, "", "")
}
