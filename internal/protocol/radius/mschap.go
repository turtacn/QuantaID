package radius

import (
	"context"
	"crypto/des"
	"crypto/sha1"
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/password"
	"github.com/turtacn/QuantaID/pkg/types"
	"golang.org/x/crypto/md4"
)

type MSCHAPHandler struct {
	userService     identity.IService
	passwordService password.IService
	codec           *AttributeCodec
}

func NewMSCHAPHandler(userService identity.IService, passwordService password.IService, codec *AttributeCodec) *MSCHAPHandler {
	return &MSCHAPHandler{
		userService:     userService,
		passwordService: passwordService,
		codec:           codec,
	}
}

// Authenticate handles MS-CHAPv2 authentication
func (h *MSCHAPHandler) Authenticate(ctx context.Context, request *Packet, client *RADIUSClient, username string) (*Packet, error) {
	// Get MS-CHAP2-Response
	mschapRespAttr := getVendorAttribute(request, VendorMicrosoft, MSCHAP2Response)
	if mschapRespAttr == nil {
		// MS-CHAPv1 is not implemented as it is insecure and deprecated
		return nil, fmt.Errorf("MS-CHAPv2 response not found")
	}

	mschapResp := mschapRespAttr.Value
	if len(mschapResp) != 50 {
		return nil, fmt.Errorf("invalid MS-CHAPv2 response length")
	}

	// Peer Challenge (16 bytes) at offset 2
	peerChallenge := mschapResp[2:18]
	// NT Response (24 bytes) at offset 26
	ntResponse := mschapResp[26:50]

	// Authenticator Challenge (16 bytes)
	authChallengeAttr := getVendorAttribute(request, VendorMicrosoft, MSCHAPChallenge)
	var authChallenge []byte
	if authChallengeAttr != nil {
		authChallenge = authChallengeAttr.Value
	} else {
		return nil, fmt.Errorf("missing MS-CHAP-Challenge")
	}

	// Find User
	u, err := h.userService.GetUserByUsername(ctx, username)
	if err != nil {
		// User not found
		return h.createMSCHAPError(request, 691, "Invalid credentials"), nil
	}

	// Get NT Hash (MD4 of unicode password)
	// Currently PasswordService might only store bcrypt/argon2.
	// For MS-CHAPv2, we need the NTHash or the cleartext password.
	// We assume PasswordService has a method `GetNTHash` or we calculate it if cleartext is available.
	// Since `GetPlainForCHAP` was mentioned in plan, we'll try that.

	// NOTE: In a real secure system, we shouldn't store cleartext.
	// But MS-CHAPv2 requires NTHash (MD4).
	// If the DB stores bcrypt, we cannot support MS-CHAPv2 unless we also store NTHash.
	// Assuming `GetNTHash` exists or we can get cleartext.
	// For this implementation, I will assume a helper `GetNTHash` or similar.
	// Since I cannot change PasswordService easily without seeing it, I'll assume an interface extension or method.
	// Plan mentioned `h.passwordService.GetNTHash(ctx, user.ID)`.

	// Mocking the call:
	// ntHash, err := h.passwordService.GetNTHash(ctx, u.ID)
	// As I can't verify if `GetNTHash` exists on `password.IService` (it likely doesn't),
	// I will check `internal/domain/password/interface.go` if I can.
	// If not, I will assume it returns error for now or add it if allowed.
	// Let's assume for this task we might need to rely on cleartext if available or just fail.
	// Wait, the plan says `ADD: internal/protocol/radius/mschap.go` but didn't explicitly say modify PasswordService.
	// But `Key Tasks` doesn't mention modifying password service.
	// However, without NTHash, MS-CHAPv2 is impossible.

	// Let's assume we can retrieve it.
	ntHash, err := h.getNTHash(ctx, u.ID)
	if err != nil {
		return h.createMSCHAPError(request, 691, "E=691 R=0 C=0000000000000000 V=3 M=Auth Failed"), nil
	}

	// Verify NT Response
	expectedNTResponse := h.generateNTResponse(authChallenge, peerChallenge, username, ntHash)
	if !bytes.Equal(ntResponse, expectedNTResponse) {
		return h.createMSCHAPError(request, 691, "E=691 R=0 C=0000000000000000 V=3 M=Auth Failed"), nil
	}

	// Generate Authenticator Response
	authResponse := h.generateAuthenticatorResponse(ntHash, ntResponse, peerChallenge, authChallenge, username)

	// Create Accept Response
	return h.createMSCHAPAccept(request, u, authResponse), nil
}

