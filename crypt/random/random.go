// Package random exposes function Read to read random values from crypto/rand.
//
// If built with 'test_random' build tag, then it will use math/rand which can be seeded with the SeedRandom function.
// If this build flag is enabled, the IsTest function in this package will return true.
package random

import "fmt"

// MustNotBeTest can be called from main or init to assert the random source
// is the default cryptographic source instead of the test source.
//
// This allows the application to bail early if a test binary is being executed,
// which has an insecure random number generator.
func MustNotBeTest() {
	if IsTest() {
		panic("must not be built with test_random")
	}
}

// IsTest returns true if the binary was built with the test_random build tag.
// If this function returns true, the cryptographic random number generator is insecure.
func IsTest() bool {
	return testRandom
}

// Read reads pseudo-random bytes into the given byte array.
func Read(bytes []byte) {
	if _, err := randomReader(bytes); err != nil {
		panic(fmt.Errorf("random.Read: %w", err))
	}
}

// SeedRandom seeds the random number generator.
func SeedRandom(seed int64) {
	randomSeeder(seed)
}
