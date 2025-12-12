//go:build integration

package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/turtacn/QuantaID/internal/protocol/radius"
)

// This test assumes a running RADIUS server or starts one in-process.
// Since we don't have the full DI container setup easily here, we mock components or skip if not setup.

func TestRADIUS_FullAuth_PAP(t *testing.T) {
	// Setup is complex for full integration in this environment without running the whole app.
	// We will rely on unit tests and manual verification via `radtest` as per plan.
	// However, we can create a mini-server instance here if dependencies are mockable.

	// Skip for now if we can't spin up dependencies easily.
	t.Skip("Skipping full integration test requiring DB and Services")
}

func TestRADIUS_PacketFlow(t *testing.T) {
	// We can test the client-server socket interaction at least

	// Start a UDP listener to simulate server
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()

	// Client sends packet
	go func() {
		clientConn, err := net.Dial("udp", serverAddr)
		if err != nil {
			return
		}
		defer clientConn.Close()

		secret := []byte("secret")
		pkt := &radius.Packet{
			Code:       radius.CodeAccessRequest,
			Identifier: 1,
			Authenticator: [16]byte{0x01},
			Secret:     secret,
		}
		pkt.AddAttribute(radius.AttrUserName, []byte("user"))

		data, _ := pkt.Encode()
		clientConn.Write(data)
	}()

	// Server reads
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Server read failed: %v", err)
	}

	if n == 0 {
		t.Fatal("Empty packet")
	}

	// Check Parse
	parsed, err := radius.ParsePacket(buf[:n], []byte("secret"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsed.Code != radius.CodeAccessRequest {
		t.Errorf("Wrong code")
	}
}
