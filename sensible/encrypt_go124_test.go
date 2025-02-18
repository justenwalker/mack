//go:build go1.24

package sensible

import (
	"testing"
)

func TestSensible_EncryptDecrypt(t *testing.T) {
	testEncryptDecrypt(t, encryptionTestConfig{
		encrypt: encryptGo124,
		decrypt: decryptGo124,
	})
}
