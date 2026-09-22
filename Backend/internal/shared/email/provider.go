package email

import (
	"context"
)

type Message struct {
	From    string
	To      string
	Subject string
	Text    string
	HTML    string
	Meta    map[string]string
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}
