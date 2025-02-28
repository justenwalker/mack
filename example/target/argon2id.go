package target

import (
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	memoryReq = 19 * 1024 // 19MiB
	timeReq   = 2
	threadReq = 1
)

// GenerateKeyArgon2ID generates a derived key from a given password, salt, and key length.
// It uses argon2id to derive the key from the arguments.
func GenerateKeyArgon2ID(password []byte, salt []byte, keyLen int) ([]byte, error) {
	if keyLen < 1 {
		return nil, fmt.Errorf("argon2id: invalid key length: %d", keyLen)
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("argon2id: empty password key length: %d", keyLen)
	}
	return argon2.IDKey(password, salt, timeReq, memoryReq, threadReq, safeUint32(keyLen)), nil
}

func safeUint32(n int) uint32 {
	if n > 1<<31 {
		return 1 << 31
	}
	if n < 0 {
		return 0
	}
	return uint32(n)
}
