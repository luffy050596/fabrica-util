package multipool

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testSizeReporter is an implementation of SizeReporter for testing purposes
type testSizeReporter struct {
	data []byte
	size int // explicitly set size for testing
}

func (t *testSizeReporter) Size() int {
	if t.size > 0 {
		return t.size
	}

	return len(t.data)
}

func newTestSizeReporter(size int) *testSizeReporter {
	return &testSizeReporter{
		data: make([]byte, size), // pre-allocate some capacity
		size: size,
	}
}

func (t *testSizeReporter) Init(n int) {
	if n <= 0 {
		return
	}

	t.data = make([]byte, n)
	t.size = n
}

func (t *testSizeReporter) Reset() {
	t.data = t.data[:0]
	t.size = 0
}

func TestMultiLayerPool_Basic(t *testing.T) {
	t.Parallel()

	pool := NewMultiLayerPool(
		func() Resetable {
			return newTestSizeReporter(64) // create a small object
		},
		func(obj Resetable) int {
			return obj.(*testSizeReporter).Size()
		},
		WithThresholds([]int{128, 256}),
	)

	// get an object and verify
	obj := pool.Get(64)
	assert.NotNil(t, obj)

	// verify the object is of the correct type and size
	bytes, ok := obj.(*testSizeReporter)
	assert.True(t, ok)
	assert.Equal(t, 64, bytes.Size())

	// put the object back
	pool.Put(obj)

	// get the statistics
	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.TotalPuts)
}

func TestMultiLayerPool_WithSizeOption(t *testing.T) {
	t.Parallel()

	small := newTestSizeReporter(64)   // small object
	medium := newTestSizeReporter(200) // medium object
	large := newTestSizeReporter(512)  // large object

	pool := NewMultiLayerPool(
		func() Resetable {
			return newTestSizeReporter(0) // create an empty object
		},
		func(obj Resetable) int {
			return obj.(*testSizeReporter).Size()
		},
		WithThresholds([]int{128, 256}),
	)

	// put the objects into the pool
	pool.Put(small)
	pool.Put(medium)
	pool.Put(large)

	// get objects
	obj1 := pool.Get(64)
	obj2 := pool.Get(200)
	obj3 := pool.Get(512)

	// verify the objects are not nil
	assert.NotNil(t, obj1)
	assert.NotNil(t, obj2)
	assert.NotNil(t, obj3)

	// put the objects back
	pool.Put(obj1)
	pool.Put(obj2)
	pool.Put(obj3)

	// verify the statistics
	stats := pool.GetStats()
	assert.Equal(t, int64(6), stats.TotalPuts)

	// verify the hits per layer
	t.Logf("Hits per layer: %v", stats.LayerHits)
}

func TestMultiLayerPool_DifferentSizes(t *testing.T) {
	t.Parallel()

	pool := NewMultiLayerPool(
		func() Resetable {
			return newTestSizeReporter(0) // create an empty object
		},
		func(obj Resetable) int {
			return obj.(*testSizeReporter).Size()
		},
		WithThresholds([]int{128, 256}),
	)

	small := newTestSizeReporter(64)   // less than the first threshold
	medium := newTestSizeReporter(200) // less than the second threshold but greater than the first
	large := newTestSizeReporter(512)  // greater than all thresholds

	pool.Put(small)
	pool.Put(medium)
	pool.Put(large)

	// verify the statistics
	stats := pool.GetStats()
	assert.Equal(t, int64(3), stats.TotalPuts)
}

// stress test: create a large number of objects and verify memory usage
func TestMultiLayerPool_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// record initial memory usage
	var initialStats runtime.MemStats

	runtime.ReadMemStats(&initialStats)

	pool := NewMultiLayerPool(
		func() Resetable {
			return newTestSizeReporter(0) // create an empty object
		},
		func(obj Resetable) int {
			return obj.(*testSizeReporter).Size()
		},
		WithThresholds([]int{128, 256, 2048, 8192, 16384}),
	)

	var wg sync.WaitGroup

	objCount := 10000
	concurrency := 10

	for c := 0; c < concurrency; c++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := 0; i < objCount/concurrency; i++ {
				var wrapper *testSizeReporter

				// set object properties and size based on index
				switch i % 3 {
				case 0: // small object
					wrapper = pool.Get(256).(*testSizeReporter)
					wrapper.data = make([]byte, 256)
					wrapper.size = 256
				case 1: // medium object
					wrapper = pool.Get(1024).(*testSizeReporter)
					wrapper.data = make([]byte, 1024)
					wrapper.size = 1024
				case 2: // large object
					wrapper = pool.Get(4096).(*testSizeReporter)
					wrapper.data = make([]byte, 4096)
					wrapper.size = 4096
				}

				// simulate using the object
				time.Sleep(time.Microsecond)

				// put the object back to the pool
				pool.Put(wrapper)
			}
		}()
	}

	// wait for all goroutines to complete
	wg.Wait()

	// force garbage collection to get a more accurate memory usage
	runtime.GC()

	// check current memory usage
	var finalStats runtime.MemStats

	runtime.ReadMemStats(&finalStats)

	// output memory usage statistics and pool statistics
	t.Logf("Initial heap alloc: %d bytes", initialStats.HeapAlloc)
	t.Logf("Final heap alloc: %d bytes", finalStats.HeapAlloc)
	t.Logf("Diff: %d bytes", finalStats.HeapAlloc-initialStats.HeapAlloc)

	stats := pool.GetStats()
	t.Logf("Pool stats - Small hits: %d, Medium hits: %d, Large hits: %d",
		stats.LayerHits[0], stats.LayerHits[1], stats.LayerHits[2])
	t.Logf("Pool stats - Small misses: %d, Medium misses: %d, Large misses: %d",
		stats.LayerMisses[0], stats.LayerMisses[1], stats.LayerMisses[2])
	t.Logf("Total puts: %d", stats.TotalPuts)
}

