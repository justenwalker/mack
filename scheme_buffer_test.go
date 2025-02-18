package mack

import (
	"crypto/sha256"
	"testing"
)

func BenchmarkGetClearBuffer(b *testing.B) {
	scheme := testHelperNoopScheme(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		bufp := scheme.getKeyBuffer()
		scheme.releaseKeyBuffer(bufp)
	}
}

func TestGetClearBuffer(t *testing.T) {
	scheme := testHelperNoopScheme(t)
	bufp := scheme.getKeyBuffer()
	if bufp == nil {
		t.Fatalf("unexpected nil buffer")
	}
	if len(*bufp) != sha256.Size {
		t.Fatalf("unexpected key size; got=%d, want=%d", len(*bufp), sha256.Size)
	}
	// verify all zeros
	for i := range *bufp {
		if (*bufp)[i] != 0 {
			t.Fatalf("unexpected non-zero byte at index %d: %x", i, *bufp)
		}
	}
	// Set the key to something
	for i := range *bufp {
		(*bufp)[i] = 0xcf
	}
	scheme.releaseKeyBuffer(bufp)
	// Get the buffer again
	bufp = scheme.getKeyBuffer()
	// verify all zeros after release and re-acquire
	for i := range *bufp {
		if (*bufp)[i] != 0 {
			t.Fatalf("unexpected non-zero byte at index %d: %x", i, *bufp)
		}
	}
}

func testHelperNoopScheme(tb testing.TB) *Scheme {
	tb.Helper()
	scheme, err := NewScheme(SchemeConfig{
		HMACScheme:           noopScheme{},
		EncryptionScheme:     noopScheme{},
		BindForRequestScheme: noopScheme{},
	})
	if err != nil {
		tb.Fatalf("failed to make noop scheme: %v", err)
	}
	return scheme
}

var _ interface {
	BindForRequestScheme
	EncryptionScheme
	HMACScheme
} = noopScheme{}

type noopScheme struct{}

func (n noopScheme) HMAC(_ []byte, _ []byte, _ []byte) error {
	panic("unimplemented")
}

func (n noopScheme) Overhead() int {
	return 0
}

func (n noopScheme) NonceSize() int {
	return 24
}

func (n noopScheme) KeySize() int {
	return 32
}

func (n noopScheme) Encrypt(_ []byte, _ []byte, _ []byte) ([]byte, error) {
	panic("implement me")
}

func (n noopScheme) Decrypt(_ []byte, _ []byte, _ []byte) ([]byte, error) {
	panic("implement me")
}

func (n noopScheme) BindForRequest(_ *Macaroon, _ []byte) error {
	panic("implement me")
}
