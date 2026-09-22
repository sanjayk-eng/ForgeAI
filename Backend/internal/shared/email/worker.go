package email

import (
	"context"
	"sync"
	"time"

	"ai-agent/internal/shared/logger"
)

const (
	maxRetries   = 3
	retryBackoff = 2 * time.Second
)

// Worker processes email jobs from the queue
type Worker struct {
	queue  Queue
	sender Sender
	logger logger.Logger
	from   string
	wg     sync.WaitGroup
}

// NewWorker creates a new email worker
func NewWorker(queue Queue, sender Sender, fromEmail string, log logger.Logger) *Worker {
	return &Worker{
		queue:  queue,
		sender: sender,
		from:   fromEmail,
		logger: log,
	}
}

// Start starts worker goroutines to process jobs
func (w *Worker) Start(ctx context.Context, workerCount int) {
	if workerCount <= 0 {
		workerCount = 3
	}

	w.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go w.process(ctx, i+1)
	}

	if w.logger != nil {
		w.logger.Info(ctx, "email workers started", "count", workerCount)
	}
}

// Stop waits for all workers to finish
func (w *Worker) Stop() {
	w.wg.Wait()
}

// process processes jobs from the queue
func (w *Worker) process(ctx context.Context, workerID int) {
	defer w.wg.Done()

	jobs := w.queue.Subscribe()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			w.handleJob(ctx, workerID, job)
		}
	}
}

// handleJob processes a single email job with retry logic
func (w *Worker) handleJob(ctx context.Context, workerID int, job Job) {
	if w.logger != nil {
		w.logger.Info(ctx, "processing email job",
			"worker", workerID,
			"type", job.Type,
			"to", job.To,
		)
	}

	message := Message{
		From:    w.from,
		To:      job.To,
		Subject: job.Subject,
		Text:    job.Text,
		HTML:    job.HTML,
	}

	// Retry logic
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = w.sender.Send(ctx, message)
		if err == nil {
			if w.logger != nil {
				w.logger.Info(ctx, "email sent successfully",
					"worker", workerID,
					"type", job.Type,
					"to", job.To,
				)
			}
			return
		}

		if w.logger != nil {
			w.logger.Warn(ctx, "email send failed",
				"worker", workerID,
				"type", job.Type,
				"to", job.To,
				"attempt", attempt,
				"error", err,
			)
		}

		if attempt < maxRetries {
			time.Sleep(retryBackoff * time.Duration(attempt))
		}
	}

	if w.logger != nil {
		w.logger.Error(ctx, "email send failed after retries",
			"worker", workerID,
			"type", job.Type,
			"to", job.To,
			"error", err,
		)
	}
}
