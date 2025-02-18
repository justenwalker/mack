package hmac_test

import (
	"bytes"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	myhmac "github.com/justenwalker/mack/crypt/internal/hmac"
)

func BenchmarkHMacSHA256(b *testing.B) {
	b.Run("zero_allocs", func(b *testing.B) {
		keyout := make([]byte, sha256.Size)
		_, _ = crand.Read(keyout)
		value := make([]byte, 1024)
		_, _ = crand.Read(value)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			myhmac.SHA256(keyout, keyout, value)
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		keyout := make([]byte, sha256.Size)
		_, _ = crand.Read(keyout)
		value := make([]byte, 1024)
		_, _ = crand.Read(value)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			stdlibHMAC(keyout, keyout, value)
		}
	})
	b.Run("stdlib-statickey-reset", func(b *testing.B) {
		keyout := make([]byte, sha256.Size)
		_, _ = crand.Read(keyout)
		value := make([]byte, 1024)
		_, _ = crand.Read(value)
		h := hmac.New(sha256.New, keyout)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			h.Reset()
			h.Write(value)
			h.Sum(keyout[:0])
		}
	})
}

func TestHMacSHA256_allocations(t *testing.T) {
	keyout := make([]byte, sha256.Size)
	value := make([]byte, 1024)
	allocs := testing.AllocsPerRun(10*1024, func() {
		myhmac.SHA256(keyout, keyout, value)
	})
	t.Logf("AllocsPerRun: %.2f", allocs)
	if allocs > 0 {
		t.Fatalf("allocations > 0: %.f", allocs)
	}
}

func FuzzHMacSHA256_compatibility(f *testing.F) {
	f.Add([]byte("key"), []byte("message"))
	f.Fuzz(func(t *testing.T, key []byte, message []byte) {
		if !hmacSHA256CompatibilityTest(key, message) {
			t.Fatalf("hmacCompatibilityTest(%x, %x) = false, want true", key, message)
		}
	})
}

func TestHMacSHA256_compatibility(t *testing.T) {
	if err := quick.Check(hmacSHA256CompatibilityTest, &quick.Config{
		Values: func(vs []reflect.Value, rand *rand.Rand) {
			keysz := rand.Int63n(16 * 1024)
			key := make([]byte, keysz)
			rand.Read(key)
			msgsz := rand.Int63n(10 * 1024 * 1024)
			message := make([]byte, msgsz)
			vs[0] = reflect.ValueOf(key)
			vs[1] = reflect.ValueOf(message)
		},
	}); err != nil {
		t.Fatalf("compatibility check failed: %v", err)
	}
}

func hmacSHA256CompatibilityTest(key []byte, message []byte) bool {
	out1 := make([]byte, sha256.Size)
	out2 := make([]byte, sha256.Size)
	myhmac.SHA256(key, out1[:0], message)
	stdlibHMAC(key, out2[:0], message)
	return bytes.Equal(out1, out2)
}

func stdlibHMAC(key []byte, out []byte, msgs ...[]byte) {
	h := hmac.New(sha256.New, key)
	for _, msg := range msgs {
		h.Write(msg)
	}
	h.Sum(out[:0])
}
