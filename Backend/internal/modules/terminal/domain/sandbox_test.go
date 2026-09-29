package domain

import (
	"encoding/json"
	"testing"
)

func TestSandboxJSONUsesFrontendFieldNames(t *testing.T) {
	encoded, err := json.Marshal(Sandbox{ID: "sandbox-id", ProjectID: "project-id", Status: StatusCreating})
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["status"]; !ok {
		t.Fatalf("sandbox JSON is missing status field: %s", encoded)
	}
	if _, ok := payload["project_id"]; !ok {
		t.Fatalf("sandbox JSON is missing project_id field: %s", encoded)
	}
}

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
