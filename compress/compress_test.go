package compress

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testWeakThreshold   = 1 << 10   // 1KB
	testStrongThreshold = 128 << 10 // 128KB
)

func TestMain(m *testing.M) {
	// 初始化测试配置
	Init(testWeakThreshold, testStrongThreshold)
	m.Run()
}

func TestCompress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dataLen int
		want    bool
	}{
		{"BelowWeak", testWeakThreshold - 1, false},
		{"EqualWeak", testWeakThreshold, true},
		{"BetweenThresholds", testWeakThreshold + 1, true},
		{"EqualStrong", testStrongThreshold, true},
		{"AboveStrong", testStrongThreshold + 1, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := make([]byte, tt.dataLen)

			_, compressed, err := Compress(data)
			require.Nil(t, err)

			assert.Equal(t, compressed, tt.want)
		})
	}
}

func TestCompressDecompress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"Small", []byte("hello world")},
		{"Medium", bytes.Repeat([]byte{0x01}, testWeakThreshold+1)},
		{"Large", bytes.Repeat([]byte{0x01}, testStrongThreshold+1024)},
		{"Random", randBytes(2 * testStrongThreshold)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			compressed, didCompress, err := Compress(tc.data)

			require.Nil(t, err)

			if didCompress {
				decompressed, err := Decompress(compressed)

				if err != nil {
					t.Fatalf("decompress failed: %v", err)
				}

				assert.Equal(t, tc.data, decompressed)
			} else {
				assert.Equal(t, tc.data, compressed)
			}
		})
	}
}

func TestErrorConditions(t *testing.T) {
	t.Parallel()

	t.Run("InvalidDecompressData", func(t *testing.T) {
		t.Parallel()

		_, err := Decompress([]byte{0x00, 0x01, 0x02})

		assert.NotNil(t, err)
	})

	t.Run("NilInput", func(t *testing.T) {
		t.Parallel()

		t.Run("Compress", func(t *testing.T) {
			t.Parallel()

			compressed, didCompress, err := Compress(nil)

			assert.Nil(t, err)
			assert.Equal(t, []byte{}, compressed)
			assert.Equal(t, false, didCompress)
		})

		t.Run("Decompress", func(t *testing.T) {
			t.Parallel()

			decompressed, err := Decompress(nil)

			assert.Nil(t, err)
			assert.Equal(t, []byte{}, decompressed)
		})
	})
}

func TestConcurrentSafety(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	const goroutines = 10

	// Test concurrent Init and Compress
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			data := randBytes(testStrongThreshold * 2)
			_, _, _ = Compress(data)
		}()
	}

	wg.Wait()
}

func BenchmarkCompress(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", 64},
		{"1KB", 1 << 10},
		{"512KB", testWeakThreshold},
		{"1MB", testStrongThreshold},
		{"4MB", 4 << 20},
	}

	for _, size := range sizes {
		size := size
		data := randBytes(size.size)

		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size.size))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _, _ = Compress(data)
			}
		})
	}
}

func BenchmarkDecompress(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", 64},
		{"1KB", 1 << 10},
		{"512KB", testWeakThreshold},
		{"1MB", testStrongThreshold},
		{"4MB", 4 << 20},
	}

	for _, size := range sizes {
		size := size
		data := randBytes(size.size)
		compressed, _, _ := Compress(data)

		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(compressed)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = Decompress(compressed)
			}
		})
	}
}

func randBytes(n int) []byte {
	data := make([]byte, n)

	_, err := cryptorand.Read(data)

	if err != nil {
		panic(err)
	}

	return data
}

func TestCompressionEfficiency(t *testing.T) {
	t.Parallel()

	type testData struct {
		ID      int64
		Name    string
		Tags    []string
		Value   float64
		Enabled bool
	}

	generateData := func(sizeKB int) []byte {
		l := sizeKB
		data := make([]byte, 0, l)
		buf := bytes.NewBuffer(data)

		for i := 0; buf.Len() < l; i++ {
			item := testData{
				ID:      int64(i),
				Name:    fmt.Sprintf("item-%d", i),
				Tags:    []string{"tag1", "tag2", "tag3"},
				Value:   cryptoRandFloat64(),
				Enabled: i%2 == 0,
			}

			b, _ := json.Marshal(item)
			buf.Write(b)
		}

		return buf.Bytes()[:l]
	}

	testCases := []struct {
		name         string
		size         int
		wantMinRatio float64 // min compression ratio
	}{
		{"SmallData", testWeakThreshold, 0.3},
		{"MediumData", testStrongThreshold << 2, 0.2},
		{"LargeData", testStrongThreshold << 4, 0.2},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := generateData(tc.size)
			origSize := len(data)

			startCompress := time.Now()
			compressed, didCompress, err := Compress(data)
			compressTime := time.Since(startCompress)

			require.NoError(t, err)
			require.True(t, didCompress)

			startDecompress := time.Now()
			decompressed, err := Decompress(compressed)
			decompressTime := time.Since(startDecompress)

			require.NoError(t, err)
			assert.Equal(t, data, decompressed)

			compressedSize := len(compressed)
			ratio := float64(compressedSize) / float64(origSize)

			t.Logf("%d KB -> %d KB (%.2f%%), ct: %.2fs (%.2f MB/s), dct: %.2fs (%.2f MB/s)",
				origSize>>10, compressedSize>>10, ratio*100,
				compressTime.Seconds(), float64(origSize)/(compressTime.Seconds()*(1<<20)),
				decompressTime.Seconds(), float64(origSize)/(decompressTime.Seconds()*(1<<20)))

			assert.Less(t, ratio, tc.wantMinRatio,
				"compression ratio too high, expected < %.2f, got %.2f",
				tc.wantMinRatio, ratio)
		})
	}
}

// cryptoRandFloat64 returns a cryptographically secure random float64 value between 0 and 1
func cryptoRandFloat64() float64 {
	var buf [8]byte

	_, err := cryptorand.Read(buf[:])

	if err != nil {
		panic(err)
	}

	// Convert to a uint64 and scale to the range [0,1)
	val := float64(byteOrder8(buf)) / float64(1<<64)

	return val
}

// byteOrder8 converts an 8-byte array to uint64
func byteOrder8(b [8]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}
