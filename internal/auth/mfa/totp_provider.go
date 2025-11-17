package mfa

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/types"
)

type TOTPProvider struct {
	issuer string
	repo   *postgresql.PostgresMFARepository
	// crypto   CryptoService // TODO: Implement crypto service
}

func (tp *TOTPProvider) Enroll(ctx context.Context, userID string, params EnrollParams) (*EnrollResult, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tp.issuer,
		AccountName: params.Email,
		SecretSize:  32,
		Algorithm:   otp.AlgorithmSHA256,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return nil, err
	}

	qrCode, err := tp.generateQRCode(key.URL())
	if err != nil {
		return nil, err
	}

	// encryptedSecret, err := tp.crypto.Encrypt(key.Secret())
	// if err != nil {
	// 	return nil, err
	// }

	backupCodes := tp.generateBackupCodes(10)

	return &EnrollResult{
		CredentialID:  uuid.New().String(),
		Secret:        key.Secret(), // TODO: use encryptedSecret
		QRCodeImage:   qrCode,
		BackupCodes:   backupCodes,
		SetupURL:      key.URL(),
	}, nil
}

func (tp *TOTPProvider) Verify(ctx context.Context, userID string, credential string) (bool, error) {
	factors, err := tp.repo.GetUserFactorsByType(ctx, types.MustParseUUID(userID), string(MFATypeTOTP))
	if err != nil {
		return false, err
	}
	if len(factors) == 0 {
		return false, fmt.Errorf("no totp factor enrolled")
	}

	// TODO: Decrypt secret
	// secret, err := tp.crypto.Decrypt(factors[0].Secret)
	// if err != nil {
	// 	return false, err
	// }
	secret := factors[0].Secret

	valid := totp.Validate(credential, secret)
	if !valid {
		// TODO: verify backup code
	}

	return valid, nil
}

func (tp *TOTPProvider) Revoke(ctx context.Context, userID string, credentialID string) error {
	// TODO:
	return nil
}

func (tp *TOTPProvider) generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = fmt.Sprintf("%08d", rand.Intn(100000000))
	}
	return codes
}

func (tp *TOTPProvider) generateQRCode(url string) (string, error) {
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(png), nil
}