func BenchmarkMultiLayerPool(b *testing.B) {
	pool := NewMultiLayerPool(
		func() Resetable {
			return newTestSizeReporter(0) // create an empty object
		},
		func(obj Resetable) int {
			return obj.(*testSizeReporter).Size()
		},
		WithThresholds([]int{128, 256, 2048, 4096}),
	)

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			obj := pool.Get(64).(*testSizeReporter)
			obj.data = make([]byte, 64)
			obj.data[0] = byte(i)

			pool.Put(obj)
		}
	})

	b.Run("WithDifferentSizes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var obj *testSizeReporter

			// simulate different sizes based on the loop index
			switch i % 3 {
			case 0:
				// small object - basic properties
				obj = pool.Get(128).(*testSizeReporter)
				obj.data = make([]byte, 128)
				obj.size = 128
			case 1:
				// medium object - add more properties
				obj = pool.Get(256).(*testSizeReporter)
				obj.data = make([]byte, 256)
				obj.size = 256
			case 2:
				// large object - more data
				obj = pool.Get(4096).(*testSizeReporter)
				obj.data = make([]byte, 4096)
				obj.size = 4096
			}

			pool.Put(obj)
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				obj := pool.Get(128).(*testSizeReporter)
				// simple modification to make the object have different sizes
				obj.data = make([]byte, 128)
				obj.size = 128

				pool.Put(obj)
			}
		})
	})

	b.Run("ParallelWithSize", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0

			for pb.Next() {
				size := 256 * (1 << (i % 3)) // 256, 512, 1024
				obj := pool.Get(size).(*testSizeReporter)

				obj.data = make([]byte, size)
				obj.size = size
				pool.Put(obj)

				i++
			}
		})
	})

	b.Run("SizeReporter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var obj *testSizeReporter

			if i%2 == 0 {
				obj = pool.Get(128).(*testSizeReporter)
				obj.data = make([]byte, 128)
				obj.size = 128
			} else {
				obj = pool.Get(256).(*testSizeReporter)
				obj.data = make([]byte, 256)
				obj.size = 256
			}

			pool.Put(obj)
		}
	})
}

func BenchmarkCompareWithStandardPool(b *testing.B) {
	standardPool := &sync.Pool{
		New: func() any {
			return newTestSizeReporter(0)
		},
	}

	multiPool := NewMultiLayerPool(
		func() Resetable {
			return newTestSizeReporter(0) // create an empty object
		},
		func(obj Resetable) int {
			return obj.(*testSizeReporter).Size()
		},
		WithThresholds([]int{128, 256, 2048, 8192, 16384}),
	)

	b.Run("NewObjectAndGC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			switch i % 8 {
			case 0:
				newTestSizeReporter(63)
			case 1:
				newTestSizeReporter(255)
			case 2:
				newTestSizeReporter(1023)
			case 3:
				newTestSizeReporter(2047)
			case 4:
				newTestSizeReporter(4095)
			case 5:
				newTestSizeReporter(8191)
			case 6:
				newTestSizeReporter(16383)
			case 7:
				newTestSizeReporter(32767)
			}

			if i%10000 == 0 {
				runtime.GC()
			}
		}
	})

	b.Run("StandardPool", func(b *testing.B) {
		obj := standardPool.Get().(*testSizeReporter)

		for i := 0; i < b.N; i++ {
			switch i % 8 {
			case 0:
				obj.Init(63)
			case 1:
				obj.Init(255)
			case 2:
				obj.Init(1023)
			case 3:
				obj.Init(2047)
			case 4:
				obj.Init(4095)
			case 5:
				obj.Init(8191)
			case 6:
				obj.Init(16383)
			case 7:
				obj.Init(32767)
			}

			obj.Reset()
			standardPool.Put(obj)
		}
	})

	b.Run("MultiLayerPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var obj *testSizeReporter

			switch i % 8 {
			case 0:
				obj = multiPool.Get(63).(*testSizeReporter)
				obj.Init(63)
			case 1:
				obj = multiPool.Get(255).(*testSizeReporter)
				obj.Init(255)
			case 2:
				obj = multiPool.Get(1023).(*testSizeReporter)
				obj.Init(1023)
			case 3:
				obj = multiPool.Get(2047).(*testSizeReporter)
				obj.Init(2047)
			case 4:
				obj = multiPool.Get(4095).(*testSizeReporter)
				obj.Init(4095)
			case 5:
				obj = multiPool.Get(8191).(*testSizeReporter)
				obj.Init(8191)
			case 6:
				obj = multiPool.Get(16383).(*testSizeReporter)
				obj.Init(16383)
			case 7:
				obj = multiPool.Get(32767).(*testSizeReporter)
				obj.Init(32767)
			}

			obj.Reset()
			multiPool.Put(obj)
		}
	})
}
