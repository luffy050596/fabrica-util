package xsync

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStopper(t *testing.T) {
	t.Parallel()

	s := NewStopper(5 * time.Second)
	require.NotNil(t, s)
	assert.Equal(t, int32(stateIdle), s.state.Load())
	assert.Equal(t, 5*time.Second, s.stopTimeout)
}

func TestTriggerStop(t *testing.T) {
	t.Parallel()

	s := NewStopper(1 * time.Second)

	require.False(t, s.IsStopTriggered())

	s.TriggerStop()
	assert.True(t, s.IsStopTriggered())

	select {
	case <-s.StopTriggered():
		t.Log("expected")
	default:
		t.Fatal("StopTriggered channel should be closed")
	}

	// Test idempotency
	s.TriggerStop()
	assert.True(t, s.IsStopTriggered())
}

func TestDoStop(t *testing.T) {
	t.Parallel()

	t.Run("NormalExecution", func(t *testing.T) {
		t.Parallel()

		s := NewStopper(1 * time.Second)
		executed := false
		stopFunc := func() {
			executed = true
		}

		err := s.DoStop(stopFunc)
		require.NoError(t, err)
		assert.True(t, executed)
		assert.True(t, s.IsStopping())

		// Should be stopped
		select {
		case <-s.Stopping():
			t.Log("expected")
		default:
			t.Fatal("Stopping channel should be closed")
		}

		<-s.stoppedChan // use channel directly to wait

		// Test idempotency
		executed = false
		err = s.DoStop(stopFunc)
		require.NoError(t, err)
		assert.False(t, executed, "stop function should not run again")
	})

	t.Run("Timeout", func(t *testing.T) {
		t.Parallel()

		s := NewStopper(50 * time.Millisecond)
		stopFunc := func() {
			time.Sleep(100 * time.Millisecond) // longer than timeout
		}

		err := s.DoStop(stopFunc)
		assert.ErrorIs(t, err, ErrStopTimeout)

		// WaitStopped should still be closed
		s.WaitStopped()
	})

	t.Run("NoTimeout", func(t *testing.T) {
		t.Parallel()

		s := NewStopper(0)
		executed := false
		stopFunc := func() {
			time.Sleep(50 * time.Millisecond)

			executed = true
		}

		err := s.DoStop(stopFunc)
		require.NoError(t, err)
		assert.True(t, executed)
	})
}

func TestFullLifecycle(t *testing.T) {
	t.Parallel()

	s := NewStopper(100 * time.Millisecond)
	require.False(t, s.IsStopTriggered())
	require.False(t, s.IsStopping())

	// 1. Trigger Stop
	s.TriggerStop()
	assert.True(t, s.IsStopTriggered())
	assert.False(t, s.IsStopping())
	<-s.StopTriggered() // ensure channel is closed

	// 2. Do Stop
	stopped := false
	err := s.DoStop(func() {
		time.Sleep(10 * time.Millisecond)

		stopped = true
	})

	assert.NoError(t, err)
	assert.True(t, stopped)
	assert.True(t, s.IsStopping())
	<-s.Stopping()

	// 3. Wait Stopped
	s.WaitStopped()
	assert.Equal(t, int32(stateStopped), s.state.Load())
}

func TestConcurrentTriggerStop(t *testing.T) {
	t.Parallel()

	s := NewStopper(1 * time.Second)
	numGoroutines := 100
	done := make(chan bool)

	for range numGoroutines {
		go func() {
			s.TriggerStop()
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	// The right check:
	<-s.StopTriggered() // This should not block
	assert.True(t, s.IsStopTriggered())

	// Test idempotency by ensuring a second call doesn't block or panic
	s.TriggerStop()
}

func TestConcurrentDoStop(t *testing.T) {
	t.Parallel()

	s := NewStopper(1 * time.Second)
	numGoroutines := 100
	done := make(chan bool)

	var executionCount int32

	for range numGoroutines {
		go func() {
			err := s.DoStop(func() {
				atomic.AddInt32(&executionCount, 1)
			})
			require.NoError(t, err)
			done <- true
		}()
	}

	for range numGoroutines {
		<-done
	}

	s.WaitStopped()
	assert.Equal(t, int32(1), atomic.LoadInt32(&executionCount), "stop function should only be executed once")
	assert.True(t, s.IsStopping())
}

// Benchmarks

// BenchmarkTriggerStop tests the performance of creating and triggering a stopper.
// It is a parallel benchmark, but each goroutine works on its own instance, so there's no contention.
func BenchmarkTriggerStop(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := NewStopper(0)
			s.TriggerStop()
		}
	})
}

// BenchmarkDoStop tests the performance of creating and stopping a stopper.
// It is a parallel benchmark, but each goroutine works on its own instance, so there's no contention.
func BenchmarkDoStop(b *testing.B) {
	stopFunc := func() {}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := NewStopper(1 * time.Second)
			err := s.DoStop(stopFunc)
			require.NoError(b, err)
		}
	})
}

// BenchmarkConcurrentTriggerStop tests the performance of TriggerStop under high contention.
// All goroutines are calling TriggerStop on the same Stopper instance.
func BenchmarkConcurrentTriggerStop(b *testing.B) {
	b.ReportAllocs()

	s := NewStopper(1 * time.Second)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.TriggerStop()
		}
	})
}

// BenchmarkConcurrentDoStop tests the performance of DoStop under high contention.
// It resets the stopper on each main iteration to allow the stop function to run.
func BenchmarkConcurrentDoStop(b *testing.B) {
	stopFunc := func() {}

	b.ReportAllocs()
	b.StopTimer()

	for range b.N {
		s := NewStopper(1 * time.Second)

		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := s.DoStop(stopFunc)
				require.NoError(b, err)
			}
		})
		b.StopTimer()
	}
}
