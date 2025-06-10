package xsync

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-pantheon/fabrica-util/errors"
)

var (
	// ErrGroupIsClosing is returned when an ErrGroup is in the process of closing
	ErrGroupIsClosing = errors.New("ErrGroup is closing")
	// ErrCloseTimeout is returned when the close function timed out
	ErrCloseTimeout = errors.New("close function timed out")
)

// Closable lifecycle close manager interface
type Closable interface {
	WaitClose
	CloseTriggerable

	// DoClose execute close function with timeout
	DoClose(f func()) error
	// ClosingStart returns channel that's closed when closing starts
	ClosingStart() <-chan struct{}
	// OnClosing checks if closing process has started
	OnClosing() bool
}

// CloseTriggerable trigger close interface
type CloseTriggerable interface {
	// TriggerClose triggers the close process
	TriggerClose()
	// CloseTriggered returns channel that's closed when close is triggered
	CloseTriggered() <-chan struct{}
	// IsCloseTriggered checks if close has been triggered
	IsCloseTriggered() bool
}

// WaitClose wait close completed interface
type WaitClose interface {
	// WaitClosed blocks until closing is complete
	WaitClosed()
}

var _ Closable = (*Closure)(nil)

// Closure implements graceful shutdown with timeout
type Closure struct {
	// State management using atomic operations for better performance
	state *atomic.Int32 // 0=idle, 1=triggered, 2=closing, 3=closed

	// Channels for notifications
	closeTrigger chan struct{} // closed when close is triggered
	closingChan  chan struct{} // closed when closing starts
	closedChan   chan struct{} // closed when closed

	// Configuration
	closeTimeout time.Duration

	// Locks for state transitions
	triggerLock sync.Mutex
	closingLock sync.Mutex
}

const (
	stateIdle = iota
	stateTriggered
	stateClosing
	stateClosed
)

// NewClosure creates a new Closure with the given timeout and options
func NewClosure(timeout time.Duration) *Closure {
	return &Closure{
		state:        &atomic.Int32{},
		closeTrigger: make(chan struct{}),
		closingChan:  make(chan struct{}),
		closedChan:   make(chan struct{}),
		closeTimeout: timeout,
	}
}

// DoClose executes the close function with timeout protection
func (s *Closure) DoClose(f func()) error {
	// Transition to closing state
	if !s.transitionToClosing() {
		return nil // Already closing or closed
	}

	defer s.transitionToClosed()

	if s.closeTimeout <= 0 {
		f()
		return nil
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.closeTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ErrCloseTimeout
	}
}

// TriggerClose triggers the stop process (idempotent)
func (s *Closure) TriggerClose() {
	if s.IsCloseTriggered() {
		return
	}

	s.triggerLock.Lock()
	defer s.triggerLock.Unlock()

	// Double-check pattern
	if s.IsCloseTriggered() {
		return
	}

	if s.state.CompareAndSwap(stateIdle, stateTriggered) {
		close(s.closeTrigger)
	}
}

// CloseTriggered returns a channel that's closed when close is triggered
func (s *Closure) CloseTriggered() <-chan struct{} {
	return s.closeTrigger
}

// IsCloseTriggered checks if close has been triggered
func (s *Closure) IsCloseTriggered() bool {
	return s.state.Load() >= stateTriggered
}

// OnClosing checks if the close process has started
func (s *Closure) OnClosing() bool {
	return s.state.Load() >= stateClosing
}

// ClosingStart returns a channel that's closed when closing starts
func (s *Closure) ClosingStart() <-chan struct{} {
	return s.closingChan
}

// WaitClosed blocks until the stopper has completed stopping
func (s *Closure) WaitClosed() {
	<-s.closedChan
}

// transitionToClosing attempts to transition to closing state
func (s *Closure) transitionToClosing() bool {
	s.closingLock.Lock()
	defer s.closingLock.Unlock()

	currentState := s.state.Load()
	if currentState >= stateClosing {
		return false // Already closing or closed
	}

	if s.state.CompareAndSwap(currentState, stateClosing) {
		close(s.closingChan)
		return true
	}

	return false
}

// transitionToClosed transitions to closed state
func (s *Closure) transitionToClosed() {
	if s.state.CompareAndSwap(stateClosing, stateClosed) {
		close(s.closedChan)
	}
}
