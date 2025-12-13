package ldap

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/password"
	"go.uber.org/zap"
)

type Server struct {
	addr        string
	baseDN      string
	tlsConfig   *tls.Config
	listener    net.Listener
	userService identity.IService
	pwdService  password.IService
	virtualTree *VirtualTree
	logger      *zap.Logger
	wg          sync.WaitGroup
	quit        chan struct{}
}

func NewServer(addr string, baseDN string, tlsConfig *tls.Config, userService identity.IService, pwdService password.IService, logger *zap.Logger) *Server {
	vt := NewVirtualTree(baseDN, userService)
	return &Server{
		addr:        addr,
		baseDN:      baseDN,
		tlsConfig:   tlsConfig,
		userService: userService,
		pwdService:  pwdService,
		virtualTree: vt,
		logger:      logger,
		quit:        make(chan struct{}),
	}
}

func (s *Server) Start() error {
	var l net.Listener
	var err error

	if s.tlsConfig != nil {
		l, err = tls.Listen("tcp", s.addr, s.tlsConfig)
	} else {
		l, err = net.Listen("tcp", s.addr)
	}

	if err != nil {
		return err
	}
	s.listener = l
	s.logger.Info("LDAP server started", zap.String("addr", s.addr), zap.String("baseDN", s.baseDN), zap.Bool("tls", s.tlsConfig != nil))

	s.wg.Add(1)
	go s.serve()
	return nil
}

func (s *Server) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				s.logger.Error("Accept error", zap.Error(err))
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	s.logger.Debug("New LDAP connection", zap.String("remote", conn.RemoteAddr().String()))

	for {
		// Read BER packet
		packet, err := ber.ReadPacket(conn)
		if err != nil {
			// EOF or error
			if err.Error() != "EOF" {
				s.logger.Debug("ReadPacket error", zap.Error(err))
			}
			return
		}

		// LDAP Message: Sequence { MessageID, ProtocolOp, ... }
		if len(packet.Children) < 2 {
			s.logger.Warn("Invalid LDAP packet: too few children")
			return
		}

		messageID, ok := packet.Children[0].Value.(int64)
		if !ok {
			s.logger.Warn("Invalid LDAP packet: MessageID not integer")
			return
		}

		protocolOp := packet.Children[1]
		// ProtocolOp tag is Application class
		if protocolOp.ClassType != ber.ClassApplication {
			s.logger.Warn("Invalid LDAP packet: ProtocolOp not Application class")
			return
		}

		ctx := context.Background() // In real app, maybe with timeout

		var resp *ber.Packet

		switch protocolOp.Tag {
		case ApplicationBindRequest:
			resp = s.HandleBind(ctx, messageID, protocolOp)
		case ApplicationSearchRequest:
			resp = s.HandleSearch(ctx, messageID, protocolOp)
		case ApplicationUnbindRequest:
			// No response needed, just close
			return
		default:
			s.logger.Warn("Unsupported LDAP operation", zap.Uint64("tag", uint64(protocolOp.Tag)))
			// Send UnwillingToPerform or similar
			resp = encodeLDAPResult(messageID, ApplicationExtendedResponse, LDAPResultUnwillingToPerform, "", "Operation not supported")
		}

		if resp != nil {
			// Check if it's our multi-response hack
			if resp.Tag == 0 && resp.ClassType == ber.ClassContext && resp.Description == "MultiResponse" {
				s.logger.Debug("Writing MultiResponse", zap.Int("count", len(resp.Children)))
				for i, child := range resp.Children {
					s.logger.Debug("Writing child packet", zap.Int("index", i), zap.Uint64("tag", uint64(child.Tag)))
					if _, err := conn.Write(child.Bytes()); err != nil {
						s.logger.Error("Write response error", zap.Error(err))
						return
					}
				}
			} else {
				s.logger.Debug("Writing SingleResponse", zap.Uint64("tag", uint64(resp.Tag)))
				if _, err := conn.Write(resp.Bytes()); err != nil {
					s.logger.Error("Write response error", zap.Error(err))
					return
				}
			}
		}
	}
}
