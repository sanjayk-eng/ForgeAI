package domain

import "testing"

func TestSandboxLifecycleTransitions(t *testing.T) {
	sandbox := Sandbox{Status: StatusCreating}
	for _, next := range []SandboxStatus{StatusCreated, StatusStarting, StatusRunning, StatusStopping, StatusStopped, StatusStarting, StatusRunning, StatusDestroying, StatusDestroyed} {
		var err error
		sandbox, err = sandbox.Transition(next)
		if err != nil {
			t.Fatalf("transition to %s: %v", next, err)
		}
	}
}

func TestSandboxRejectsInvalidTransition(t *testing.T) {
	_, err := (Sandbox{Status: StatusCreated}).Transition(StatusRunning)
	if err != ErrInvalidTransition {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func TestSandboxLifecycleAllowsRecoveryFromFailure(t *testing.T) {
	sandbox := Sandbox{Status: StatusStarting}
	var err error
	sandbox, err = sandbox.Transition(StatusFailed)
	if err != nil {
		t.Fatalf("transition to failed: %v", err)
	}
	if _, err = sandbox.Transition(StatusStarting); err != nil {
		t.Fatalf("recover from failed state: %v", err)
	}
}
