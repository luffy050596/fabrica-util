package xsync

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelayer_BasicFunctionality(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()
	defer delayer.Stop()

	// Test setting expiry time
	expiryTime := time.Now().Add(100 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime)

	// Verify expiry time is set correctly
	if !delayer.ExpiryTime().Equal(expiryTime) {
		t.Errorf("Expected expiry time %v, got %v", expiryTime, delayer.ExpiryTime())
	}

	// Wait for expiry
	select {
	case <-delayer.Tick():
		// Expected behavior
	case <-time.After(200 * time.Millisecond):
		t.Error("delayer did not expire within expected time")
	}

	// Check if expired
	if !delayer.IsExpired() {
		t.Error("delayer should be expired")
	}
}

func TestDelayer_Reset(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()
	defer delayer.Stop()

	// Set expiry time
	expiryTime := time.Now().Add(100 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime)

	// Reset before expiry
	delayer.Reset()

	// Verify expiry time is reset
	if !delayer.ExpiryTime().IsZero() {
		t.Error("Expiry time should be zero after reset")
	}

	// Verify no tick is received
	select {
	case <-delayer.Tick():
		t.Error("Should not receive tick after reset")
	case <-time.After(150 * time.Millisecond):
		t.Log("Should not receive tick after reset")
	}
}

func TestDelayer_Stop(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()

	// Set expiry time
	expiryTime := time.Now().Add(100 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime)

	// Stop immediately
	delayer.Stop()

	// Verify no tick is received
	select {
	case <-delayer.Tick():
		t.Error("Should not receive tick after stop")
	case <-time.After(150 * time.Millisecond):
		t.Log("Should not receive tick after stop")
	}
}

func TestDelayer_TimeRemaining(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()
	defer delayer.Stop()

	// Set expiry time 200ms from now
	expiryTime := time.Now().Add(200 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime)

	// Check initial remaining time
	remaining := delayer.TimeRemaining()
	if remaining <= 150*time.Millisecond || remaining > 200*time.Millisecond {
		t.Errorf("Expected remaining time around 200ms, got %v", remaining)
	}

	// Wait a bit and check again
	time.Sleep(50 * time.Millisecond)

	remaining = delayer.TimeRemaining()
	if remaining <= 100*time.Millisecond || remaining > 150*time.Millisecond {
		t.Errorf("Expected remaining time around 150ms, got %v", remaining)
	}

	// Wait for expiry
	<-delayer.Tick()

	// Check remaining time after expiry
	remaining = delayer.TimeRemaining()
	if remaining != 0 {
		t.Errorf("Expected remaining time to be 0 after expiry, got %v", remaining)
	}
}

func TestDelayer_ImmediateExpiry(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()
	defer delayer.Stop()

	// Set expiry time in the past
	expiryTime := time.Now().Add(-100 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime)

	// Should receive immediate tick
	select {
	case <-delayer.Tick():
		// Expected behavior
	case <-time.After(50 * time.Millisecond):
		t.Error("Should receive immediate tick for past expiry time")
	}

	// Should be expired
	assert.True(t, delayer.IsExpired(), "Should be expired for past expiry time")
}

func TestDelayer_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()
	defer delayer.Stop()

	var wg sync.WaitGroup

	const numGoroutines = 10

	// Start multiple goroutines that set expiry times
	for range numGoroutines {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := range 10 {
				expiryTime := time.Now().Add(time.Duration(j+1) * 10 * time.Millisecond)
				delayer.SetExpiryTime(expiryTime)
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	// Start goroutines that read expiry time
	for range numGoroutines {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 20 {
				_ = delayer.ExpiryTime()
				_ = delayer.IsExpired()
				_ = delayer.TimeRemaining()

				time.Sleep(2 * time.Millisecond)
			}
		}()
	}

	// Start goroutines that reset and stop
	for range 5 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 5 {
				delayer.Reset()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
}

func TestDelayer_MultipleSetExpiryTime(t *testing.T) {
	t.Parallel()

	delayer := NewDelayer()
	defer delayer.Stop()

	// Set initial expiry time
	expiryTime1 := time.Now().Add(200 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime1)

	// Immediately set a new expiry time (shorter)
	expiryTime2 := time.Now().Add(100 * time.Millisecond)
	delayer.SetExpiryTime(expiryTime2)

	// Should expire at the second time
	start := time.Now()

	<-delayer.Tick()

	elapsed := time.Since(start)

	if elapsed > 150*time.Millisecond {
		t.Errorf("Expected to expire around 100ms, took %v", elapsed)
	}

	// Verify final expiry time
	if !delayer.ExpiryTime().Equal(expiryTime2) {
		t.Errorf("Expected final expiry time %v, got %v", expiryTime2, delayer.ExpiryTime())
	}
}

// BenchmarkDelayer_SetExpiryTime benchmarks the SetExpiryTime operation
func BenchmarkDelayer_SetExpiryTime(b *testing.B) {
	delayer := NewDelayer()
	defer delayer.Stop()

	b.ResetTimer()

	for i := range b.N {
		expiryTime := time.Now().Add(time.Duration(i%1000) * time.Millisecond)
		delayer.SetExpiryTime(expiryTime)
	}
}

// BenchmarkDelayer_ConcurrentAccess benchmarks concurrent access
func BenchmarkDelayer_ConcurrentAccess(b *testing.B) {
	delayer := NewDelayer()
	defer delayer.Stop()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			expiryTime := time.Now().Add(100 * time.Millisecond)
			delayer.SetExpiryTime(expiryTime)
			_ = delayer.ExpiryTime()
			_ = delayer.IsExpired()
			_ = delayer.TimeRemaining()
		}
	})
}
