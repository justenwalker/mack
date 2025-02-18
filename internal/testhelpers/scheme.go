package testhelpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/justenwalker/mack"
)

type testScheme struct {
	TB    testing.TB
	debug bool
}

const (
	gcmStandardNonceSize = 12
	gcmTagSize           = 16
)

func NewScheme(tb testing.TB) *mack.Scheme {
	tb.Helper()
	ts := &testScheme{TB: tb}
	s, err := mack.NewScheme(mack.SchemeConfig{
		HMACScheme:           ts,
		EncryptionScheme:     ts,
		BindForRequestScheme: ts,
	})
	if err != nil {
		tb.Fatalf("NewScheme failed: %s", err)
	}
	return s
}

func (t *testScheme) BindForRequest(ts *mack.Macaroon, sig []byte) error {
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
	return gcmTagSize + gcmStandardNonceSize
}

func (t *testScheme) NonceSize() int {
	return gcmStandardNonceSize
}

func (t *testScheme) Encrypt(out []byte, in []byte, key []byte) ([]byte, error) {
	var (
		keyStr = printableString(key)
		inStr  = printableString(in)
	)
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	sz := len(in) + gcmTagSize + gcmStandardNonceSize
	buf, _ := sliceForAppend(out, sz)
	copy(buf[gcmStandardNonceSize:], in)
	in = buf[gcmStandardNonceSize : gcmStandardNonceSize+len(in)]
	nonce := buf[:gcmStandardNonceSize]
	_, _ = ReadRandom(nonce)
	nonceStr := printableString(nonce)
	sealed := gcm.Seal(buf[gcmStandardNonceSize:gcmStandardNonceSize], nonce, in, nil)
	if t.debug {
		t.TB.Logf("Encrypt(\n\tKEY=%s,\n\tNONCE=%s,\n\tPLAINTEXT=%s\n): %s", keyStr, nonceStr, inStr, printableString(out))
	}
	return buf[:len(sealed)+gcmStandardNonceSize], nil
}

func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}

func (t *testScheme) Decrypt(out []byte, in []byte, key []byte) ([]byte, error) {
	var (
		keyStr = printableString(key)
		inStr  = printableString(in)
	)
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	if len(in) < gcmTagSize+gcmStandardNonceSize {
		return nil, fmt.Errorf("ciphertext too short: %s", inStr)
	}
	nonce := in[:gcmStandardNonceSize]
	plaintext := in[gcmStandardNonceSize:]
	sz := len(in) - gcmTagSize - gcmStandardNonceSize
	if cap(out) < sz {
		out = make([]byte, 0, sz)
	}
	nonceStr := printableString(nonce)
	out, err = gcm.Open(out[:0], nonce, plaintext, nil)
	if err != nil {
		return nil, err
	}
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
	mack.BindForRequestScheme
	mack.EncryptionScheme
	mack.HMACScheme
} = (*testScheme)(nil)
