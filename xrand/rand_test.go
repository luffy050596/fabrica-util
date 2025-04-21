package xrand

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntN(t *testing.T) {
	t.Parallel()

	// Test normal cases
	t.Run("normal cases", func(t *testing.T) {
		t.Parallel()
		// Test with common values
		for _, n := range []int{10, 100, 1000} {
			for i := 0; i < 100; i++ {
				result := IntN(n)
				assert.True(t, result >= 0 && result < n)
			}
		}
	})

	// Test boundary cases
	t.Run("boundary cases", func(t *testing.T) {
		t.Parallel()
		// Test with n=1
		result := IntN(1)
		assert.Equal(t, result, 0)

		// Test with n=2 (multiple times to ensure both 0 and 1 are possible)
		seen0, seen1 := false, false
		for i := 0; i < 100 && (!seen0 || !seen1); i++ {
			result = IntN(2)
			switch result {
			case 0:
				seen0 = true
			case 1:
				seen1 = true
			default:
				t.Errorf("IntN(2) returned %d, expected 0 or 1", result)
			}
		}

		assert.True(t, seen0)
		assert.True(t, seen1)
	})
}

func TestInt64(t *testing.T) {
	t.Parallel()

	// Generate multiple values to check for reasonable distribution
	const iterations = 1000

	var (
		sum    int64
		minVal int64 = math.MaxInt64
		maxVal int64 = math.MinInt64
	)

	for i := 0; i < iterations; i++ {
		val := Int64()
		sum += val

		if val < minVal {
			minVal = val
		}

		if val > maxVal {
			maxVal = val
		}
	}

	assert.NotEqual(t, minVal, maxVal)
}

func TestInt64N(t *testing.T) {
	t.Parallel()

	// Test normal cases
	t.Run("normal cases", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int64{10, 1000, math.MaxInt64 / 2} {
			for i := 0; i < 100; i++ {
				result := Int64N(n)
				assert.True(t, result >= 0 && result < n)
			}
		}
	})

	// Test boundary cases
	t.Run("boundary cases", func(t *testing.T) {
		t.Parallel()
		// Test with n=1
		result := Int64N(1)
		assert.Equal(t, result, int64(0))

		// Test with n=2 (multiple times to ensure both 0 and 1 are possible)
		seen0, seen1 := false, false
		for i := 0; i < 100 && (!seen0 || !seen1); i++ {
			result = Int64N(2)
			switch result {
			case 0:
				seen0 = true
			case 1:
				seen1 = true
			default:
				t.Errorf("Int64N(2) returned %d, expected 0 or 1", result)
			}
		}

		assert.True(t, seen0)
		assert.True(t, seen1)
	})
}

func TestUint32N(t *testing.T) {
	t.Parallel()

	// Test normal cases
	t.Run("normal cases", func(t *testing.T) {
		t.Parallel()

		for _, n := range []uint32{10, 1000, math.MaxUint32 / 2} {
			for i := 0; i < 100; i++ {
				result := Uint32N(n)
				assert.True(t, result < n)
			}
		}
	})

	// Test boundary cases
	t.Run("boundary cases", func(t *testing.T) {
		t.Parallel()
		// Test with n=1
		result := Uint32N(1)
		if result != 0 {
			t.Errorf("Uint32N(1) returned %d, expected 0", result)
		}

		// Test with n=2 (multiple times to ensure both 0 and 1 are possible)
		seen0, seen1 := false, false
		for i := 0; i < 100 && (!seen0 || !seen1); i++ {
			result = Uint32N(2)
			switch result {
			case 0:
				seen0 = true
			case 1:
				seen1 = true
			default:
				t.Errorf("Uint32N(2) returned %d, expected 0 or 1", result)
			}
		}

		assert.True(t, seen0)
		assert.True(t, seen1)
	})
}

func TestFloat64(t *testing.T) {
	t.Parallel()

	// Test that values are in the range [0.0, 1.0)
	for i := 0; i < 1000; i++ {
		val := Float64()
		assert.True(t, val >= 0.0 && val < 1.0)
	}

	// Check distribution by dividing the range into buckets
	const buckets = 10
	counts := make([]int, buckets)

	const iterations = 10000

	for i := 0; i < iterations; i++ {
		val := Float64()
		bucket := int(val * buckets)

		if bucket == buckets {
			bucket = buckets - 1 // Handle the edge case of val being very close to 1.0
		}

		counts[bucket]++
	}

	// Check that each bucket has a reasonable number of values
	// Expected is iterations/buckets, allow for 30% deviation
	expected := iterations / buckets
	minExpected := int(float64(expected) * 0.7)
	maxExpected := int(float64(expected) * 1.3)

	for _, count := range counts {
		assert.True(t, count >= minExpected && count <= maxExpected)
	}
}

// Benchmarks
func BenchmarkIntN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IntN(100)
	}
}

func BenchmarkInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int64()
	}
}

func BenchmarkInt64N(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int64N(1000)
	}
}

func BenchmarkUint32N(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Uint32N(1000)
	}
}

func BenchmarkFloat64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Float64()
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("IntN - zero", func(t *testing.T) {
		t.Parallel()

		defer func() {
			assert.NotNil(t, recover())
		}()
		IntN(0)
	})

	t.Run("Int64N - zero", func(t *testing.T) {
		t.Parallel()

		defer func() {
			assert.NotNil(t, recover())
		}()
		Int64N(0)
	})

	t.Run("Uint32N - zero", func(t *testing.T) {
		t.Parallel()

		defer func() {
			assert.NotNil(t, recover())
		}()
		Uint32N(0)
	})

	t.Run("Int64N - max int64", func(t *testing.T) {
		t.Parallel()

		val := Int64N(math.MaxInt64)
		assert.True(t, val >= 0 && val < math.MaxInt64)
	})
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	defer func() {
		assert.Nil(t, recover())
	}()

	const goroutines = 10

	const iterations = 1000

	done := make(chan bool, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			for i := 0; i < iterations; i++ {
				IntN(100)
				Int64()
				Int64N(100)
				Uint32N(100)
				Float64()
				BytesN(16)
			}
			done <- true
		}()
	}

	for g := 0; g < goroutines; g++ {
		<-done
	}
}

func BenchmarkRandomGenerationComparison(b *testing.B) {
	b.Run("IntN", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IntN(100)
		}
	})

	b.Run("Int64N", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Int64N(100)
		}
	})
}
