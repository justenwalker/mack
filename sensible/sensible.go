// Package sensible exports a *mack.Scheme with sensible implementations for each cryptographic parameters.
//
// HMACScheme       : HMAC-SHA256
// EncryptionScheme : AES-256-GCM
// BindForRequest   : HMAC(M.sig, sig)
package sensible

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/crypt"
)

var (
	scheme     *mack.Scheme
	schemeOnce sync.Once
)

const (
	gcmStandardNonceSize = 12
	gcmTagSize           = 16
)

// Scheme constructs a mack.Scheme with sensible defaults.
func Scheme() *mack.Scheme {
	schemeOnce.Do(func() {
		var s Sensible
		var err error
		scheme, err = mack.NewScheme(mack.SchemeConfig{
			HMACScheme:           s,
			EncryptionScheme:     s,
			BindForRequestScheme: s,
		})
		if err != nil {
			panic(fmt.Errorf("sensible.Scheme: should not fail to construct. %w", err))
		}
	})
	return scheme
}

func HMAC(key []byte, out []byte, data []byte) error {
	return crypt.HmacSha256Z(key, out, data)
}

func BindForRequest(ts *mack.Macaroon, sig []byte) error {
	return crypt.BindForRequestHmacSHA256(ts, sig)
}

type Sensible struct{}

func (Sensible) HMAC(key []byte, out []byte, data []byte) error {
	return HMAC(key, out, data)
}

func (Sensible) Overhead() int {
	return gcmTagSize + gcmStandardNonceSize
}

func (Sensible) KeySize() int {
	return sha256.Size
}

func (Sensible) Encrypt(out []byte, in []byte, key []byte) ([]byte, error) {
	return Encrypt(out, in, key)
}

func (Sensible) Decrypt(out []byte, in []byte, key []byte) ([]byte, error) {
	return Decrypt(out, in, key)
}

func (Sensible) BindForRequest(ts *mack.Macaroon, sig []byte) error {
	return BindForRequest(ts, sig)
}

// Encrypt encrypts the plaintext using AES-GCM-256 with a randomly generated nonce.
//
// To reuse plaintext's storage for the encrypted output, use plaintext[:0]
// as dst. Otherwise, the remaining capacity of dst must not overlap plaintext.
// dst and additionalData may not overlap.
func Encrypt(dst []byte, plaintext []byte, key []byte) ([]byte, error) {
	return encryptFunc(dst, plaintext, key)
}

// Decrypt decrypts a previously encrypted plaintext produced by Encrypt.
//
// To reuse ciphertext's storage for the decrypted output, use ciphertext[:0]
// as dst. Otherwise, the remaining capacity of dst must not overlap ciphertext.
// dst and additionalData may not overlap.
//
// Even if the function fails, the contents of dst, up to its capacity,
// may be overwritten.
func Decrypt(dst []byte, ciphertext []byte, key []byte) ([]byte, error) {
	return decryptFunc(dst, ciphertext, key)
}

var (
	_ mack.HMACScheme           = Sensible{}
	_ mack.EncryptionScheme     = Sensible{}
	_ mack.BindForRequestScheme = Sensible{}
)
