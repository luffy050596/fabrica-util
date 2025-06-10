// Package xsync provides extended synchronization primitives and utilities
package xsync

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

// ErrDelayerExpired is returned when a delayer has expired
var ErrDelayerExpired = errors.New("delayer expired")

// Delayable the interface for delayers
type Delayable interface {
	WorkerDelayable

	Wait() chan struct{}
	Close()
	IsExpired() bool
	TimeRemaining() time.Duration
}

// WorkerDelayable the interface for worker delayers
type WorkerDelayable interface {
	ExpiryTime() time.Time
	SetExpiryTime(time.Time)
	Reset()
}

var _ Delayable = (*delayer)(nil)

// delayer implements the delayer
type delayer struct {
	mu         sync.RWMutex
	expiryTime time.Time
	timer      *time.Timer
	tick       chan struct{}
	stopCh     chan struct{}
	stopped    bool
}

// NewDelayer creates a new delayer
func NewDelayer() Delayable {
	return &delayer{
		expiryTime: time.Time{},
		tick:       make(chan struct{}, 1), // buffered to prevent blocking
		stopCh:     make(chan struct{}),
		stopped:    false,
	}
}

func (c *delayer) ExpiryTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.expiryTime
}

func (c *delayer) SetExpiryTime(newTime time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop existing timer if any
	if c.timer != nil {
		c.timer.Stop()
	}

	c.stopped = false

	// Clear any pending tick signals
	select {
	case <-c.tick:
	default:
	}

	// Calculate duration until expiry
	duration := time.Until(newTime)
	if duration <= 0 {
		// Already expired, send immediate signal
		select {
		case c.tick <- struct{}{}:
		default:
		}

		c.expiryTime = newTime

		return
	}

	// Create new timer
	c.timer = time.NewTimer(duration)
	c.expiryTime = newTime

	go c.handleExpiry()
}

func (c *delayer) handleExpiry() {
	c.mu.RLock()
	timer := c.timer
	c.mu.RUnlock()

	if timer == nil {
		return
	}

	select {
	case <-timer.C:
		// Timer expired, send tick signal
		c.mu.RLock()
		if !c.stopped {
			select {
			case c.tick <- struct{}{}:
			default:
			}
		}
		c.mu.RUnlock()
	case <-c.stopCh:
		// Timer was stopped
		return
	}
}

func (c *delayer) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	c.expiryTime = time.Time{}
	c.stopped = true

	// Clear any pending tick signals
	select {
	case <-c.tick:
	default:
	}
}

func (c *delayer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	c.stopped = true

	// Signal stop to any running goroutines
	select {
	case c.stopCh <- struct{}{}:
	default:
	}

	// Clear any pending tick signals
	select {
	case <-c.tick:
	default:
	}
}

func (c *delayer) Wait() chan struct{} {
	return c.tick
}

// IsExpired checks if the delayer has expired
func (c *delayer) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.expiryTime.IsZero() {
		return false
	}

	return time.Now().After(c.expiryTime)
}

// TimeRemaining returns the remaining time until expiry
func (c *delayer) TimeRemaining() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.expiryTime.IsZero() {
		return 0
	}

	remaining := time.Until(c.expiryTime)
	if remaining < 0 {
		return 0
	}

	return remaining
}
