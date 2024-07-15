//go:build test_random

package random

import (
	"math/rand"
	"time"
)

const testRandom = true

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
