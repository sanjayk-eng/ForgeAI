package terminal

import (
	"sync"

	"ai-agent/internal/shared/realtime"
)

const realtimeBufferSize = 128

type EventHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan realtime.Event]struct{}
	bufferSize  int
}

func NewEventHub() *EventHub {
	return newEventHub(realtimeBufferSize)
}

func newEventHub(bufferSize int) *EventHub {
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &EventHub{
		subscribers: make(map[string]map[chan realtime.Event]struct{}),
		bufferSize:  bufferSize,
	}
}

func (hub *EventHub) Subscribe(projectID string) (<-chan realtime.Event, func()) {
	events := make(chan realtime.Event, hub.bufferSize)
	hub.mu.Lock()
	if hub.subscribers[projectID] == nil {
		hub.subscribers[projectID] = make(map[chan realtime.Event]struct{})
	}
	hub.subscribers[projectID][events] = struct{}{}
	hub.mu.Unlock()

	var once sync.Once
	return events, func() {
		once.Do(func() {
			hub.mu.Lock()
			delete(hub.subscribers[projectID], events)
			if len(hub.subscribers[projectID]) == 0 {
				delete(hub.subscribers, projectID)
			}
			close(events)
			hub.mu.Unlock()
		})
	}
}

func (hub *EventHub) Publish(event realtime.Event) {
	if event.ProjectID == "" {
		return
	}
	if event.Version == 0 {
		event.Version = 1
	}

	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for events := range hub.subscribers[event.ProjectID] {
		select {
		case events <- event:
		default:
			select {
			case <-events:
			default:
			}
			resync := realtime.Event{Version: 1, Event: "sync.required", ProjectID: event.ProjectID, WorkspaceID: event.WorkspaceID, SandboxID: event.SandboxID}
			select {
			case events <- resync:
			default:
			}
		}
	}
}
