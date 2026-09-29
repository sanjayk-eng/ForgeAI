package worker

import (
	"context"
	"sync"
	"testing"
)

func TestSandboxWorkerStopIsSafeWhilePublishing(t *testing.T) {
	worker := NewSandboxWorker(nil, nil, nil)
	worker.Start(context.Background(), 1)

	const publisherCount = 8
	var ready sync.WaitGroup
	var publishers sync.WaitGroup
	ready.Add(publisherCount)
	publishers.Add(publisherCount)
	start := make(chan struct{})
	for range publisherCount {
		go func() {
			defer publishers.Done()
			ready.Done()
			<-start
			for range 1_000 {
				worker.Publish(ProjectEvent{})
			}
		}()
	}
	ready.Wait()
	close(start)
	worker.Stop()
	publishers.Wait()
	worker.Publish(ProjectEvent{})
}
