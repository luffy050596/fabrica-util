package xrand

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandAlphaNumString(t *testing.T) {
	t.Parallel()

	t.Run("normal cases", func(t *testing.T) {
		t.Parallel()
		// Test generating random strings of different lengths
		lengths := []int{4, 8, 16, 32, 64}
		for _, length := range lengths {
			t.Run("length_"+strconv.Itoa(length), func(t *testing.T) {
				t.Parallel()
				// Test generating random strings of different lengths
				ck := make(map[string]struct{}, 1000)

				for range 1000 {
					s, err := RandAlphaNumString(length)
					assert.Nil(t, err)
					assert.Equal(t, length, len(s))

					// Verify uniqueness for longer strings
					if length >= 16 {
						_, ok := ck[s]
						assert.False(t, ok, "duplicate string generated: %s", s)

						ck[s] = struct{}{}
					}
				}
			})
		}
	})

	t.Run("boundary cases", func(t *testing.T) {
		t.Parallel()
		// Test boundary values
		testCases := []struct {
			name        string
			length      int
			shouldError bool
		}{
			{"zero length", 0, true},
			{"negative length", -1, true},
			{"very large length", 1 << 20, false}, // 1MB length
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s, err := RandAlphaNumString(tc.length)

				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, tc.length, len(s))
				}
			})
		}
	})
}

func TestBytesN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		length int
	}{
		{"normal length", 16},
		{"zero length", 0},
		{"large length", 1 << 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := BytesN(tc.length)
			assert.Equal(t, tc.length, len(b))
		})
	}
}

func BenchmarkRandAlphaNumString(b *testing.B) {
	benchCases := []struct {
		name   string
		length int
	}{
		{"tiny", 4},
		{"small", 8},
		{"medium", 16},
		{"large", 32},
		{"huge", 64},
		{"massive", 128},
		{"massive", 256},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := RandAlphaNumString(bc.length)
				assert.Nil(b, err)
			}
		})
	}
}

func BenchmarkBytesN(b *testing.B) {
	lengths := []int{16, 32, 64, 256, 512, 1024}
	for _, length := range lengths {
		b.Run(fmt.Sprintf("%d bytes", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = BytesN(length)
			}
		})
	}
}
