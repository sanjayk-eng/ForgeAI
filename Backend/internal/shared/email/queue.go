package email

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrQueueFull    = errors.New("email queue is full")
	ErrQueueStopped = errors.New("email queue is stopped")
)

// Queue is a simple FIFO queue for email jobs
type Queue interface {
	Publish(ctx context.Context, job Job) error
	Subscribe() <-chan Job
	Stop()
}

// InMemoryQueue is an in-memory implementation of Queue
type InMemoryQueue struct {
	jobs     chan Job
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewQueue creates a new in-memory email queue
func NewQueue(bufferSize int) *InMemoryQueue {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &InMemoryQueue{
		jobs:     make(chan Job, bufferSize),
		stopChan: make(chan struct{}),
	}
}

// Publish adds a job to the queue
func (q *InMemoryQueue) Publish(ctx context.Context, job Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.stopChan:
		return ErrQueueStopped
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

// Subscribe returns a channel to receive jobs from
func (q *InMemoryQueue) Subscribe() <-chan Job {
	return q.jobs
}

// Stop closes the queue
func (q *InMemoryQueue) Stop() {
	q.stopOnce.Do(func() {
		close(q.stopChan)
		close(q.jobs)
	})
}
