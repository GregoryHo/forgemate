package sidecar

import (
	"math"
	"sync"
	"time"
)

// Status exposes sidecar health from the Go control-plane perspective.
type Status struct {
	State          string    `json:"state"`
	FailureCount   int       `json:"failureCount"`
	BreakerOpen    bool      `json:"breakerOpen"`
	BreakerOpened  time.Time `json:"breakerOpened,omitempty"`
	LastFailure    time.Time `json:"lastFailure,omitempty"`
	RestartBackoff string    `json:"restartBackoff"`
}

// Supervisor tracks restart policy using exponential backoff plus breaker window.
type Supervisor struct {
	mu            sync.RWMutex
	state         string
	failureCount  int
	lastFailure   time.Time
	breakerOpen   bool
	breakerOpened time.Time
	window        time.Duration
	threshold     int
	baseBackoff   time.Duration
	maxBackoff    time.Duration
}

func NewSupervisor() *Supervisor {
	return &Supervisor{
		state:       "starting",
		window:      10 * time.Minute,
		threshold:   5,
		baseBackoff: 1 * time.Second,
		maxBackoff:  30 * time.Second,
	}
}

func (s *Supervisor) MarkHealthy() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = "running"
	s.failureCount = 0
	s.lastFailure = time.Time{}
	s.breakerOpen = false
	s.breakerOpened = time.Time{}
}

func (s *Supervisor) MarkFailed(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lastFailure.IsZero() && now.Sub(s.lastFailure) > s.window {
		s.failureCount = 0
		s.breakerOpen = false
		s.breakerOpened = time.Time{}
	}

	s.failureCount++
	s.lastFailure = now
	s.state = "degraded"

	if s.failureCount >= s.threshold {
		s.breakerOpen = true
		s.breakerOpened = now
		s.state = "breaker-open"
	}
}

func (s *Supervisor) CanRestart(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.breakerOpen {
		return true
	}
	return now.Sub(s.breakerOpened) > s.window
}

func (s *Supervisor) NextBackoff() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return calculateBackoff(s.failureCount, s.baseBackoff, s.maxBackoff)
}

func (s *Supervisor) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	backoff := s.baseBackoff
	if s.failureCount > 0 {
		backoff = calculateBackoff(s.failureCount, s.baseBackoff, s.maxBackoff)
	}

	return Status{
		State:          s.state,
		FailureCount:   s.failureCount,
		BreakerOpen:    s.breakerOpen,
		BreakerOpened:  s.breakerOpened,
		LastFailure:    s.lastFailure,
		RestartBackoff: backoff.String(),
	}
}

func calculateBackoff(failureCount int, baseBackoff time.Duration, maxBackoff time.Duration) time.Duration {
	exponent := math.Max(float64(failureCount-1), 0)
	value := float64(baseBackoff) * math.Pow(2, exponent)
	if value > float64(maxBackoff) {
		value = float64(maxBackoff)
	}
	return time.Duration(value)
}
