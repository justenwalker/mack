package crypt

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func TestXSalsa20_quick(t *testing.T) {
	fn := func(key []byte, nonce []byte, plaintext []byte) bool {
		in := plaintext
		out := make([]byte, len(plaintext))
		copy(out, in)
		err := testEncryptDecryptXSalsa20(out, in, nonce, key)
		if err != nil {
			t.Error(err)
			return false
		}
		return true
	}
	err := quick.Check(fn, &quick.Config{
		Values: func(values []reflect.Value, rand *rand.Rand) {
			key := make([]byte, sha256.Size)
			nonce := make([]byte, XSalsa20NonceSize)
			dataSz := rand.Int63n(65536)
			rand.Read(key)
			rand.Read(nonce)
			data := make([]byte, dataSz)
			rand.Read(data)
			values[0] = reflect.ValueOf(key)
			values[1] = reflect.ValueOf(nonce)
			values[2] = reflect.ValueOf(data)
		},
	})
	if err != nil {
		t.Fatalf("HmacSha256 quick check failed: %v", err)
	}
}

func TestXSSalsa20(t *testing.T) {
	tests := []struct {
		name      string
		in        []byte
		out       []byte
		nonce     []byte
		key       []byte
		expectErr bool
	}{
		{
			name:      "ok",
			in:        []byte(`hello`),
			out:       []byte(`00000`),
			nonce:     []byte(`abcd1234abcd1234abcd1234`),
			key:       []byte(`00001111222233334444555566667777`),
			expectErr: false,
		},
		{
			name:      "ok-noout",
			in:        []byte(`hello`),
			nonce:     []byte(`abcd1234abcd1234abcd1234`),
			key:       []byte(`00001111222233334444555566667777`),
			expectErr: false,
		},
		{
			name:      "short-key",
			in:        []byte(`hello`),
			out:       []byte(`00000`),
			nonce:     []byte(`abcd1234abcd1234abcd1234`),
			key:       []byte(`abc`),
			expectErr: true,
		},
		{
			name:      "short-nonce",
			in:        []byte(`hello`),
			out:       []byte(`00000`),
			nonce:     []byte(`abc`),
			key:       []byte(`00001111222233334444555566667777`),
			expectErr: true,
		},
		{
			name:      "long-key",
			in:        []byte(`hello`),
			out:       []byte(`00000`),
			nonce:     []byte(`abcd1234abcd1234abcd1234`),
			key:       []byte(`0000111122223333444455556666777700001111222233334444555566667777`),
			expectErr: true,
		},
		{
			name:      "long-nonce",
			in:        []byte(`hello`),
			out:       []byte(`00000`),
			key:       []byte(`00001111222233334444555566667777`),
			nonce:     []byte(`abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234`),
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testEncryptDecryptXSalsa20(tt.out, tt.in, tt.nonce, tt.key)
			if err != nil && !tt.expectErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
		})
	}
}

func testEncryptDecryptXSalsa20(out, in, nonce, key []byte) error {
	enc, err := EncryptDecryptXSalsa20(out, in, nonce, key)
	if err != nil {
		return fmt.Errorf("EncryptDecryptXSalsa20: Encrypt error %w", err)
	}
	if bytes.Equal(enc, in) {
		return errors.New("EncryptDecryptXSalsa20: plaintext = ciphertext")
	}
	dec := make([]byte, len(in))
	dec, err = EncryptDecryptXSalsa20(dec, enc, nonce, key)
	if err != nil {
		return fmt.Errorf("EncryptDecryptXSalsa20: Decrypt error %w", err)
	}
	if !bytes.Equal(dec, in) {
		return errors.New("EncryptDecryptXSalsa20: plaintext != decrypted")
	}
	return nil
}
