//go:build !test_random

package random

import "crypto/rand"

const testRandom = false

var (
	randomReader = rand.Read
	randomSeeder = func(int64) {}
)
