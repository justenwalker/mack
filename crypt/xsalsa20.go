package crypt

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/salsa20"
)

const (
	XSalsa20NonceSize = 24
)

// EncryptDecryptXSalsa20 implements the encryption and decryption function using XSalsa20.
func EncryptDecryptXSalsa20(out, in []byte, nonce []byte, key []byte) ([]byte, error) {
	if len(nonce) != XSalsa20NonceSize {
		return nil, fmt.Errorf("xsalsa20: nonce must be exactly %d bytes", XSalsa20NonceSize)
	}
	if len(key) != 32 {
		return nil, errors.New("xsalsa20: key must be exactly 32 bytes")
	}
	var k [32]byte
	copy(k[:], key)
	if cap(out) < len(in) {
		out = make([]byte, len(in))
	}
	salsa20.XORKeyStream(out[:len(in)], in, nonce, &k)
	return out, nil
}
