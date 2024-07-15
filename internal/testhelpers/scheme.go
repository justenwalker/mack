package testhelpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"testing"

	"golang.org/x/crypto/chacha20"

	"github.com/justenwalker/mack/macaroon"
)

type testScheme struct {
	TB    testing.TB
	debug bool
}

func NewScheme(tb testing.TB) *macaroon.Scheme {
	tb.Helper()
	ts := &testScheme{TB: tb}
	s, err := macaroon.NewScheme(macaroon.SchemeConfig{
		HMACScheme:           ts,
		EncryptionScheme:     ts,
		BindForRequestScheme: ts,
	})
	if err != nil {
		tb.Fatalf("NewScheme failed: %s", err)
	}
	return s
}

func (t *testScheme) BindForRequest(ts *macaroon.Macaroon, sig []byte) error {
	var (
		tsIDStr  = printableString(ts.ID())
		tsSigStr = printableString(ts.Signature())
		sigStr   = printableString(sig)
	)
	h := sha256.New()
	h.Write(ts.Signature())
	h.Write(sig)
	sig = h.Sum(sig[:0])
	if t.debug {
		t.TB.Logf("BindForRequest(\n\tID=%s,\n\tTARGET_SIG=%s,\n\tDISCHARGE_SIG=%s\n): %s", tsIDStr, tsSigStr, sigStr, printableString(sig))
	}
	return nil
}

func (t *testScheme) Overhead() int {
	return 0
}

func (t *testScheme) NonceSize() int {
	return chacha20.NonceSizeX
}

func (t *testScheme) Encrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error) {
	var (
		keyStr   = printableString(key)
		nonceStr = printableString(nonce)
		inStr    = printableString(in)
	)
	c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return nil, err
	}
	c.XORKeyStream(out, in)
	if t.debug {
		t.TB.Logf("Encrypt(\n\tKEY=%s,\n\tNONCE=%s,\n\tPLAINTEXT=%s\n): %s", keyStr, nonceStr, inStr, printableString(out))
	}
	return out, nil
}

func (t *testScheme) Decrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error) {
	var (
		keyStr   = printableString(key)
		nonceStr = printableString(nonce)
		inStr    = printableString(in)
	)
	c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return nil, err
	}
	c.XORKeyStream(out, in)
	if t.debug {
		t.TB.Logf("Decrypt(\n\tKEY=%s,\n\tNONCE=%s,\n\tCIPHERTEXT=%s\n): %s", keyStr, nonceStr, inStr, printableString(out))
	}
	return out, nil
}

func (t *testScheme) KeySize() int {
	return sha256.Size
}

func (t *testScheme) HMAC(key []byte, out []byte, data []byte) error {
	var (
		keyStr  = printableString(key)
		dataStr = printableString(data)
	)
	h := hmac.New(sha256.New, key)
	h.Write(data)
	out = h.Sum(out[:0])
	if t.debug {
		t.TB.Logf("HMAC-SHA256(\n\tKEY=%s,\n\tDATA=%s\n): %s", keyStr, dataStr, printableString(out))
	}
	return nil
}

var _ interface {
	macaroon.BindForRequestScheme
	macaroon.EncryptionScheme
	macaroon.HMACScheme
} = (*testScheme)(nil)
