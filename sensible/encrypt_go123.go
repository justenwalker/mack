//go:build !go1.24

package sensible

var (
	decryptFunc = decryptGo123
	encryptFunc = encryptGo123
)

func encryptGo123(dst []byte, plaintext []byte, key []byte) ([]byte, error) {
	if alias.InexactOverlap(dst, plaintext) {
		panic("invalid buffer overlap of dst and plaintext")
	}
	ret, out := sliceForAppend(dst, len(plaintext)+gcmTagSize+gcmStandardNonceSize)
	nonce := out[:gcmStandardNonceSize]
	ciphertext := out[gcmStandardNonceSize:]
	if alias.AnyOverlap(out, plaintext) {
		copy(ciphertext, plaintext)
		plaintext = ciphertext[:len(plaintext)]
	}
	crand.Read(nonce)
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("sensible.Encrypt: failed to create aes cipher: %w", err)
	}
	a, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("sensible.Encrypt: failed to create new aes-gcm cipher: %w", err)
	}
	dst = a.Seal(ciphertext[:0], nonce, plaintext, nil)
	ret = ret[:len(dst)+gcmStandardNonceSize]
	return ret, nil
}

func decryptGo123(dst []byte, ciphertext []byte, key []byte) ([]byte, error) {
	if alias.InexactOverlap(dst, ciphertext) {
		panic("invalid buffer overlap of dst and plaintext")
	}
	var nonce []byte
	ret, out := sliceForAppend(dst, len(ciphertext)-gcmStandardNonceSize-gcmTagSize)
	if alias.AnyOverlap(out, ciphertext) {
		nonce = make([]byte, gcmStandardNonceSize)
		copy(nonce, ciphertext)
		copy(out[:len(ciphertext)], ciphertext[gcmStandardNonceSize:])
		ciphertext = out[:len(ciphertext)-gcmStandardNonceSize]
	} else {
		nonce = ciphertext[:gcmStandardNonceSize]
		ciphertext = ciphertext[gcmStandardNonceSize:]
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("sensible.Decrypt: failed to create aes cipher: %w", err)
	}
	a, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("sensible.Decrypt: failed to create new aes-gcm cipher: %w", err)
	}
	_, err = a.Open(out[:0], nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// sliceForAppend takes a slice and a requested number of bytes. It returns a
// slice with the contents of the given slice followed by that many bytes and a
// second slice that aliases into it and contains only the extra bytes. If the
// original slice has sufficient capacity then no allocation is performed.
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
