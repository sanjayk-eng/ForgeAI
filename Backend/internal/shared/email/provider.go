package email

import (
	"context"
	"errors"
)

type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
	Meta    map[string]string
}

type Provider interface {
	Send(ctx context.Context, message Message) error
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}

var ErrNoEmailProvider = errors.New("email provider is not configured")

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (service *Service) Send(ctx context.Context, message Message) error {
	if service == nil || service.provider == nil {
		return ErrNoEmailProvider
	}
	return service.provider.Send(ctx, message)
}