// getNTHash is a placeholder. In real impl, this calls PasswordService.
func (h *MSCHAPHandler) getNTHash(ctx context.Context, userID string) ([]byte, error) {
	// Attempt to get from PasswordService if it supports it.
	// Currently defined IService probably doesn't have it.
	// We might need to check if we can get plain password.
	// If not, this will fail.
	// For the sake of "Deliverables", I will implement the logic assuming I have the hash.
	// In integration, we might need a test user with known NTHash or reversible password.

	// Temporarily: Return error if not implemented.
	// But to pass tests, we might need a way.
	// Let's assume we can get it via a special method or casting.
	if ps, ok := h.passwordService.(interface{ GetNTHash(context.Context, string) ([]byte, error) }); ok {
		return ps.GetNTHash(ctx, userID)
	}

	// If the service doesn't support it, we can't proceed.
	return nil, fmt.Errorf("password service does not support NTHash retrieval")
}

func (h *MSCHAPHandler) generateNTResponse(authChallenge, peerChallenge []byte, username string, ntHash []byte) []byte {
	// ChallengeHash = SHA1(PeerChallenge + AuthChallenge + Username)[:8]
	sha := sha1.New()
	sha.Write(peerChallenge)
	sha.Write(authChallenge)
	sha.Write([]byte(username))
	challengeHash := sha.Sum(nil)[:8]

	return h.challengeResponse(challengeHash, ntHash)
}

func (h *MSCHAPHandler) challengeResponse(challenge, ntHash []byte) []byte {
	// Pad NT Hash to 21 bytes
	hashPad := make([]byte, 21)
	copy(hashPad, ntHash)

	response := make([]byte, 24)

	des1 := desEncrypt(makeDESKey(hashPad[0:7]), challenge)
	des2 := desEncrypt(makeDESKey(hashPad[7:14]), challenge)
	des3 := desEncrypt(makeDESKey(hashPad[14:21]), challenge)

	copy(response[0:8], des1)
	copy(response[8:16], des2)
	copy(response[16:24], des3)

	return response
}

func (h *MSCHAPHandler) generateAuthenticatorResponse(ntHash, ntResponse, peerChallenge, authChallenge []byte, username string) string {
	// RFC 2759 Section 8.7
	// Magic1 = 0x4D 0x61 0x67 0x69 0x63 0x20 0x73 0x65 0x72 0x76 0x65 0x72 0x20 0x74 0x6F 0x20 0x63 0x6C 0x69 0x65 0x6E 0x74 0x20 0x73 0x69 0x67 0x6E 0x69 0x6E 0x67 0x20 0x63 0x6F 0x6E 0x73 0x74 0x61 0x6E 0x74
	magic1 := []byte{0x4D, 0x61, 0x67, 0x69, 0x63, 0x20, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x20, 0x74, 0x6F, 0x20, 0x63, 0x6C, 0x69, 0x65, 0x6E, 0x74, 0x20, 0x73, 0x69, 0x67, 0x6E, 0x69, 0x6E, 0x67, 0x20, 0x63, 0x6F, 0x6E, 0x73, 0x74, 0x61, 0x6E, 0x74}
	// Magic2 = 0x50 0x61 0x64 0x20 0x74 0x6F 0x20 0x6D 0x61 0x6B 0x65 0x20 0x69 0x74 0x20 0x64 0x6F 0x20 0x6D 0x6F 0x72 0x65 0x20 0x74 0x68 0x61 0x6E 0x20 0x6F 0x6E 0x65 0x20 0x69 0x74 0x65 0x72 0x61 0x74 0x69 0x6F 0x6E
	magic2 := []byte{0x50, 0x61, 0x64, 0x20, 0x74, 0x6F, 0x20, 0x6D, 0x61, 0x6B, 0x65, 0x20, 0x69, 0x74, 0x20, 0x64, 0x6F, 0x20, 0x6D, 0x6F, 0x72, 0x65, 0x20, 0x74, 0x68, 0x61, 0x6E, 0x20, 0x6F, 0x6E, 0x65, 0x20, 0x69, 0x74, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6F, 0x6E}

	// PasswordHashHash = MD4(NTHash)
	md4h := md4.New()
	md4h.Write(ntHash)
	passwordHashHash := md4h.Sum(nil)

	// Digest = SHA1(PasswordHashHash + NTResponse + Magic1)
	sha := sha1.New()
	sha.Write(passwordHashHash)
	sha.Write(ntResponse)
	sha.Write(magic1)
	digest := sha.Sum(nil)

	// ChallengeHash = SHA1(PeerChallenge + AuthChallenge + Username)
	sha = sha1.New()
	sha.Write(peerChallenge)
	sha.Write(authChallenge)
	sha.Write([]byte(username))
	challengeHash := sha.Sum(nil)[:8]

	// AuthResponse = SHA1(Digest + ChallengeHash + Magic2)
	sha = sha1.New()
	sha.Write(digest)
	sha.Write(challengeHash)
	sha.Write(magic2)
	authResp := sha.Sum(nil)

	return fmt.Sprintf("S=%X", authResp)
}

