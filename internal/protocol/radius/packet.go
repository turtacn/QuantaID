package radius

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

const (
	CodeAccessRequest      = 1
	CodeAccessAccept       = 2
	CodeAccessReject       = 3
	CodeAccountingRequest  = 4
	CodeAccountingResponse = 5
	CodeAccessChallenge    = 11

	MaxPacketSize = 4096
	HeaderSize    = 20
)

type Packet struct {
	Code          byte
	Identifier    byte
	Length        uint16
	Authenticator [16]byte
	Attributes    []Attribute
	Secret        []byte
}

type Attribute struct {
	Type   byte
	Length byte
	Value  []byte
}

// ParsePacket parses a raw RADIUS packet.
func ParsePacket(data []byte, secret []byte) (*Packet, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("packet too short")
	}

	packet := &Packet{
		Code:       data[0],
		Identifier: data[1],
		Length:     binary.BigEndian.Uint16(data[2:4]),
		Secret:     secret,
	}

	copy(packet.Authenticator[:], data[4:20])

	if int(packet.Length) > len(data) {
		return nil, fmt.Errorf("invalid length")
	}

	// Parse Attributes
	var err error
	packet.Attributes, err = parseAttributes(data[20:packet.Length])
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func parseAttributes(data []byte) ([]Attribute, error) {
	var attrs []Attribute
	offset := 0
	for offset < len(data) {
		if len(data)-offset < 2 {
			break
		}
		attrType := data[offset]
		attrLen := data[offset+1]

		if attrLen < 2 {
			return nil, fmt.Errorf("invalid attribute length")
		}
		if offset+int(attrLen) > len(data) {
			return nil, fmt.Errorf("attribute length overflow")
		}

		attr := Attribute{
			Type:   attrType,
			Length: attrLen,
			Value:  make([]byte, int(attrLen)-2),
		}
		copy(attr.Value, data[offset+2:offset+int(attrLen)])
		attrs = append(attrs, attr)
		offset += int(attrLen)
	}
	return attrs, nil
}

// Encode serializes the packet to bytes.
func (p *Packet) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(p.Code)
	buf.WriteByte(p.Identifier)

	// Placeholder for length
	binary.Write(buf, binary.BigEndian, uint16(0))
	buf.Write(p.Authenticator[:])

	for _, attr := range p.Attributes {
		buf.WriteByte(attr.Type)
		buf.WriteByte(attr.Length)
		buf.Write(attr.Value)
	}

	data := buf.Bytes()
	binary.BigEndian.PutUint16(data[2:4], uint16(len(data)))
	return data, nil
}

// GetAttribute returns the first attribute of the given type.
func (p *Packet) GetAttribute(attrType byte) *Attribute {
	for _, attr := range p.Attributes {
		if attr.Type == attrType {
			// Return a copy or pointer to the item in slice
			// Since we iterate by value, `attr` is a copy.
			// Let's return a pointer to a new struct to avoid issues if we modify it (though we shouldn't).
			return &attr
		}
	}
	return nil
}

// GetString returns the string value of the first attribute of the given type.
func (p *Packet) GetString(attrType byte) string {
	attr := p.GetAttribute(attrType)
	if attr == nil {
		return ""
	}
	return string(attr.Value)
}

// AddAttribute adds an attribute to the packet.
func (p *Packet) AddAttribute(attrType byte, value []byte) {
	attr := Attribute{
		Type:   attrType,
		Length: byte(2 + len(value)),
		Value:  value,
	}
	p.Attributes = append(p.Attributes, attr)
}

// CreateResponse creates a response packet corresponding to the request.
func (p *Packet) CreateResponse(code byte) *Packet {
	return &Packet{
		Code:       code,
		Identifier: p.Identifier,
		Secret:     p.Secret,
		// Authenticator needs to be calculated later or copied depending on logic.
		// Usually for response it is calculated.
	}
}

// CalculateResponseAuthenticator calculates the authenticator for a response packet.
func (p *Packet) CalculateResponseAuthenticator(requestAuth [16]byte) [16]byte {
	// ResponseAuth = MD5(Code+ID+Length+RequestAuth+Attributes+Secret)
	encoded, _ := p.Encode()
	// Replace the authenticator in the encoded bytes with the request's authenticator
	// This is required for the calculation per RFC 2865
	copy(encoded[4:20], requestAuth[:])

	h := md5.New()
	h.Write(encoded)
	h.Write(p.Secret)

	var auth [16]byte
	copy(auth[:], h.Sum(nil))
	return auth
}
