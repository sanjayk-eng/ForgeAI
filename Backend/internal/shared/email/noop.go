package email

import "context"

type NoopProvider struct{}

func (provider *NoopProvider) Send(ctx context.Context, message Message) error {
	return nil
}
