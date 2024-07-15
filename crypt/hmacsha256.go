package crypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"

	myhmac "github.com/justenwalker/mack/crypt/internal/hmac"
	"github.com/justenwalker/mack/macaroon"
)

// HmacSha256Z implements the HMAC function using SHA-256 with zero allocations.
func HmacSha256Z(key []byte, out []byte, data []byte) error {
	myhmac.SHA256(key, out[:0], data)
	return nil
}

// HmacSha256 implements the HMAC function using SHA-256 and the standard library crypto/hmac.
func HmacSha256(key []byte, out []byte, data []byte) error {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	h.Sum(out[:0])
	return nil
}

// BindForRequestHmacSHA256 implements BindForRequest by using HMAC-SHA256.
// sig = HMacSHA256(tm.Sig, sig).
func BindForRequestHmacSHA256(tm *macaroon.Macaroon, sig []byte) error {
	if len(sig) < sha256.Size {
		return errors.New("sig too short, must be at least 32 bytes")
	}
	myhmac.SHA256(tm.Signature(), sig[0:], sig)
	return nil
}

// BindForRequestSHA256 implements BindForRequest by using SHA256.
// sig = SHA256(sig :: tm.Sig).
func BindForRequestSHA256(tm *macaroon.Macaroon, sig []byte) error {
	if len(sig) < sha256.Size {
		return errors.New("sig too short, must be at least 32 bytes")
	}
	h := sha256.New()
	h.Write(sig)
	h.Write(tm.Signature())
	h.Sum(sig[:0])
	return nil
}
