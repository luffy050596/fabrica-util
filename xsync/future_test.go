package xsync

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFuture_Basic(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()

	assert.False(t, f.IsComplete())

	// Complete the future
	f.Complete(42, nil)

	assert.True(t, f.IsComplete())

	// Test Get
	value, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestFuture_Error(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()
	expectedErr := errors.New("test error")

	f.Complete(0, expectedErr)

	value, err := f.Get()
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 0, value)
}

func TestFuture_Cancel(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()

	f.Cancel()

	value, err := f.Get()
	assert.Equal(t, ErrFutureCancelled, err)
	assert.Equal(t, 0, value)
}

func TestFuture_Timeout(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()

	// Test timeout
	value, err := f.GetWithTimeout(100 * time.Millisecond)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Equal(t, 0, value)
}

func TestFuture_Context(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	value, err := f.GetWithContext(ctx)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, 0, value)
}

func TestFuture_Then(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()

	// Create a chained future
	chained := Then(f, func(value int) (string, error) {
		return "value: " + string(rune(value)), nil
	})

	// Complete the original future
	f.Complete(42, nil)

	// Check the chained future
	value, err := chained.Get()
	assert.NoError(t, err)
	assert.Equal(t, "value: *", value)
}

func TestFuture_DoubleComplete(t *testing.T) {
	t.Parallel()

	f := NewFuture[int]()

	// First completion
	f.Complete(42, nil)

	// Second completion should be ignored
	f.Complete(100, errors.New("ignored"))

	value, err := f.Get()
	if err != nil {
		t.Errorf("Get() returned unexpected error: %v", err)
	}

	if value != 42 {
		t.Errorf("Get() returned %v, want 42", value)
	}
}
