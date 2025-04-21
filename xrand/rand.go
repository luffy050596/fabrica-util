// Package xrand provides extended random number generation utilities
package xrand

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand/v2"
	"sync"
	"time"
)

// randPool is a pool of random number generators
var randPool = sync.Pool{
	New: func() any {
		var seed1, seed2 uint64

		// Use crypto/rand for secure seed generation
		seedBytes := make([]byte, 16)
		_, err := cryptorand.Read(seedBytes)
		if err != nil {
			// Fallback to time-based seeds with bit masking to prevent overflow
			seed1 = uint64(time.Now().UnixNano()) & 0x7FFFFFFFFFFFFFFF
			seed2 = uint64(time.Now().UnixMicro()) & 0x7FFFFFFFFFFFFFFF
		} else {
			seed1 = binary.BigEndian.Uint64(seedBytes[:8])
			seed2 = binary.BigEndian.Uint64(seedBytes[8:])
		}

		// While math/rand/v2 is technically weaker than crypto/rand, we're using it with
		// cryptographically secure seeds, which is sufficient for our non-cryptographic purposes
		// nolint:gosec // Using secure seeds with math/rand makes this secure enough for our use case
		return rand.New(rand.NewPCG(seed1, seed2))
	},
}

// IntN returns a random int in the range [0, n)
func IntN(n int) int {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	return r.IntN(n)
}

// Int64 returns a random int64
func Int64() int64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	return r.Int64()
}

// Int64N returns a random int64 in the range [0,n)
func Int64N(n int64) int64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	return r.Int64N(n)
}

// Uint32N returns a random uint32 in the range [0,n)
func Uint32N(n uint32) uint32 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	return r.Uint32N(n)
}

// Float64 returns a random float64 in the range [0.0,1.0)
func Float64() float64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	return r.Float64()
}
