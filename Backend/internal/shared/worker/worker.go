package worker

import (
	"context"
	"errors"
	"sync"
)

type Job func(context.Context) error

type Worker interface {
	Enqueue(ctx context.Context, job Job) error
	Run(ctx context.Context) error
}

var ErrWorkerStopped = errors.New("worker is stopped")

type InMemoryWorker struct {
	jobs chan Job
	wg   sync.WaitGroup
	once sync.Once
	stop chan struct{}
}

func NewInMemoryWorker(buffer int) *InMemoryWorker {
	if buffer <= 0 {
		buffer = 10
	}
	return &InMemoryWorker{
		jobs: make(chan Job, buffer),
		stop: make(chan struct{}),
	}
}

func (worker *InMemoryWorker) Enqueue(ctx context.Context, job Job) error {
	if worker == nil || job == nil {
		return errors.New("job is invalid")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-worker.stop:
		return ErrWorkerStopped
	case worker.jobs <- job:
		return nil
	}
}

func (worker *InMemoryWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-worker.stop:
			return nil
		case job := <-worker.jobs:
			if job == nil {
				continue
			}
			worker.wg.Add(1)
			go func(current Job) {
				defer worker.wg.Done()
				_ = current(ctx)
			}(job)
		}
	}
}

func (worker *InMemoryWorker) Stop() {
	worker.once.Do(func() {
		close(worker.stop)
	})
	worker.wg.Wait()
}
