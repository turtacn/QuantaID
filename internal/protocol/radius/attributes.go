package radius

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

// Standard Attribute Types (RFC 2865)
const (
	AttrUserName          = 1
	AttrUserPassword      = 2
	AttrCHAPPassword      = 3
	AttrNASIPAddress      = 4
	AttrNASPort           = 5
	AttrServiceType       = 6
	AttrFramedProtocol    = 7
	AttrFramedIPAddress   = 8
	AttrFramedIPNetmask   = 9
	AttrFramedRouting     = 10
	AttrFilterId          = 11
	AttrFramedMTU         = 12
	AttrFramedCompression = 13
	AttrLoginIPHost       = 14
	AttrLoginService      = 15
	AttrLoginTCPPort      = 16
	AttrReplyMessage      = 18
	AttrCallbackNumber    = 19
	AttrCallbackId        = 20
	AttrFramedRoute       = 22
	AttrFramedIPXNetwork  = 23
	AttrState             = 24
	AttrClass             = 25
	AttrVendorSpecific    = 26
	AttrSessionTimeout    = 27
	AttrIdleTimeout       = 28
	AttrTerminationAction = 29
	AttrCalledStationId   = 30
	AttrCallingStationId  = 31
	AttrNASIdentifier     = 32
	AttrProxyState        = 33
	AttrLoginLATService   = 34
	AttrLoginLATNode      = 35
	AttrLoginLATGroup     = 36
	AttrFramedAppleTalkLink    = 37
	AttrFramedAppleTalkNetwork = 38
	AttrFramedAppleTalkZone    = 39
	AttrCHAPChallenge     = 60
	AttrNASPortType       = 61
	AttrPortLimit         = 62
	AttrLoginLATPort      = 63
	AttrTunnelType        = 64
	AttrTunnelMediumType  = 65
	AttrTunnelClientEndpoint = 66
	AttrTunnelServerEndpoint = 67
	AttrConnectInfo       = 77
	AttrEAPMessage        = 79
	AttrMessageAuthenticator = 80
	AttrTunnelPrivateGroupID = 81
	AttrNASPortId         = 87
)

// Accounting Attributes (RFC 2866)
const (
	AttrAcctStatusType    = 40
	AttrAcctDelayTime     = 41
	AttrAcctInputOctets   = 42
	AttrAcctOutputOctets  = 43
	AttrAcctSessionId     = 44
	AttrAcctAuthentic     = 45
	AttrAcctSessionTime   = 46
	AttrAcctInputPackets  = 47
	AttrAcctOutputPackets = 48
	AttrAcctTerminateCause = 49
	AttrAcctMultiSessionId = 50
	AttrAcctLinkCount     = 51
	AttrAcctInputGigawords = 52
	AttrAcctOutputGigawords = 53
)

// Accounting Status Types
const (
	AcctStatusStart       = 1
	AcctStatusStop        = 2
	AcctStatusInterimUpdate = 3
	AcctStatusAccountingOn = 7
	AcctStatusAccountingOff = 8
)

// MS-CHAP Attributes (RFC 2548)
const (
	VendorMicrosoft       = 311
	MSCHAPChallenge       = 11
	MSCHAPResponse        = 1
	MSCHAP2Response       = 25
	MSCHAP2Success        = 26
	MSCHAPError           = 2
	MSCHAPDomain          = 10
	MSCHAPMPPEKeys        = 12
	MSMPPESendKey         = 16
	MSMPPERecvKey         = 17
	MSMPPEEncryptionPolicy = 7
	MSMPPEEncryptionTypes = 8
)

type AttrValueType int

const (
	AttrTypeString AttrValueType = iota
	AttrTypeInteger
	AttrTypeIPAddr
	AttrTypeDate
	AttrTypeOctets
)

type AttrDefinition struct {
	Name    string
	Type    AttrValueType
	Encrypt bool // Whether to encrypt (like User-Password)
}

type AttributeCodec struct {
	dictionary map[byte]AttrDefinition
	vendors    map[uint32]map[byte]AttrDefinition
}

// NewAttributeCodec creates a new attribute codec with standard definitions.
func NewAttributeCodec() *AttributeCodec {
	c := &AttributeCodec{
		dictionary: make(map[byte]AttrDefinition),
		vendors:    make(map[uint32]map[byte]AttrDefinition),
	}
	c.loadStandardAttributes()
	c.loadMicrosoftAttributes()
	return c
}

