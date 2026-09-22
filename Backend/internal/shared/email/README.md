# Email Module

A clean, queue-based email service for sending emails asynchronously with background workers.

## Architecture

```
Service (API) → Queue (Channel) → Workers → Email Provider (Resend)
```

## Components

### 1. **Queue** (`queue.go`)
- Simple FIFO channel-based queue
- Holds email jobs waiting to be processed
- Thread-safe publish/subscribe pattern

### 2. **Job** (`job.go`)
- Represents an email task
- Types: welcome, verification, password_reset, workspace_invite, notification, generic

### 3. **Service** (`service.go`)
- High-level API for sending emails
- Validates email data before queuing
- Methods: `SendWelcome()`, `SendVerification()`, `SendWorkspaceInvite()`, etc.

### 4. **Worker** (`worker.go`)
- Background workers that process jobs from the queue
- Automatic retry logic (3 attempts with backoff)
- Multiple workers run concurrently for throughput

### 5. **Provider** (`provider.go`, `resend.go`)
- Interface for email providers
- Currently supports Resend API
- Extensible for other providers (SendGrid, Mailgun, etc.)

### 6. **Module** (`module.go`)
- Bootstraps all components together
- Manages lifecycle (start/stop)

## Usage

### Initialize in main.go

```go
// Create email module
emailModule, err := email.NewModule(email.Config{
    Provider:    "resend",
    APIKey:      "re_xxxxx",
    FromEmail:   "noreply@yourdomain.com",
    QueueSize:   100,     // Buffer size
    WorkerCount: 3,       // Number of concurrent workers
    Logger:      logger,
})
if err != nil {
    log.Fatal(err)
}

// Start workers
ctx := context.Background()
emailModule.Start(ctx)
defer emailModule.Stop()
```

### Use in Services

```go
// In your service constructor
type YourService struct {
    email EmailService
}

// Define interface for decoupling
type EmailService interface {
    SendWelcome(ctx context.Context, to, userName string) error
    SendWorkspaceInvite(ctx context.Context, to, workspace, inviter, link, role string) error
}

// Use it
func (s *YourService) RegisterUser(ctx context.Context, email, name string) error {
    // ... create user ...
    
    // Send welcome email (non-blocking)
    if err := s.email.SendWelcome(ctx, email, name); err != nil {
        log.Warn("failed to queue welcome email", err)
    }
    
    return nil
}
```

## Available Methods

```go
// Pre-built email types
service.SendWelcome(ctx, "user@example.com", "John")
service.SendVerification(ctx, "user@example.com", "https://verify-link")
service.SendPasswordReset(ctx, "user@example.com", "https://reset-link")
service.SendWorkspaceInvite(ctx, "user@example.com", "Workspace", "Inviter", "link", "ADMIN")
service.SendNotification(ctx, "user@example.com", "Subject", "Message")

// Generic email
service.Send(ctx, "to@example.com", "Subject", "Text content", "<html>HTML content</html>")
```

## Features

✅ **Asynchronous**: Non-blocking email sending via queue  
✅ **Concurrent**: Multiple workers process emails in parallel  
✅ **Reliable**: Automatic retry logic with exponential backoff  
✅ **Clean**: Interface-based design for testability  
✅ **Simple**: No complex event bus, just a queue and workers  
✅ **Extensible**: Easy to add new providers or email types  

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Provider` | string | `"resend"` | Email provider (currently only "resend") |
| `APIKey` | string | required | API key for the provider |
| `FromEmail` | string | required | Sender email address |
| `QueueSize` | int | `100` | Queue buffer size |
| `WorkerCount` | int | `3` | Number of concurrent workers |
| `Logger` | Logger | `nil` | Optional logger for debugging |

## Retry Logic

- **Max Retries**: 3 attempts
- **Backoff**: 2 seconds × attempt number × 2
- Failed emails after max retries are logged as errors

## Testing

```go
// Mock the EmailService interface
type MockEmailService struct{}

func (m *MockEmailService) SendWelcome(ctx context.Context, to, name string) error {
    return nil
}

// Use in tests
service := NewYourService(&MockEmailService{})
```

## Adding New Email Types

1. Add new `JobType` constant in `job.go`
2. Add method in `service.go`:

```go
func (s *Service) SendCustomEmail(ctx context.Context, to, data string) error {
    subject := "Custom Subject"
    text := fmt.Sprintf("Text: %s", data)
    html := fmt.Sprintf("<p>HTML: %s</p>", data)
    return s.send(ctx, JobTypeCustom, to, subject, text, html)
}
```

3. Update interface in consuming services if needed

## Adding New Providers

Implement the `Sender` interface:

```go
type CustomProvider struct{}

func (p *CustomProvider) Send(ctx context.Context, msg Message) error {
    // Implement sending logic
    return nil
}
```

Then update `module.go`:

```go
case "custom":
    sender = NewCustomProvider(config.APIKey, config.FromEmail)
```
