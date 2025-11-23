package types

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/kms"
)

// GlobalKMS is the global KMS provider used by EncryptedString.
// It must be set at application startup.
var globalKMS kms.KeyProvider

// SetGlobalKMS sets the global KMS provider.
func SetGlobalKMS(k kms.KeyProvider) {
	globalKMS = k
}

// EncryptedString is a string that is encrypted when stored in the database
// and decrypted when read from the database.
type EncryptedString string

// Value implements the driver.Valuer interface.
// It encrypts the string using the global KMS and returns it as a base64 encoded string.
func (es EncryptedString) Value() (driver.Value, error) {
	if globalKMS == nil {
		return nil, errors.New("global KMS not initialized")
	}

	if es == "" {
		return "", nil
	}

	ciphertext, _, err := globalKMS.Encrypt(context.Background(), []byte(es))
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Scan implements the sql.Scanner interface.
// It decrypts the base64 encoded string using the global KMS.
func (es *EncryptedString) Scan(value interface{}) error {
	if globalKMS == nil {
		return errors.New("global KMS not initialized")
	}

	if value == nil {
		*es = ""
		return nil
	}

	var ciphertextBase64 string
	switch v := value.(type) {
	case string:
		ciphertextBase64 = v
	case []byte:
		ciphertextBase64 = string(v)
	default:
		return fmt.Errorf("failed to scan EncryptedString: unexpected type %T", value)
	}

	if ciphertextBase64 == "" {
		*es = ""
		return nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	plaintext, err := globalKMS.Decrypt(context.Background(), ciphertext)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	*es = EncryptedString(plaintext)
	return nil
}