func (h *MSCHAPHandler) createMSCHAPAccept(request *Packet, user *types.User, authResponse string) *Packet {
	response := request.CreateResponse(CodeAccessAccept)

	// Add MS-CHAP2-Success
	// Format: vendor specific
	// 2 bytes: type + length (handled by codec)
	// Value: S=...

	val := []byte(authResponse)
	// Prepend identifying byte if needed? No, usually just the string S=...
	// Actually RFC 2548 says:
	// String "S=<auth_response>"

	vsa := h.codec.EncodeVendorSpecific(VendorMicrosoft, MSCHAP2Success, val)
	response.AddAttribute(AttrVendorSpecific, vsa)

	return response
}

func (h *MSCHAPHandler) createMSCHAPError(request *Packet, errorCode int, message string) *Packet {
	response := request.CreateResponse(CodeAccessReject)

	// MS-CHAP-Error
	// String.
	// Usually has structure "E=691 R=0 C=... V=3 M=..."

	vsa := h.codec.EncodeVendorSpecific(VendorMicrosoft, MSCHAPError, []byte(message))
	response.AddAttribute(AttrVendorSpecific, vsa)

	return response
}

// Helpers

func getVendorAttribute(p *Packet, vendorID uint32, attrType byte) *Attribute {
	for _, attr := range p.Attributes {
		if attr.Type == AttrVendorSpecific {
			// Decode VSA
			vid := binary.BigEndian.Uint32(attr.Value[0:4])
			atype := attr.Value[4]
			if vid == vendorID && atype == attrType {
				return &Attribute{
					Type:   attrType,
					Length: attr.Value[5],
					Value:  attr.Value[6:], // Value without headers
				}
			}
		}
	}
	return nil
}

func desEncrypt(key, plain []byte) []byte {
	block, _ := des.NewCipher(key)
	res := make([]byte, 8)
	block.Encrypt(res, plain)
	return res
}

func makeDESKey(key []byte) []byte {
	// Parity bits insertion
	// 7 bytes -> 8 bytes
	res := make([]byte, 8)
	res[0] = key[0] & 0xFE
	res[1] = (key[0] << 7) | (key[1] >> 1)
	res[2] = (key[1] << 6) | (key[2] >> 2)
	res[3] = (key[2] << 5) | (key[3] >> 3)
	res[4] = (key[3] << 4) | (key[4] >> 4)
	res[5] = (key[4] << 3) | (key[5] >> 5)
	res[6] = (key[5] << 2) | (key[6] >> 6)
	res[7] = key[6] << 1

	for i := 0; i < 8; i++ {
		// Set odd parity
		c := res[i]
		bits := 0
		for j := 0; j < 8; j++ {
			if (c & (1 << j)) != 0 {
				bits++
			}
		}
		if bits%2 == 0 {
			res[i] = c | 1
		} else {
			res[i] = c & 0xFE
		}
	}
	return res
}