func (c *AttributeCodec) loadStandardAttributes() {
	c.dictionary = map[byte]AttrDefinition{
		AttrUserName:         {Name: "User-Name", Type: AttrTypeString},
		AttrUserPassword:     {Name: "User-Password", Type: AttrTypeString, Encrypt: true},
		AttrCHAPPassword:     {Name: "CHAP-Password", Type: AttrTypeOctets},
		AttrNASIPAddress:     {Name: "NAS-IP-Address", Type: AttrTypeIPAddr},
		AttrNASPort:          {Name: "NAS-Port", Type: AttrTypeInteger},
		AttrServiceType:      {Name: "Service-Type", Type: AttrTypeInteger},
		AttrFramedIPAddress:  {Name: "Framed-IP-Address", Type: AttrTypeIPAddr},
		AttrReplyMessage:     {Name: "Reply-Message", Type: AttrTypeString},
		AttrState:            {Name: "State", Type: AttrTypeOctets},
		AttrClass:            {Name: "Class", Type: AttrTypeOctets},
		AttrSessionTimeout:   {Name: "Session-Timeout", Type: AttrTypeInteger},
		AttrIdleTimeout:      {Name: "Idle-Timeout", Type: AttrTypeInteger},
		AttrNASIdentifier:    {Name: "NAS-Identifier", Type: AttrTypeString},
		AttrCHAPChallenge:    {Name: "CHAP-Challenge", Type: AttrTypeOctets},
		AttrNASPortType:      {Name: "NAS-Port-Type", Type: AttrTypeInteger},
		AttrAcctStatusType:   {Name: "Acct-Status-Type", Type: AttrTypeInteger},
		AttrAcctSessionId:    {Name: "Acct-Session-Id", Type: AttrTypeString},
		AttrAcctSessionTime:  {Name: "Acct-Session-Time", Type: AttrTypeInteger},
	}
}

func (c *AttributeCodec) loadMicrosoftAttributes() {
	// Microsoft Vendor attributes could be added here if we need to parse them by name.
	// For now, we often handle VSA manually or by ID.
	c.vendors[VendorMicrosoft] = map[byte]AttrDefinition{
		MSCHAPChallenge: {Name: "MS-CHAP-Challenge", Type: AttrTypeOctets},
		MSCHAP2Response: {Name: "MS-CHAP2-Response", Type: AttrTypeOctets},
		MSCHAP2Success:  {Name: "MS-CHAP2-Success", Type: AttrTypeOctets},
		MSCHAPError:     {Name: "MS-CHAP-Error", Type: AttrTypeString},
	}
}

// EncodePassword encodes the User-Password attribute according to RFC 2865.
func (c *AttributeCodec) EncodePassword(password string, authenticator [16]byte, secret []byte) []byte {
	pass := []byte(password)

	// Pad to multiple of 16
	if len(pass) == 0 {
		pass = make([]byte, 16)
	} else if len(pass)%16 != 0 {
		padLen := 16 - (len(pass) % 16)
		pass = append(pass, make([]byte, padLen)...)
	}

	result := make([]byte, len(pass))
	prevBlock := authenticator[:]

	for i := 0; i < len(pass); i += 16 {
		h := md5.New()
		h.Write(secret)
		h.Write(prevBlock)
		hash := h.Sum(nil)

		for j := 0; j < 16; j++ {
			result[i+j] = pass[i+j] ^ hash[j]
		}
		prevBlock = result[i : i+16]
	}
	return result
}

// DecodePassword decodes the User-Password attribute.
func (c *AttributeCodec) DecodePassword(encoded []byte, authenticator [16]byte, secret []byte) string {
	if len(encoded) == 0 || len(encoded)%16 != 0 {
		return ""
	}

	result := make([]byte, len(encoded))
	prevBlock := authenticator[:]

	for i := 0; i < len(encoded); i += 16 {
		h := md5.New()
		h.Write(secret)
		h.Write(prevBlock)
		hash := h.Sum(nil)

		for j := 0; j < 16; j++ {
			result[i+j] = encoded[i+j] ^ hash[j]
		}
		prevBlock = encoded[i : i+16]
	}

	return string(bytes.TrimRight(result, "\x00"))
}

// EncodeVendorSpecific creates a Vendor-Specific Attribute (VSA) value.
func (c *AttributeCodec) EncodeVendorSpecific(vendorID uint32, attrType byte, value []byte) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, vendorID)
	buf.WriteByte(attrType)
	buf.WriteByte(byte(2 + len(value)))
	buf.Write(value)
	return buf.Bytes()
}

// DecodeVendorSpecific decodes a VSA value.
func (c *AttributeCodec) DecodeVendorSpecific(data []byte) (vendorID uint32, attrType byte, value []byte, err error) {
	if len(data) < 6 {
		return 0, 0, nil, fmt.Errorf("invalid VSA length")
	}
	vendorID = binary.BigEndian.Uint32(data[0:4])
	attrType = data[4]
	vsaLen := data[5]

	if int(vsaLen) > len(data)-4 {
		return 0, 0, nil, fmt.Errorf("invalid VSA internal length")
	}

	// VSA Value starts at offset 6 (4 byte vendor ID + 1 byte type + 1 byte len)
	// But wait, standard says: Vendor-Id (4), Vendor-Type (1), Vendor-Length (1), Attribute-Specific...
	// The `data` passed here is the Value part of the RADIUS Attribute (Type 26).
	// So data[0:4] is Vendor ID.
	// data[4] is Vendor Type.
	// data[5] is Vendor Length.
	// value is data[6:4+vsaLen]
	// Note: vsaLen includes the type and length fields (2 bytes), so actual value length is vsaLen - 2.

	valStart := 6
	valEnd := 4 + int(vsaLen) // 4 bytes offset for vendor ID + vsaLen

	if valEnd > len(data) {
		 return 0, 0, nil, fmt.Errorf("VSA length out of bounds")
	}

	value = data[valStart:valEnd]
	return vendorID, attrType, value, nil
}
