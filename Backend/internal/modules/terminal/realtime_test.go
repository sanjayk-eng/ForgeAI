package terminal

import (
	"testing"
	"time"

	"ai-agent/internal/shared/realtime"
)

func TestEventHubScopesEventsByProject(t *testing.T) {
	hub := NewEventHub()
	projectA, closeA := hub.Subscribe("project-a")
	defer closeA()
	projectB, closeB := hub.Subscribe("project-b")
	defer closeB()

	hub.Publish(realtime.Event{Event: "file.changed", ProjectID: "project-a", Path: "src/app.go"})

	select {
	case event := <-projectA:
		if event.ProjectID != "project-a" || event.Version != 1 {
			t.Fatalf("unexpected event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("project A did not receive its event")
	}
	select {
	case event := <-projectB:
		t.Fatalf("project B received project A event: %+v", event)
	default:
	}
}

func TestEventHubRequestsResyncWhenSubscriberFallsBehind(t *testing.T) {
	hub := newEventHub(1)
	events, unsubscribe := hub.Subscribe("project-a")
	defer unsubscribe()

	hub.Publish(realtime.Event{Event: "file.changed", ProjectID: "project-a", Path: "one.go"})
	hub.Publish(realtime.Event{Event: "file.changed", ProjectID: "project-a", Path: "two.go"})

	select {
	case event := <-events:
		if event.Event != "sync.required" {
			t.Fatalf("event = %q, want sync.required", event.Event)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive resync event")
	}
}
