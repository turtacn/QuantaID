package totp

import (
    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
)

type TOTPProvider struct{}

func (p *TOTPProvider) GenerateSecret(issuer, accountName string) (*otp.Key, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      issuer,      // "QuantaID"
        AccountName: accountName, // user email
        SecretSize:  32,
    })
    if err != nil {
        return nil, err
    }
    return key, nil
}

func (p *TOTPProvider) VerifyCode(secret, code string) bool {
    return totp.Validate(code, secret)
}

func (p *TOTPProvider) GenerateQRCodeURL(key *otp.Key) string {
    return key.URL() // otpauth://totp/QuantaID:user@example.com?secret=...&issuer=QuantaID
}
