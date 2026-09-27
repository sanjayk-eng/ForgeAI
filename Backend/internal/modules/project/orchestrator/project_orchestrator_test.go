package orchestrator

import (
	"context"
	"testing"
)

func TestProjectOrchestrator_TriggerOnBranchUpdated(t *testing.T) {
	o := &ProjectOrchestrator{}
	called := false

	o.SetOnBranchUpdated(func(ctx context.Context, projectID string) error {
		called = true
		if projectID != "project-123" {
			t.Fatalf("unexpected project id: %s", projectID)
		}
		return nil
	})

	if err := o.TriggerOnBranchUpdated(context.Background(), "project-123"); err != nil {
		t.Fatalf("TriggerOnBranchUpdated() error = %v", err)
	}
	if !called {
		t.Fatal("expected branch-update callback to be invoked")
	}
}
