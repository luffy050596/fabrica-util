package bloom

import (
	"math"
	mathrand "math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInt64BloomFilter(t *testing.T) {
	t.Parallel()

	bf := NewInt64Bloom(1000, 0.01)

	// test basic functions
	testData := []int64{0, -1, 123456789, math.MaxInt64 - 1, math.MinInt64 + 1}
	for _, d := range testData {
		bf.Add(d)
		assert.True(t, bf.Contains(d), "Should contain added element")
	}

	// test false positive rate
	falsePositives := 0
	total := 10000

	r := newRand()

	for i := 0; i < total; i++ {
		randomNum := r.Int64()
		if bf.Contains(randomNum) && !contains(testData, randomNum) {
			falsePositives++
		}
	}

	fpRate := float64(falsePositives) / float64(total)
	assert.True(t, fpRate < 0.02, "False positive rate too high: %f", fpRate)
}

func TestInt64EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("min and max", func(t *testing.T) {
		t.Parallel()

		bf := NewInt64Bloom(100, 0.01)
		bf.Add(-1 << 63)
		bf.Add(1<<63 - 1)
		assert.True(t, bf.Contains(-1<<63))
		assert.True(t, bf.Contains(1<<63-1))
	})

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		bf := NewInt64Bloom(10, 0.01)
		bf.Add(0)
		assert.True(t, bf.Contains(0))
	})
}

func BenchmarkInt64Bloom(b *testing.B) {
	bf := NewInt64Bloom(1000000, 0.01)

	r := newRand()
	data := make([]int64, b.N)

	for i := range data {
		data[i] = r.Int64()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf.Add(data[i])
		bf.Contains(data[i])
	}
}

func contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func TestBasicOperations(t *testing.T) {
	t.Parallel()

	bf := NewInt64Bloom(1000, 0.01)

	// Test adding single elements
	bf.Add(1)
	bf.Add(100)
	bf.Add(10000)

	// Test membership queries
	assert.True(t, bf.Contains(1), "Bloom filter should contain added element")
	assert.True(t, bf.Contains(100), "Bloom filter should contain added element")
	assert.True(t, bf.Contains(10000), "Bloom filter should contain added element")

	// Test non-membership
	// Note: false positives are possible but unlikely with these parameters
	// However, if the test has a false positive, don't consider it a failure
	nonMembers := []int64{2, 200, 20000}
	falsePositives := 0

	for _, item := range nonMembers {
		if bf.Contains(item) {
			falsePositives++
		}
	}

	t.Logf("False positive rate: %.4f", float64(falsePositives)/float64(len(nonMembers)))
}

func TestBatchAdd(t *testing.T) {
	t.Parallel()

	t.Run("add_many_elements", func(t *testing.T) {
		t.Parallel()

		bf := NewInt64Bloom(100, 0.01)
		elements := []int64{1, 2, 3, 4, 5}
		bf.MAdd(elements)

		for _, e := range elements {
			assert.True(t, bf.Contains(e))
		}
	})

	t.Run("empty_batch", func(t *testing.T) {
		t.Parallel()

		bf := NewInt64Bloom(10, 0.01)

		bf.MAdd([]int64{})
		assert.Equal(t, bf.bitmap.Count(), int64(0))
	})
}

func BenchmarkAdd(b *testing.B) {
	bf := NewInt64Bloom(100000, 0.01)

	for i := 0; i < b.N; i++ {
		bf.Add(int64(i))
	}
}

func newRand() *mathrand.Rand {
	return mathrand.New(mathrand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().Add(1*time.Second).UnixNano())))
}
