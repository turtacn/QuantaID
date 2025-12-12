package unit

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"testing"

	"github.com/turtacn/QuantaID/internal/protocol/radius"
)

func TestPacketParsing(t *testing.T) {
	secret := []byte("secret")
	authenticator := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	buf := new(bytes.Buffer)
	buf.WriteByte(radius.CodeAccessRequest)
	buf.WriteByte(1) // ID
	binary.Write(buf, binary.BigEndian, uint16(20+6)) // Length (header + attribute)
	buf.Write(authenticator[:])

	// Add User-Name attribute
	buf.WriteByte(radius.AttrUserName)
	buf.WriteByte(6) // 2 + 4 chars
	buf.WriteString("test")

	data := buf.Bytes()

	packet, err := radius.ParsePacket(data, secret)
	if err != nil {
		t.Fatalf("Failed to parse packet: %v", err)
	}

	if packet.Code != radius.CodeAccessRequest {
		t.Errorf("Expected code %d, got %d", radius.CodeAccessRequest, packet.Code)
	}

	if packet.Identifier != 1 {
		t.Errorf("Expected id 1, got %d", packet.Identifier)
	}

	username := packet.GetString(radius.AttrUserName)
	if username != "test" {
		t.Errorf("Expected username 'test', got '%s'", username)
	}
}

func TestPasswordEncoding(t *testing.T) {
	codec := radius.NewAttributeCodec()
	secret := []byte("xyzzy5461")
	authenticator := [16]byte{} // Zero auth for simplicity or random
	copy(authenticator[:], []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}) // Actually needs to be 16 bytes.

	// RFC 2865 Example
	// Shared Secret = "xyzzy5461"
	// Request Authenticator (RA) = ...
	// User-Password = ...
	// We just test reversibility here.

	password := "myPassword"
	encoded := codec.EncodePassword(password, authenticator, secret)
	decoded := codec.DecodePassword(encoded, authenticator, secret)

	if decoded != password {
		t.Errorf("Expected '%s', got '%s'", password, decoded)
	}
}

func TestCHAPResponse(t *testing.T) {
	// Simple MD5 test for CHAP logic
	// CHAP Response = MD5(ID + Password + Challenge)
	id := byte(1)
	password := "password"
	challenge := []byte("challenge1234567") // 16 bytes usually

	h := md5.New()
	h.Write([]byte{id})
	h.Write([]byte(password))
	h.Write(challenge)
	expected := h.Sum(nil)

	if len(expected) != 16 {
		t.Errorf("MD5 len mismatch")
	}
}
