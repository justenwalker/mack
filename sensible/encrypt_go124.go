//go:build go1.24

package sensible

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

var (
	decryptFunc = decryptGo124
	encryptFunc = encryptGo124
)

func encryptGo124(dst []byte, plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("sensible.Encrypt: failed to create aes cipher: %w", err)
	}
	a, err := cipher.NewGCMWithRandomNonce(c)
	if err != nil {
		return nil, fmt.Errorf("sensible.Encrypt: failed to create new aes-gcm cipher: %w", err)
	}
	return a.Seal(dst, nil, plaintext, nil), nil
}

func decryptGo124(dst []byte, ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("sensible.Decrypt: failed to create aes cipher: %w", err)
	}
	a, err := cipher.NewGCMWithRandomNonce(c)
	if err != nil {
		return nil, fmt.Errorf("sensible.Decrypt: failed to create new aes-gcm cipher: %w", err)
	}
	return a.Open(dst, nil, ciphertext, nil)
}
