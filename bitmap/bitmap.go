// Package bitmap provides thread-safe bitmap implementation using byte arrays
package bitmap

import (
	"math/bits"
	"sync"
)

// Bitmap represents a thread-safe bitmap using a byte array
type Bitmap struct {
	mutex sync.Mutex
	bits  []byte
	size  int64 // Track original size for bounds checking
}

// NewBitmap creates a new Bitmap with the given size (in bits)
func NewBitmap(size int64) *Bitmap {
	if size < 0 {
		panic("bitmap size must be non-negative")
	}

	return &Bitmap{
		bits: make([]byte, (size+7)/8),
		size: size,
	}
}

// Set sets the bit at the given index to 1
func (b *Bitmap) Set(index int64) {
	b.validateIndex(index)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.bits[index/8] |= 1 << (index % 8)
}

// MSet sets multiple bits at the given indexes to 1
func (b *Bitmap) MSet(indexes []int64) {
	if len(indexes) == 0 {
		return
	}

	for _, index := range indexes {
		b.validateIndex(index)
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, index := range indexes {
		b.bits[index/8] |= 1 << (index % 8)
	}
}

// Clear clears the bit at the given index to 0
func (b *Bitmap) Clear(index int64) {
	b.validateIndex(index)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.bits[index/8] &^= 1 << (index % 8)
}

// IsSet checks if the bit at the given index is set to 1
func (b *Bitmap) IsSet(index int64) bool {
	b.validateIndex(index)
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.bits[index/8]&(1<<(index%8)) != 0
}

// Count returns the number of bits set to 1 using efficient bit counting
func (b *Bitmap) Count() int64 {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	count := int64(0)

	for _, byteVal := range b.bits {
		count += int64(bits.OnesCount8(byteVal))
	}

	return count
}

// Size returns the capacity of the bitmap in bits
func (b *Bitmap) Size() int64 {
	return b.size
}

// validateIndex checks if index is within valid range
func (b *Bitmap) validateIndex(index int64) {
	if index < 0 || index >= b.size {
		panic("bitmap index out of range")
	}
}
