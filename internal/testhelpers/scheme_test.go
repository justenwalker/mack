package testhelpers

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

func TestEncryptDecrypt_quick(t *testing.T) {
	fn := func(key []byte, plaintext []byte) bool {
		in := plaintext
		out := make([]byte, len(plaintext))
		copy(out, in)
		err := testEncryptDecrypt(t, out, in, key)
		if err != nil {
			t.Error(err)
			return false
		}
		return true
	}
	err := quick.Check(fn, &quick.Config{
		Values: func(values []reflect.Value, rand *rand.Rand) {
			key := make([]byte, sha256.Size)
			dataSz := rand.Int63n(16)
			rand.Read(key)
			data := make([]byte, dataSz)
			rand.Read(data)
			values[0] = reflect.ValueOf(key)
			values[1] = reflect.ValueOf(data)
		},
	})
	if err != nil {
		t.Fatalf("TestEncryptDecrypt quick check failed: %v", err)
	}
}

func testEncryptDecrypt(tb testing.TB, out, in, key []byte) error {
	tb.Helper()
	ts := &testScheme{TB: tb}
	enc, err := ts.Encrypt(out, in, key)
	if err != nil {
		return fmt.Errorf("testEncryptDecrypt: Encrypt error %w", err)
	}
	if bytes.Equal(enc, in) {
		return errors.New("testEncryptDecrypt: plaintext = ciphertext")
	}
	dec := make([]byte, len(in))
	dec, err = ts.Decrypt(dec, enc, key)
	if err != nil {
		return fmt.Errorf("testEncryptDecrypt: Decrypt error %w", err)
	}
	if !bytes.Equal(dec, in) {
		return errors.New("testEncryptDecrypt: plaintext != decrypted")
	}
	return nil
}
