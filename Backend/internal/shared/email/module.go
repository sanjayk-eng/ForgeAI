package email

import (
	"context"
	"errors"
	"net/http"

	"ai-agent/internal/shared/logger"
)

var ErrInvalidConfig = errors.New("invalid email configuration")

// Module encapsulates the complete email system
type Module struct {
	Service     *Service
	Worker      *Worker
	Queue       Queue
	Sender      Sender
	WorkerCount int
}

// Config holds email module configuration
type Config struct {
	Provider      string // "resend"
	APIKey        string
	FromEmail     string
	QueueSize     int
	WorkerCount   int
	HTTPClient    *http.Client
	Logger        logger.Logger
}

// NewModule creates and initializes the email module
func NewModule(config Config) (*Module, error) {
	// Validate config
	if config.FromEmail == "" {
		return nil, ErrInvalidConfig
	}
	if config.Provider == "" {
		config.Provider = "resend"
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}
	if config.WorkerCount <= 0 {
		config.WorkerCount = 3
	}

	// Create sender based on provider
	var sender Sender
	switch config.Provider {
	case "resend":
		if config.APIKey == "" {
			return nil, ErrInvalidConfig
		}
		sender = NewResendProvider(config.APIKey, config.FromEmail, config.HTTPClient)
	default:
		return nil, ErrInvalidConfig
	}

	// Create queue
	queue := NewQueue(config.QueueSize)

	// Create worker
	worker := NewWorker(queue, sender, config.FromEmail, config.Logger)

	// Create service
	service := NewService(queue)

	return &Module{
		Service:     service,
		Worker:      worker,
		Queue:       queue,
		Sender:      sender,
		WorkerCount: config.WorkerCount,
	}, nil
}

// Start starts the email workers
func (m *Module) Start(ctx context.Context) {
	if m == nil || m.Worker == nil {
		return
	}
	m.Worker.Start(ctx, m.WorkerCount)
}

// Stop gracefully stops the email module
func (m *Module) Stop() {
	if m == nil {
		return
	}
	if m.Queue != nil {
		m.Queue.Stop()
	}
	if m.Worker != nil {
		m.Worker.Stop()
	}
}
