package sidecar

import (
	"testing"
	"time"
)

func TestSupervisorMarkHealthyResetsFailureState(t *testing.T) {
	s := NewSupervisor()
	now := time.Unix(1000, 0)
	s.MarkFailed(now)
	s.MarkHealthy()

	status := s.Status()
	if status.State != "running" {
		t.Fatalf("expected running state, got %q", status.State)
	}
	if status.FailureCount != 0 {
		t.Fatalf("expected failure count reset, got %d", status.FailureCount)
	}
	if status.BreakerOpen {
		t.Fatal("expected breaker to be closed")
	}
}

func TestSupervisorOpensBreakerAtThreshold(t *testing.T) {
	s := NewSupervisor()
	base := time.Unix(2000, 0)
	for i := 0; i < 5; i++ {
		s.MarkFailed(base.Add(time.Duration(i) * time.Second))
	}

	status := s.Status()
	if !status.BreakerOpen {
		t.Fatal("expected breaker to be open")
	}
	if status.State != "breaker-open" {
		t.Fatalf("expected breaker-open state, got %q", status.State)
	}
	if s.CanRestart(base.Add(5 * time.Second)) {
		t.Fatal("expected restart to be blocked while breaker is open")
	}
	if !s.CanRestart(base.Add(11 * time.Minute)) {
		t.Fatal("expected restart to be allowed after cooldown window")
	}
}

func TestSupervisorBackoffCapsAtMax(t *testing.T) {
	s := NewSupervisor()
	base := time.Unix(3000, 0)
	for i := 0; i < 10; i++ {
		s.MarkFailed(base.Add(time.Duration(i) * time.Second))
	}

	if backoff := s.NextBackoff(); backoff != 30*time.Second {
		t.Fatalf("expected max backoff 30s, got %s", backoff)
	}
}
