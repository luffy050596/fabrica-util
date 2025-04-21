package xid

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	mathrand "math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCodecID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id int64
	}{
		{id: int64(0)},
		{id: int64(1)},
		{id: int64(2)},
		{id: int64(3)},
		{id: int64(65534)},
		{id: int64(65535)},
		{id: int64(65536)},
		{id: math.MaxInt64},
		{id: math.MaxInt64 - 1},
		{id: int64(-1)},
		{id: -math.MaxInt64},
		{id: -(math.MaxInt64 - 1)},
	}

	for _, tt := range tests {
		str, _ := EncodeID(tt.id)
		id2, _ := DecodeID(str)
		assert.Equal(t, tt.id, id2)
	}
}

// check 5 millions users' id encode str is unique
// func TestIDUnique(t *testing.T) {
// 	v := make(map[string]struct{}, math.MaxInt64)
// 	for id := int64(0); id < 5_000_000; id++ {
// 		str, _ := EncodeID(id)
// 		_, ok := v[str]
// 		assert.False(t, ok)
// 		id2, _ := DecodeID(str)
// 		assert.Equal(t, id, id2)
// 	}
// }

func TestCombineZoneID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		zoneID int64
		zone   uint8
		want   int64
	}{
		{
			name:   "zero values",
			zoneID: 0,
			zone:   0,
			want:   0,
		},
		{
			name:   "small values",
			zoneID: 1,
			zone:   2,
			want:   258, // (1 << 8) | 2 = 256 + 2 = 258
		},
		{
			name:   "max zone value",
			zoneID: 100,
			zone:   MaxZone,
			want:   (100 << zoneBit) | int64(MaxZone),
		},
		{
			name:   "large zoneID",
			zoneID: 1<<55 - 1, // Test with a large but valid zoneID
			zone:   123,
			want:   ((1<<55 - 1) << zoneBit) | 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := CombineZoneID(tt.zoneID, tt.zone); got != tt.want {
				t.Errorf("CombineZoneID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int64
		wantZoneID int64
		wantZone   uint8
	}{
		{
			name:       "zero value",
			id:         0,
			wantZoneID: 0,
			wantZone:   0,
		},
		{
			name:       "small value",
			id:         258, // (1 << 8) | 2 = 258
			wantZoneID: 1,
			wantZone:   2,
		},
		{
			name:       "max zone value",
			id:         (100 << zoneBit) | int64(MaxZone),
			wantZoneID: 100,
			wantZone:   MaxZone,
		},
		{
			name:       "large zoneID",
			id:         ((1<<55 - 1) << zoneBit) | 123,
			wantZoneID: 1<<55 - 1,
			wantZone:   123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotZoneID, gotZone := SplitID(tt.id)

			if gotZoneID != tt.wantZoneID {
				t.Errorf("SplitID() gotZoneID = %v, want %v", gotZoneID, tt.wantZoneID)
			}

			if gotZone != tt.wantZone {
				t.Errorf("SplitID() gotZone = %v, want %v", gotZone, tt.wantZone)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		zoneID int64
		zone   uint8
	}{
		{
			name:   "zero values",
			zoneID: 0,
			zone:   0,
		},
		{
			name:   "small values",
			zoneID: 1,
			zone:   2,
		},
		{
			name:   "max zone value",
			zoneID: 100,
			zone:   MaxZone,
		},
		{
			name:   "large zoneID",
			zoneID: 1<<55 - 1,
			zone:   123,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			combined := CombineZoneID(tc.zoneID, tc.zone)
			gotZoneID, gotZone := SplitID(combined)

			if gotZoneID != tc.zoneID {
				t.Errorf("RoundTrip zoneID = %v, want %v", gotZoneID, tc.zoneID)
			}

			if gotZone != tc.zone {
				t.Errorf("RoundTrip zone = %v, want %v", gotZone, tc.zone)
			}
		})
	}
}

func newRand() *mathrand.Rand {
	// Use crypto/rand for secure seed generation
	var seed1, seed2 uint64

	seedBytes := make([]byte, 16)
	_, err := rand.Read(seedBytes)

	if err != nil {
		// Fallback to time-based seeds with bit masking
		seed1 = uint64(time.Now().UnixMicro()) & 0x7FFFFFFFFFFFFFFF //nolint:gosec // acceptable for tests
		seed2 = uint64(time.Now().UnixMilli()) & 0x7FFFFFFFFFFFFFFF //nolint:gosec // acceptable for tests
	} else {
		seed1 = binary.BigEndian.Uint64(seedBytes[:8])
		seed2 = binary.BigEndian.Uint64(seedBytes[8:])
	}

	return mathrand.New(mathrand.NewPCG(seed1, seed2)) //nolint:gosec // test code with secure seed
}

func BenchmarkEncodeID(b *testing.B) {
	id := newRand().Int64N(math.MaxInt64)
	for i := 0; i < b.N; i++ {
		_, _ = EncodeID(id)
	}
}

func BenchmarkDecodeID(b *testing.B) {
	id := newRand().Int64N(65535)
	for i := 0; i < b.N; i++ {
		str, _ := EncodeID(id)
		_, _ = DecodeID(str)
	}
}
