// Package bloom provides Bloom filter implementations for efficient set membership testing
package bloom

import (
	"math"

	"github.com/go-pantheon/fabrica-util/bitmap"
)

// Int64BloomFilter optimized Bloom filter for int64
type Int64BloomFilter struct {
	bitmap   *bitmap.Bitmap
	hashFunc []func(int64) int64
	size     int64
}

// NewInt64Bloom create int64 optimized Bloom filter
// n: expected element count
// p: expected false positive rate (0 < p < 1)
func NewInt64Bloom(n int64, p float64) *Int64BloomFilter {
	m, k := estimateParameters(n, p)
	// limit max hash function count to 8
	if k > 8 {
		k = 8
	}

	return &Int64BloomFilter{
		bitmap:   bitmap.NewBitmap(m),
		hashFunc: createInt64HashFunctions(k),
		size:     m,
	}
}

// Add add int64 element
func (bf *Int64BloomFilter) Add(data int64) {
	for _, fn := range bf.hashFunc {
		h := fn(data) % bf.size
		bf.bitmap.Set(h)
	}
}

// MAdd add multiple int64 elements
func (bf *Int64BloomFilter) MAdd(data []int64) {
	if len(data) == 0 {
		return
	}

	indexes := make([]int64, 0, len(data)*len(bf.hashFunc))

	for _, d := range data {
		for _, fn := range bf.hashFunc {
			h := fn(d) % bf.size
			indexes = append(indexes, h)
		}
	}

	bf.bitmap.MSet(indexes)
}

// Contains check if the element may exist
func (bf *Int64BloomFilter) Contains(data int64) bool {
	for _, fn := range bf.hashFunc {
		h := fn(data) % bf.size
		if !bf.bitmap.IsSet(h) {
			return false
		}
	}

	return true
}

// estimateParameters calculate optimal parameters (m: array size, k: hash function count)
func estimateParameters(n int64, p float64) (int64, int64) {
	m := int64(math.Ceil(-float64(n) * math.Log(p) / (math.Pow(math.Log(2), 2))))
	k := int64(math.Ceil(math.Log(2) * float64(m) / float64(n)))

	return m, k
}

// create int64 optimized hash functions
func createInt64HashFunctions(k int64) []func(int64) int64 {
	return func() []func(int64) int64 {
		fns := make([]func(int64) int64, k)

		primes := []int64{
			7540381766041133433, // replace overflow constant
			4354685564936845354,
			6742423442829656945,
			5864445385546375633,
			6955415921471435831,
			2270897969802886123,
			6620516959819538809,
			4354685564936845354,
		}

		for i := int64(0); i < k && i < int64(len(primes)); i++ {
			seed := primes[i]
			fns[i] = func(data int64) int64 {
				h := data * seed
				h ^= h >> 33
				h *= 0x5136ead0f35b37d // replace overflow constant
				h ^= h >> 33
				h *= 0x4cf5ad432745937 // replace overflow constant
				h ^= h >> 33

				return h
			}
		}

		// if k is greater than the number of predefined primes, use the variant
		for i := int64(len(primes)); i < k; i++ {
			seed := primes[i%int64(len(primes))] ^ (i * 0x1234567)
			fns[i] = func(data int64) int64 {
				h := data ^ seed
				h ^= h >> 33
				h *= 0x5136ead0f35b37d // replace overflow constant
				h ^= h >> 33
				h *= 0x4cf5ad432745937 // replace overflow constant
				h ^= h >> 33

				return h
			}
		}

		return fns
	}()
}
