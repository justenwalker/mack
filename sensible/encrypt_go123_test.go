//go:build !go1.24

package sensible

import "testing"

func TestSensible_EncryptDecrypt_go1_23(t *testing.T) {
	testEncryptDecrypt(t, encryptionTestConfig{
		encrypt: encryptGo123,
		decrypt: decryptGo123,
	})
}
