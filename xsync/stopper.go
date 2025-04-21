package xsync

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

// ErrGroupStopping is returned when an ErrGroup is in the process of stopping
var ErrGroupStopping = errors.New("ErrGroup is stopping") // Stoppable is stopping signal

// Stoppable lifecycle stop manager interface
type Stoppable interface {
	WaitStoppable
	StopTriggerable

	// DoStop execute stop
	DoStop(f func())
	// Stopping listen stop is started
	Stopping() <-chan struct{}
	// IsStopping check stop is started
	IsStopping() bool
}

// StopTriggerable trigger stop interface
type StopTriggerable interface {
	// TriggerStop trigger stop
	TriggerStop()
	// StopTriggered listen stop is triggered
	StopTriggered() <-chan struct{}
}

// WaitStoppable wait stop completed
type WaitStoppable interface {
	WaitStopped()
}

var _ Stoppable = (*Stopper)(nil)

// StopperOption is a function that configures a Stopper
type StopperOption func(*Stopper)

// WithLog sets a logger for the Stopper
func WithLog(log *log.Helper) StopperOption {
	return func(s *Stopper) {
		s.log = log
	}
}

// Stopper implements Stoppable interface
type Stopper struct {
	log *log.Helper

	_triggerLock  sync.Mutex
	stopTrigger   chan struct{} // the notification of stop triggered
	stopTriggered *atomic.Bool  // stop is triggered

	_stoppingLock sync.Mutex
	stoppingChan  chan struct{} // the notification of starting to stop
	isStopping    *atomic.Bool  // stop is started

	stoppedChan chan struct{} // the notification of stopping completed
	stopTimeout time.Duration // the timeout of stop
}

// NewStopper creates a new Stopper with the given timeout and options
func NewStopper(stopTimeout time.Duration, opts ...StopperOption) *Stopper {
	s := &Stopper{
		stopTrigger:   make(chan struct{}),
		stopTriggered: atomic.NewBool(false),

		stoppingChan: make(chan struct{}),
		isStopping:   atomic.NewBool(false),

		stoppedChan: make(chan struct{}),
		stopTimeout: stopTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// DoStop executes the stop function if the stopper is not already stopping
func (s *Stopper) DoStop(f func()) {
	if s.IsStopping() {
		return
	}

	func() {
		s._stoppingLock.Lock()
		defer s._stoppingLock.Unlock()

		if s.IsStopping() {
			return
		}

		close(s.stoppingChan)
		s.isStopping.Store(true)
	}()

	defer close(s.stoppedChan)

	ctx, cancel := context.WithTimeout(context.Background(), s.stopTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		f()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		s.log.Error("Stopper stop timed out after timeout")
	}
}

// TriggerStop triggers the stop process if it hasn't been triggered already
func (s *Stopper) TriggerStop() {
	if s.isStopTriggered() {
		return
	}

	s._triggerLock.Lock()
	defer s._triggerLock.Unlock()

	if s.isStopTriggered() {
		return
	}

	close(s.stopTrigger)
	s.stopTriggered.Store(true)
}

// StopTriggered returns a channel that is closed when stop is triggered
func (s *Stopper) StopTriggered() <-chan struct{} {
	return s.stopTrigger
}

// IsStopping returns true if the stopper is in the process of stopping
func (s *Stopper) IsStopping() bool {
	return s.isStopping.Load()
}

// Stopping returns a channel that is closed when stopping has started
func (s *Stopper) Stopping() <-chan struct{} {
	return s.stoppingChan
}

// WaitStopped blocks until the stopper has completed stopping
func (s *Stopper) WaitStopped() {
	<-s.stoppedChan
}

func (s *Stopper) isStopTriggered() bool {
	return s.stopTriggered.Load()
}
