package crypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/justenwalker/mack"
)

func TestHmacSha256(t *testing.T) {
	fn := func(key []byte, data []byte) bool {
		out1 := make([]byte, sha256.Size)
		out2 := make([]byte, sha256.Size)
		if err := HmacSha256(key, out1, data); err != nil {
			t.Errorf("HmacSha256: %v", err)
			return false
		}
		if err := HmacSha256Z(key, out2, data); err != nil {
			t.Errorf("HmacSha256Z: %v", err)
			return false
		}
		return hmac.Equal(out1, out2)
	}
	err := quick.Check(fn, &quick.Config{
		Values: func(values []reflect.Value, rand *rand.Rand) {
			key := make([]byte, sha256.Size)
			dataSz := rand.Int63n(65536)
			rand.Read(key)
			data := make([]byte, dataSz)
			rand.Read(data)
			values[0] = reflect.ValueOf(key)
			values[1] = reflect.ValueOf(data)
		},
	})
	if err != nil {
		t.Fatalf("HmacSha256 quick check failed: %v", err)
	}
}

func TestBindForRequestHmacSHA256(t *testing.T) {
	tm := mack.NewFromRaw(mack.Raw{
		ID:        []byte(`id`),
		Signature: []byte(`sig`),
	})
	tests := []struct {
		name      string
		sig       []byte
		expected  string
		expectErr bool
	}{
		{
			name:     "ok",
			sig:      []byte(`abcd1234abcd1234abcd1234abcd1234`),
			expected: "50b429aaf441cf35dc37ee0d1aaf522551e58b276dc968984d9ad37f402960e8",
		},
		{
			name:      "short",
			sig:       []byte(`sig`),
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := make([]byte, len(tt.sig))
			copy(sig, tt.sig)
			err := BindForRequestSHA256(&tm, sig)
			if err != nil && !tt.expectErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			actual := hex.EncodeToString(sig)
			if tt.expected != actual {
				t.Fatalf("sig does not match: expected: %v, actual: %v", tt.expected, actual)
			}
		})
	}
}

func TestBindForRequestSHA256(t *testing.T) {
	tests := []struct {
		name      string
		key       []byte
		sig       []byte
		expected  string
		expectErr bool
	}{
		{
			name:     "ok",
			key:      []byte(`00001111222233334444555566667777`),
			sig:      []byte(`abcd1234abcd1234abcd1234abcd1234`),
			expected: "f8aa005becfae3d7b6fd08ee19f518ca57aed80c2472fadfd1b0bc17117eb21e",
		},
		{
			name:      "short-sig",
			key:       []byte(`00001111222233334444555566667777`),
			sig:       []byte(`sig`),
			expectErr: true,
		},
		{
			name:      "short-key",
			key:       []byte(`key`),
			sig:       []byte(`sig`),
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := mack.NewFromRaw(mack.Raw{
				ID:        []byte(`id`),
				Signature: tt.sig,
			})
			sig := make([]byte, len(tt.sig))
			copy(sig, tt.sig)
			err := BindForRequestHmacSHA256(&tm, sig)
			if err != nil && !tt.expectErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			actual := hex.EncodeToString(sig)
			if tt.expected != actual {
				t.Fatalf("sig does not match: expected: %v, actual: %v", tt.expected, actual)
			}
		})
	}
}
