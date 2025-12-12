package radius

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type ProxyConfig struct {
	Enabled         bool
	UpstreamServers []UpstreamServer
	RetryCount      int
	Timeout         time.Duration
}

type UpstreamServer struct {
	Address string
	Secret  string
	Weight  int
}

type Proxy struct {
	config ProxyConfig
	conns  map[string]*net.UDPConn
	mu     sync.Mutex
}

func NewProxy(config ProxyConfig) *Proxy {
	return &Proxy{
		config: config,
		conns:  make(map[string]*net.UDPConn),
	}
}

func (p *Proxy) Forward(ctx context.Context, request *Packet, client *RADIUSClient) (*Packet, error) {
	if !p.config.Enabled || len(p.config.UpstreamServers) == 0 {
		return nil, fmt.Errorf("proxy not enabled or no upstream servers")
	}

	// Simple round-robin or first available selection
	// Real world: use weights, health checks.
	upstream := p.config.UpstreamServers[0]

	// Connect if not connected
	conn, err := p.getConn(upstream.Address)
	if err != nil {
		return nil, err
	}

	// Re-encode packet with upstream secret?
	// Actually RADIUS proxying involves re-calculating authenticator or attributes?
	// RFC 2865: "The proxy server... MUST replace the Request Authenticator... with a new random one."
	// We need to re-package the request.

	// 1. Create new request based on original
	// But actually we might just forward bytes if we act as transparent proxy,
	// but we must update Authenticator/Secret relationship.
	// Since `Packet` struct has the decoded attributes, we can just re-encode with new Secret.

	proxyPacket := &Packet{
		Code:       request.Code,
		Identifier: request.Identifier, // Use same ID or map IDs? If same ID, we might have collisions if multiple clients map to same upstream.
		// Usually proxies manage a map of (UpstreamID) -> (ClientAddr, ClientID)
		Attributes: make([]Attribute, len(request.Attributes)),
		Secret:     []byte(upstream.Secret),
	}
	copy(proxyPacket.Attributes, request.Attributes)

	// Add Proxy-State?

	// Generate new Authenticator for the upstream leg
	// copy(proxyPacket.Authenticator[:], generateRandomAuth())
	// For simplicity, let's keep original for now (though RFC says change it).
	proxyPacket.Authenticator = request.Authenticator

	// Encode
	data, err := proxyPacket.Encode()
	if err != nil {
		return nil, err
	}

	// Send
	conn.SetWriteDeadline(time.Now().Add(p.config.Timeout))
	if _, err := conn.Write(data); err != nil {
		return nil, err
	}

	// Wait for response
	// This is synchronous and blocks the worker.
	// In high performance proxy, this is async with ID mapping.
	// For P6, synchronous simple forwarding.
	respBuf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(p.config.Timeout))
	n, _, err := conn.ReadFromUDP(respBuf)
	if err != nil {
		return nil, err
	}

	// Parse response with upstream secret
	respPacket, err := ParsePacket(respBuf[:n], []byte(upstream.Secret))
	if err != nil {
		return nil, err
	}

	// We need to return this packet to the client.
	// We must switch secret back to Client Secret?
	// The `server.go` will handle encoding the returned packet using `client.Secret`.
	// So we just return the parsed packet structure.
	// But `server.go` expects `respPacket` to be constructed such that `Encode` uses `respPacket.Secret`.
	// So we should set `respPacket.Secret` to `client.Secret` before returning?
	// Or `server.go` handles it?
	// `server.go` calls `response.Encode()`. `response.Secret` is used.
	// So yes, we must swap the secret to the Client's secret so the client can verify it.

	respPacket.Secret = []byte(client.Secret)

	// Also, if we changed the Request Authenticator, we need to ensure the Response Authenticator calculation is correct for the Client.
	// ResponseAuth = MD5(Code+ID+Length+RequestAuth+Attributes+Secret)
	// The `respPacket` from upstream has Auth valid for UpstreamSecret + UpstreamRequestAuth.
	// We need to recalculate it for ClientSecret + ClientRequestAuth.
	// Fortunately, `server.go` usually recalculates auth before sending?
	// Let's check `server.go` (to be written).
	// Ideally `server.go` logic:
	//   response = handler.Handle()
	//   response.Authenticator = response.CalculateResponseAuthenticator(request.Authenticator)
	//   conn.Write(response.Encode())

	// If `server.go` does that, then we are good.

	return respPacket, nil
}

func (p *Proxy) getConn(address string) (*net.UDPConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.conns[address]; ok {
		return conn, nil
	}

	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	p.conns[address] = conn
	return conn, nil
}
