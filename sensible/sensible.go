// Package sensible exports a *macaroon.Scheme with sensible implementations for each cryptographic parameters.
//
// HMACScheme       : HMAC-SHA256
// EncryptionScheme : XSalsa20
// BindForRequest   : HMAC(M.sig, sig)
package sensible

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/justenwalker/mack/crypt"
	"github.com/justenwalker/mack/macaroon"
)

var (
	scheme     *macaroon.Scheme
	schemeOnce sync.Once
)

// Scheme constructs a macaroon.Scheme with sensible defaults.
func Scheme() *macaroon.Scheme {
	schemeOnce.Do(func() {
		var s Sensible
		var err error
		scheme, err = macaroon.NewScheme(macaroon.SchemeConfig{
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

func Encrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error) {
	return crypt.EncryptDecryptXSalsa20(out, in, nonce, key)
}

func Decrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error) {
	return crypt.EncryptDecryptXSalsa20(out, in, nonce, key)
}

func BindForRequest(ts *macaroon.Macaroon, sig []byte) error {
	return crypt.BindForRequestHmacSHA256(ts, sig)
}

type Sensible struct{}

func (Sensible) HMAC(key []byte, out []byte, data []byte) error {
	return HMAC(key, out, data)
}

func (Sensible) Overhead() int {
	return 0
}

func (Sensible) NonceSize() int {
	return crypt.XSalsa20NonceSize
}

func (Sensible) KeySize() int {
	return sha256.Size
}

func (Sensible) Encrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error) {
	return Encrypt(out, in, nonce, key)
}

func (Sensible) Decrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error) {
	return Decrypt(out, in, nonce, key)
}

func (Sensible) BindForRequest(ts *macaroon.Macaroon, sig []byte) error {
	return BindForRequest(ts, sig)
}

var (
	_ macaroon.HMACScheme           = Sensible{}
	_ macaroon.EncryptionScheme     = Sensible{}
	_ macaroon.BindForRequestScheme = Sensible{}
)
