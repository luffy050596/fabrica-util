package xsync

import (
	"context"
	"sync"
	"time"

	"github.com/go-pantheon/fabrica-util/errors"
	"go.uber.org/atomic"
)

// ErrGroupStopping is returned when an ErrGroup is in the process of stopping
var (
	ErrGroupStopping = errors.New("ErrGroup is stopping")
	ErrStopTimeout   = errors.New("stop function timed out")
)

// Stoppable lifecycle stop manager interface
type Stoppable interface {
	WaitStoppable
	StopTriggerable

	// DoStop execute stop function with timeout
	DoStop(f func()) error
	// Stopping returns channel that's closed when stopping starts
	Stopping() <-chan struct{}
	// IsStopping checks if stop process has started
	IsStopping() bool
}

// StopTriggerable trigger stop interface
type StopTriggerable interface {
	// TriggerStop triggers the stop process
	TriggerStop()
	// StopTriggered returns channel that's closed when stop is triggered
	StopTriggered() <-chan struct{}
	// IsStopTriggered checks if stop has been triggered
	IsStopTriggered() bool
}

// WaitStoppable wait stop completed interface
type WaitStoppable interface {
	// WaitStopped blocks until stopping is complete
	WaitStopped()
}

var _ Stoppable = (*Stopper)(nil)

// Stopper implements graceful shutdown with timeout
type Stopper struct {
	// State management using atomic operations for better performance
	state *atomic.Int32 // 0=idle, 1=triggered, 2=stopping, 3=stopped

	// Channels for notifications
	stopTrigger  chan struct{} // closed when stop is triggered
	stoppingChan chan struct{} // closed when stopping starts
	stoppedChan  chan struct{} // closed when stopped

	// Configuration
	stopTimeout time.Duration

	// Locks for state transitions
	triggerLock  sync.Mutex
	stoppingLock sync.Mutex
}

const (
	stateIdle = iota
	stateTriggered
	stateStopping
	stateStopped
)

// NewStopper creates a new Stopper with the given timeout and options
func NewStopper(stopTimeout time.Duration) *Stopper {
	return &Stopper{
		state:        atomic.NewInt32(stateIdle),
		stopTrigger:  make(chan struct{}),
		stoppingChan: make(chan struct{}),
		stoppedChan:  make(chan struct{}),
		stopTimeout:  stopTimeout,
	}
}

// DoStop executes the stop function with timeout protection
func (s *Stopper) DoStop(f func()) error {
	// Transition to stopping state
	if !s.transitionToStopping() {
		return nil // Already stopping or stopped
	}

	defer s.transitionToStopped()

	if s.stopTimeout <= 0 {
		f()
		return nil
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.stopTimeout)
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
		return ErrStopTimeout
	}
}

// TriggerStop triggers the stop process (idempotent)
func (s *Stopper) TriggerStop() {
	if s.IsStopTriggered() {
		return
	}

	s.triggerLock.Lock()
	defer s.triggerLock.Unlock()

	// Double-check pattern
	if s.IsStopTriggered() {
		return
	}

	if s.state.CompareAndSwap(stateIdle, stateTriggered) {
		close(s.stopTrigger)
	}
}

// StopTriggered returns a channel that's closed when stop is triggered
func (s *Stopper) StopTriggered() <-chan struct{} {
	return s.stopTrigger
}

// IsStopTriggered checks if stop has been triggered
func (s *Stopper) IsStopTriggered() bool {
	return s.state.Load() >= stateTriggered
}

// IsStopping checks if the stop process has started
func (s *Stopper) IsStopping() bool {
	return s.state.Load() >= stateStopping
}

// Stopping returns a channel that's closed when stopping starts
func (s *Stopper) Stopping() <-chan struct{} {
	return s.stoppingChan
}

// WaitStopped blocks until the stopper has completed stopping
func (s *Stopper) WaitStopped() {
	<-s.stoppedChan
}

// transitionToStopping attempts to transition to stopping state
func (s *Stopper) transitionToStopping() bool {
	s.stoppingLock.Lock()
	defer s.stoppingLock.Unlock()

	currentState := s.state.Load()
	if currentState >= stateStopping {
		return false // Already stopping or stopped
	}

	if s.state.CompareAndSwap(currentState, stateStopping) {
		close(s.stoppingChan)
		return true
	}

	return false
}

// transitionToStopped transitions to stopped state
func (s *Stopper) transitionToStopped() {
	if s.state.CompareAndSwap(stateStopping, stateStopped) {
		close(s.stoppedChan)
	}
}
