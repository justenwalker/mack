package sensible

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
)

// testEncryptDecrypt helps make sure that the implementation of AES-GCM-256
// both works, and its outputs match the reference implementation in the standard lib.
func testEncryptDecrypt(t *testing.T, tc encryptionTestConfig) {
	t.Helper()
	tests := []struct {
		name  string
		check func(v *encryptionTestCase) (result encResult, err error)
	}{
		{
			name:  "testNil",
			check: testNil,
		},
		{
			name:  "testOverlap",
			check: testOverlap,
		},
		{
			name:  "testInsufficientSize",
			check: testInsufficientSize,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := quickCheckFunc(t, tc, tt.check)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

type encryptionTestConfig struct {
	encrypt func(dst []byte, plaintext []byte, key []byte) ([]byte, error)
	decrypt func(dst []byte, ciphertext []byte, key []byte) ([]byte, error)
}

func quickCheckFunc(t *testing.T, tc encryptionTestConfig, check func(tc *encryptionTestCase) (result encResult, err error)) error {
	t.Helper()
	return quick.Check(func(keyHex string, dataHex string) bool {
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			t.Fatalf("hex.DecodeString(key:%s): %testCase", keyHex, err)
			return false
		}
		data, err := hex.DecodeString(dataHex)
		if err != nil {
			t.Fatalf("hex.DecodeString(data:%s): err: %testCase", dataHex, err)
			return false
		}
		testCase := &encryptionTestCase{
			key:     key,
			data:    data,
			encrypt: tc.encrypt,
			decrypt: tc.decrypt,
		}
		result, err := check(testCase)
		if err != nil {
			t.Errorf("implementation failed: key=%s, data=%s, err=%testCase", keyHex, dataHex, err)
			return false
		}
		assertReferenceResult(t, testCase, &result)
		if err != nil {
			t.Errorf("reference result failed: key=%s, data=%s, err=%testCase", keyHex, dataHex, err)
			return false
		}
		return true
	}, &quick.Config{
		Values: func(values []reflect.Value, rand *rand.Rand) {
			key := make([]byte, 32)
			rand.Read(key)
			dataSz := int(rand.Int63n(32))
			data := make([]byte, dataSz)
			rand.Read(data)
			values[0] = reflect.ValueOf(hex.EncodeToString(key))
			values[1] = reflect.ValueOf(hex.EncodeToString(data))
		},
	})
}

type encryptionTestCase struct {
	key     []byte
	data    []byte
	encrypt func(dst []byte, plaintext []byte, key []byte) ([]byte, error)
	decrypt func(dst []byte, ciphertext []byte, key []byte) ([]byte, error)
}

func assertReferenceResult(t *testing.T, v *encryptionTestCase, result *encResult) {
	t.Helper()
	c, err := aes.NewCipher(v.key)
	if err != nil {
		t.Fatalf("aes.NewCipher(key): %v", err)
	}
	a, err := cipher.NewGCM(c)
	if err != nil {
		t.Fatalf("cipher.NewGCM(c): %v", err)
	}
	ciphertext := a.Seal(nil, result.nonce, result.plaintext, nil)
	if diff := cmp.Diff(hex.EncodeToString(ciphertext), hex.EncodeToString(result.ciphertext)); diff != "" {
		t.Errorf("gcm.Seal(nil,%s,%s,nil) mismatch (-want +got):\n%s", hex.EncodeToString(result.nonce), hex.EncodeToString(result.plaintext), diff)
	}

	plaintext, err := a.Open(nil, result.nonce, result.ciphertext, nil)
	if err != nil {
		t.Fatalf("gcm.Open(nil,%s,%s,nil): %v", hex.EncodeToString(result.nonce), hex.EncodeToString(ciphertext), err)
	}
	if diff := cmp.Diff(hex.EncodeToString(plaintext), hex.EncodeToString(result.plaintext)); diff != "" {
		t.Errorf("gcm.Open(nil,%s,%s,nil) mismatch (-want +got):\n%s", hex.EncodeToString(result.nonce), hex.EncodeToString(ciphertext), diff)
	}
}

func testInsufficientSize(v *encryptionTestCase) (result encResult, err error) {
	data := make([]byte, len(v.data))
	copy(data, v.data)

	result.key = make([]byte, len(v.key))
	copy(result.key, v.key)

	result.ciphertext = data[:0]
	result.ciphertext, err = v.encrypt(result.ciphertext, data, v.key)
	if err != nil {
		return result, fmt.Errorf("testEncryptDecrypt: encrypt error %w", err)
	}
	enc := make([]byte, len(result.ciphertext))
	copy(enc, result.ciphertext)
	result.nonce = result.ciphertext[:gcmStandardNonceSize]
	result.ciphertext = result.ciphertext[gcmStandardNonceSize:]

	result.plaintext, err = v.decrypt(enc[:0], enc, v.key)
	if err != nil {
		return result, fmt.Errorf("testEncryptDecrypt: decrypt error %w", err)
	}
	return result, nil
}

func testNil(v *encryptionTestCase) (result encResult, err error) {
	data := make([]byte, len(v.data), cap(v.data))
	copy(data, v.data)

	result.key = make([]byte, len(v.key))
	copy(result.key, v.key)

	result.ciphertext, err = v.encrypt(nil, data, v.key)
	if err != nil {
		return result, fmt.Errorf("testEncryptDecrypt: encrypt error %w", err)
	}
	enc := make([]byte, len(result.ciphertext))
	copy(enc, result.ciphertext)
	result.nonce = result.ciphertext[:gcmStandardNonceSize]
	result.ciphertext = result.ciphertext[gcmStandardNonceSize:]

	result.plaintext, err = v.decrypt(nil, enc, v.key)
	if err != nil {
		return result, fmt.Errorf("testEncryptDecrypt: decrypt error %w", err)
	}
	return result, nil
}

func testOverlap(v *encryptionTestCase) (result encResult, err error) {
	data := make([]byte, len(v.data), len(v.data)+gcmStandardNonceSize+gcmTagSize)
	copy(data, v.data)

	result.key = make([]byte, len(v.key))
	copy(result.key, v.key)

	result.ciphertext = data[:0]
	result.ciphertext, err = v.encrypt(result.ciphertext, data, v.key)
	if err != nil {
		return result, fmt.Errorf("testEncryptDecrypt: encrypt error %w", err)
	}
	enc := make([]byte, len(result.ciphertext))
	copy(enc, result.ciphertext)
	result.nonce = result.ciphertext[:gcmStandardNonceSize]
	result.ciphertext = result.ciphertext[gcmStandardNonceSize:]

	result.plaintext, err = v.decrypt(enc[:0], enc, v.key)
	if err != nil {
		return result, fmt.Errorf("testEncryptDecrypt: decrypt error %w", err)
	}
	return result, nil
}

type encResult struct {
	key        []byte
	nonce      []byte
	ciphertext []byte
	plaintext  []byte
}
