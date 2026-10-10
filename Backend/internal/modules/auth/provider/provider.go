package provider

import (
	"context"
)

type ServiceProvider interface {
	ExchangeCode(ctx context.Context, code string) (ServiceUser, error)
}

var _ ServiceProvider = (*GoogleProvider)(nil)
var _ ServiceProvider = (*GitHubProvider)(nil)
