package testhelpers

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	randomReader func([]byte) (int, error)
	randomSeeder func(int64)
)

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	randomSeeder = r.Seed
	randomReader = r.Read
}

// ReadRandom reads pseudo-random bytes into the given byte array.
func ReadRandom(bytes []byte) (int, error) {
	return randomReader(bytes)
}

// MustReadRandom reads pseudo-random bytes into the given byte array.
// This variant panics if it is unable to or reads less than the desireed bytes.
func MustReadRandom(bytes []byte) {
	n, err := ReadRandom(bytes)
	if err != nil {
		panic(fmt.Errorf("MustReadRandom: failed to read random bytes: %w", err))
	}
	if n != len(bytes) {
		panic(fmt.Errorf("MustReadRandom: failed to read random %d bytes: read %d bytes", len(bytes), n))
	}
}

// SeedRandom seeds the random number generator.
func SeedRandom(seed int64) {
	randomSeeder(seed)
}
