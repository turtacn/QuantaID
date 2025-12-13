package ldap

import (
	"context"

	ber "github.com/go-asn1-ber/asn1-ber"
)

// Handler interface for extensibility
type Handler interface {
	Handle(ctx context.Context, req *ber.Packet) *ber.Packet
}

// RequestHandler dispatches requests
type RequestHandler struct {
	server *Server
}
