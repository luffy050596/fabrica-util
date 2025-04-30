package multipool

import (
	"sync"
	"sync/atomic"
)

var (
	// defaultThresholds is the default size thresholds for the object pool
	defaultThresholds = []int{256, 1024, 4096, 16384}
)

// Resetable define a method for object to report its memory size
type Resetable interface {
	Reset()
}

// MultiLayerPool implements a multi-level object pool based on object size
// Time complexity: O(1) to get and put objects
// Space complexity: O(n) where n is the total number of objects in all pool layers
type MultiLayerPool struct {
	// The size thresholds for each object pool (bytes)
	thresholds []int
	// Multiple object pools, layered by object size
	pools []sync.Pool
	// Record the number of hits for each pool
	hits []atomic.Int64
	// Record the number of misses for each pool
	misses []atomic.Int64
	// Record the total number of objects put back
	puts atomic.Int64
	// The function to create a new object
	newFunc func() Resetable

	sizeFunc func(obj Resetable) int
}

// MultiLayerPoolOption define the type of the configuration option function
type MultiLayerPoolOption func(*MultiLayerPool)

// WithThresholds set the size thresholds for the object, in bytes
// For example: []int{128, 256, 512} will create 4 pools:
// - Pool 0: objects <=128 bytes
// - Pool 1: objects >128 and <=256 bytes
// - Pool 2: objects >256 and <=512 bytes
// - Pool 3: objects >512 bytes
func WithThresholds(thresholds []int) MultiLayerPoolOption {
	return func(mp *MultiLayerPool) {
		mp.thresholds = thresholds
	}
}

// NewMultiLayerPool create a new multi-level object pool
func NewMultiLayerPool(newFunc func() Resetable, sizeFunc func(obj Resetable) int, opts ...MultiLayerPoolOption) *MultiLayerPool {
	mp := &MultiLayerPool{
		thresholds: defaultThresholds,
		newFunc:    newFunc,
		sizeFunc:   sizeFunc,
	}

	for _, opt := range opts {
		opt(mp)
	}

	// Initialize the object pools, one more pool is added to accommodate objects larger than the maximum threshold
	poolCount := len(mp.thresholds) + 1
	mp.pools = make([]sync.Pool, poolCount)
	mp.hits = make([]atomic.Int64, poolCount)
	mp.misses = make([]atomic.Int64, poolCount)

	for i := 0; i < poolCount; i++ {
		poolIndex := i
		mp.pools[i].New = func() any {
			mp.misses[poolIndex].Add(1)
			return mp.newFunc()
		}
	}

	return mp
}

// Get get an object from the object pool
// First try to get an object using the estimated size, then reallocate it based on the actual size (through the Size() method)
func (mp *MultiLayerPool) Get(size int) Resetable {
	poolIndex := mp.getPoolIndex(size)

	obj := mp.pools[poolIndex].Get()
	if newReporter, ok := obj.(Resetable); ok {
		mp.hits[poolIndex].Add(1)
		return newReporter
	}

	mp.misses[poolIndex].Add(1)
	return mp.newFunc()
}

// Put put an object back to the appropriate object pool
func (mp *MultiLayerPool) Put(obj Resetable) {
	if obj == nil {
		return
	}

	mp.puts.Add(1)

	size := mp.sizeFunc(obj)
	poolIndex := mp.getPoolIndex(size)

	obj.Reset()
	mp.pools[poolIndex].Put(obj)
}

// getPoolIndex get the index of the appropriate object pool based on the size of the object
func (mp *MultiLayerPool) getPoolIndex(size int) int {
	for i, threshold := range mp.thresholds {
		if size <= threshold {
			return i
		}
	}
	return len(mp.thresholds)
}

// Stats return the usage statistics of the pool
type Stats struct {
	LayerHits   []int64
	LayerMisses []int64
	TotalPuts   int64
	Thresholds  []int
}

// GetStats return the usage statistics of the pool
func (mp *MultiLayerPool) GetStats() Stats {
	stats := Stats{
		LayerHits:   make([]int64, len(mp.hits)),
		LayerMisses: make([]int64, len(mp.misses)),
		TotalPuts:   mp.puts.Load(),
		Thresholds:  mp.thresholds,
	}

	for i := range mp.hits {
		stats.LayerHits[i] = mp.hits[i].Load()
		stats.LayerMisses[i] = mp.misses[i].Load()
	}

	return stats
}
