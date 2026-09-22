package worker

import (
	"context"
	"errors"
	"sync"
)

type Job func(context.Context) error

type Publisher interface {
	Publish(ctx context.Context, job Job) error
}

type Consumer interface {
	Consume(ctx context.Context, workerCount int)
}

type Worker interface {
	Publisher
	Consumer
	Stop()
}

var ErrWorkerStopped = errors.New("worker is stopped")
var ErrWorkerQueueFull = errors.New("worker queue is full")

type InMemoryPubSub struct {
	jobs     chan Job
	wg       sync.WaitGroup
	consume  sync.Once
	stopOnce sync.Once
	stop     chan struct{}
}

func NewInMemoryPubSub(buffer int) *InMemoryPubSub {
	if buffer <= 0 {
		buffer = 10
	}
	return &InMemoryPubSub{
		jobs: make(chan Job, buffer),
		stop: make(chan struct{}),
	}
}

func (pubsub *InMemoryPubSub) Publish(ctx context.Context, job Job) error {
	if pubsub == nil || job == nil {
		return errors.New("job is invalid")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-pubsub.stop:
		return ErrWorkerStopped
	case pubsub.jobs <- job:
		return nil
	default:
		return ErrWorkerQueueFull
	}
}

func (pubsub *InMemoryPubSub) Consume(ctx context.Context, workerCount int) {
	if pubsub == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	pubsub.consume.Do(func() {
		pubsub.wg.Add(workerCount)
		for workerID := 1; workerID <= workerCount; workerID++ {
			go pubsub.consumeNode(ctx, workerID)
		}
	})
}

func (pubsub *InMemoryPubSub) Stop() {
	if pubsub == nil {
		return
	}
	pubsub.stopOnce.Do(func() { close(pubsub.stop) })
	pubsub.wg.Wait()
}

func (pubsub *InMemoryPubSub) consumeNode(ctx context.Context, workerID int) {
	defer pubsub.wg.Done()
	for {
		select {
		case <-ctx.Done():
			pubsub.stopOnce.Do(func() { close(pubsub.stop) })
			return
		case <-pubsub.stop:
			return
		case job, ok := <-pubsub.jobs:
			if !ok {
				return
			}
			if job != nil {
				_ = job(ctx)
			}
		}
	}
}
