package xsync

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClosure(t *testing.T) {
	t.Parallel()

	s := NewClosure(5 * time.Second)
	require.NotNil(t, s)
	assert.Equal(t, int32(stateIdle), s.state.Load())
	assert.Equal(t, 5*time.Second, s.closeTimeout)
}

func TestTriggerClosure(t *testing.T) {
	t.Parallel()

	s := NewClosure(1 * time.Second)

	require.False(t, s.IsCloseTriggered())

	s.TriggerClose()
	assert.True(t, s.IsCloseTriggered())

	select {
	case <-s.CloseTriggered():
		t.Log("expected")
	default:
		t.Fatal("CloseTriggered channel should be closed")
	}

	// Test idempotency
	s.TriggerClose()
	assert.True(t, s.IsCloseTriggered())
}

func TestDoClose(t *testing.T) {
	t.Parallel()

	t.Run("NormalExecution", func(t *testing.T) {
		t.Parallel()

		s := NewClosure(1 * time.Second)
		executed := false
		closeFunc := func() {
			executed = true
		}

		err := s.DoClose(closeFunc)
		require.NoError(t, err)
		assert.True(t, executed)
		assert.True(t, s.OnClosing())

		// Should be stopped
		select {
		case <-s.ClosingStart():
			t.Log("expected")
		default:
			t.Fatal("Stopping channel should be closed")
		}

		<-s.closedChan // use channel directly to wait

		// Test idempotency
		executed = false
		err = s.DoClose(closeFunc)
		require.NoError(t, err)
		assert.False(t, executed, "stop function should not run again")
	})

	t.Run("Timeout", func(t *testing.T) {
		t.Parallel()

		s := NewClosure(50 * time.Millisecond)
		stopFunc := func() {
			time.Sleep(100 * time.Millisecond) // longer than timeout
		}

		err := s.DoClose(stopFunc)
		assert.ErrorIs(t, err, ErrCloseTimeout)

		// WaitStopped should still be closed
		s.WaitClosed()
	})

	t.Run("NoTimeout", func(t *testing.T) {
		t.Parallel()

		s := NewClosure(0)
		executed := false
		stopFunc := func() {
			time.Sleep(50 * time.Millisecond)

			executed = true
		}

		err := s.DoClose(stopFunc)
		require.NoError(t, err)
		assert.True(t, executed)
	})
}

func TestFullLifecycle(t *testing.T) {
	t.Parallel()

	s := NewClosure(100 * time.Millisecond)
	require.False(t, s.IsCloseTriggered())
	require.False(t, s.OnClosing())

	// 1. Trigger Stop
	s.TriggerClose()
	assert.True(t, s.IsCloseTriggered())
	assert.False(t, s.OnClosing())
	<-s.CloseTriggered() // ensure channel is closed

	// 2. Do Stop
	stopped := false
	err := s.DoClose(func() {
		time.Sleep(10 * time.Millisecond)

		stopped = true
	})

	assert.NoError(t, err)
	assert.True(t, stopped)
	assert.True(t, s.OnClosing())
	<-s.ClosingStart()

	// 3. Wait Stopped
	s.WaitClosed()
	assert.Equal(t, int32(stateClosed), s.state.Load())
}

func TestConcurrentTriggerStop(t *testing.T) {
	t.Parallel()

	s := NewClosure(1 * time.Second)
	numGoroutines := 100
	done := make(chan bool)

	for range numGoroutines {
		go func() {
			s.TriggerClose()
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	// The right check:
	<-s.CloseTriggered() // This should not block
	assert.True(t, s.IsCloseTriggered())

	// Test idempotency by ensuring a second call doesn't block or panic
	s.TriggerClose()
}

func TestConcurrentDoStop(t *testing.T) {
	t.Parallel()

	s := NewClosure(1 * time.Second)
	numGoroutines := 100
	done := make(chan bool)

	var executionCount int32

	for range numGoroutines {
		go func() {
			err := s.DoClose(func() {
				atomic.AddInt32(&executionCount, 1)
			})
			require.NoError(t, err)
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	s.WaitClosed()
	assert.Equal(t, int32(1), atomic.LoadInt32(&executionCount), "stop function should only be executed once")
	assert.True(t, s.OnClosing())
}

// Benchmarks

// BenchmarkTriggerStop tests the performance of creating and triggering a stopper.
// It is a parallel benchmark, but each goroutine works on its own instance, so there's no contention.
func BenchmarkTriggerStop(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := NewClosure(0)
			s.TriggerClose()
		}
	})
}

// BenchmarkDoClose tests the performance of creating and stopping a stopper.
// It is a parallel benchmark, but each goroutine works on its own instance, so there's no contention.
func BenchmarkDoClose(b *testing.B) {
	stopFunc := func() {}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := NewClosure(1 * time.Second)
			err := s.DoClose(stopFunc)
			require.NoError(b, err)
		}
	})
}

// BenchmarkConcurrentTriggerClose tests the performance of TriggerStop under high contention.
// All goroutines are calling TriggerStop on the same Stopper instance.
func BenchmarkConcurrentTriggerClose(b *testing.B) {
	b.ReportAllocs()

	s := NewClosure(1 * time.Second)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.TriggerClose()
		}
	})
}

// BenchmarkConcurrentDoClose tests the performance of DoStop under high contention.
// It resets the stopper on each main iteration to allow the stop function to run.
func BenchmarkConcurrentDoClose(b *testing.B) {
	stopFunc := func() {}

	b.ReportAllocs()
	b.StopTimer()

	for range b.N {
		s := NewClosure(1 * time.Second)

		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := s.DoClose(stopFunc)
				require.NoError(b, err)
			}
		})
		b.StopTimer()
	}
}
