// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NOTE: This implementation has been altered from the Go source code to reduce allocations
// as a trade-off with only supporting SHA-256 and not implementing io.Writer / Sum.

/*
Package hmac implements the Keyed-Hash Message Authentication Code (HMAC) as
defined in U.S. Federal Information Processing Standards Publication 198.
An HMAC is a cryptographic hash that uses a key to sign a message.
The receiver verifies the hash by recomputing it using the same key.

Receivers should be careful to use Equal to compare MACs in order to avoid
timing side-channels:

	// ValidMAC reports whether messageMAC is a valid HMAC tag for message.
	func ValidMAC(message, messageMAC, key []byte) bool {
		mac := hmac.New(sha256.New, key)
		mac.Write(message)
		expectedMAC := mac.Sum(nil)
		return hmac.Equal(messageMAC, expectedMAC)
	}
*/
package hmac

import (
	"crypto/sha256"
	"fmt"
)

// FIPS 198-1:
// https://csrc.nist.gov/publications/fips/fips198-1/FIPS-198-1_final.pdf

// key is zero padded to the block size of the hash function
// ipad = 0x36 byte repeated for key length
// opad = 0x5c byte repeated for key length
// hmac = H([key ^ opad] H([key ^ ipad] text))

// SHA256 computes the HMAC with SHA-256 hash algorithm.
// the hash is written to the `out` parameter.
// the key and the out parameter can overlap, however this will modify the key parameter as well.
// this function will panic if the capacity of the out slice is less than sha256.BlockSize (64 bytes).
func SHA256(key []byte, out []byte, data []byte) []byte {
	outer := sha256.New()
	inner := sha256.New()
	var ipad [sha256.BlockSize]byte
	var opad [sha256.BlockSize]byte

	// omit uniqueness check since we are not parameterizing the hash function
	if len(key) > sha256.BlockSize {
		// If key is too big, hash it.
		outer.Write(key)
		key = outer.Sum(nil)
	}
	if cap(out) < sha256.Size {
		panic(fmt.Errorf("HMACSHA256: out capacity too small to contain SHA-256 (%d bytes), was=%d bytes", sha256.Size, cap(out)))
	}
	copy(ipad[:], key)
	copy(opad[:], key)
	for i := range ipad {
		ipad[i] ^= 0x36
	}
	for i := range opad {
		opad[i] ^= 0x5c
	}
	inner.Write(ipad[:])
	inner.Write(data)
	in := inner.Sum(out[:0])
	outer.Reset()
	outer.Write(opad[:])
	outer.Write(in[0:])
	return outer.Sum(in[:0])
}
