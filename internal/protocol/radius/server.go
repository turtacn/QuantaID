package radius

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ServerConfig struct {
	AuthPort      int
	AcctPort      int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	WorkerCount   int
	MaxPacketSize int
}

type Server struct {
	authConn      *net.UDPConn
	acctConn      *net.UDPConn
	authenticator *Authenticator
	accounting    *AccountingHandler
	clientManager *ClientManager
	proxy         *Proxy
	config        ServerConfig
	stopCh        chan struct{}
	wg            sync.WaitGroup
	logger        *zap.Logger
}

func NewServer(
	authenticator *Authenticator,
	accounting *AccountingHandler,
	clientManager *ClientManager,
	proxy *Proxy,
	config ServerConfig,
	logger *zap.Logger,
) *Server {
	if config.MaxPacketSize == 0 {
		config.MaxPacketSize = MaxPacketSize
	}
	return &Server{
		authenticator: authenticator,
		accounting:    accounting,
		clientManager: clientManager,
		proxy:         proxy,
		config:        config,
		logger:        logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Start Auth Port
	authAddr := &net.UDPAddr{Port: s.config.AuthPort}
	var err error
	s.authConn, err = net.ListenUDP("udp", authAddr)
	if err != nil {
		return fmt.Errorf("auth port listen failed: %w", err)
	}

	// Start Acct Port
	acctAddr := &net.UDPAddr{Port: s.config.AcctPort}
	s.acctConn, err = net.ListenUDP("udp", acctAddr)
	if err != nil {
		s.authConn.Close()
		return fmt.Errorf("acct port listen failed: %w", err)
	}

	s.stopCh = make(chan struct{})

	// Start Workers
	for i := 0; i < s.config.WorkerCount; i++ {
		s.wg.Add(2)
		go s.authWorker(ctx, i)
		go s.acctWorker(ctx, i)
	}

	s.logger.Info("RADIUS server started",
		zap.Int("auth_port", s.config.AuthPort),
		zap.Int("acct_port", s.config.AcctPort))
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	close(s.stopCh)
	if s.authConn != nil {
		s.authConn.Close()
	}
	if s.acctConn != nil {
		s.acctConn.Close()
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) authWorker(ctx context.Context, id int) {
	defer s.wg.Done()
	buf := make([]byte, s.config.MaxPacketSize)

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			// s.authConn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
			// Removing strict ReadDeadline loop or handling it carefully to avoid CPU spin on timeout if checking stopCh
			// Or just set it before Read.

			// Using a short deadline to allow checking stopCh periodically?
			// Or just rely on Close() to unblock Read.
			// Standard pattern: ReadFromUDP blocks. Close unblocks it returning error.

			n, remoteAddr, err := s.authConn.ReadFromUDP(buf)
			if err != nil {
				// Check if closed
				select {
				case <-s.stopCh:
					return
				default:
				}
				s.logger.Debug("read error", zap.Error(err))
				continue
			}

			// Process packet in a goroutine? Or inline?
			// If we process inline, this worker is blocked.
			// With multiple workers reading from same conn (is that safe? UDPConn is thread-safe).
			// We are passing buf slice. We need to copy it if we go async.
			// But here we are "worker", so we process it.

			data := make([]byte, n)
			copy(data, buf[:n])

			s.handleAuthRequest(ctx, data, remoteAddr)
		}
	}
}

func (s *Server) acctWorker(ctx context.Context, id int) {
	defer s.wg.Done()
	buf := make([]byte, s.config.MaxPacketSize)

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			n, remoteAddr, err := s.acctConn.ReadFromUDP(buf)
			if err != nil {
				select {
				case <-s.stopCh:
					return
				default:
				}
				s.logger.Debug("read error", zap.Error(err))
				continue
			}

			data := make([]byte, n)
			copy(data, buf[:n])

			s.handleAcctRequest(ctx, data, remoteAddr)
		}
	}
}

func (s *Server) handleAuthRequest(ctx context.Context, data []byte, addr *net.UDPAddr) {
	client, err := s.clientManager.GetByIP(ctx, addr.IP.String())
	if err != nil || client == nil {
		s.logger.Warn("unknown NAS client", zap.String("ip", addr.IP.String()))
		return
	}

	packet, err := ParsePacket(data, []byte(client.Secret))
	if err != nil {
		s.logger.Error("packet parse error", zap.Error(err))
		return
	}

	// Should check Message-Authenticator here if required?
	// RFC 3579 requires Message-Authenticator for EAP. We aren't doing EAP yet.

	if packet.Code != CodeAccessRequest {
		return
	}

	var response *Packet

	// Proxy check
	// if s.proxy != nil && s.proxy.ShouldProxy(packet) { ... }

	if s.authenticator != nil {
		response, err = s.authenticator.Authenticate(ctx, packet, client)
		if err != nil {
			s.logger.Error("auth error", zap.Error(err))
			response = s.createRejectResponse(packet, "Internal error")
		}
	} else {
		response = s.createRejectResponse(packet, "Server not configured")
	}

	// Calculate Response Authenticator
	response.Authenticator = response.CalculateResponseAuthenticator(packet.Authenticator)

	respData, err := response.Encode()
	if err != nil {
		s.logger.Error("response encode error", zap.Error(err))
		return
	}

	s.authConn.WriteToUDP(respData, addr)
}

func (s *Server) handleAcctRequest(ctx context.Context, data []byte, addr *net.UDPAddr) {
	client, err := s.clientManager.GetByIP(ctx, addr.IP.String())
	if err != nil || client == nil {
		s.logger.Warn("unknown NAS client", zap.String("ip", addr.IP.String()))
		return
	}

	packet, err := ParsePacket(data, []byte(client.Secret))
	if err != nil {
		s.logger.Error("packet parse error", zap.Error(err))
		return
	}

	if packet.Code != CodeAccountingRequest {
		return
	}

	var response *Packet
	if s.accounting != nil {
		response, err = s.accounting.Handle(ctx, packet, client, addr)
		if err != nil {
			s.logger.Error("accounting error", zap.Error(err))
			return // Drop or send error? RADIUS accounting doesn't usually send NAK, just doesn't Ack.
		}
	}

	if response != nil {
		response.Authenticator = response.CalculateResponseAuthenticator(packet.Authenticator)
		respData, _ := response.Encode()
		s.acctConn.WriteToUDP(respData, addr)
	}
}

func (s *Server) createRejectResponse(request *Packet, message string) *Packet {
	response := request.CreateResponse(CodeAccessReject)
	response.AddAttribute(AttrReplyMessage, []byte(message))
	return response
}
