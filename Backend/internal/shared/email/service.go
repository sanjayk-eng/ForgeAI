package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidRecipient = errors.New("invalid email recipient")
	ErrInvalidSubject   = errors.New("email subject is required")
	ErrInvalidContent   = errors.New("email content is required")
)

// Service provides high-level email sending API
type Service struct {
	queue Queue
}

// NewService creates a new email service
func NewService(queue Queue) *Service {
	return &Service{queue: queue}
}

// SendWelcome sends a welcome email
func (s *Service) SendWelcome(ctx context.Context, to, userName string) error {
	subject := "Welcome to ForgeAI!"
	text := fmt.Sprintf("Hello %s,\n\nWelcome to ForgeAI! We're excited to have you on board.", userName)
	html := fmt.Sprintf("<h1>Welcome to ForgeAI!</h1><p>Hello %s,</p><p>We're excited to have you on board.</p>", userName)

	return s.send(ctx, JobTypeWelcome, to, subject, text, html)
}

// SendVerification sends an email verification link
func (s *Service) SendVerification(ctx context.Context, to, verificationLink string) error {
	subject := "Verify your email address"
	text := fmt.Sprintf("Please verify your email by clicking: %s", verificationLink)
	html := fmt.Sprintf("<h2>Verify your email</h2><p><a href='%s'>Verify Email</a></p>", verificationLink)

	return s.send(ctx, JobTypeVerification, to, subject, text, html)
}

// SendPasswordReset sends a password reset link
func (s *Service) SendPasswordReset(ctx context.Context, to, resetLink string) error {
	subject := "Reset your password"
	text := fmt.Sprintf("Click this link to reset your password: %s", resetLink)
	html := fmt.Sprintf("<h2>Reset your password</h2><p><a href='%s'>Reset Password</a></p>", resetLink)

	return s.send(ctx, JobTypePasswordReset, to, subject, text, html)
}

// SendWorkspaceInvite sends a workspace invitation
func (s *Service) SendWorkspaceInvite(ctx context.Context, to, workspaceName, inviterName, inviteLink, role string) error {
	subject := fmt.Sprintf("You've been invited to join %s", workspaceName)
	text := fmt.Sprintf("%s invited you to join %s as %s. Accept: %s", inviterName, workspaceName, role, inviteLink)
	html := fmt.Sprintf(
		"<h2>Workspace Invitation</h2><p><strong>%s</strong> invited you to join <strong>%s</strong> as <strong>%s</strong>.</p><p><a href='%s'>Accept Invitation</a></p>",
		inviterName, workspaceName, role, inviteLink,
	)

	return s.send(ctx, JobTypeWorkspaceInvite, to, subject, text, html)
}

// SendNotification sends a notification email
func (s *Service) SendNotification(ctx context.Context, to, subject, message string) error {
	html := fmt.Sprintf("<p>%s</p>", message)
	return s.send(ctx, JobTypeNotification, to, subject, message, html)
}

// Send sends a generic email
func (s *Service) Send(ctx context.Context, to, subject, text, html string) error {
	return s.send(ctx, JobTypeGeneric, to, subject, text, html)
}

// send validates and publishes an email job to the queue
func (s *Service) send(ctx context.Context, jobType JobType, to, subject, text, html string) error {
	if err := s.validate(to, subject, text, html); err != nil {
		return err
	}

	job := Job{
		Type:    jobType,
		To:      to,
		Subject: subject,
		Text:    text,
		HTML:    html,
	}

	return s.queue.Publish(ctx, job)
}

// validate validates email parameters
func (s *Service) validate(to, subject, text, html string) error {
	to = strings.TrimSpace(to)
	if to == "" || !strings.Contains(to, "@") {
		return ErrInvalidRecipient
	}

	if strings.TrimSpace(subject) == "" {
		return ErrInvalidSubject
	}

	if strings.TrimSpace(text) == "" && strings.TrimSpace(html) == "" {
		return ErrInvalidContent
	}

	return nil
}
